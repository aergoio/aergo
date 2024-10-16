package verificationstore

import (
	"sync"

	"github.com/aergoio/aergo/v2/types"
)

var (
	store = &verificationStore{
		results: make(map[types.TxID]string),
	}
)

type verificationStore struct {
	sync.RWMutex
	results map[types.TxID]string
}

func StoreResult(txID types.TxID, result string) {
	store.Lock()
	defer store.Unlock()
	store.results[txID] = result
}

func GetResult(txID types.TxID) (string, bool) {
	store.RLock()
	defer store.RUnlock()
	result, exists := store.results[txID]
	return result, exists
}

func DeleteResult(txID types.TxID) {
	store.Lock()
	defer store.Unlock()
	delete(store.results, txID)
}