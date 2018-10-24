package dpos

import (
	"bytes"
	"encoding/gob"
	"io"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/types"
)

// Status manages DPoS-related infomations like LIB.
type Status struct {
	sync.RWMutex
	bestBlock *types.Block
	libState  *libStatus
	done      bool
}

// NewStatus returns a newly allocated Status.
func NewStatus(confirmsRequired uint16) *Status {
	return &Status{
		libState: newLibStatus(confirmsRequired),
	}
}

// load restores the last LIB status by using the informations loaded from the
// DB.
func (s *Status) load() {
	if s.done {
		return
	}

	s.bestBlock = libLoader.bestBlock()

	s.libState = libLoader.ls

	if libLoader.ls != nil {
		s.libState = libLoader.ls
	}

	genesisBlock := libLoader.genesisBlock()
	s.libState.genesisInfo = &blockInfo{
		BlockHash: genesisBlock.ID(),
		BlockNo:   genesisBlock.BlockNo(),
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
		s.libState.addConfirmInfo(block)

		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("block no", block.BlockNo()).
			Msg("update LIB status")

		// Block connected
		if lib := s.libState.update(); lib != nil {
			s.updateLIB(lib)
		}

	} else {
		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("target block no", block.BlockNo()).
			Msg("rollback LIB status")

		// Block reorganized. TODO: update consensus status, correctly.
		if err := s.libState.rollbackStatusTo(block, s.libState.Lib); err != nil {
			logger.Debug().Err(err).Msg("failed to rollback DPoS status")
			panic(err)
		}
	}

	s.bestBlock = block
}

func (s *Status) updateLIB(lib *blockInfo) {
	s.libState.Lib = lib
	s.libState.gc(lib)

	logger.Debug().
		Str("block hash", s.libState.Lib.BlockHash).
		Uint64("block no", s.libState.Lib.BlockNo).
		Int("confirms len", s.libState.confirms.Len()).
		Msg("last irreversible block (BFT) updated")
}

// Save saves the consensus status information for the later recovery.
func (s *Status) Save(tx db.Transaction) error {
	if err := s.libState.save(tx); err != nil {
		return err
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
	s.RLock()
	defer s.RUnlock()

	if s.libState.Lib == nil {
		logger.Debug().Uint64("branch root no", rootNo).Msg("no LIB")
		return true
	}

	libNo := s.libState.Lib.BlockNo

	reorganizable := rootNo >= libNo
	if !reorganizable {
		logger.Info().
			Uint64("LIB", libNo).
			Uint64("branch root no", rootNo).
			Msg("reorganization beyond LIB is not allowed")
	}

	return reorganizable
}

// Init recovers the last DPoS status including pre-LIB map and confirms
// list between LIB and the best block.
func (s *Status) Init(cdb consensus.ChainDbReader) {
	genesis, err := cdb.GetBlockByNo(0)
	if err != nil {
		panic(err)
	}

	best, err := cdb.GetBestBlock()
	if err != nil {
		best = genesis
	}

	libLoader = &bootLoader{
		ls:      newLibStatus(defaultConsensusCount),
		best:    best,
		genesis: genesis,
		cdb:     cdb,
	}

	libLoader.load()
}
