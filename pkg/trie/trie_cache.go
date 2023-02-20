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
	// nodes marked to be deleted from the database
	deletedNodes map[Hash]bool
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
	// whether this is a light node
	lightnode bool
}

// commit adds updatedNodes to the given database transaction.
func (c *CacheDB) commit(txn *DbTx) {
	c.updatedMux.Lock()
	defer c.updatedMux.Unlock()
	// add all updated nodes
	for key, batch := range c.updatedNodes {
		var node []byte
		node = append(node, key[:]...)
		logger.Debug().Msgf("commit - saving: %x", node)
		// add the node to the database
		(*txn).Set(node, c.serializeBatch(batch))
		// if the node is on the list of deleted nodes, remove it
		if c.lightnode && c.deletedNodes[key] {
		  logger.Debug().Msgf("commit - disabling node from deletion: %x", node)
		  c.deletedNodes[key] = false
		}
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
