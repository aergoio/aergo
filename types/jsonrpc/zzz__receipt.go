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
	ContractAddress   string
	Status            string
	Ret               string
	TxHash            string
	FeeUsed           string
	CumulativeFeeUsed string        `json:",omitempty"`
	Bloom             []byte        `json:",omitempty"`
	Events            []*InOutEvent `json:",omitempty"`
	BlockHash         string
	BlockNo           uint64
	TxIndex           int32
	From              string
	To                string
	FeeDelegation     bool
	GasUsed           uint64
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
	ContractAddress string
	EventName       string
	JsonArgs        string
	EventIdx        int32
	TxHash          string
	BlockHash       string
	BlockNo         uint64
	TxIndex         int32
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
	Version        string
	Language       string
	Functions      []*InOutFunction
	StateVariables []*InOutStateVar
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
	Name          string
	Arguments     []*InOutFunctionArgument
	Payable       bool
	View          bool
	FeeDelegation bool
}

func ConvFunctionArgument(msg *types.FnArgument) *InOutFunctionArgument {
	return &InOutFunctionArgument{
		Name: msg.Name,
	}
}

type InOutFunctionArgument struct {
	Name string
}

func ConvStateVar(msg *types.StateVar) *InOutStateVar {
	return &InOutStateVar{
		Name: msg.Name,
		Type: msg.Type,
		Len:  msg.Len,
	}
}

type InOutStateVar struct {
	Name string
	Type string
	Len  int32
}
