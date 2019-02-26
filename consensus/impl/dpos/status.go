package dpos

import (
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/impl/dpos/bp"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var bsLoader *bootLoader

// Status manages DPoS-related infomations like LIB.
type Status struct {
	sync.RWMutex
	done      bool
	bestBlock *types.Block
	libState  *libStatus
	bps       *bp.Snapshots
}

// NewStatus returns a newly allocated Status.
func NewStatus(c bp.ClusterMember, cdb consensus.ChainDB, sdb *state.ChainStateDB, resetHeight types.BlockNo) *Status {
	s := &Status{
		libState: newLibStatus(consensusBlockCount(c.Size())),
		bps:      bp.NewSnapshots(c, cdb, sdb),
	}
	s.init(cdb, resetHeight)

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

		s.bps.AddSnapshot(block.BlockNo())
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

		// Rollback BP list. -- BP list is alos affected by a fork.
		s.bps.UpdateCluster(block.BlockNo())
	}

	s.libState.gc()

	s.bestBlock = block
}

func (s *Status) libBlockNo() types.BlockNo {
	s.RLock()
	defer s.RUnlock()
	return s.libState.libNo()
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
	s.Lock()
	defer s.Unlock()

	if err := s.libState.save(tx); err != nil {
		return err
	}

	return nil
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
func (s *Status) init(cdb consensus.ChainDB, resetHeight types.BlockNo) {
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

	bsLoader.load(resetHeight)
}

type bootLoader struct {
	ls               *libStatus
	best             *types.Block
	genesis          *types.Block
	cdb              consensus.ChainDB
	confirmsRequired uint16
}

func (bs *bootLoader) load(resetHeight types.BlockNo) {
	if ls := bs.loadLibStatus(); ls != nil {
		bs.ls = ls
		logger.Debug().Int("proposed lib len", len(ls.Prpsd)).Msg("lib status loaded from DB")

		for id, p := range ls.Prpsd {
			if p == nil {
				continue
			}

			// Remove the too high pre-LIBs when the ForceResetHeight parameter is set.
			if resetHeight > 0 && (p.Plib.BlockNo > resetHeight || p.PlibBy.BlockNo > resetHeight) {
				delete(ls.Prpsd, id)
			}

			logger.Debug().Str("BPID", id).
				Str("confirmed hash", p.Plib.Hash()).
				Str("confirmedBy hash", p.PlibBy.Hash()).
				Msg("pre-LIB entry")
		}

		// Reset the LIB when the ForceResetHeight parameter is set.
		if resetHeight > 0 && ls.Lib.BlockNo > resetHeight {
			ls.Lib = &blockInfo{
				BlockHash: bs.genesisBlock().ID(),
				BlockNo:   bs.genesisBlock().BlockNo(),
			}

			tx := bs.cdb.NewTx()
			reset(tx)
			tx.Commit()
		}
	}
}
