package raftv2

import (
	"errors"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/types"
	"github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
)

var (
	ErrInvalidEntry = errors.New("Invalid raftpb.entry")
)

type WalDB struct {
	consensus.ChainWAL
}

func NewWalDB(chainWal consensus.ChainWAL) *WalDB {
	return &WalDB{chainWal}
}

func (wal *WalDB) SaveEntry(state raftpb.HardState, entries []raftpb.Entry) error {
	if raft.IsEmptyHardState(state) && len(entries) == 0 {
		return nil
	}

	// save hardstate
	if err := wal.WriteHardState(&state); err != nil {
		return err
	}

	walEnts, blocks := wal.convertFromRaft(entries)

	if err := wal.WriteRaftEntry(walEnts, blocks); err != nil {
		return err
	}

	return nil
}

func (wal *WalDB) convertFromRaft(entries []raftpb.Entry) ([]*consensus.WalEntry, []*types.Block) {
	lenEnts := len(entries)
	if lenEnts == 0 {
		return []*consensus.WalEntry{}, nil
	}

	getWalEntryType := func(entry *raftpb.Entry) consensus.EntryType {
		switch entry.Type {
		case raftpb.EntryNormal:
			if entry.Data != nil {
				return consensus.EntryBlock
			} else {
				return consensus.EntryEmpty
			}
		case raftpb.EntryConfChange:
			return consensus.EntryConfChange
		default:
			panic("not support raftpb entrytype")
		}
	}

	getWalData := func(entry *raftpb.Entry) (*types.Block, []byte, error) {
		if entry.Type == raftpb.EntryNormal && entry.Data != nil {
			block, err := unmarshalEntryData(entry.Data)
			if err != nil {
				return nil, nil, ErrInvalidEntry
			}

			return block, block.BlockHash(), nil
		} else {
			return nil, entry.Data, nil
		}

		return nil, nil, nil
	}

	blocks := make([]*types.Block, lenEnts)
	walents := make([]*consensus.WalEntry, lenEnts)

	var (
		data  []byte
		block *types.Block
		err   error
	)
	for i, entry := range entries {
		if block, data, err = getWalData(&entry); err != nil {
			panic("entry unmarshalEntryData error")
		}

		blocks[i] = block

		walents[i] = &consensus.WalEntry{
			Type:  getWalEntryType(&entry),
			Term:  entry.Term,
			Index: entry.Index,
			Data:  data,
		}
	}

	return walents, blocks
}
