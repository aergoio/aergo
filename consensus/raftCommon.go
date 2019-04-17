package consensus

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
	"github.com/aergoio/etcd/raft/raftpb"
)

type EntryType int8

const (
	EntryBlock EntryType = iota
	EntryEmpty           // it is generated when node becomes leader
	EntryConfChange
)

var (
	WalEntryType_name = map[EntryType]string{
		0: "EntryBlock",
		1: "EntryEmpty",
		2: "EntryConfChange",
	}

	ErrEmptySnapData = errors.New("failed to decode snapshot data. encoded data is empty")
)

type WalEntry struct {
	Type  EntryType
	Term  uint64
	Index uint64
	Data  []byte // hash is set if Type is EntryBlock
}

func (we *WalEntry) ToBytes() ([]byte, error) {
	var val bytes.Buffer
	enc := gob.NewEncoder(&val)
	if err := enc.Encode(we); err != nil {
		panic("raft entry to bytes error")
		return nil, err
	}

	return val.Bytes(), nil
}

func (we *WalEntry) ToString() string {
	if we == nil {
		return "wal entry is nil"
	}
	return fmt.Sprintf("wal entry[type:%s, index:%d, term:%d", WalEntryType_name[we.Type], we.Index, we.Term)
}

type ChainWAL interface {
	ChainDB

	IsWALInited() bool
	GetBlock(blockHash []byte) (*types.Block, error)
	ReadAll() (state raftpb.HardState, ents []raftpb.Entry, err error)
	WriteRaftEntry([]*WalEntry, []*types.Block) error
	GetRaftEntry(idx uint64) (*WalEntry, error)
	GetRaftEntryLastIdx() (uint64, error)
	GetHardState() (*raftpb.HardState, error)
	WriteHardState(hardstate *raftpb.HardState) error
	WriteSnapshot(snap *raftpb.Snapshot, done bool) error
	GetSnapshot() (*raftpb.Snapshot, error)
}

type ChainSnapshot struct {
	No   types.BlockNo
	Hash []byte
}

func NewChainSnapshot(block *types.Block) *ChainSnapshot {
	if block == nil {
		return nil
	}

	return &ChainSnapshot{No: block.BlockNo(), Hash: block.BlockHash()}
}

func (csnap *ChainSnapshot) ToBytes() ([]byte, error) {
	var val bytes.Buffer

	enc := gob.NewEncoder(&val)
	if err := enc.Encode(csnap); err != nil {
		logger.Fatal().Err(err).Msg("failed to encode chainsnap")
		return nil, err
	}

	return val.Bytes(), nil
}

func (csnap *ChainSnapshot) ToString() string {
	if csnap == nil || csnap.Hash == nil {
		return "csnap: empty"
	}
	return fmt.Sprintf("chainsnap:(no=%d, hash=%s)", csnap.No, enc.ToString(csnap.Hash))
}

func DecodeChainSnapshot(data []byte) (*ChainSnapshot, error) {
	var snap ChainSnapshot
	var b bytes.Buffer
	b.Write(data)

	if data == nil {
		return nil, ErrEmptySnapData
	}

	decoder := gob.NewDecoder(&b)
	if err := decoder.Decode(&snap); err != nil {
		logger.Fatal().Err(err).Msg("failed to decode chainsnap")
		return nil, err
	}

	return &snap, nil
}

func ConfStateToString(conf *raftpb.ConfState) string {
	var buf string

	buf = fmt.Sprintf("node")
	for _, node := range conf.Nodes {
		buf = buf + fmt.Sprintf("[%d]", node)
	}

	buf = buf + fmt.Sprintf("\nlearner")
	for _, learner := range conf.Learners {
		buf = buf + fmt.Sprintf("[%d]", learner)
	}
	return buf
}

func SnapToString(snap *raftpb.Snapshot, chainSnap *ChainSnapshot) string {
	var buf string
	buf = buf + fmt.Sprintf("snap=[index:%d term:%d conf:%s]", snap.Metadata.Index, snap.Metadata.Term, ConfStateToString(&snap.Metadata.ConfState))

	if chainSnap != nil {
		buf = buf + fmt.Sprintf(", chain=[no:%d hash:%s]", chainSnap.No, enc.ToString(chainSnap.Hash))
	}

	return buf
}
