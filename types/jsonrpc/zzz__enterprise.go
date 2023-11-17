package jsonrpc

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvEnterpriseTxStatus(msg *types.EnterpriseTxStatus) *InOutEnterpriseTxStatus {
	if msg == nil {
		return nil
	}

	ets := &InOutEnterpriseTxStatus{}
	ets.Status = msg.Status
	ets.Ret = msg.Ret
	ets.CCStatus = *ConvChangeClusterStatus(msg.CCStatus)
	return ets
}

type InOutEnterpriseTxStatus struct {
	Status   string                   `json:"status"`
	Ret      string                   `json:"ret"`
	CCStatus InOutChangeClusterStatus `json:"change_cluster,omitempty"`
}

func ConvChangeClusterStatus(msg *types.ChangeClusterStatus) *InOutChangeClusterStatus {
	if msg == nil {
		return nil
	}

	ccs := &InOutChangeClusterStatus{}
	ccs.State = msg.State
	ccs.Error = msg.Error
	ccs.Members = make([]*InOutMemberAttr, len(msg.Members))
	for i, m := range msg.Members {
		ccs.Members[i] = ConvMemberAttr(m)
	}
	return ccs
}

type InOutChangeClusterStatus struct {
	State   string             `json:"status"`
	Error   string             `json:"error"`
	Members []*InOutMemberAttr `json:"members"`
}

func ConvMemberAttr(msg *types.MemberAttr) *InOutMemberAttr {
	return &InOutMemberAttr{
		ID:      msg.ID,
		Name:    msg.Name,
		Address: msg.Address,
		PeerID:  base58.Encode(msg.PeerID),
	}
}

type InOutMemberAttr struct {
	ID      uint64 `json:"ID,omitempty"`
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
	PeerID  string `json:"peerID,omitempty"`
}
