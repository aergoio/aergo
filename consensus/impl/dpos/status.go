package dpos

import (
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/impl/dpos/bp"
	"github.com/aergoio/aergo/types"
)

var bsLoader *bootLoader

// Status manages DPoS-related infomations like LIB.
type Status struct {
	sync.RWMutex
	bestBlock *types.Block
	libState  *libStatus
	bps       *bp.Snapshots
	done      bool
}

// NewStatus returns a newly allocated Status.
func NewStatus(bpCount uint16, cdb consensus.ChainDbReader) *Status {
	s := &Status{
		libState: newLibStatus(consensusBlockCount(bpCount)),
		bps:      bp.NewSnapshots(bpCount),
	}
	s.init(cdb)

	return s
}

// load restores the last LIB status by using the informations loaded from the
// DB.
func (s *Status) load() {
	if s.done {
		return
	}

	s.bestBlock = bsLoader.bestBlock()

	s.libState = bsLoader.ls

	if bsLoader.ls != nil {
		s.libState = bsLoader.ls
	}

	genesisBlock := bsLoader.genesisBlock()
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

	// TODO: move the lib status loading to dpos.NewStatus.
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
		// Rollback resulting from a reorganization.
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

	s.libState.gc()

	s.bestBlock = block
}

func (s *Status) updateLIB(lib *blockInfo) {
	s.libState.Lib = lib

	logger.Debug().
		Str("block hash", s.libState.Lib.BlockHash).
		Uint64("block no", s.libState.Lib.BlockNo).
		Int("confirms len", s.libState.confirms.Len()).
		Msg("last irreversible block (BFT) updated")
}

// Save saves the consensus status information for the later recovery.
func (s *Status) Save(tx db.Transaction) error {
	return s.libState.save(tx)
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

// init recovers the last DPoS status including pre-LIB map and confirms
// list between LIB and the best block.
func (s *Status) init(cdb consensus.ChainDbReader) {
	if cdb == nil {
		return
	}

	genesis, err := cdb.GetBlockByNo(0)
	if err != nil {
		panic(err)
	}

	best, err := cdb.GetBestBlock()
	if err != nil {
		best = genesis
	}

	bsLoader = &bootLoader{
		ls:               newLibStatus(s.libState.confirmsRequired),
		best:             best,
		genesis:          genesis,
		cdb:              cdb,
		confirmsRequired: s.libState.confirmsRequired,
	}

	bsLoader.load()
}

type bootLoader struct {
	ls               *libStatus
	best             *types.Block
	genesis          *types.Block
	bpIDs            []string
	cdb              consensus.ChainDbReader
	confirmsRequired uint16
}

func (bs *bootLoader) load() {
	if ls := bs.loadLibStatus(); ls != nil {
		bs.ls = ls
		logger.Debug().Int("proposed lib len", len(ls.Prpsd)).Msg("lib status loaded from DB")
		for id, p := range ls.Prpsd {
			if p == nil {
				continue
			}
			logger.Debug().Str("BPID", id).
				Str("confirmed hash", p.Plib.Hash()).
				Str("confirmedBy hash", p.PlibBy.Hash()).
				Msg("pre-LIB entry")
		}
	}

	if gi := bs.cdb.GetGenesisInfo(); gi != nil {
		bs.bpIDs = gi.BPs
		for i, bp := range bs.bpIDs {
			logger.Info().Int("index", i).Str("BPID", bp).Msg("initial BP")
		}
	}
}

// GetInitialBPs returns the initial BP IDs, which are loaded from Genesis
// info in the chain DB.
func GetInitialBPs() []string {
	return bsLoader.bpIDs
}
