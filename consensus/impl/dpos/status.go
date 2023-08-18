package dpos

import (
	"encoding/json"
	"sync"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/consensus/impl/dpos/bp"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

var bsLoader *bootLoader

// Status manages DPoS-related information like LIB.
type Status struct {
	sync.RWMutex
	done      bool
	bestBlock *types.Block
	libState  *libStatus
	bps       *bp.Snapshots
	sdb       *state.ChainStateDB
}

// NewStatus returns a newly allocated Status.
func NewStatus(c bp.ClusterMember, cdb consensus.ChainDB, sdb *state.ChainStateDB, resetHeight types.BlockNo) *Status {
	s := &Status{
		libState: newLibStatus(c.Size()),
		bps:      bp.NewSnapshots(c, cdb, sdb),
		sdb:      sdb,
	}
	s.init(cdb, resetHeight)

	return s
}

// load restores the last LIB status by using the information loaded from the
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

	var bps []string

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

		bps, _ = s.bps.AddSnapshot(block.BlockNo())
	} else {
		// Rollback resulting from a reorganization: The code below assumes
		// that there is no block-by-block rollback; it assumes that the
		// rollback procedure is performed by simply replacing the current
		// state DB into that of the branch block.
		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("target block no", block.BlockNo()).
			Msg("rollback LIB status")

		// Block reorganized. TODO: update consensus status, correctly.
		if err := s.libState.rollbackStatusTo(block, s.libState.Lib); err != nil {
			logger.Fatal().Err(err).Msg("failed to rollback DPoS status")
		}

		// Rollback BP list. -- BP list is alos affected by a fork.
		bps = s.bps.UpdateCluster(block.BlockNo())

		// Rollback Voting Power Rank: the snapshot fully re-loaded from the
		// branch block. TODO: let's find a smarter way or use parallel
		// loading.
		if err := InitVPR(s.sdb.OpenNewStateDB(block.GetHeader().GetBlocksRootHash())); err != nil {
			logger.Fatal().Err(err).Msg("failed to rollback Voting Power Rank")
		} else {
			logger.Debug().Uint64("from block no", block.BlockNo()).Msg("VPR reloaded")
		}
	}

	s.libState.gc(bps)
	s.libState.setConfirmsRequired(s.bps.Size())

	s.bestBlock = block
}

func (s *Status) libNo() types.BlockNo {
	s.RLock()
	defer s.RUnlock()
	return s.libState.libNo()
}

func (s *Status) lib() *blockInfo {
	s.RLock()
	defer s.RUnlock()
	return s.libState.lib()
}

func (s *Status) libAsJSON() *json.RawMessage {
	lib := s.lib()
	if lib == nil || lib.BlockNo == 0 {
		return nil
	}

	l := &struct {
		LibHash string
		LibNo   types.BlockNo
	}{
		LibHash: lib.BlockHash,
		LibNo:   lib.BlockNo,
	}

	if b, err := json.Marshal(l); err == nil {
		m := json.RawMessage(b)
		return &m
	}

	return nil
}

func (s *Status) updateLIB(lib *blockInfo) {
	s.libState.Lib = lib

	logger.Debug().
		Str("block hash", s.libState.Lib.BlockHash).
		Uint64("block no", s.libState.Lib.BlockNo).
		Int("confirms len", s.libState.confirms.Len()).
		Int("pm len", len(s.libState.Prpsd)).
		Msg("last irreversible block (BFT) updated")
}

// Save saves the consensus status information for the later recovery.
func (s *Status) Save(tx consensus.TxWriter) error {
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

// Info returns the current last irreversible block information as a JSON
// string.
func (s *Status) Info() string {
	return s.String()
}

// String returns the current LIB as a JSON string.
func (s *Status) String() string {
	info := consensus.NewInfo(GetName())
	info.Status = s.libAsJSON()

	return info.AsJSON()
}

func (s *Status) lpbNo() types.BlockNo {
	return s.libState.LpbNo
}

// init recovers the last DPoS status including pre-LIB map and confirms
// list between LIB and the best block.
func (s *Status) init(cdb consensus.ChainDB, resetHeight types.BlockNo) {
	if cdb == nil {
		return
	}

	genesis, err := cdb.GetBlockByNo(0)
	if err != nil {
		logger.Panic().Err(err).Msg("failed to get genesis block")
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
