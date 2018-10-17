package dpos

import (
	"bytes"
	"encoding/gob"
	"io"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
)

// Status manages DPoS-related infomations like LIB.
type Status struct {
	sync.RWMutex
	bestBlock *types.Block
	pls       *pLibStatus
	lib       *blockInfo
	done      bool
}

// NewStatus returns a newly allocated Status.
func NewStatus(confirmsRequired uint16) *Status {
	return &Status{
		pls: newPlibStatus(confirmsRequired),
	}
}

// load restores the last LIB status by using the informations loaded from the
// DB.
func (s *Status) load() {
	if s.done {
		return
	}

	s.bestBlock = bootState.bestBlock()

	genesisBlock := bootState.genesisBlock()
	s.pls.genesisInfo = &blockInfo{
		BlockHash: genesisBlock.ID(),
		BlockNo:   genesisBlock.BlockNo(),
	}

	//s.pls.addConfirmInfo(s.bestBlock)

	s.lib = bootState.lib

	if len(bootState.plib) != 0 {
		s.pls.plib = bootState.plib
	}

	if bootState.confirms != nil {
		s.pls.confirms = bootState.confirms
		//dumpConfirmInfo("XXX CONFIRMS XXX", s.pls.confirms)
	}
	s.done = true
}

// Update updates the last irreversible block (LIB).
func (s *Status) Update(block *types.Block) {
	s.Lock()
	defer s.Unlock()

	s.load()

	curBestID := s.bestBlock.ID()
	if curBestID == block.PrevID() {
		s.pls.addConfirmInfo(block)

		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("block no", block.BlockNo()).
			Msg("update LIB status")

		// Block connected
		if lib := s.pls.update(); lib != nil {
			s.updateLIB(lib)
		}

	} else {
		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("target block no", block.BlockNo()).
			Msg("rollback LIB status")

		// Block reorganized. TODO: update consensus status, correctly.
		if err := s.pls.rollbackStatusTo(block, s.lib); err != nil {
			logger.Debug().Err(err).Msg("failed to rollback DPoS status")
			panic(err)
		}
	}

	s.bestBlock = block
}

func (s *Status) updateLIB(lib *blockInfo) {
	s.lib = lib
	s.pls.gc(lib)

	logger.Debug().
		Str("block hash", s.lib.BlockHash).
		Uint64("block no", s.lib.BlockNo).
		Int("confirms len", s.pls.confirms.Len()).
		Msg("last irreversible block (BFT) updated")
}

// Save saves the consensus status information for the later recovery.
func (s *Status) Save(tx db.Transaction) error {
	if len(s.pls.plib) != 0 {
		buf, err := encode(s.pls.plib)
		if err != nil {
			return err
		}
		plib := buf.Bytes()

		tx.Set(statusKeyPreLIB, plib)
	}

	if s.lib != nil {
		buf, err := encode(s.lib)
		if err != nil {
			return err
		}
		lib := buf.Bytes()

		tx.Set(statusKeyLIB, lib)
	}

	return nil
}

func encode(e interface{}) (bytes.Buffer, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(e)

	return buf, err
}

func decode(r io.Reader, e interface{}) error {
	dec := gob.NewDecoder(r)
	return dec.Decode(e)
}

// NeedReorganization reports whether reorganization is needed or not.
func (s *Status) NeedReorganization(rootNo types.BlockNo) bool {
	return true

	/*
			s.RLock()
			defer s.RUnlock()

			if s.lib == nil {
				logger.Debug().Uint64("branch root no", rootNo).Msg("no LIB")
				return true
			}

			libNo := s.lib.BlockNo

			reorganizable := rootNo >= libNo
			if reorganizable {
				logger.Info().
					Uint64("LIB", libNo).
					Uint64("branch root no", rootNo).
					Msg("not reorganizable - the current main branch has a LIB.")
			}

		return reorganizable
	*/
}

// Init recovers the last DPoS status including pre-LIB map and confirms
// list between LIB and the best block.
func (s *Status) Init(genesis, best *types.Block, get func([]byte) []byte,
	getBlock func(types.BlockNo) (*types.Block, error)) {

	bootState = &bootingStatus{
		plib:     make(preLIB),
		lib:      &blockInfo{},
		best:     best,
		genesis:  genesis,
		get:      get,
		getBlock: getBlock,
	}

	bootState.load()
	bootState.loadConfirms()
}
