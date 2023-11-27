package jsonrpc

import (
	"math/big"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvReceipt(msg *types.Receipt) *InOutReceipt {
	r := &InOutReceipt{}
	r.ContractAddress = types.EncodeAddress(msg.ContractAddress)
	r.Status = msg.Status
	r.Ret = msg.Ret
	r.TxHash = base58.Encode(msg.TxHash)
	r.FeeUsed = new(big.Int).SetBytes(msg.FeeUsed).String()
	r.CumulativeFeeUsed = new(big.Int).SetBytes(msg.CumulativeFeeUsed).String()
	r.Bloom = msg.Bloom
	if msg.Events != nil {
		r.Events = make([]*InOutEvent, len(msg.Events))
		for i, e := range msg.Events {
			r.Events[i] = ConvEvent(e)
		}
	}
	r.BlockHash = base58.Encode(msg.BlockHash)
	r.BlockNo = msg.BlockNo
	r.TxIndex = msg.TxIndex
	r.From = types.EncodeAddress(msg.From)
	r.To = types.EncodeAddress(msg.To)
	r.FeeDelegation = msg.FeeDelegation
	r.GasUsed = msg.GasUsed
	return r
}

type InOutReceipt struct {
	ContractAddress   string        `json:"contractAddress"`
	Status            string        `json:"status"`
	Ret               string        `json:"ret"`
	TxHash            string        `json:"txHash"`
	FeeUsed           string        `json:"feeUsed"`
	CumulativeFeeUsed string        `json:"cumulativeFeeUsed,omitempty"`
	Bloom             []byte        `json:"bloom,omitempty"`
	Events            []*InOutEvent `json:"events,omitempty"`
	BlockHash         string        `json:"blockHash,omitempty"`
	BlockNo           uint64        `json:"blockNo,omitempty"`
	TxIndex           int32         `json:"txIndex,omitempty"`
	From              string        `json:"from,omitempty"`
	To                string        `json:"to,omitempty"`
	FeeDelegation     bool          `json:"feeDelegation,omitempty"`
	GasUsed           uint64        `json:"gasUsed,omitempty"`
}

func ConvEvent(msg *types.Event) *InOutEvent {
	return &InOutEvent{
		ContractAddress: types.EncodeAddress(msg.ContractAddress),
		EventName:       msg.EventName,
		JsonArgs:        msg.JsonArgs,
		EventIdx:        msg.EventIdx,
		TxHash:          base58.Encode(msg.TxHash),
		BlockHash:       base58.Encode(msg.BlockHash),
		BlockNo:         msg.BlockNo,
		TxIndex:         msg.TxIndex,
	}
}

type InOutEvent struct {
	ContractAddress string `json:"contractAddress"`
	EventName       string `json:"eventName"`
	JsonArgs        string `json:"jsonArgs"`
	EventIdx        int32  `json:"eventIdx"`
	TxHash          string `json:"txHash"`
	BlockHash       string `json:"blockHash"`
	BlockNo         uint64 `json:"blockNo"`
	TxIndex         int32  `json:"txIndex"`
}

func ConvAbi(msg *types.ABI) *InOutAbi {
	abi := &InOutAbi{}
	abi.Version = msg.Version
	abi.Language = msg.Language
	abi.Functions = make([]*InOutFunction, len(msg.Functions))
	for i, fn := range msg.Functions {
		abi.Functions[i] = ConvFunction(fn)
	}
	abi.StateVariables = make([]*InOutStateVar, len(msg.StateVariables))
	for i, sv := range msg.StateVariables {
		abi.StateVariables[i] = ConvStateVar(sv)
	}
	return abi
}

type InOutAbi struct {
	Version        string           `json:"version"`
	Language       string           `json:"language"`
	Functions      []*InOutFunction `json:"functions"`
	StateVariables []*InOutStateVar `json:"stateVariables"`
}

func ConvFunction(msg *types.Function) *InOutFunction {
	fn := &InOutFunction{}
	fn.Name = msg.Name
	fn.Arguments = make([]*InOutFunctionArgument, len(msg.Arguments))
	for i, arg := range msg.Arguments {
		fn.Arguments[i] = ConvFunctionArgument(arg)
	}
	fn.Payable = msg.Payable
	fn.View = msg.View
	fn.FeeDelegation = msg.FeeDelegation
	return fn
}

type InOutFunction struct {
	Name          string                   `json:"name"`
	Arguments     []*InOutFunctionArgument `json:"arguments"`
	Payable       bool                     `json:"payable"`
	View          bool                     `json:"view"`
	FeeDelegation bool                     `json:"feeDelegation"`
}

func ConvFunctionArgument(msg *types.FnArgument) *InOutFunctionArgument {
	return &InOutFunctionArgument{
		Name: msg.Name,
	}
}

type InOutFunctionArgument struct {
	Name string `json:"name"`
}

func ConvStateVar(msg *types.StateVar) *InOutStateVar {
	return &InOutStateVar{
		Name: msg.Name,
		Type: msg.Type,
		Len:  msg.Len,
	}
}

type InOutStateVar struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Len  int32  `json:"len"`
}

func ConvReceipts(msg *types.Receipts) *InOutReceipts {
	rs := &InOutReceipts{}
	rs.BlockNo = msg.BlockNo
	rs.Receipts = make([]*InOutReceipt, len(msg.Receipts))
	for i, receipt := range msg.Receipts {
		rs.Receipts[i] = ConvReceipt(receipt)
	}
	return rs
}

type InOutReceipts struct {
	Receipts       	[]*InOutReceipt		`json:"receipts"`
	BlockNo        	uint64				`json:"blockNo,omitempty"`
}

func ConvEvents(msg *types.EventList) *InOutEventList {
	rs := &InOutEventList{}
	rs.Events = make([]*InOutEvent, len(msg.Events))
	for i, event := range msg.Events {
		rs.Events[i] = ConvEvent(event)
	}
	return rs
}

type InOutEventList struct {
	Events []*InOutEvent `json:"events,omitempty"`
}