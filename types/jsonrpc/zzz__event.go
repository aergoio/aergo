package jsonrpc

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvEvent(msg *types.Event) *InOutEvent {
	e := &InOutEvent{}
	e.ContractAddress = base58.Encode(msg.ContractAddress)
	e.EventName = msg.EventName
	e.JsonArgs = msg.JsonArgs
	e.EventIdx = msg.EventIdx
	e.TxHash = base58.Encode(msg.TxHash)
	e.BlockHash = base58.Encode(msg.BlockHash)
	e.BlockNo = msg.BlockNo
	e.TxIndex = msg.TxIndex
	return e
}

type InOutEvent struct {
	ContractAddress string
	EventName       string
	JsonArgs        string
	EventIdx        int32
	TxHash          string
	BlockHash       string
	BlockNo         uint64
	TxIndex         int32
}
