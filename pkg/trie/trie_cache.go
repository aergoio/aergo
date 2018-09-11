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
	liveCache map[Hash][][]byte
	// liveMux is a lock for liveCache
	liveMux sync.RWMutex
	// updatedNodes that have will be flushed to disk
	updatedNodes map[Hash][][]byte
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
	for key, batch := range db.updatedNodes {
		//node := key
		var node []byte
		txn.Set(append(node, key[:]...), db.serializeBatch(batch))
		//txn.Set(node[:], db.serializeBatch(batch))
	}
	txn.Commit()
}

// serializeBatch serialises the 2D [][]byte into a []byte for db
func (db *CacheDB) serializeBatch(batch [][]byte) []byte {
	serialized := make([]byte, 4) //, 30*33)
	if batch[0][0] == 1 {
		// the batch node is a shortcut
		bitSet(serialized, 31)
	}
	for i := 1; i < 31; i++ {
		if len(batch[i]) != 0 {
			bitSet(serialized, uint64(i-1))
			serialized = append(serialized, batch[i]...)
		}
	}
	return serialized
}
