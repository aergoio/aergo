package dpos

import (
	"container/list"
	"fmt"
	"sort"
	"sync"

	"github.com/aergoio/aergo/types"
)

type errLibUpdate struct {
	current string
	parent  string
	oldBest string
}

func (e errLibUpdate) Error() string {
	return fmt.Sprintf(
		"current block %v (parent %v) inconsistent with old best %v",
		e.current, e.parent, e.oldBest)
}

// Status manages DPoS-related infomations like LIB.
type Status struct {
	sync.RWMutex
	bestBlock *types.Block
	pls       *pLibStatus
	lib       *blockInfo
}

// NewStatus returns a newly allocated Status.
func NewStatus(confirmsRequired uint16) *Status {
	return &Status{
		pls: newPlibStatus(confirmsRequired),
	}
}

type pLibStatus struct {
	confirmsRequired uint16
	confirms         *list.List
	plib             map[string]*blockInfo // BP-wise proposed LIB map
}

func newPlibStatus(confirmsRequired uint16) *pLibStatus {
	return &pLibStatus{
		confirmsRequired: confirmsRequired,
		confirms:         list.New(),
		plib:             make(map[string]*blockInfo),
	}
}

func (pls *pLibStatus) init() {
	pls.confirms.Init()
}

func (pls *pLibStatus) updateStatus(block *types.Block) *blockInfo {
	bp := block.BPID2Str()
	ci := newConfirmInfo(block, pls.confirmsRequired)
	pls.confirms.PushBack(ci)

	logger.Debug().Str("BP", bp).
		Str("hash", block.ID()).Uint64("no", block.BlockNo()).
		Msg("new confirm info added")

	if bi := pls.getPreLIB(ci.blockInfo); bi != nil {
		pls.plib[bp] = bi

		logger.Debug().Str("BP", bp).
			Str("hash", bi.blockHash).Uint64("no", bi.blockNo).
			Msg("proposed LIB map updated")

		return pls.calcLIB()
	}

	return nil
}

func (pls *pLibStatus) rollbackStatusTo(block *types.Block) error {
	// XXX Do the real status rollback instead of init.
	pls.init()

	return nil
}

func (pls *pLibStatus) getPreLIB(curBlockInfo *blockInfo) (bi *blockInfo) {
	var (
		prev *list.Element
		del  = false
		e    = pls.confirms.Back()
		cr   = curBlockInfo.confirmRange
	)

	for e != nil && cr > 0 {
		prev = e.Prev()
		cr--

		if !del {
			c := e.Value.(*confirmInfo)
			c.confirmsLeft--
			if c.confirmsLeft == 0 {
				// proposed LIB info to return
				bi = c.blockInfo
				del = true
			}
		}

		// Delete all the previous elements including the one corresponding to
		// a block to be finalized (c.confirmsLeft == 0). They are not
		// necessary any more, since all the blocks before a finalized block
		// are also final.
		if del {
			pls.confirms.Remove(e)
		}

		e = prev
	}

	return
}

func (pls *pLibStatus) calcLIB() *blockInfo {
	libInfos := make([]*blockInfo, 0, len(pls.plib))
	for _, l := range pls.plib {
		libInfos = append(libInfos, l)
	}

	sort.Slice(libInfos, func(i, j int) bool {
		return libInfos[i].blockNo < libInfos[j].blockNo
	})

	lib := libInfos[(len(libInfos)-1)/(3*int(bpConsensusCount))]
	if lib != nil {
		for _, l := range pls.plib {
			if l.blockNo < lib.blockNo {
				delete(pls.plib, l.blockHash)
			}
		}
	}

	return lib
}

type confirmInfo struct {
	*blockInfo
	confirmsLeft uint16
}

func newConfirmInfo(block *types.Block, confirmsRequired uint16) *confirmInfo {
	return &confirmInfo{
		blockInfo: &blockInfo{
			blockHash:    block.ID(),
			blockNo:      block.BlockNo(),
			confirmRange: block.GetHeader().GetConfirms(),
		},
		confirmsLeft: confirmsRequired,
	}
}

type blockInfo struct {
	blockHash    string
	blockNo      uint64
	confirmRange uint64
}

// UpdateStatus updates the last irreversible block (LIB).
func (s *Status) UpdateStatus(block *types.Block) {
	s.Lock()
	defer s.Unlock()

	if s.bestBlock == nil {
		s.bestBlock = block
		return
	}

	curBestID := s.bestBlock.ID()

	if curBestID == block.PrevID() {
		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("block no", block.BlockNo()).
			Msg("update LIB status")

		// Block connected
		if lib := s.pls.updateStatus(block); lib != nil {
			s.updateLIB(lib)
		}
	} else {
		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("target block no", block.BlockNo()).
			Msg("rollback LIB status")

		// Block reorganized. TODO: update consensus status, correctly.
		if err := s.pls.rollbackStatusTo(block); err != nil {
			panic(err)
		}
	}

	s.bestBlock = block
}

func (s *Status) updateLIB(lib *blockInfo) {
	s.lib = lib
	logger.Debug().
		Str("block hash", s.lib.blockHash).
		Uint64("block no", s.lib.blockNo).
		Msg("last irreversible block (BFT) updated")
}

// NeedReorganization reports whether reorganization is needed or not.
func (s *Status) NeedReorganization(rootNo, bestNo types.BlockNo) bool {
	return true
	// Disable until the reorganization logic is correctly implmented.
	/*
		s.RLock()
		defer s.RUnlock()

		if s.lib == nil {
			logger.Debug().Uint64("branch root no", rootNo).Msg("no LIB")
			return true
		}

		libNo := s.lib.blockNo

		reorganizable := rootNo < libNo && bestNo > libNo
		if reorganizable {
			logger.Info().
				Uint64("LIB", libNo).
				Uint64("branch root no", rootNo).
				Uint64("best no", bestNo).
				Msg("not reorganizable - the current main branch has a LIB.")
		}

		return reorganizable
	*/
}
