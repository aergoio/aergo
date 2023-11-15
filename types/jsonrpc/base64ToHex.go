package jsonrpc

import (
	"encoding/json"

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

func ConvHexBlockchainStatus(in *types.BlockchainStatus) string {
	out := &InOutBlockchainStatus{}
	out.Hash = hex.Encode(in.BestBlockHash)
	out.Height = in.BestHeight
	out.ChainIdHash = hex.Encode(in.BestChainIdHash)
	jsonout, err := json.Marshal(out)
	if err != nil {
		return ""
	}
	return string(jsonout)
}
