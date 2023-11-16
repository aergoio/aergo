package jsonrpc

import (
	"encoding/json"
	"math/big"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/types"
)

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

func (ci *InOutChainInfo) FromProto(msg *types.ChainInfo) {
	if msg.Id != nil {
		ci.Chainid.FromProto(msg.Id)
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
}

func (ci *InOutChainInfo) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(ci, "", " ")
}

type InOutChainId struct {
	Magic     string
	Public    bool
	Mainnet   bool
	Consensus string
	Version   int32
}

func (ci *InOutChainId) FromProto(msg *types.ChainId) {
	ci.Magic = msg.Magic
	ci.Public = msg.Public
	ci.Mainnet = msg.Mainnet
	ci.Consensus = msg.Consensus
	ci.Version = msg.Version
}
