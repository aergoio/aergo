package chain

import (
	"errors"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/enc/gob"
	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
	"github.com/aergoio/etcd/raft/raftpb"
)

var (
	ErrMismatchedEntry    = errors.New("mismatched entry")
	ErrNoWalEntry         = errors.New("no entry")
	ErrEncodeRaftIdentity = errors.New("failed encoding of raft identity")
	ErrDecodeRaftIdentity = errors.New("failed decoding of raft identity")
	ErrNoWalEntryForBlock = errors.New("no raft entry for block")
	ErrNilHardState       = errors.New("hardstateinfo must not be nil")
)

func (cdb *ChainDB) ResetWAL(hardStateInfo *types.HardStateInfo) error {
	if hardStateInfo == nil {
		return ErrNilHardState
	}

	logger.Info().Str("hardstate", hardStateInfo.ToString()).Msg("reset wal with given hardstate")

	cdb.ClearWAL()

	if err := cdb.WriteHardState(&raftpb.HardState{Term: hardStateInfo.Term, Commit: hardStateInfo.Commit}); err != nil {
		return err
	}

	// build snapshot
	var (
		snapBlock *types.Block
		err       error
	)
	if snapBlock, err = cdb.GetBestBlock(); err != nil {
		return err
	}

	snapData := consensus.NewSnapshotData(nil, nil, snapBlock)
	if snapData == nil {
		logger.Panic().Uint64("SnapBlockNo", snapBlock.BlockNo()).Msg("new snap failed")
	}

	data, err := snapData.Encode()
	if err != nil {
		return err
	}

	tmpSnapshot := raftpb.Snapshot{
		Metadata: raftpb.SnapshotMetadata{Index: hardStateInfo.Commit, Term: hardStateInfo.Term},
		Data:     data,
	}

	if err := cdb.WriteSnapshot(&tmpSnapshot); err != nil {
		logger.Fatal().Err(err).Msg("failed to save snapshot to wal")
	}

	// write initial values
	// last entry index = commit
	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	cdb.writeRaftEntryLastIndex(dbTx, hardStateInfo.Commit)

	dbTx.Commit()

	return nil
}

// ClearWAL removes all data used by raft
func (cdb *ChainDB) ClearWAL() {
	logger.Info().Msg("clear all data used by raft")

	removeAllRaftEntries := func(lastIdx uint64) {
		logger.Debug().Uint64("last", lastIdx).Msg("reset raft entries from datafiles")

		bulk := cdb.store.NewBulk()
		defer bulk.DiscardLast()

		for i := lastIdx; i >= 1; i-- {
			bulk.Delete(dbkey.RaftEntry(i))
		}

		bulk.Delete(dbkey.RaftEntryLastIdx())

		bulk.Flush()
	}

	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	dbTx.Delete(dbkey.RaftIdentity())
	// remove hardstate

	dbTx.Delete(dbkey.RaftState())

	// remove snapshot
	dbTx.Delete(dbkey.RaftSnap())

	logger.Debug().Msg("reset identify, hardstate, snapshot from datafiles")

	dbTx.Commit()

	// remove raft entries
	if last, err := cdb.GetRaftEntryLastIdx(); err == nil {
		// remove 1 ~ last raft entry
		removeAllRaftEntries(last)
	}

	logger.Debug().Msg("clear WAL done")
}

func (cdb *ChainDB) WriteHardState(hardstate *raftpb.HardState) error {
	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	var data []byte
	var err error

	logger.Info().Uint64("term", hardstate.Term).Str("vote", types.Uint64ToHexaString(hardstate.Vote)).Uint64("commit", hardstate.Commit).Msg("save hard state")

	if data, err = proto.Encode(hardstate); err != nil {
		logger.Panic().Msg("failed to marshal raft state")
	}
	dbTx.Set(dbkey.RaftState(), data)
	dbTx.Commit()

	return nil
}

func (cdb *ChainDB) GetHardState() (*raftpb.HardState, error) {
	data := cdb.store.Get(dbkey.RaftState())

	if len(data) == 0 {
		return nil, ErrWalNoHardState
	}

	state := &raftpb.HardState{}
	if err := proto.Decode(data, state); err != nil {
		logger.Panic().Msg("failed to unmarshal raft state")
	}

	logger.Info().Uint64("term", state.Term).Str("vote", types.Uint64ToHexaString(state.Vote)).Uint64("commit", state.Commit).Msg("load hard state")

	return state, nil
}

