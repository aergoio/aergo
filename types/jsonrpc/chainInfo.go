package jsonrpc

import (
	"encoding/json"
	"math/big"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/types"
)

func ConvChainInfo(msg *types.ChainInfo) *InOutChainInfo {
	if msg == nil {
		return nil
	}

	ci := &InOutChainInfo{}
	ci.Id = ConvChainId(msg.Id)
	ci.BpNumber = msg.BpNumber
	ci.MaxBlockSize = msg.Maxblocksize
	ci.MaxTokens = new(big.Int).SetBytes(msg.Maxtokens).String()
	if ci.Id != nil && consensus.IsDposName(ci.Id.Consensus) {
		ci.StakingMinimum = new(big.Int).SetBytes(msg.Stakingminimum).String()
		ci.Totalstaking = new(big.Int).SetBytes(msg.Totalstaking).String()
	}
	ci.GasPrice = new(big.Int).SetBytes(msg.Gasprice).String()
	ci.NamePrice = new(big.Int).SetBytes(msg.Nameprice).String()
	ci.TotalVotingPower = new(big.Int).SetBytes(msg.Totalvotingpower).String()
	ci.VotingReward = new(big.Int).SetBytes(msg.Votingreward).String()
	ci.Hardfork = msg.Hardfork
	return ci
}

type InOutChainInfo struct {
	Id               *InOutChainId     `json:"id,omitempty"`
	BpNumber         uint32            `json:"bpNumber,omitempty"`
	MaxBlockSize     uint64            `json:"maxblocksize,omitempty"`
	MaxTokens        string            `json:"maxtokens,omitempty"`
	StakingMinimum   string            `json:"stakingminimum,omitempty"`
	Totalstaking     string            `json:"totalstaking,omitempty"`
	GasPrice         string            `json:"gasprice,omitempty"`
	NamePrice        string            `json:"nameprice,omitempty"`
	TotalVotingPower string            `json:"totalvotingpower,omitempty"`
	VotingReward     string            `json:"votingreward,omitempty"`
	Hardfork         map[string]uint64 `json:"hardfork,omitempty"`
}

func ConvChainId(msg *types.ChainId) *InOutChainId {
	if msg == nil {
		return nil
	}
	return &InOutChainId{
		Magic:     msg.Magic,
		Public:    msg.Public,
		Mainnet:   msg.Mainnet,
		Consensus: msg.Consensus,
		Version:   msg.Version,
	}
}

type InOutChainId struct {
	Magic     string `json:"magic,omitempty"`
	Public    bool   `json:"public,omitempty"`
	Mainnet   bool   `json:"mainnet,omitempty"`
	Consensus string `json:"consensus,omitempty"`
	Version   int32  `json:"version,omitempty"`
}

func ConvBlockchainStatus(msg *types.BlockchainStatus) *InOutBlockchainStatus {
	if msg == nil {
		return nil
	}

	bs := &InOutBlockchainStatus{}
	bs.Hash = base58.Encode(msg.BestBlockHash)
	bs.Height = msg.BestHeight
	bs.ChainIdHash = base58.Encode(msg.BestChainIdHash)

	toJRM := func(s string) *json.RawMessage {
		if len(s) > 0 {
			m := json.RawMessage(s)
			return &m
		}
		return nil
	}
	bs.ConsensusInfo = toJRM(msg.ConsensusInfo)
	if msg.ChainInfo != nil {
		bs.ChainInfo = ConvChainInfo(msg.ChainInfo)
	}
	return bs
}

func ConvHexBlockchainStatus(msg *types.BlockchainStatus) *InOutBlockchainStatus {
	if msg == nil {
		return nil
	}

	bs := &InOutBlockchainStatus{}
	bs.Hash = hex.Encode(msg.BestBlockHash)
	bs.Height = msg.BestHeight
	bs.ChainIdHash = hex.Encode(msg.BestChainIdHash)
	return bs
}

type InOutBlockchainStatus struct {
	Hash          string           `json:"hash"`
	Height        uint64           `json:"height"`
	ConsensusInfo *json.RawMessage `json:"consensusInfo,omitempty"`
	ChainIdHash   string           `json:"chainIdHash"`
	ChainStat     *json.RawMessage `json:"chainStat,omitempty"`
	ChainInfo     *InOutChainInfo  `json:"chainInfo,omitempty"`
}

func ConvChainStat(msg *types.ChainStats) *InOutChainStats {
	if msg == nil {
		return nil
	}

	cs := &InOutChainStats{}
	_ = json.Unmarshal([]byte(msg.GetReport()), &cs.Report)
	return cs
}

type InOutChainStats struct {
	Report interface{} `json:"report,omitempty"`
}

func ConvConsensusInfo(msg *types.ConsensusInfo) *InOutConsensusInfo {
	if msg == nil {
		return nil
	}

	ci := &InOutConsensusInfo{}
	ci.Type = msg.GetType()
	_ = json.Unmarshal([]byte(msg.GetInfo()), &ci.Info)

	ci.Bps = make([]interface{}, len(msg.Bps))
	for i, bps := range msg.Bps {
		_ = json.Unmarshal([]byte(bps), &ci.Bps[i])
	}
	return ci
}

type InOutConsensusInfo struct {
	Type string        `json:"type,omitempty"`
	Info interface{}   `json:"info,omitempty"`
	Bps  []interface{} `json:"bps,omitempty"`
}
