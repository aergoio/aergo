/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"sync"

	"github.com/aergoio/aergo-lib/db"
)

type CacheDB struct {
	// liveCache contains the first levels of the trie (nodes that have 2 non default children)
	liveCache map[Hash][]byte
	// liveMux is a lock for liveCache
	liveMux sync.RWMutex
	// updatedNodes that have will be flushed to disk
	updatedNodes map[Hash][]byte
	// updatedMux is a lock for updatedNodes
	updatedMux sync.RWMutex
	// lock for CacheDB
	lock sync.RWMutex
	// store is the interface to disk db
	store db.DB
}

// commit stores the updated nodes to disk.
func (db *CacheDB) commit() {
	db.updatedMux.Lock()
	defer db.updatedMux.Unlock()
	txn := db.store.NewTx(true)
	// NOTE The tx interface doesnt handle ErrTxnTooBig
	for key, value := range db.updatedNodes {
		// txn.Set(key[:], value) doesn't work with a transaction but does with db.store.Set(key[:], value)
		var node [32]byte
		copy(node[:], key[:])
		txn.Set(node[:], value)
	}
	txn.Commit()
}
