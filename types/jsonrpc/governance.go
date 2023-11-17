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
	Staking InOutStaking
	Voting  []*InOutVoteInfo
}

func ConvStaking(msg *types.Staking) *InOutStaking {
	return &InOutStaking{
		Amount: new(big.Int).SetBytes(msg.Amount).String(),
		When:   msg.When,
	}
}

type InOutStaking struct {
	Amount string
	When   uint64
}

func ConvVoteInfo(msg *types.VoteInfo) *InOutVoteInfo {
	return &InOutVoteInfo{
		Id:         msg.Id,
		Candidates: msg.Candidates,
		Amount:     msg.Amount,
	}
}

type InOutVoteInfo struct {
	Id         string
	Candidates []string
	Amount     string
}
