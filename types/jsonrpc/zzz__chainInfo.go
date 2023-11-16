package jsonrpc

import (
	"math/big"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/types"
)

func ConvChainInfo(msg *types.ChainInfo) *InOutChainInfo {
	ci := &InOutChainInfo{}
	if msg.Id != nil {
		ci.Chainid = *ConvChainId(msg.Id)
	}
	ci.MaxTokens = new(big.Int).SetBytes(msg.Maxtokens).String()
	if consensus.IsDposName(msg.Id.Consensus) {
		ci.StakingMinimum = new(big.Int).SetBytes(msg.Stakingminimum).String()
		ci.StakingTotal = new(big.Int).SetBytes(msg.Totalstaking).String()
	}
	ci.GasPrice = new(big.Int).SetBytes(msg.Gasprice).String()
	ci.NamePrice = new(big.Int).SetBytes(msg.Nameprice).String()
	ci.TotalVotingPower = new(big.Int).SetBytes(msg.Totalvotingpower).String()
	ci.VotingReward = new(big.Int).SetBytes(msg.Votingreward).String()
	return ci
}

type InOutChainInfo struct {
	Chainid          InOutChainId
	BpNumber         uint32
	MaxBlockSize     uint64
	MaxTokens        string
	StakingMinimum   string `json:",omitempty"`
	StakingTotal     string `json:",omitempty"`
	GasPrice         string `json:",omitempty"`
	NamePrice        string `json:",omitempty"`
	TotalVotingPower string `json:",omitempty"`
	VotingReward     string `json:",omitempty"`
}

func ConvChainId(msg *types.ChainId) *InOutChainId {
	ci := &InOutChainId{}
	ci.Magic = msg.Magic
	ci.Public = msg.Public
	ci.Mainnet = msg.Mainnet
	ci.Consensus = msg.Consensus
	ci.Version = msg.Version
	return ci
}

type InOutChainId struct {
	Magic     string
	Public    bool
	Mainnet   bool
	Consensus string
	Version   int32
}
