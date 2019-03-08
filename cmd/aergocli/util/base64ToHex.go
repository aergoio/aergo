package util

import (
	"encoding/hex"
	"encoding/json"

	"github.com/aergoio/aergo/types"
)

type InOutBlockchainStatus struct {
	Hash          string
	Height        uint64
	ConsensusInfo *json.RawMessage `json:",omitempty"`
}

func ConvHexBlockchainStatus(in *types.BlockchainStatus) string {
	out := &InOutBlockchainStatus{}
	out.Hash = hex.EncodeToString(in.BestBlockHash)
	out.Height = in.BestHeight
	jsonout, err := json.Marshal(out)
	if err != nil {
		return ""
	}
	return string(jsonout)
}
