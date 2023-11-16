package jsonrpc

import (
	"encoding/json"

	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/types"
)

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
