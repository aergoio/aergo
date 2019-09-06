package util

import (
	"encoding/json"
	"math/big"

	"github.com/aergoio/aergo/consensus"

	"github.com/aergoio/aergo/types"
)

type InOutChainId struct {
	Magic     string
	Public    bool
	Mainnet   bool
	Consensus string
}

type InOutChainInfo struct {
	Chainid        InOutChainId
	BpNumber       uint32
	MaxBlockSize   uint64
	MaxTokens      string
	StakingMinimum string `json:",omitempty"`
	StakingTotal   string `json:",omitempty"`
	GasPrice       string `json:",omitempty"`
	NamePrice      string `json:",omitempty"`
}

func ConvChainInfoMsg(msg *types.ChainInfo) string {
	jsonout, err := json.MarshalIndent(convChainInfo(msg), "", " ")
	if err != nil {
		return ""
	}
	return string(jsonout)
}

func convChainInfo(msg *types.ChainInfo) *InOutChainInfo {
	out := &InOutChainInfo{}
	out.Chainid.Magic = msg.Id.Magic
	out.Chainid.Public = msg.Id.Public
	out.Chainid.Mainnet = msg.Id.Mainnet
	out.Chainid.Consensus = msg.Id.Consensus
	out.BpNumber = msg.BpNumber
	out.MaxBlockSize = msg.Maxblocksize
	out.MaxTokens = new(big.Int).SetBytes(msg.Maxtokens).String()

	if consensus.IsDposName(msg.Id.Consensus) {
		out.StakingMinimum = new(big.Int).SetBytes(msg.Stakingminimum).String()
		out.StakingTotal = new(big.Int).SetBytes(msg.Totalstaking).String()
	}

	out.GasPrice = new(big.Int).SetBytes(msg.Gasprice).String()
	out.NamePrice = new(big.Int).SetBytes(msg.Nameprice).String()
	return out
}
