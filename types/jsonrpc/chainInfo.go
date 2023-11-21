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
		ci.Id = *ConvChainId(msg.Id)
	}
	ci.MaxTokens = new(big.Int).SetBytes(msg.Maxtokens).String()
	if consensus.IsDposName(msg.Id.Consensus) {
		ci.StakingMinimum = new(big.Int).SetBytes(msg.Stakingminimum).String()
		ci.Totalstaking = new(big.Int).SetBytes(msg.Totalstaking).String()
	}
	ci.GasPrice = new(big.Int).SetBytes(msg.Gasprice).String()
	ci.NamePrice = new(big.Int).SetBytes(msg.Nameprice).String()
	ci.TotalVotingPower = new(big.Int).SetBytes(msg.Totalvotingpower).String()
	ci.VotingReward = new(big.Int).SetBytes(msg.Votingreward).String()
	return ci
}

type InOutChainInfo struct {
	Id               InOutChainId `json:"id,omitempty"`
	BpNumber         uint32       `json:"bpNumber,omitempty"`
	MaxBlockSize     uint64       `json:"maxblocksize,omitempty"`
	MaxTokens        string       `json:"maxtokens,omitempty"`
	StakingMinimum   string       `json:"stakingminimum,omitempty"`
	Totalstaking     string       `json:"totalstaking,omitempty"`
	GasPrice         string       `json:"gasprice,omitempty"`
	NamePrice        string       `json:"nameprice,omitempty"`
	TotalVotingPower string       `json:"totalvotingpower,omitempty"`
	VotingReward     string       `json:"votingreward,omitempty"`
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
	Magic     string `json:"magic,omitempty"`
	Public    bool   `json:"public,omitempty"`
	Mainnet   bool   `json:"mainnet,omitempty"`
	Consensus string `json:"consensus,omitempty"`
	Version   int32  `json:"version,omitempty"`
}

func ConvBlockchainStatus(msg *types.BlockchainStatus) *InOutBlockchainStatus {
	bs := &InOutBlockchainStatus{}
	bs.BestBlockHash = base58.Encode(msg.BestBlockHash)
	bs.BestHeight = msg.BestHeight
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
	bs.BestBlockHash = hex.Encode(msg.BestBlockHash)
	bs.BestHeight = msg.BestHeight
	bs.ChainIdHash = hex.Encode(msg.BestChainIdHash)
	return bs
}

type InOutBlockchainStatus struct {
	BestBlockHash string           `json:"best_block_hash,omitempty"`
	BestHeight    uint64           `json:"best_height,omitempty"`
	ConsensusInfo *json.RawMessage `json:"consensus_info,omitempty"`
	ChainIdHash   string           `json:"best_chain_id_hash,omitempty"`
	ChainStat     *json.RawMessage `json:"chain_stat,omitempty"`
	ChainInfo     *InOutChainInfo  `json:"chain_info,omitempty"`
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
	Peers []*InOutPeerMetric `json:"peers,omitempty"`
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
	PeerID string `json:"peerID,omitempty"`
	SumIn  int64  `json:"sumIn,omitempty"`
	AvrIn  int64  `json:"avrIn,omitempty"`
	SumOut int64  `json:"sumOut,omitempty"`
	AvrOut int64  `json:"avrOut,omitempty"`
}

func ConvChainStat(msg *types.ChainStats) *InOutChainStats {
	return &InOutChainStats{
		Report: msg.GetReport(),
	}
}

type InOutChainStats struct {
	Report string `json:"report,omitempty"`
}

func ConvConsensusInfo(msg *types.ConsensusInfo) *InOutConsensusInfo {
	
	ci := &InOutConsensusInfo{}
	ci.Type = msg.GetType()
	ci.Info = msg.GetInfo()
	ci.Bps = make([]string, len(msg.Bps))
	for i, bps := range msg.Bps {
		ci.Bps[i] = bps
	}
	return ci
}

type InOutConsensusInfo struct {
	Type string   `json:"type,omitempty"`
	Info string   `json:"info,omitempty"`
	Bps  []string `json:"bps,omitempty"`
}