func (cdb *ChainDB) WriteRaftEntry(ents []*consensus.WalEntry, blocks []*types.Block, ccProposes []*raftpb.ConfChange) error {
	var data []byte
	var err error
	var lastIdx uint64

	// truncate conflicting entry
	last, err := cdb.GetRaftEntryLastIdx()
	if err != nil {
		return err
	}

	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	if ents[0].Index <= last {
		logger.Debug().Uint64("from", ents[0].Index).Uint64("to", last).Msg("truncate conflicting index")

		for i := ents[0].Index; i <= last; i++ {
			// delete ents[0].Index ~ lastIndex of wal
			dbTx.Delete(dbkey.RaftEntry(i))
		}
	}

	for i, entry := range ents {
		var targetNo uint64

		if entry.Type == consensus.EntryBlock {
			if err := cdb.addBlock(dbTx, blocks[i]); err != nil {
				logger.Panic().Err(err).Uint64("BlockNo", blocks[i].BlockNo()).Msg("failed to add block entry")
			}

			targetNo = blocks[i].BlockNo()
		}

		if data, err = entry.ToBytes(); err != nil {
			logger.Panic().Err(err).Uint64("BlockNo", blocks[i].BlockNo()).Uint64("index", entry.Index).Msg("failed to convert entry to bytes")
		}

		lastIdx = entry.Index
		dbTx.Set(dbkey.RaftEntry(entry.Index), data)

		// invert key to search raft entry corresponding to block hash
		if entry.Type == consensus.EntryBlock {
			dbTx.Set(dbkey.RaftEntryInvert(blocks[i].BlockHash()), types.Uint64ToBytes(entry.Index))
		}

		if entry.Type == consensus.EntryConfChange {
			if ccProposes[i] == nil {
				logger.Fatal().Str("entry", entry.ToString()).Msg("confChangePropose must not be nil")
			}
			if err := cdb.writeConfChangeProgress(dbTx, ccProposes[i].ID,
				&types.ConfChangeProgress{State: types.ConfChangeState_CONF_CHANGE_STATE_SAVED, Err: ""}); err != nil {
				return err
			}

			targetNo = ccProposes[i].ID
		}

		logger.Info().Str("type", consensus.WalEntryType_name[entry.Type]).Uint64("Index", entry.Index).Uint64("term", entry.Term).Uint64("blockNo/requestID", targetNo).Msg("add raft log entry")
	}

	// set lastindex
	cdb.writeRaftEntryLastIndex(dbTx, lastIdx)

	dbTx.Commit()

	return nil
}

func (cdb *ChainDB) writeRaftEntryLastIndex(dbTx db.Transaction, lastIdx uint64) {
	logger.Debug().Uint64("index", lastIdx).Msg("set last wal entry")

	// lastIdx is not block number but raft entry index.
	dbTx.Set(dbkey.RaftEntryLastIdx(), types.Uint64ToBytes(lastIdx))
}

func (cdb *ChainDB) GetRaftEntry(idx uint64) (*consensus.WalEntry, error) {
	data := cdb.store.Get(dbkey.RaftEntry(idx))
	if len(data) == 0 {
		return nil, ErrNoWalEntry
	}

	var entry consensus.WalEntry
	if err := gob.Decode(data, &entry); err != nil {
		return nil, err
	}

	if entry.Index != idx {
		logger.Error().Uint64("entry", entry.Index).Uint64("req", idx).Msg("mismatched wal entry")
		return nil, ErrMismatchedEntry
	}

	return &entry, nil
}

func (cdb *ChainDB) GetRaftEntryIndexOfBlock(hash []byte) (uint64, error) {
	data := cdb.store.Get(dbkey.RaftEntryInvert(hash))
	if len(data) == 0 {
		return 0, ErrNoWalEntryForBlock
	}

	idx := types.BytesToUint64(data)
	if idx == 0 {
		return 0, ErrNoWalEntryForBlock
	}

	return idx, nil
}

func (cdb *ChainDB) GetRaftEntryOfBlock(hash []byte) (*consensus.WalEntry, error) {
	idx, err := cdb.GetRaftEntryIndexOfBlock(hash)
	if err != nil {
		return nil, err
	}

	return cdb.GetRaftEntry(idx)
}

func (cdb *ChainDB) GetRaftEntryLastIdx() (uint64, error) {
	lastBytes := cdb.store.Get(dbkey.RaftEntryLastIdx())
	if lastBytes == nil || len(lastBytes) == 0 {
		return 0, nil
	}

	return types.BlockNoFromBytes(lastBytes), nil
}

var (
	ErrWalNotEqualIdentityName   = errors.New("name of identity is not equal")
	ErrWalNotEqualIdentityPeerID = errors.New("peerid of identity is not equal")
)

// HasWal checks chaindb has valid status of Raft WAL.
// 1. compare identity with config
// 2. check if hardstate exists
// 3. check if last raft entry index exists
// last entry index can be 0 if first sync has failed
func (cdb *ChainDB) HasWal(identity consensus.RaftIdentity) (bool, error) {
	var (
		id   *consensus.RaftIdentity
		last uint64
		hs   *raftpb.HardState
		err  error
	)

	if id, err = cdb.GetIdentity(); err != nil || id == nil {
		return false, err
	}

	if id.Name != identity.Name {
		logger.Debug().Str("config name", identity.Name).Str("saved id", id.Name).Msg("unmatched name of identity")
		return false, ErrWalNotEqualIdentityName
	}

	if id.PeerID != identity.PeerID {
		logger.Debug().Str("config peerid", identity.PeerID).Str("saved id", id.PeerID).Msg("unmatched peerid of identity")
		return false, ErrWalNotEqualIdentityPeerID
	}

	if hs, err = cdb.GetHardState(); err != nil {
		return false, err
	}

	if last, err = cdb.GetRaftEntryLastIdx(); err != nil {
		return false, err
	}

	logger.Info().Str("identity", id.ToString()).Str("hardstate", types.RaftHardStateToString(*hs)).Uint64("lastidx", last).Msg("existing wal status")

	return true, nil
}

