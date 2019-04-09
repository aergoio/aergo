package consensus

import (
	"bytes"
	"encoding/gob"
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
)

type WalEntry struct {
	Type  EntryType
	Term  uint64
	Index uint64
	Data  []byte // hash is set if Type is EntryBlock
}

func (we *WalEntry) ToBytes() ([]byte, error) {
	var val bytes.Buffer
	gob := gob.NewEncoder(&val)
	if err := gob.Encode(we); err != nil {
		panic("raft entry to bytes error")
		return nil, err
	}

	return val.Bytes(), nil
}

type ChainWAL interface {
	ChainDB

	IsNew() bool
	ReadAll() (state raftpb.HardState, ents []raftpb.Entry, err error)
	WriteRaftEntry([]*WalEntry, []*types.Block) error
	GetHardState() (*raftpb.HardState, error)
	WriteHardState(hardstate *raftpb.HardState) error
}
