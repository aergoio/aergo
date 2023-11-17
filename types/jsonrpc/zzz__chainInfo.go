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
	return &InOutChainId{
		Magic:     msg.Magic,
		Public:    msg.Public,
		Mainnet:   msg.Mainnet,
		Consensus: msg.Consensus,
		Version:   msg.Version,
	}
}

type InOutChainId struct {
	Magic     string
	Public    bool
	Mainnet   bool
	Consensus string
	Version   int32
}

func ConvBlockchainStatus(msg *types.BlockchainStatus) *InOutBlockchainStatus {
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
	bs := &InOutBlockchainStatus{}
	bs.Hash = hex.Encode(msg.BestBlockHash)
	bs.Height = msg.BestHeight
	bs.ChainIdHash = hex.Encode(msg.BestChainIdHash)
	return bs
}

type InOutBlockchainStatus struct {
	Hash          string
	Height        uint64
	ConsensusInfo *json.RawMessage `json:",omitempty"`
	ChainIdHash   string
	ChainStat     *json.RawMessage `json:",omitempty"`
	ChainInfo     *InOutChainInfo  `json:",omitempty"`
}

func ConvMetrics(msg *types.Metrics) *InOutMetrics {
	m := &InOutMetrics{}
	m.Peers = make([]*InOutPeerMetric, len(msg.Peers))
	for i, peer := range msg.Peers {
		m.Peers[i] = ConvPeerMetric(peer)
	}
	return m
}

type InOutMetrics struct {
	Peers []*InOutPeerMetric
}

func ConvPeerMetric(msg *types.PeerMetric) *InOutPeerMetric {
	return &InOutPeerMetric{
		PeerID: base58.Encode(msg.PeerID),
		SumIn:  msg.SumIn,
		AvrIn:  msg.AvrIn,
		SumOut: msg.SumOut,
		AvrOut: msg.AvrOut,
	}
}

type InOutPeerMetric struct {
	PeerID string
	SumIn  int64 `json:",omitempty"`
	AvrIn  int64 `json:",omitempty"`
	SumOut int64 `json:",omitempty"`
	AvrOut int64 `json:",omitempty"`
}
