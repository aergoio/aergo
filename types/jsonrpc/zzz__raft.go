package jsonrpc

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/etcd/raft/raftpb"
)

func ConvConfChange(msg *raftpb.ConfChange) *InOutConfChange {
	return &InOutConfChange{
		ID:      msg.ID,
		Type:    msg.Type,
		NodeID:  msg.NodeID,
		Context: base58.Encode(msg.Context),
	}
}

type InOutConfChange struct {
	ID      uint64
	Type    raftpb.ConfChangeType
	NodeID  uint64
	Context string `json:"Context,omitempty"`
}