/*
func encodeBool(v bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, v)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decodeBool(data []byte) (bool, error) {
	var val bool
	bufreader := bytes.NewReader(data)
	if err := binary.Read(bufreader, binary.LittleEndian, &val); err != nil {
		return false, err
	}

	return val, nil
}
*/

func (cdb *ChainDB) WriteSnapshot(snap *raftpb.Snapshot) error {
	var snapdata = consensus.SnapshotData{}
	err := snapdata.Decode(snap.Data)
	if err != nil {
		logger.Fatal().Msg("failed to unmarshal snapshot data to write")
		return err
	}

	logger.Debug().Str("snapshot", consensus.SnapToString(snap, &snapdata)).Msg("write snapshot to wal")
	data, err := proto.Encode(snap)
	if err != nil {
		return err
	}

	dbTx := cdb.store.NewTx()
	dbTx.Set(dbkey.RaftSnap(), data)
	dbTx.Commit()

	return nil
}

/*
	func (cdb *ChainDB) WriteSnapshotDone() error {
		data, err := encodeBool(true)
		if err != nil {
			return err
		}

		dbTx := cdb.store.NewTx()
		dbTx.Set(raftSnapStatusKey, data)
		dbTx.Commit()

		return nil
	}

	func (cdb *ChainDB) GetSnapshotDone() (bool, error) {
		data := cdb.store.Get(raftSnapStatusKey)
		if len(data) == 0 {
			return false, nil
		}

		val, err := decodeBool(data)
		if err != nil {
			return false, err
		}

		return val, nil
	}
*/

func (cdb *ChainDB) GetSnapshot() (*raftpb.Snapshot, error) {
	data := cdb.store.Get(dbkey.RaftSnap())
	if len(data) == 0 {
		return nil, nil
	}

	snap := &raftpb.Snapshot{}
	if err := proto.Decode(data, snap); err != nil {
		logger.Panic().Msg("failed to unmarshal raft snap")
		return nil, ErrInvalidRaftSnapshot
	}

	if snap.Data == nil {
		logger.Panic().Msg("raft snap data is nil")
		return nil, ErrInvalidRaftSnapshot
	}

	return snap, nil
}

func (cdb *ChainDB) WriteIdentity(identity *consensus.RaftIdentity) error {
	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	logger.Info().Str("id", identity.ToString()).Msg("save raft identity")

	enc, err := gob.Encode(identity)
	if err != nil {
		return ErrEncodeRaftIdentity
	}

	dbTx.Set(dbkey.RaftIdentity(), enc)
	dbTx.Commit()

	return nil
}

func (cdb *ChainDB) GetIdentity() (*consensus.RaftIdentity, error) {
	data := cdb.store.Get(dbkey.RaftIdentity())
	if len(data) == 0 {
		return nil, nil
	}

	var id consensus.RaftIdentity
	if err := gob.Decode(data, &id); err != nil {
		return nil, ErrDecodeRaftIdentity
	}

	logger.Info().Str("id", types.Uint64ToHexaString(id.ID)).Str("name", id.Name).Str("peerid", id.PeerID).Msg("get raft identity")

	return &id, nil
}

func (cdb *ChainDB) WriteConfChangeProgress(id uint64, progress *types.ConfChangeProgress) error {
	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	if err := cdb.writeConfChangeProgress(dbTx, id, progress); err != nil {
		return err
	}

	dbTx.Commit()

	return nil
}

func (cdb *ChainDB) writeConfChangeProgress(dbTx db.Transaction, id uint64, progress *types.ConfChangeProgress) error {
	if id == 0 {
		// it's for intial member's for startup
		return nil
	}

	// Make CC Data
	var data []byte
	var err error

	if data, err = proto.Encode(progress); err != nil {
		logger.Error().Msg("failed to marshal confChangeProgress")
		return err
	}

	dbTx.Set(dbkey.RaftConfChangeProgress(id), data)

	return nil
}

func (cdb *ChainDB) GetConfChangeProgress(id uint64) (*types.ConfChangeProgress, error) {
	data := cdb.store.Get(dbkey.RaftConfChangeProgress(id))
	if len(data) == 0 {
		return nil, nil
	}

	var progress types.ConfChangeProgress

	if err := proto.Decode(data, &progress); err != nil {
		logger.Error().Msg("failed to unmarshal raft state")
		return nil, ErrInvalidCCProgress
	}

	logger.Info().Uint64("id", id).Str("status", progress.ToString()).Msg("get conf change status")

	return &progress, nil
}
