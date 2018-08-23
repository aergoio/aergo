package dpos

import (
	"container/list"
	"sync"

	"github.com/aergoio/aergo/consensus/impl/dpos"
	"github.com/aergoio/aergo/types"
)

const (
	confirmsForPLIB = dpos.BlockProducers*2/3 + 1
)

// Status manages DPoS-related infomations like LIB.
type Status struct {
	sync.RWMutex
	plibConfirms *list.List        // confirm counts to become proposed LIB
	plib         map[string]uint64 // BP-wise proposed LIB map
	bestBlock    *types.Block
}

type plibConfirm struct {
	hash         string
	confirmsLeft uint8
}

// NewStatus returns a newly allocated Status.
func NewStatus() *Status {
	return &Status{
		plibConfirms: list.New(),
	}
}

func newPLibConfirms(block *types.Block) *plibConfirm {
	return &plibConfirm{
		hash:         block.ID(),
		confirmsLeft: confirmsForPLIB,
	}
}

// StatusUpdate updates the last irreversible block (LIB).
func (s *Status) StatusUpdate(block *types.Block) {
	s.Lock()
	defer s.Unlock()

	if block.PrevID() != s.bestBlock.ID() {
		// Chain reorganization happened.
		s.plibConfirms.Init()
		// XXX More work may be need here!!!
	}

	s.bestBlock = block
}
