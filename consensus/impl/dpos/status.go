package dpos

import (
	"container/list"
	"fmt"
	"sort"
	"sync"

	"github.com/aergoio/aergo/chain"
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
	genesisInfo      *blockInfo
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

func (pls *pLibStatus) addConfirmInfo(block *types.Block) {
	ci := newConfirmInfo(block, pls.confirmsRequired)
	pls.confirms.PushBack(ci)

	bi := ci.blockInfo

	// Initialize an empty pre-LIB map entry with genesis block info.
	if _, exist := pls.plib[bi.bpID]; !exist {
		pls.updatePreLIB(bi.bpID, pls.genesisInfo)
	}

	logger.Debug().Str("BP", bi.bpID).
		Str("hash", bi.blockHash).Uint64("no", bi.blockNo).
		Msg("new confirm info added")
}

func (pls *pLibStatus) updateStatus() *blockInfo {
	if bi := pls.getPreLIB(); bi != nil {
		pls.updatePreLIB(bi.bpID, bi)
	}

	return pls.calcLIB()
}

func (pls *pLibStatus) updatePreLIB(bpID string, bi *blockInfo) {
	pls.plib[bi.bpID] = bi
	logger.Debug().Str("BP", bi.bpID).
		Str("hash", bi.blockHash).Uint64("no", bi.blockNo).
		Msg("proposed LIB map updated")
}

func (pls *pLibStatus) rollbackStatusTo(block *types.Block) error {
	// XXX Do the real status rollback instead of init.
	pls.init()

	return nil
}

func (pls *pLibStatus) getPreLIB() (bi *blockInfo) {
	cInfo := func(e *list.Element) *confirmInfo {
		return e.Value.(*confirmInfo)
	}

	bInfo := func(c *confirmInfo) *blockInfo {
		return c.blockInfo
	}

	var (
		prev *list.Element
		del  = false
		e    = pls.confirms.Back()
		cr   = bInfo(cInfo(e)).confirmRange
	)

	for e != nil && cr > 0 {
		prev = e.Prev()
		cr--

		if !del {
			c := cInfo(e)
			c.confirmsLeft--
			if c.confirmsLeft == 0 {
				// proposed LIB info to return
				bi = bInfo(c)
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

	// TODO: check the correctness of the formula.
	lib := libInfos[(len(libInfos)-1)/3]
	// Comment out until it proves to be necessary. TODO: check whether the
	// cleanup below is correct.
	//
	// if lib != nil {
	// 	for _, l := range pls.plib {
	// 		if l.blockNo < lib.blockNo {
	// 			delete(pls.plib, l.blockHash)
	// 		}
	// 	}
	// }

	return lib
}

type confirmInfo struct {
	*blockInfo
	confirmsLeft uint16
}

func newConfirmInfo(block *types.Block, confirmsRequired uint16) *confirmInfo {
	return &confirmInfo{
		blockInfo: &blockInfo{
			bpID:         block.BPID2Str(),
			blockHash:    block.ID(),
			blockNo:      block.BlockNo(),
			confirmRange: block.GetHeader().GetConfirms(),
		},
		confirmsLeft: confirmsRequired,
	}
}

type blockInfo struct {
	bpID         string
	blockHash    string
	blockNo      uint64
	confirmRange uint64
}

// UpdateStatus updates the last irreversible block (LIB).
func (s *Status) UpdateStatus(block *types.Block) {
	s.Lock()
	defer s.Unlock()

	if s.pls.genesisInfo == nil {
		if genesisBlock := chain.GetGenesisBlock(); genesisBlock != nil {
			s.pls.genesisInfo = &blockInfo{
				blockHash: genesisBlock.ID(),
				blockNo:   genesisBlock.BlockNo(),
			}

			// Temporarily set s.bestBlock to genesisBlock whenever the server
			// is started. TODO: Do the status reovery correctly.
			if s.bestBlock == nil {
				s.bestBlock = genesisBlock
			}
		}
	}

	curBestID := s.bestBlock.ID()
	if curBestID == block.PrevID() {
		s.pls.addConfirmInfo(block)

		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("block no", block.BlockNo()).
			Msg("update LIB status")

		// Block connected
		if lib := s.pls.updateStatus(); lib != nil {
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
