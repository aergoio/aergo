package contract

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
)

type InternalOperation struct {
	Id        int64    `json:"-"`
	Operation string   `json:"op"`
	Amount    string   `json:"amount"`
	Args      []string `json:"args"`
	Result    string   `json:"result"`
	Call      InternalCall `json:"call"`
}

type InternalCall struct {
	Contract  string   `json:"contract"`
	Function  string   `json:"function"`
	Args      string   `json:"args"`
	Amount    string   `json:"amount"`
	Operations []InternalOperation `json:"operations"`
}

type InternalOperations struct {
	TxHash    string   `json:"txhash"`
	Contract  string   `json:"contract"`
	Operations []InternalOperation `json:"operations"`
}

var (
	opsLock      sync.Mutex
	nextOpId     int64
)

func doNotLog(ctx *vmContext) bool {
	return ctx.isQuery
}

func getCurrentCall(ctx *vmContext) *InternalCall {
	var depth int32 = 1
	opCall := &ctx.internalOpsCall
	for {
		if opCall == nil {
			break
		}
		if depth == ctx.callDepth {
			return opCall
		}
		if len(opCall.Operations) == 0 {
			break
		}
		opCall = &opCall.Operations[len(opCall.Operations)-1].Call
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

	opCall := getCurrentCall(ctx)
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

	opCall := getCurrentCall(ctx)
	if opCall == nil {
		log.Printf("no call found")
		return
	}

	for i := range opCall.Operations {
		if opCall.Operations[i].Id == operationId {
			opCall.Operations[i].Result = result
			return
		}
	}

	log.Printf("No operation found with ID %d", operationId)
}

func logInternalCall(ctx *vmContext, contract string, function string, args string) error {
	if doNotLog(ctx) {
		return nil
	}

	opCall := getCurrentCall(ctx)
	if opCall == nil {
		log.Printf("no call found")
		return nil
	}

	// get the last operation
	op := &opCall.Operations[len(opCall.Operations)-1]

	// add this call to the last operation
	op.Call = InternalCall{
		Contract: contract,
		Function: function,
		Args: args,
	}

	return nil
}

func getInternalOperations(ctx *vmContext) string {
	if doNotLog(ctx) {
		return ""
	}
	if len(ctx.internalOpsCall.Operations) == 0 {
		return ""
	}

	opsLock.Lock()
	defer opsLock.Unlock()

	internalOps := InternalOperations{
		TxHash: base58.Encode(ctx.txHash),
		Contract: types.EncodeAddress(ctx.curContract.contractId),
		Operations: ctx.internalOpsCall.Operations,
	}

	data, err := json.Marshal(internalOps)
	if err != nil {
		log.Fatal("Failed to marshal operations:", err)
		return ""
	}

	return string(data)
}
