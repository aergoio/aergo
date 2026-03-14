/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/types/dbkey"
)

// DbTx represents Set and Delete interface to store data
type DbTx interface {
	Set(key, value []byte)
	Delete(key []byte)
}

// nodeChange tracks a node's batch data and reference count changes within a block.
// refcount > 0: node was created/referenced, needs Set() calls
// refcount < 0: node was dereferenced, needs Delete() calls
// refcount == 0: balanced creates and deletes, no DB operation needed
type nodeChange struct {
	batch    [][]byte
	refcount int
}

type CacheDB struct {
	// liveCache contains the first levels of the trie (nodes that have 2 non default children)
	liveCache map[Hash][][]byte
	// liveMux is a lock for liveCache
	liveMux sync.RWMutex
	// nodeChanges tracks nodes that were created or deleted in the current block
	// with an in-memory reference count to properly handle the DB's reference counter
	nodeChanges map[Hash]*nodeChange
	// nodeChangesMux is a lock for nodeChanges
	nodeChangesMux sync.RWMutex
	// nodesToRevert will be deleted from db
	nodesToRevert [][]byte
	// revertMux is a lock for nodesToRevert
	revertMux sync.RWMutex
	// lock for CacheDB
	lock sync.RWMutex
	// store is the interface to disk db
	Store db.DB
}

// updatedCount returns the number of nodes with positive refcount (nodes to be saved).
// This is used for testing to maintain compatibility with old tests.
func (c *CacheDB) updatedCount() int {
	c.nodeChangesMux.RLock()
	defer c.nodeChangesMux.RUnlock()
	count := 0
	for _, change := range c.nodeChanges {
		if change.refcount > 0 {
			count++
		}
	}
	return count
}

// getUpdatedNodes returns a map of nodes with positive refcount.
// This is used for testing to maintain compatibility with old tests.
func (c *CacheDB) getUpdatedNodes() map[Hash]bool {
	c.nodeChangesMux.RLock()
	defer c.nodeChangesMux.RUnlock()
	result := make(map[Hash]bool)
	for hash, change := range c.nodeChanges {
		if change.refcount > 0 {
			result[hash] = true
		}
	}
	return result
}

// commit processes nodeChanges and applies them to the database transaction.
// For light nodes (deldeldb):
// - refcount > 0: call Set() refcount times (to increment DB refcount)
// - refcount < 0: call Delete() |refcount| times (to queue deletions)
// - refcount == 0: do nothing (balanced creates and deletes)
// For non-light nodes:
// - refcount > 0: call Set() once (original behavior)
// - refcount <= 0: do nothing (node was deleted or balanced)
func (c *CacheDB) commit(txn *DbTx) {
	c.nodeChangesMux.Lock()
	defer c.nodeChangesMux.Unlock()

	isLightNode := c.Store.Type() == "deldeldb"

	for key, change := range c.nodeChanges {
		if change.refcount > 0 {
			serialized := c.serializeBatch(change.batch)
			if isLightNode {
				// Light nodes: call Set for each reference to increment DB refcount
				for i := 0; i < change.refcount; i++ {
					(*txn).Set(dbkey.Trie(key[:]), serialized)
				}
			} else {
				// Non-light nodes: call Set once (original behavior)
				(*txn).Set(dbkey.Trie(key[:]), serialized)
			}
		} else if change.refcount < 0 && isLightNode {
			// Node was dereferenced - queue deletions (only for light nodes)
			for i := 0; i < -change.refcount; i++ {
				(*txn).Delete(dbkey.Trie(key[:]))
			}
		}
		// refcount == 0: balanced, no action needed
	}
}

// serializeBatch serialises the 2D [][]byte into a []byte for db
func (c *CacheDB) serializeBatch(batch [][]byte) []byte {
	serialized := make([]byte, 4) //, 30*33)
	if batch[0][0] == 1 {
		// the batch node is a shortcut
		bitSet(serialized, 31)
	}
	for i := 1; i < 31; i++ {
		if len(batch[i]) != 0 {
			bitSet(serialized, i-1)
			serialized = append(serialized, batch[i]...)
		}
	}
	return serialized
}
