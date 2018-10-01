/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"fmt"
)

// LoadCache loads the first layers of the merkle tree given a root
// This is called after a node restarts so that it doesnt become slow with db reads
// LoadCache also updates the Root with the given root.
func (s *Trie) LoadCache(root []byte) error {
	if s.db.store == nil {
		return fmt.Errorf("DB not connected to trie")
	}
	s.loadDefaultHashes()
	ch := make(chan error, 1)
	s.loadCache(root, nil, 0, s.TrieHeight, ch)
	s.Root = root
	return <-ch
}

// loadCache loads the first layers of the merkle tree given a root
func (s *Trie) loadCache(root []byte, batch [][]byte, iBatch, height uint64, ch chan<- (error)) {
	if height < s.CacheHeightLimit || len(root) == 0 {
		ch <- nil
		return
	}
	if height%4 == 0 {
		// Load the node from db
		s.db.lock.Lock()
		dbval := s.db.store.Get(root[:HashLength])
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
		// Check if node is default
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
func (s *Trie) get(root, key []byte, batch [][]byte, iBatch, height uint64) ([]byte, error) {
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

// TrieRootExists returns true if the root exists in Database.
func (s *Trie) TrieRootExists(root []byte) bool {
	s.db.lock.RLock()
	dbval := s.db.store.Get(root)
	s.db.lock.RUnlock()
	if len(dbval) != 0 {
		return true
	}
	return false
}

// DefaultHash is a getter for the defaultHashes array
func (s *Trie) DefaultHash(height uint64) []byte {
	return s.defaultHashes[height]
}

// Commit stores the updated nodes to disk.
// Commit should be called for every block otherwise past tries are not recorded and it is not possible to revert to them
func (s *Trie) Commit() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.db.store == nil {
		return fmt.Errorf("DB not connected to trie")
	}
	// Commit the new nodes to database, clear updatedNodes and store the Root in history for reverts.
	if !s.atomicUpdate {
		// if previously AtomicUpdate was called, then past tries is already updated
		if len(s.pastTries) >= maxPastTries {
			copy(s.pastTries, s.pastTries[1:])
			s.pastTries[len(s.pastTries)-1] = s.Root
		} else {
			s.pastTries = append(s.pastTries, s.Root)
		}
	}
	s.db.commit()
	s.db.updatedNodes = make(map[Hash][][]byte)
	return nil
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
	return nil
}
