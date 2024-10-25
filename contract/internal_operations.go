package contract

import (
	"encoding/json"
	"log"
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
}

type InternalCall struct {
	Contract  string   `json:"contract,omitempty"`
	Function  string   `json:"function,omitempty"`
	Args      string   `json:"args,omitempty"`
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
	return ctx.isQuery
}

func getCurrentCall(ctx *vmContext, callDepth int32) *InternalCall {
	var depth int32 = 1
	opCall := &ctx.internalOpsCall
	for {
		if opCall == nil {
			log.Printf("no call found at depth %d", depth)
			break
		}
		if depth == callDepth {
			return opCall
		}
		if len(opCall.Operations) == 0 {
			log.Printf("no operations found at depth %d", depth)
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
		log.Printf("no call found")
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

	log.Printf("no operation found with ID %d to store result", operationId)
}

func logInternalCall(ctx *vmContext, contract string, function string, args string) error {
	if doNotLog(ctx) {
		return nil
	}

	opCall := getCurrentCall(ctx, ctx.callDepth-1)
	if opCall == nil {
		log.Printf("no call found")
		return nil
	}

	// get the last operation
	op := &opCall.Operations[len(opCall.Operations)-1]

	// add this call to the last operation
	op.Call = &InternalCall{
		Contract: contract,
		Function: function,
		Args: args,
	}

	return nil
}

func logFirstCall(ctx *vmContext, contract string, function string, args string) {
	ctx.internalOpsCall.Contract = contract
	ctx.internalOpsCall.Function = function
	ctx.internalOpsCall.Args = args
}

func logCall(ctx *vmContext, contract string, function string, args string) {
	if ctx.internalOpsCall.Contract == "" {
		logFirstCall(ctx, contract, function, args)
	} else {
		logInternalCall(ctx, contract, function, args)
	}
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
		log.Fatal("Failed to marshal operations:", err)
		return ""
	}

	return string(data)
}
