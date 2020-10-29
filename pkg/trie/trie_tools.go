/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/aergoio/aergo-lib/db"
)

// LoadCache loads the first layers of the merkle tree given a root
// This is called after a node restarts so that it doesnt become slow with db reads
// LoadCache also updates the Root with the given root.
func (s *Trie) LoadCache(root []byte) error {
	if s.db.Store == nil {
		return fmt.Errorf("DB not connected to trie")
	}
	s.db.liveCache = make(map[Hash][][]byte)
	ch := make(chan error, 1)
	s.loadCache(root, nil, 0, s.TrieHeight, ch)
	s.Root = root
	return <-ch
}

// loadCache loads the first layers of the merkle tree given a root
func (s *Trie) loadCache(root []byte, batch [][]byte, iBatch, height int, ch chan<- (error)) {
	if height < s.CacheHeightLimit || len(root) == 0 {
		ch <- nil
		return
	}
	if height%4 == 0 {
		// Load the node from db
		s.db.lock.Lock()
		dbval := s.db.Store.Get(root[:HashLength])
		s.db.lock.Unlock()
		if len(dbval) == 0 {
			ch <- fmt.Errorf("the trie node %x is unavailable in the disk db, db may be corrupted", root)
			return
		}
		//Store node in cache.
		var node Hash
		copy(node[:], root)
		batch = s.parseBatch(dbval)
		s.db.liveMux.Lock()
		s.db.liveCache[node] = batch
		s.db.liveMux.Unlock()
		iBatch = 0
		if batch[0][0] == 1 {
			// if height == 0 this will also return
			ch <- nil
			return
		}
	}
	if iBatch != 0 && batch[iBatch][HashLength] == 1 {
		// Check if node is a leaf node
		ch <- nil
	} else {
		// Load subtree
		lnode, rnode := batch[2*iBatch+1], batch[2*iBatch+2]

		lch := make(chan error, 1)
		rch := make(chan error, 1)
		go s.loadCache(lnode, batch, 2*iBatch+1, height-1, lch)
		go s.loadCache(rnode, batch, 2*iBatch+2, height-1, rch)
		if err := <-lch; err != nil {
			ch <- err
			return
		}
		if err := <-rch; err != nil {
			ch <- err
			return
		}
		ch <- nil
	}
}

// Get fetches the value of a key by going down the current trie root.
func (s *Trie) Get(key []byte) ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.atomicUpdate = false
	return s.get(s.Root, key, nil, 0, s.TrieHeight)
}

// get fetches the value of a key given a trie root
func (s *Trie) get(root, key []byte, batch [][]byte, iBatch, height int) ([]byte, error) {
	if len(root) == 0 {
		// the trie does not contain the key
		return nil, nil
	}
	// Fetch the children of the node
	batch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(root, height, iBatch, batch)
	if err != nil {
		return nil, err
	}
	if isShortcut {
		if bytes.Equal(lnode[:HashLength], key) {
			return rnode[:HashLength], nil
		}
		// also returns nil if height 0 is not a shortcut
		return nil, nil
	}
	if bitIsSet(key, s.TrieHeight-height) {
		return s.get(rnode, key, batch, 2*iBatch+2, height-1)
	}
	return s.get(lnode, key, batch, 2*iBatch+1, height-1)
}

// WalkResult contains the key and value obtained with a Walk() operation
type WalkResult struct {
	Value []byte
	Key   []byte
}

// Walk finds all the trie stored values from left to right and calls callback.
// If callback returns true, the walk will stop, else it will continue.
func (s *Trie) Walk(callback func(*WalkResult) bool) error {
	walkc := make(chan *WalkResult)
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.atomicUpdate = false
	close := make(chan (bool), 1)
	stop := false
	wg := sync.WaitGroup{}
	go func() {
		for {
			select {
			case <-close:
				break
			case value := <-walkc:
				wg.Add(1)
				if stop = callback(value); stop {
					wg.Done()
					break
				}
				wg.Done()
			}
		}
	}()
	err := s.walk(walkc, &stop, s.Root, nil, 0, s.TrieHeight)
	close <- true
	wg.Wait()
	return err
}

// walk fetches the value of a key given a trie root
func (s *Trie) walk(walkc chan (*WalkResult), stop *bool, root []byte, batch [][]byte, iiBatch, height int) error {
	if len(root) == 0 || *stop {
		// the trie does not contain the key or stop bool is set tu true
		return nil
	}
	// Fetch the children of the node
	nbatch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(root, height, iiBatch, batch)
	if err != nil {
		return err
	}
	if isShortcut {
		var key []byte
		if len(rnode) <= HashLength {
			return nil
		}
		if string(nbatch[0]) == string([]byte{0x01}) {
			key = nbatch[iBatch+1][:HashLength]
		} else {
			key = batch[2*iiBatch+1][:HashLength]
		}
		walkc <- &WalkResult{Value: rnode[:HashLength], Key: key}
		return nil
	}
	// Go left
	if err := s.walk(walkc, stop, lnode, nbatch, 2*iBatch+1, height-1); err != nil {
		return err
	}
	// Go Right
	if err := s.walk(walkc, stop, rnode, nbatch, 2*iBatch+2, height-1); err != nil {
		return err
	}
	return nil
}

// TrieRootExists returns true if the root exists in Database.
func (s *Trie) TrieRootExists(root []byte) bool {
	s.db.lock.RLock()
	dbval := s.db.Store.Get(root)
	s.db.lock.RUnlock()
	if len(dbval) != 0 {
		return true
	}
	return false
}

// Commit stores the updated nodes to disk.
// Commit should be called for every block otherwise past tries
// are not recorded and it is not possible to revert to them
// (except if AtomicUpdate is used, which records every state).
func (s *Trie) Commit() error {
	if s.db.Store == nil {
		return fmt.Errorf("DB not connected to trie")
	}
	// NOTE The tx interface doesnt handle ErrTxnTooBig
	txn := s.db.Store.NewTx().(DbTx)
	s.StageUpdates(txn)
	txn.(db.Transaction).Commit()
	return nil
}

// StageUpdates requires a database transaction as input
// Unlike Commit(), it doesnt commit the transaction
// the database transaction MUST be commited otherwise the
// state ROOT will not exist.
func (s *Trie) StageUpdates(txn DbTx) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// Commit the new nodes to database, clear updatedNodes and store the Root in pastTries for reverts.
	if !s.atomicUpdate {
		// if previously AtomicUpdate was called, then past tries is already updated
		s.updatePastTries()
	}
	s.db.commit(&txn)

	s.db.updatedNodes = make(map[Hash][][]byte)
	s.prevRoot = s.Root
}

// Stash rolls back the changes made by previous updates
// and loads the cache from before the rollback.
func (s *Trie) Stash(rollbackCache bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Root = s.prevRoot
	if rollbackCache {
		// Making a temporary liveCache requires it to be copied, so it's quicker
		// to just load the cache from DB if a block state root was incorrect.
		s.db.liveCache = make(map[Hash][][]byte)
		ch := make(chan error, 1)
		s.loadCache(s.Root, nil, 0, s.TrieHeight, ch)
		err := <-ch
		if err != nil {
			return err
		}
	} else {
		s.db.liveCache = make(map[Hash][][]byte)
	}
	s.db.updatedNodes = make(map[Hash][][]byte)
	// also stash past tries created by Atomic update
	for i := len(s.pastTries) - 1; i >= 0; i-- {
		if bytes.Equal(s.pastTries[i], s.Root) {
			break
		} else {
			// remove from past tries
			s.pastTries = s.pastTries[:len(s.pastTries)-1]
		}
	}
	return nil
}
