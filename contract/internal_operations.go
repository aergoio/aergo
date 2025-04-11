package contract

import (
	"encoding/json"
	"sync"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
)

type InternalOperation struct {
	Id        int64    `json:"-"`
	Operation string   `json:"op"`
	Amount    string   `json:"amount,omitempty"`
	Args      []string `json:"args"`
	Result    string   `json:"result,omitempty"`
	Call      *InternalCall `json:"call,omitempty"`
	Reverted  bool     `json:"reverted,omitempty"`
}

type InternalCall struct {
	Contract  string   `json:"contract,omitempty"`
	Function  string   `json:"function,omitempty"`
	Args      []interface{} `json:"args,omitempty"`
	Amount    string   `json:"amount,omitempty"`
	Operations []InternalOperation `json:"operations,omitempty"`
}

type InternalOperations struct {
	TxHash    string   `json:"txhash"`
	Call      InternalCall `json:"call"`
}

var (
	opsLock      sync.Mutex
	nextOpId     int64
)

func doNotLog(ctx *vmContext) bool {
	if logInternalOperations == false {
		return true
	}
	return ctx.isQuery
}

func getCurrentCall(ctx *vmContext, callDepth int32) *InternalCall {
	var depth int32 = 1
	opCall := &ctx.internalOpsCall
	for {
		if opCall == nil {
			ctrLgr.Printf("no call found at depth %d", depth)
			break
		}
		if depth == callDepth {
			return opCall
		}
		if len(opCall.Operations) == 0 {
			ctrLgr.Printf("no operations found at depth %d", depth)
			break
		}
		opCall = opCall.Operations[len(opCall.Operations)-1].Call
		depth++
	}
	return nil
}

func logOperation(ctx *vmContext, amount string, operation string, args ...string) int64 {
	if doNotLog(ctx) {
		return 0
	}

	ctrLgr.Printf("logOperation: depth: %d, amount: %s, operation: %s, args: %v", ctx.callDepth, amount, operation, args)

	opsLock.Lock()
	defer opsLock.Unlock()

	nextOpId++
	if nextOpId > 1000000000000000000 {
		nextOpId = 1
	}

	op := InternalOperation{
		Id:        nextOpId,
		Operation: operation,
		Amount:    amount,
		Args:      args,
	}

	opCall := getCurrentCall(ctx, ctx.callDepth)
	if opCall == nil {
		ctrLgr.Printf("no call found")
		return 0
	}
	// add the operation to the list
	opCall.Operations = append(opCall.Operations, op)

	return op.Id
}

func logOperationResult(ctx *vmContext, operationId int64, result string) {
	if doNotLog(ctx) {
		return
	}

	ctrLgr.Printf("logOperationResult: depth: %d, operationId: %d, result: %s", ctx.callDepth, operationId, result)

	opsLock.Lock()
	defer opsLock.Unlock()

	// try with the last and the previous call depth
	for callDepth := ctx.callDepth; callDepth >= ctx.callDepth - 1; callDepth-- {
		opCall := getCurrentCall(ctx, callDepth)
		if opCall == nil {
			continue
		}
		for i := range opCall.Operations {
			if opCall.Operations[i].Id == operationId {
				opCall.Operations[i].Result = result
				return
			}
		}
	}

	ctrLgr.Printf("no operation found with ID %d to store result", operationId)
}

func markOperationsAsReverted(opCall *InternalCall, startOp int) {
	for i := startOp; i < len(opCall.Operations); i++ {
		opCall.Operations[i].Reverted = true
	}
}

func logInternalCall(ctx *vmContext, contract string, function string, args []interface{}, amount string) error {
	if doNotLog(ctx) {
		return nil
	}

	ctrLgr.Printf("logInternalCall: depth: %d, contract: %s, function: %s, args: %s, amount: %s", ctx.callDepth, contract, function, argsToJson(args), amount)

	opCall := getCurrentCall(ctx, ctx.callDepth-1)
	if opCall == nil {
		ctrLgr.Printf("no call found")
		return nil
	}

	// get the last operation
	op := &opCall.Operations[len(opCall.Operations)-1]

	// add this call to the last operation
	op.Call = &InternalCall{
		Contract: contract,
		Function: function,
		Args: args,
		Amount: amount,
	}

	return nil
}

func logFirstCall(ctx *vmContext, contract string, function string, args []interface{}, amount string) {
	ctrLgr.Printf("logFirstCall: depth: %d, contract: %s, function: %s, args: %s, amount: %s", ctx.callDepth, contract, function, argsToJson(args), amount)
	ctx.internalOpsCall.Contract = contract
	ctx.internalOpsCall.Function = function
	ctx.internalOpsCall.Args = args
	ctx.internalOpsCall.Amount = amount
}

func logCall(ctx *vmContext, contract string, function string, args []interface{}, amount string) {
	if amount == "0" {
		amount = ""
	}
	if ctx.internalOpsCall.Contract == "" {
		logFirstCall(ctx, contract, function, args, amount)
	} else {
		logInternalCall(ctx, contract, function, args, amount)
	}
}

func argsToJson(argsList []interface{}) (string) {
	if argsList == nil {
		return ""
	}
	args, err := json.Marshal(argsList)
	if err != nil {
		return ""
	}
	return string(args)
}

func getInternalOperations(ctx *vmContext) string {
	if doNotLog(ctx) {
		return ""
	}
	if ctx.internalOpsCall.Contract == "" {
		return ""
	}

	opsLock.Lock()
	defer opsLock.Unlock()

	internalOps := InternalOperations{
		TxHash: base58.Encode(ctx.txHash),
		Call: ctx.internalOpsCall,
	}

	data, err := json.Marshal(internalOps)
	if err != nil {
		ctrLgr.Fatal().Err(err).Msg("Failed to marshal operations")
		return ""
	}

	return string(data)
}
