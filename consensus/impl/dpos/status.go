package dpos

import (
	"container/list"
	"sort"
	"sync"

	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
)

// Status manages DPoS-related infomations like LIB.
type Status struct {
	sync.RWMutex
	bestBlock        *types.Block
	confirmsRequired uint16
	plibConfirms     *list.List            // confirm counts to become proposed LIB
	plib             map[string]*blockInfo // BP-wise proposed LIB map
	lib              *blockInfo
}

type blockInfo struct {
	blockHash string
	blockNo   uint64
}

type plibConfirm struct {
	*blockInfo
	confirmsLeft uint16
}

// NewStatus returns a newly allocated Status.
func NewStatus(confirmsRequired uint16) *Status {
	return &Status{
		confirmsRequired: confirmsRequired,
		plibConfirms:     list.New(),
		plib:             make(map[string]*blockInfo),
	}
}

func newPLibConfirm(block *types.Block, confirmsRequired uint16) *plibConfirm {
	return &plibConfirm{
		blockInfo: &blockInfo{
			blockHash: block.ID(),
			blockNo:   block.BlockNo(),
		},
		confirmsLeft: confirmsRequired,
	}
}

// UpdateStatus updates the last irreversible block (LIB).
func (s *Status) UpdateStatus(block *types.Block) {
	s.Lock()
	defer s.Unlock()

	if s.bestBlock == nil {
		s.bestBlock = block
		return
	}

	oldBest := s.bestBlock
	s.bestBlock = block
	if block.PrevID() != oldBest.ID() {
		logger.Debug().
			Int("len", s.plibConfirms.Len()).
			Msg("plibConfirms reset due to chain reorganization")

		// Chain reorganized.
		s.plibConfirms.Init()
		// TODO: Rebuild LIB status & update a new LIB (if needed) after the
		// last LIB, of which infomation should be able to retrieve from DB.
		return
	}

	libInfo := calcProposedLIB(s.plibConfirms, block, s.confirmsRequired)
	if libInfo != nil {
		bp := blockBP(block)
		logger.Debug().Str("BP", bp).
			Str("lib hash", libInfo.blockHash).Uint64("lib no", libInfo.blockNo).
			Str("best block hash", block.ID()).Uint64("best block no", block.BlockNo()).
			Msg("proposed LIB map updated")
		s.updateLIB(bp, libInfo)
	}
}

func calcProposedLIB(confirms *list.List, block *types.Block, confirmsRequired uint16) (bi *blockInfo) {
	pc := newPLibConfirm(block, confirmsRequired)
	confirms.PushBack(pc)
	logger.Debug().Int("len", confirms.Len()).Str("hash", spew.Sdump(pc)).Msg("new plibConfirm added")

	var c *plibConfirm
	var remEnd *list.Element
	for e := confirms.Back(); e != nil; e = e.Prev() {
		c = e.Value.(*plibConfirm)
		// TODO: Apply DPoS 3.0 modification (ranged confirmation)
		c.confirmsLeft--
		if c.confirmsLeft == 0 {
			// proposed LIB info to return
			bi = c.blockInfo
			remEnd = e.Next()
			break
		}
	}

	if c.confirmsLeft == 0 {
		// Remove all the blockInfos before remEnd. They correspond to lower
		// heights than the proposed LIB's height so we don't neet them
		// anymore.
		var next *list.Element
		for e := confirms.Front(); e != remEnd; e = next {
			next = e.Next()
			confirms.Remove(e)
		}
	}

	return
}

func (s *Status) updateLIB(bp string, libInfo *blockInfo) {
	s.plib[bp] = libInfo

	libInfos := make([]*blockInfo, 0, len(s.plib))
	for _, l := range s.plib {
		libInfos = append(libInfos, l)
	}
	// TODO: find better method.
	sort.Slice(libInfos, func(i, j int) bool {
		return libInfos[i].blockNo < libInfos[j].blockNo
	})

	s.lib = libInfos[(len(libInfos)-1)/3]
	logger.Debug().
		Str("block hash", s.lib.blockHash).
		Uint64("block no", s.lib.blockNo).
		Msg("last irreversible block (BFT) updated")
}

func blockBP(block *types.Block) string {
	return block.BPID2Str()
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
