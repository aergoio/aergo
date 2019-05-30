package chain

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/gogo/protobuf/proto"
)

var (
	ErrMismatchedEntry    = errors.New("mismatched entry")
	ErrNoWalEntry         = errors.New("no entry")
	ErrEncodeRaftIdentity = errors.New("failed encoding of raft identity")
	ErrDecodeRaftIdentity = errors.New("failed decoding of raft identity")
)

func (cdb *ChainDB) ResetWAL() {
	logger.Info().Msg("reset given datafiles to use in joined node")

	removeAllRaftEntries := func(lastIdx uint64) {
		bulk := cdb.store.NewBulk()
		defer bulk.DiscardLast()

		for i := lastIdx; i >= 0; i-- {
			bulk.Delete(getRaftEntryKey(i))
		}

		bulk.Delete(raftEntryLastIdxKey)

		bulk.Flush()

		logger.Debug().Msg("reset raft entries from datafiles")
	}

	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	dbTx.Delete(raftIdentityKey)
	// remove hardstate

	dbTx.Delete(raftStateKey)

	// remove snapshot
	dbTx.Delete(raftSnapKey)

	logger.Debug().Msg("reset identify, hardstate, snapshot from datafiles")

	dbTx.Commit()

	// remove raft entries
	if last, err := cdb.GetRaftEntryLastIdx(); err == nil {
		// remove 1 ~ last raft entry
		removeAllRaftEntries(last)
	}
}

func (cdb *ChainDB) WriteHardState(hardstate *raftpb.HardState) error {
	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	var data []byte
	var err error

	logger.Info().Uint64("term", hardstate.Term).Uint64("vote", hardstate.Vote).Uint64("commit", hardstate.Commit).Msg("save hard state")

	if data, err = proto.Marshal(hardstate); err != nil {
		logger.Panic().Msg("failed to marshal raft state")
		return err
	}
	dbTx.Set(raftStateKey, data)
	dbTx.Commit()

	return nil
}

func (cdb *ChainDB) GetHardState() (*raftpb.HardState, error) {
	data := cdb.store.Get(raftStateKey)

	if len(data) == 0 {
		return nil, ErrWalNoHardState
	}

	state := &raftpb.HardState{}
	if err := proto.Unmarshal(data, state); err != nil {
		logger.Panic().Msg("failed to unmarshal raft state")
		return nil, ErrInvalidHardState
	}

	logger.Info().Uint64("term", state.Term).Uint64("vote", state.Vote).Uint64("commit", state.Commit).Msg("load hard state")

	return state, nil
}

func getRaftEntryKey(idx uint64) []byte {
	var key bytes.Buffer
	key.Write(raftEntryPrefix)
	l := make([]byte, 8)
	binary.LittleEndian.PutUint64(l[:], idx)
	key.Write(l)
	return key.Bytes()
}

func (cdb *ChainDB) WriteRaftEntry(ents []*consensus.WalEntry, blocks []*types.Block) error {
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
			dbTx.Delete(getRaftEntryKey(i))
		}
	}

	for i, entry := range ents {
		logger.Debug().Str("type", consensus.WalEntryType_name[entry.Type]).Uint64("Index", entry.Index).Uint64("term", entry.Term).Msg("add raft log entry")

		if entry.Type == consensus.EntryBlock {
			if err := cdb.addBlock(&dbTx, blocks[i]); err != nil {
				panic("add block entry")
				return err
			}
		}

		if data, err = entry.ToBytes(); err != nil {
			return err
		}

		lastIdx = entry.Index
		dbTx.Set(getRaftEntryKey(entry.Index), data)
	}

	// set lastindex
	logger.Debug().Uint64("index", lastIdx).Msg("set last wal entry")

	dbTx.Set(raftEntryLastIdxKey, types.BlockNoToBytes(lastIdx))

	dbTx.Commit()

	return nil
}

func (cdb *ChainDB) GetRaftEntry(idx uint64) (*consensus.WalEntry, error) {
	data := cdb.store.Get(getRaftEntryKey(idx))
	if len(data) == 0 {
		return nil, ErrNoWalEntry
	}

	var entry consensus.WalEntry
	var b bytes.Buffer
	b.Write(data)
	decoder := gob.NewDecoder(&b)
	if err := decoder.Decode(&entry); err != nil {
		return nil, err
	}

	if entry.Index != idx {
		logger.Error().Uint64("entry", entry.Index).Uint64("req", idx).Msg("mismatched wal entry")
		return nil, ErrMismatchedEntry
	}

	return &entry, nil
}

func (cdb *ChainDB) GetRaftEntryLastIdx() (uint64, error) {
	lastBytes := cdb.store.Get(raftEntryLastIdxKey)
	if lastBytes == nil || len(lastBytes) == 0 {
		return 0, nil
	}

	return types.BlockNoFromBytes(lastBytes), nil
}

var (
	ErrWalNotEqualIdentity = errors.New("identity of wal is not equal")
)

// HasWal checks chaindb has valid status of Raft WAL.
// 1. compare identity with config
// 2. check if hardstate exists
// 3. check if last raft entiry index exists
func (cdb *ChainDB) HasWal(identity consensus.RaftIdentity) (bool, error) {
	var (
		id   *consensus.RaftIdentity
		last uint64
		err  error
	)

	if id, err = cdb.GetIdentity(); err != nil {
		return false, err
	}

	if id.Name != identity.Name {
		return false, ErrWalNotEqualIdentity
	}

	if _, err = cdb.GetHardState(); err != nil {
		return false, err
	}

	if last, err = cdb.GetRaftEntryLastIdx(); err != nil {
		return false, err
	}

	if last > 0 {
		return true, nil
	}

	return false, nil
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
	data, err := proto.Marshal(snap)
	if err != nil {
		return err
	}

	dbTx := cdb.store.NewTx()
	dbTx.Set(raftSnapKey, data)
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
	data := cdb.store.Get(raftSnapKey)
	if len(data) == 0 {
		return nil, nil
	}

	snap := &raftpb.Snapshot{}
	if err := proto.Unmarshal(data, snap); err != nil {
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

	var val bytes.Buffer

	gob := gob.NewEncoder(&val)
	if err := gob.Encode(identity); err != nil {
		return ErrEncodeRaftIdentity
	}

	dbTx.Set(raftIdentityKey, val.Bytes())
	dbTx.Commit()

	return nil
}

func (cdb *ChainDB) GetIdentity() (*consensus.RaftIdentity, error) {
	data := cdb.store.Get(raftIdentityKey)
	if len(data) == 0 {
		return nil, nil
	}

	var id consensus.RaftIdentity
	var b bytes.Buffer
	b.Write(data)
	decoder := gob.NewDecoder(&b)
	if err := decoder.Decode(&id); err != nil {
		return nil, ErrDecodeRaftIdentity
	}

	logger.Info().Uint64("id", id.ID).Str("name", id.Name).Msg("save raft identity")

	return &id, nil
}
