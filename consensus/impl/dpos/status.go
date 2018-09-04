package dpos

import (
	"container/list"
	"sync"

	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
)

// Status manages DPoS-related infomations like LIB.
type Status struct {
	sync.RWMutex
	confirmsRequired uint16
	plibConfirms     *list.List            // confirm counts to become proposed LIB
	plib             map[string]*blockInfo // BP-wise proposed LIB map
	lib              *blockInfo
	bestBlock        *types.Block
}

type blockInfo struct {
	hash  string
	blkNo uint64
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
			hash:  block.ID(),
			blkNo: block.BlockNo(),
		},
		confirmsLeft: confirmsRequired,
	}
}

// StatusUpdate updates the last irreversible block (LIB).
func (s *Status) StatusUpdate(block *types.Block) {
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
		s.plib[bp] = libInfo
		logger.Debug().Str("BP", bp).
			Str("lib hash", libInfo.hash).Uint64("lib no", libInfo.blkNo).
			Str("best block hash", block.ID()).Uint64("best block no", block.BlockNo()).
			Msg("proposed LIB map updated")
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

func blockBP(block *types.Block) string {
	return block.BPID2Str()
}
