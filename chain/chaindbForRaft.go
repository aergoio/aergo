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
	ErrInvalidRaftEntry = errors.New("invalid raft entry")
	ErrMismatchedEntry  = errors.New("mismatched entry")
)

// implement ChainWAL interface
func (cdb *ChainDB) IsNew() bool {
	//TODO
	return true
}

func (cdb *ChainDB) ReadAll() (state raftpb.HardState, ents []raftpb.Entry, err error) {
	//TODO
	return raftpb.HardState{}, nil, nil
}

func (cdb *ChainDB) WriteHardState(hardstate *raftpb.HardState) error {
	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	var data []byte
	var err error

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

	state := &raftpb.HardState{}
	if err := proto.Unmarshal(data, state); err != nil {
		logger.Panic().Msg("failed to unmarshal raft state")
		return nil, ErrInvalidHardState
	}

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

	dbTx := cdb.store.NewTx()
	for i, entry := range ents {
		logger.Debug().Str("type", consensus.WalEntryType_name[entry.Type]).Uint64("Index", entry.Index).Uint64("term", entry.Term).Msg("add raft log entry")

		if entry.Type == consensus.EntryBlock {
			if err := cdb.addBlock(&dbTx, blocks[i]); err != nil {
				dbTx.Discard()
				panic("add block entry")
				return err
			}
		}

		if data, err = entry.ToBytes(); err != nil {
			dbTx.Discard()
			return err
		}

		lastIdx = entry.Index
		dbTx.Set(getRaftEntryKey(entry.Index), data)
	}

	// set lastindex
	dbTx.Set(raftEntryLastIdxKey, types.BlockNoToBytes(lastIdx))

	dbTx.Commit()

	return nil
}

func (cdb *ChainDB) GetRaftEntry(idx uint64) (*consensus.WalEntry, error) {
	data := cdb.store.Get(getRaftEntryKey(idx))

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

func (cdb *ChainDB) WriteSnapshot(snap *raftpb.Snapshot) error {
	data, err := proto.Marshal(snap)
	if err != nil {
		return err
	}

	falseBytes, err := encodeBool(false)
	if err != nil {
		return err
	}

	dbTx := cdb.store.NewTx()
	dbTx.Set(raftSnapKey, data)
	dbTx.Set(raftSnapStatusKey, falseBytes)
	dbTx.Commit()

	return nil
}

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
