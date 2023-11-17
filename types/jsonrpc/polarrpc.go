package jsonrpc

import "github.com/aergoio/aergo/v2/types"

func ConvBLConfEntries(msg *types.BLConfEntries) *InOutBLConfEntries {
	return &InOutBLConfEntries{
		Enabled: msg.Enabled,
		Entries: msg.Entries,
	}
}

type InOutBLConfEntries struct {
	Enabled bool     `json:"enabled,omitempty"`
	Entries []string `json:"entries,omitempty"`
}
