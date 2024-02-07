package jsonrpc

import (
	"github.com/aergoio/aergo/v2/types"
)

func ConvAbi(msg *types.ABI) *InOutAbi {
	if msg == nil {
		return nil
	}
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
	if msg == nil {
		return nil
	}
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
	if msg == nil {
		return nil
	}
	return &InOutFunctionArgument{
		Name: msg.Name,
	}
}

type InOutFunctionArgument struct {
	Name string `json:"name"`
}

func ConvStateVar(msg *types.StateVar) *InOutStateVar {
	if msg == nil {
		return nil
	}
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
	if msg == nil {
		return nil
	}

	rs := &InOutReceipts{}
	rs.BlockNo = msg.GetBlockNo()
	rs.Receipts = make([]*types.Receipt, len(msg.Get()))
	for i, receipt := range msg.Get() {
		rs.Receipts[i] = receipt
	}
	return rs
}

type InOutReceipts struct {
	Receipts []*types.Receipt `json:"receipts"`
	BlockNo  uint64           `json:"blockNo,omitempty"`
}

func ConvReceiptsPaged(msg *types.ReceiptsPaged) *InOutReceiptsPaged {
	if msg == nil {
		return nil
	}

	rp := &InOutReceiptsPaged{}
	rp.Total = msg.GetTotal()
	rp.Offset = msg.GetOffset()
	rp.Size = msg.GetSize()
	rp.BlockNo = msg.GetBlockNo()
	rp.Receipts = make([]*types.Receipt, len(msg.Get()))
	for i, receipt := range msg.Get() {
		rp.Receipts[i] = receipt
	}

	return rp
}

type InOutReceiptsPaged struct {
	Total    uint32           `json:"total,omitempty"`
	Offset   uint32           `json:"offset,omitempty"`
	Size     uint32           `json:"size,omitempty"`
	Receipts []*types.Receipt `json:"receipts"`
	BlockNo  uint64           `json:"blockNo,omitempty"`
}

func ConvEvents(msg *types.EventList) *InOutEventList {
	if msg == nil {
		return nil
	}

	rs := &InOutEventList{}
	rs.Events = make([]*types.Event, len(msg.Events))
	for i, event := range msg.Events {
		rs.Events[i] = event
	}
	return rs
}

type InOutEventList struct {
	Events []*types.Event `json:"events,omitempty"`
}
