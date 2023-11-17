package jsonrpc

import (
	"math/big"

	"github.com/aergoio/aergo/v2/types"
)

func ConvInOutAccountVoteInfo(msg *types.AccountVoteInfo) *InOutAccountVoteInfo {
	avi := &InOutAccountVoteInfo{}
	avi.Staking = *ConvStaking(msg.Staking)
	avi.Voting = make([]*InOutVoteInfo, len(msg.Voting))
	for i, v := range msg.Voting {
		avi.Voting[i] = ConvVoteInfo(v)
	}
	return avi
}

type InOutAccountVoteInfo struct {
	Staking InOutStaking     `json:"staking,omitempty"`
	Voting  []*InOutVoteInfo `json:"voting,omitempty"`
}

func ConvStaking(msg *types.Staking) *InOutStaking {
	return &InOutStaking{
		Amount: new(big.Int).SetBytes(msg.Amount).String(),
		When:   msg.When,
	}
}

type InOutStaking struct {
	Amount string `json:"amount,omitempty"`
	When   uint64 `json:"when,omitempty"`
}

func ConvVoteInfo(msg *types.VoteInfo) *InOutVoteInfo {
	return &InOutVoteInfo{
		Id:         msg.Id,
		Candidates: msg.Candidates,
		Amount:     msg.Amount,
	}
}

type InOutVoteInfo struct {
	Id         string   `json:"id,omitempty"`
	Candidates []string `json:"candidates,omitempty"`
	Amount     string   `json:"amount,omitempty"`
}
