/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"sync"

	"github.com/aergoio/aergo-lib/db"
)

// DbTx represents Set and Delete interface to store data
type DbTx interface {
	Set(key, value []byte)
	Delete(key []byte)
}

type CacheDB struct {
	// liveCache contains the first levels of the trie (nodes that have 2 non default children)
	liveCache map[Hash][][]byte
	// liveMux is a lock for liveCache
	liveMux sync.RWMutex
	// updatedNodes that have will be flushed to disk
	updatedNodes map[Hash][][]byte
	// nodes that should not be deleted from the database
	fixedNodes map[Hash]bool
	// nodes marked to be deleted from the database
	deletedNodes map[Hash]bool
	// updatedMux is a lock for updatedNodes
	updatedMux sync.RWMutex
	// nodesToRevert will be deleted from db
	nodesToRevert [][]byte
	// revertMux is a lock for updatedNodes
	revertMux sync.RWMutex
	// lock for CacheDB
	lock sync.RWMutex
	// store is the interface to disk db
	Store db.DB
}

// load the list/map of fixed nodes from the database
// they are stored on a single key named 'fixedNodes' as a list of hashes
func (c *CacheDB) loadFixedNodes(txn *DbTx) {
	c.fixedNodes = make(map[Hash]bool)
	fixedNodes := c.Store.Get([]byte("fixedNodes"))
	if fixedNodes != nil {
		for i := 0; i < len(fixedNodes); i += HashLength {
			var node Hash
			copy(node[:], fixedNodes[i:i+HashLength])
			c.fixedNodes[node] = true
		}
	}
}

// commit adds updatedNodes to the given database transaction.
func (c *CacheDB) commit(txn *DbTx) {
	c.updatedMux.Lock()
	defer c.updatedMux.Unlock()
	hasNewFixedNode := false
	// load the list of fixed nodes if not already loaded
	if c.fixedNodes == nil {
		c.loadFixedNodes(txn)
	}
	// add all updated nodes
	for key, batch := range c.updatedNodes {
		var node []byte
		node = append(node, key[:]...)
		// if the key is already on the database, mark it as fixed (non-deletable)
		if c.Store.Get(node) != nil {
			c.fixedNodes[key] = true
			hasNewFixedNode = true
		}
		// add the node to the database
		(*txn).Set(node, c.serializeBatch(batch))
		// if the node is on the list of deleted nodes, remove it
		//delete(c.deletedNodes, key)
		// this may be faster
		if c.deletedNodes[key] {
			c.deletedNodes[key] = false
		}
	}
	// delete nodes marked for deletion, unless they are fixed
	for key, deleted := range c.deletedNodes {
		if deleted && !c.fixedNodes[key] {
			var node []byte
			node = append(node, key[:]...)
			(*txn).Delete(node)
		}
	}
	// add the list of fixed nodes to the database
	if hasNewFixedNode {
		fixedNodes := make([]byte, 0, len(c.fixedNodes)*HashLength)
		for key := range c.fixedNodes {
			fixedNodes = append(fixedNodes, key[:]...)
		}
		(*txn).Set([]byte("fixedNodes"), fixedNodes)
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
