package jsonrpc

import (
	"encoding/json"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/types"
)

type InOutBlockchainStatus struct {
	Hash          string
	Height        uint64
	ConsensusInfo *json.RawMessage `json:",omitempty"`
	ChainIdHash   string
	ChainStat     *json.RawMessage `json:",omitempty"`
	ChainInfo     *InOutChainInfo  `json:",omitempty"`
}

func (bs *InOutBlockchainStatus) FromProtoHex(msg *types.BlockchainStatus) {
	bs.Hash = hex.Encode(msg.BestBlockHash)
	bs.Height = msg.BestHeight
	bs.ChainIdHash = hex.Encode(msg.BestChainIdHash)
}

func (bs *InOutBlockchainStatus) FromProtoBase58(msg *types.BlockchainStatus) {
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
		bs.ChainInfo.FromProto(msg.ChainInfo)
	}
}
