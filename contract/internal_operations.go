package contract

import (
	"encoding/json"
	"log"
	"sync"
)

type InternalOperation struct {
	Id        int64    `json:"-"`
	Level     int      `json:"level"`
	Contract  string   `json:"contract"`
	Operation string   `json:"operation"`
	Amount    string   `json:"amount"`
	Args      []string `json:"args"`
	Results   []string `json:"results"`
}

var (
	internalOperations []InternalOperation
	txHash             []byte
	opsLock            sync.Mutex // To handle concurrent updates
	nextOpId           int64
)

func initInternalOps(newTxHash []byte) {
	opsLock.Lock()
	defer opsLock.Unlock()

	internalOperations = nil // Reinitialize the slice for new transaction
	txHash = newTxHash
	nextOpId = 1 // Reset nextOpId for each new transaction
}

func logOperation(ctx *vmContext, amount string, operation string, args ...string) int64 {
	opsLock.Lock()
	defer opsLock.Unlock()

	op := InternalOperation{
		Id:        nextOpId,
		Level:     ctx.CallDepth(),
		Contract:  ctx.Contract,
		Operation: operation,
		Amount:    amount,
		Args:      args,
	}
	internalOperations = append(internalOperations, op)
	nextOpId++
	return op.Id
}

func logOperationResult(operationId int64, results ...string) {
	opsLock.Lock()
	defer opsLock.Unlock()

	for i, op := range internalOperations {
		if op.Id == operationId {
			internalOperations[i].Results = append(internalOperations[i].Results, results...)
			return
		}
	}
	log.Printf("No operation found with ID %d", operationId)
}

func saveOperations() {
	opsLock.Lock()
	defer opsLock.Unlock()

	data, err := json.Marshal(internalOperations)
	if err != nil {
		log.Fatal("Failed to marshal operations:", err)
		return
	}

	// Simulated database set function (adjust as needed for actual database implementation)
	err = db.Set(txHash, data) // Assuming db.Set exists and works with byte keys and value
	if err != nil {
		log.Fatal("Failed to save operations to database:", err)
	}
}
