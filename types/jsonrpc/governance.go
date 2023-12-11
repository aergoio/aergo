package jsonrpc

import (
	"math/big"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvInOutAccountVoteInfo(msg *types.AccountVoteInfo) *InOutAccountVoteInfo {
	if msg == nil {
		return nil
	}

	avi := &InOutAccountVoteInfo{}	
	staking := ConvStaking(msg.Staking)
	if staking != nil {		
		avi.Staking = *staking		
	}	
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
	if msg == nil {
		return nil
	}

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

func ConvVote(msg *types.Vote) *InOutVote {
	return &InOutVote{
		Candidate: base58.Encode(msg.Candidate),
		Amount:    msg.GetAmountBigInt().String(),
	}
}

type InOutVote struct {
	Candidate string `json:"candidate,omitempty"`
	Amount    string `json:"amount,omitempty"`
}

func ConvVotes(msg *types.VoteList) *InOutVotes {
	vs := &InOutVotes{}
	vs.Id = msg.GetId()

	vs.Votes = make([]*InOutVote, len(msg.Votes))
	for i, vote := range msg.Votes {
		vs.Votes[i] = ConvVote(vote)
	}

	return vs
}

type InOutVotes struct {
	Votes []*InOutVote `json:"votes,omitempty"`
	Id    string       `json:"id,omitempty"`
}
