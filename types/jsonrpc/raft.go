package jsonrpc

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/etcd/raft/raftpb"
)

func ConvConfChange(msg *raftpb.ConfChange) *InOutConfChange {
	if msg == nil {
		return nil
	}
	return &InOutConfChange{
		ID:      msg.ID,
		Type:    msg.Type,
		NodeID:  msg.NodeID,
		Context: base58.Encode(msg.Context),
	}
}

type InOutConfChange struct {
	ID      uint64                `json:"ID,omitempty"`
	Type    raftpb.ConfChangeType `json:"type,omitempty"`
	NodeID  uint64                `json:"nodeID,omitempty"`
	Context string                `json:"context,omitempty"`
}
