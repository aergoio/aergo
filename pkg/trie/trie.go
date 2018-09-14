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

// Trie is a modified sparse Merkle tree.
// Instead of storing values at the leaves of the tree,
// the values are stored at the highest subtree root that contains only that value.
// If the tree is sparse, this requires fewer hashing operations.
type Trie struct {
	db *CacheDB
	// Root is the current root of the smt.
	Root []byte
	// lock is for the whole struct
	lock sync.RWMutex
	hash func(data ...[]byte) []byte
	// KeySize is the size in bytes corresponding to TrieHeight, is size of an address.
	//KeySize       uint64
	TrieHeight    uint64
	defaultHashes [][]byte
	// LoadDbCounter counts the nb of db reads in on update
	LoadDbCounter uint64
	// loadDbMux is a lock for LoadDbCounter
	loadDbMux sync.RWMutex
	// LoadCacheCounter counts the nb of cache reads in on update
	LoadCacheCounter uint64
	// liveCountMux is a lock fo LoadCacheCounter
	liveCountMux sync.RWMutex
	// counterOn is used to enable/diseable for efficiency
	counterOn bool
	// CacheHeightLimit is the number of tree levels we want to store in cache
	CacheHeightLimit uint64
	// pastTries stores the past maxPastTries trie roots to revert
	pastTries [][]byte
}

// NewSMT creates a new SMT given a keySize and a hash function.
func NewTrie(root []byte, hash func(data ...[]byte) []byte, store db.DB) *Trie {
	s := &Trie{
		hash:             hash,
		TrieHeight:       uint64(len(hash([]byte("height"))) * 8), // hash any string to get output length
		CacheHeightLimit: 233,                                     // 246, //234, // based on the number of nodes we can keep in memory.
		counterOn:        false,
	}
	s.db = &CacheDB{
		liveCache:    make(map[Hash][][]byte),
		updatedNodes: make(map[Hash][][]byte),
		store:        store,
	}
	s.Root = root
	s.loadDefaultHashes()
	return s
}

// loadDefaultHashes creates the default hashes and stores them in cache
func (s *Trie) loadDefaultHashes() []byte {
	s.defaultHashes = make([][]byte, s.TrieHeight+1)
	s.defaultHashes[0] = DefaultLeaf
	var h []byte
	for i := 1; i <= int(s.TrieHeight); i++ {
		h = s.hash(s.defaultHashes[i-1], s.defaultHashes[i-1])
		s.defaultHashes[i] = h
	}
	return h
}

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
func (s *Trie) loadCache(root []byte, batch [][]byte, iBatch uint8, height uint64, ch chan<- (error)) {
	if height <= s.CacheHeightLimit+1 || len(root) == 0 {
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

// Update adds and deletes a sorted list of keys and their values to the trie
// Adding and deleting can be simultaneous as long as keys are sorted.
// To delete, set the value to DefaultLeaf.
// Make sure you don't delete keys that don't exist,
// the root hash would become invalid.
func (s *Trie) Update(keys, values [][]byte) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.LoadDbCounter = 0
	s.LoadCacheCounter = 0
	ch := make(chan mresult, 1)
	s.update(s.Root, keys, values, nil, 0, s.TrieHeight, ch)
	result := <-ch
	if result.err != nil {
		return nil, result.err
	}
	s.Root = result.update[:HashLength]
	return s.Root, nil
}

// mresult is used to contain the result of goroutines and is sent through a channel.
type mresult struct {
	update []byte
	// flag if a node was deleted and a shortcut node maybe has to move up the tree
	deleted bool
	err     error
}

// update adds and deletes a sorted list of keys and their values to the trie.
// Adding and deleting can be simultaneous as long as keys are sorted.
// To delete, set the value to DefaultLeaf.
// It returns the root of the updated tree.
// A DefaultLeaf shouldn't be updated to a DefaultLeaf as shortcut nodes
// could be moved down the tree resulting in an invalid root
func (s *Trie) update(root []byte, keys, values, batch [][]byte, iBatch uint8, height uint64, ch chan<- (mresult)) {
	if height == 0 {
		if bytes.Equal(DefaultLeaf, values[0]) {
			//TODO
			// Delete the key-value from the trie if it is being set to DefaultLeaf
			// The value will be set to [] in batch by maybeMoveupShortcut or interiorHash
			s.deleteCacheNode(root, height)
			ch <- mresult{nil, true, nil}
		} else {
			// create a new shortcut batch.
			// simply storing the value will make it hard to move up the
			// shortcut in case of sibling deletion
			batch = make([][]byte, 31, 31)
			node := s.leafHash(keys[0], values[0], batch, 0, height, root)
			ch <- mresult{node, false, nil}
		}
		return
	}

	// Load the node to update
	batch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(root, height, batch, iBatch, true)
	if err != nil {
		ch <- mresult{nil, false, err}
		return
	}

	// Check if the keys are updating the shortcut node
	if isShortcut {
		keys, values = s.maybeAddShortcutToKV(keys, values, lnode[:HashLength], rnode[:HashLength])
		if iBatch == 0 {
			s.deleteCacheNode(root, height)
		}
		// The shortcut node was added to keys and values so consider this subtree default.
		root = nil
		lnode, rnode = nil, nil
		// update in the batch (set key, value to default so the next loadChildren is correct)
		batch[2*iBatch+1] = lnode
		batch[2*iBatch+2] = rnode
		if len(keys) == 0 {
			// Set true so that a potential sibling shortcut may move up.
			ch <- mresult{root, true, nil}
			return
		}
	}
	// Store shortcut node
	if (len(lnode) == 0) && (len(rnode) == 0) && (len(keys) == 1) {
		// We are adding 1 key to an empty subtree so store it as a shortcut
		node := s.leafHash(keys[0], values[0], batch, iBatch, height, root)
		ch <- mresult{node, false, nil}
		return
	}

	// Split the keys array so each branch can be updated in parallel
	lkeys, rkeys := s.splitKeys(keys, s.TrieHeight-height)
	splitIndex := len(lkeys)
	lvalues, rvalues := values[:splitIndex], values[splitIndex:]

	switch {
	case len(lkeys) == 0 && len(rkeys) > 0:
		s.updateRight(lnode, rnode, root, keys, values, batch, iBatch, height, ch)
	case len(lkeys) > 0 && len(rkeys) == 0:
		s.updateLeft(lnode, rnode, root, keys, values, batch, iBatch, height, ch)
	default:
		s.updateParallel(lnode, rnode, root, lkeys, rkeys, lvalues, rvalues, batch, iBatch, height, ch)
	}
}

// updateRight updates the right side of the tree
func (s *Trie) updateRight(lnode, rnode, root []byte, keys, values, batch [][]byte, iBatch uint8, height uint64, ch chan<- (mresult)) {
	// all the keys go in the right subtree
	newch := make(chan mresult, 1)
	s.update(rnode, keys, values, batch, 2*iBatch+2, height-1, newch)
	result := <-newch
	if result.err != nil {
		ch <- mresult{nil, false, result.err}
		return
	}
	// Move up a shortcut node if necessary.
	if result.deleted {
		if s.maybeMoveUpShortcut(lnode, result.update, root, batch, iBatch, height, ch) {
			return
		}
	}
	node := s.interiorHash(lnode, result.update, batch, iBatch, height, root)
	ch <- mresult{node, false, nil}
}

// updateLeft updates the left side of the tree
func (s *Trie) updateLeft(lnode, rnode, root []byte, keys, values, batch [][]byte, iBatch uint8, height uint64, ch chan<- (mresult)) {
	// all the keys go in the left subtree
	newch := make(chan mresult, 1)
	s.update(lnode, keys, values, batch, 2*iBatch+1, height-1, newch)
	result := <-newch
	if result.err != nil {
		ch <- mresult{nil, false, result.err}
		return
	}
	// Move up a shortcut node if necessary.
	if result.deleted {
		if s.maybeMoveUpShortcut(result.update, rnode, root, batch, iBatch, height, ch) {
			return
		}
	}
	node := s.interiorHash(result.update, rnode, batch, iBatch, height, root)
	ch <- mresult{node, false, nil}
}

// updateParallel updates both sides of the trie simultaneously
func (s *Trie) updateParallel(lnode, rnode, root []byte, lkeys, rkeys, lvalues, rvalues, batch [][]byte, iBatch uint8, height uint64, ch chan<- (mresult)) {
	lch := make(chan mresult, 1)
	rch := make(chan mresult, 1)
	go s.update(lnode, lkeys, lvalues, batch, 2*iBatch+1, height-1, lch)
	go s.update(rnode, rkeys, rvalues, batch, 2*iBatch+2, height-1, rch)
	lresult := <-lch
	rresult := <-rch
	if lresult.err != nil {
		ch <- mresult{nil, false, lresult.err}
		return
	}
	if rresult.err != nil {
		ch <- mresult{nil, false, rresult.err}
		return
	}

	// Move up a shortcut node if it's sibling is default
	if lresult.deleted || rresult.deleted {
		if s.maybeMoveUpShortcut(lresult.update, rresult.update, root, batch, iBatch, height, ch) {
			return
		}
	}
	node := s.interiorHash(lresult.update, rresult.update, batch, iBatch, height, root)
	ch <- mresult{node, false, nil}
}

// deleteCacheNode deletes the node from liveCache
func (s *Trie) deleteCacheNode(root []byte, height uint64) {
	if height >= s.CacheHeightLimit {
		var node Hash
		copy(node[:], root)
		s.db.liveMux.Lock()
		delete(s.db.liveCache, node)
		s.db.liveMux.Unlock()
	}
}

// splitKeys devides the array of keys into 2 so they can update left and right branches in parallel
func (s *Trie) splitKeys(keys [][]byte, height uint64) ([][]byte, [][]byte) {
	for i, key := range keys {
		if bitIsSet(key, height) {
			return keys[:i], keys[i:]
		}
	}
	return keys, nil
}

// maybeMoveUpShortcut moves up a shortcut if it's sibling node is default
func (s *Trie) maybeMoveUpShortcut(left, right, root []byte, batch [][]byte, iBatch uint8, height uint64, ch chan<- (mresult)) bool {
	if len(left) == 0 && len(right) == 0 {
		// Both update and sibling are deleted subtrees
		if iBatch == 0 {
			// If the deleted subtrees are at the root, then delete it.
			s.deleteCacheNode(root, height)
		} else {
			batch[2*iBatch+1] = nil
			batch[2*iBatch+2] = nil
		}
		//ch <- mresult{s.defaultHashes[height], false, nil}
		ch <- mresult{nil, false, nil}
		return true
	} else if len(left) == 0 {
		// If right is a shortcut move it up
		if right[HashLength] == 1 {
			return s.moveUpShortcut(right, root, batch, iBatch, 2*iBatch+2, height, ch)
		}
	} else if len(right) == 0 {
		// If left is a shortcut move it up
		if left[HashLength] == 1 {
			return s.moveUpShortcut(left, root, batch, iBatch, 2*iBatch+1, height, ch)
		}
	}
	return false
}

func (s *Trie) moveUpShortcut(shortcut, root []byte, batch [][]byte, iBatch, iShortcut uint8, height uint64, ch chan<- (mresult)) bool {
	_, _, shortcutKey, shortcutVal, _, err := s.loadChildren(shortcut, height-1, batch, iShortcut, false)

	// when moving up the shortcut, it's hash will change because height is +1
	newShortcut := s.hash(shortcutKey[:HashLength], shortcutVal[:HashLength], []byte{byte(height)})
	newShortcut = append(newShortcut, byte(1))
	if err != nil {
		ch <- mresult{nil, false, err}
		return false
	}
	if iBatch == 0 {
		// Modify batch to a shortcut batch
		batch[0] = []byte{1}
		batch[2*iBatch+1] = shortcutKey
		batch[2*iBatch+2] = shortcutVal
		batch[2*iShortcut+1] = nil
		batch[2*iShortcut+2] = nil
		s.storeNode(batch, newShortcut, root, height)
		s.deleteCacheNode(root, height)
	} else if (height-1)%4 == 0 {
		// move up shortcut and delete old batch
		batch[2*iBatch+1] = shortcutKey
		batch[2*iBatch+2] = shortcutVal
		s.deleteCacheNode(shortcut, height-1)
	} else {
		//move up shortcut
		batch[2*iBatch+1] = shortcutKey
		batch[2*iBatch+2] = shortcutVal
		batch[2*iShortcut+1] = nil
		batch[2*iShortcut+2] = nil
	}
	// Return the left sibling node to move it up
	ch <- mresult{newShortcut, true, nil}
	return true
}

// maybeAddShortcutToKV adds a shortcut key to the keys array to be updated.
// this is used when a subtree containing a shortcut node is being updated
func (s *Trie) maybeAddShortcutToKV(keys, values [][]byte, shortcutKey, shortcutVal []byte) ([][]byte, [][]byte) {
	newKeys := make([][]byte, 0, len(keys)+1)
	newVals := make([][]byte, 0, len(keys)+1)

	if bytes.Compare(shortcutKey, keys[0]) < 0 {
		newKeys = append(newKeys, shortcutKey)
		newKeys = append(newKeys, keys...)
		newVals = append(newVals, shortcutVal)
		newVals = append(newVals, values...)
	} else if bytes.Compare(shortcutKey, keys[len(keys)-1]) > 0 {
		newKeys = append(newKeys, keys...)
		newKeys = append(newKeys, shortcutKey)
		newVals = append(newVals, values...)
		newVals = append(newVals, shortcutVal)
	} else {
		higher := false
		for i, key := range keys {
			if bytes.Equal(shortcutKey, key) {
				if bytes.Equal(DefaultLeaf, values[i]) {
					// Delete shortcut if it is updated to DefaultLeaf
					keys = append(keys[:i], keys[i+1:]...)
					values = append(values[:i], values[i+1:]...)
				}
				return keys, values
			}
			if bytes.Compare(shortcutKey, key) > 0 {
				higher = true
			}
			if higher && bytes.Compare(shortcutKey, key) < 0 {
				// insert shortcut in slices
				newKeys = append(newKeys, keys[:i]...)
				newKeys = append(newKeys, shortcutKey)
				newKeys = append(newKeys, keys[i:]...)
				newVals = append(newVals, values[:i]...)
				newVals = append(newVals, shortcutVal)
				newVals = append(newVals, values[i:]...)
			}
		}
	}
	return newKeys, newVals
}

// loadChildren looks for the children of a node.
// if the node is not stored in cache, it will be loaded from db.
func (s *Trie) loadChildren(root []byte, height uint64, batch [][]byte, iBatch uint8, updateSafe bool) ([][]byte, uint8, []byte, []byte, bool, error) {
	isShortcut := false
	if height%4 == 0 {
		if len(root) == 0 {
			// create a new default batch
			batch = make([][]byte, 31, 31)
			batch[0] = []byte{0}
		} else {
			var err error
			batch, err = s.loadBatch(root, updateSafe)
			if err != nil {
				return nil, 0, nil, nil, false, err
			}
		}
		iBatch = 0
		if batch[0][0] == 1 {
			isShortcut = true
		}
	} else {
		if len(batch[iBatch]) != 0 && batch[iBatch][HashLength] == 1 {
			isShortcut = true
		}
	}
	return batch, iBatch, batch[2*iBatch+1], batch[2*iBatch+2], isShortcut, nil
}

// loadBatch fetches a batch of nodes in cache or db
func (s *Trie) loadBatch(root []byte, updateSafe bool) ([][]byte, error) {
	var node Hash
	copy(node[:], root)

	s.db.liveMux.RLock()
	val, exists := s.db.liveCache[node]
	s.db.liveMux.RUnlock()
	if exists {
		if s.counterOn {
			s.liveCountMux.Lock()
			s.LoadCacheCounter++
			s.liveCountMux.Unlock()
		}
		if updateSafe {
			// Return a copy so that Commit() doesnt have to be called at
			// each block and still commit every state transition.
			newVal := make([][]byte, 31, 31)
			copy(newVal, val)
			return newVal, nil
		}
		return val, nil
	}

	// checking updated nodes is useful if get() or update() is called twice in a row without db commit
	s.db.updatedMux.RLock()
	val, exists = s.db.updatedNodes[node]
	s.db.updatedMux.RUnlock()
	if exists {
		if updateSafe {
			// Return a copy so that Commit() doesnt have to be called at
			// each block and still commit every state transition.
			newVal := make([][]byte, 31, 31)
			copy(newVal, val)
			return newVal, nil
		}
		return val, nil
	}
	//Fetch node in disk database
	if s.db.store == nil {
		return nil, fmt.Errorf("DB not connected to trie")
	}
	if s.counterOn {
		s.loadDbMux.Lock()
		s.LoadDbCounter++
		s.loadDbMux.Unlock()
	}
	s.db.lock.Lock()
	dbval := s.db.store.Get(root[:HashLength])
	s.db.lock.Unlock()
	nodeSize := len(dbval)
	if nodeSize != 0 {
		return s.parseBatch(dbval), nil
	}
	return nil, fmt.Errorf("the trie node %x is unavailable in the disk db, db may be corrupted", root)
}

// parseBatch decodes the byte data into a slice of nodes and bitmap
func (s *Trie) parseBatch(val []byte) [][]byte {
	batch := make([][]byte, 31, 31)
	bitmap := val[:4]
	// check if the batch root is a shortcut
	if bitIsSet(val, 31) {
		batch[0] = []byte{1}
		batch[1] = val[4 : 4+33]
		batch[2] = val[4+33 : 4+33*2]
	} else {
		batch[0] = []byte{0}
		j := 0
		for i := 1; i <= 30; i++ {
			if bitIsSet(bitmap, uint64(i-1)) {
				batch[i] = val[4+33*j : 4+33*(j+1)]
				j++
			}
		}
	}
	return batch
}

// Get fetches the value of a key by going down the current trie root.
func (s *Trie) Get(key []byte) ([]byte, error) {
	return s.get(s.Root, key, nil, 0, s.TrieHeight)
}

// get fetches the value of a key given a trie root
func (s *Trie) get(root []byte, key []byte, batch [][]byte, iBatch uint8, height uint64) ([]byte, error) {
	if len(root) == 0 {
		// the trie does not contain the key
		return nil, nil
	}
	// Fetch the children of the node
	batch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(root, height, batch, iBatch, false)
	// TODO make a loadChildren without copy
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

// DefaultHash is a getter for the defaultHashes array
func (s *Trie) DefaultHash(height uint64) []byte {
	return s.defaultHashes[height]
}

// leafHash returns the hash of key_value_byte(height) concatenated, stores it in the updatedNodes and maybe in liveCache.
// keys of go mappings cannot be byte slices so the hash is copied to a byte array
func (s *Trie) leafHash(key, value []byte, batch [][]byte, iBatch uint8, height uint64, oldRoot []byte) []byte {
	// byte(height) is here for 2 reasons.
	// 1- to prevent potential problems with merkle proofs where if an account
	// has the same address as a node, it would be possible to prove a
	// different value for the account.
	// 2- when accounts are added to the trie, accounts on their path get pushed down the tree
	// with them. if an old account changes position from a shortcut batch to another
	// shortcut batch of different height, if would be deleted when reverting.
	h := s.hash(key, value, []byte{byte(height)})
	h = append(h, byte(1)) // byte(1) is a flag for the shortcut
	batch[2*iBatch+2] = append(value, byte(2))
	batch[2*iBatch+1] = append(key, byte(2))
	if height%4 == 0 {
		batch[0] = []byte{1} // byte(1) is a flag for the shortcut batch
		s.storeNode(batch, h, oldRoot, height)
	}
	return h
}

// storeNode stores a batch and deleted the old node from cache
func (s *Trie) storeNode(batch [][]byte, h, oldRoot []byte, height uint64) {
	// store new node in db
	var node Hash
	copy(node[:], h)
	s.db.updatedMux.Lock()
	s.db.updatedNodes[node] = batch
	s.db.updatedMux.Unlock()
	// Cache the shortcut node if it's height is over CacheHeightLimit
	if height >= s.CacheHeightLimit {
		s.db.liveMux.Lock()
		s.db.liveCache[node] = batch
		if (len(oldRoot) != 0) && !bytes.Equal(h, oldRoot) {
			// Delete old liveCache node if it has been updated and is not default
			var node Hash
			copy(node[:], oldRoot)
			delete(s.db.liveCache, node)
		}
		s.db.liveMux.Unlock()
	}
}

// interiorHash hashes 2 children to get the parent hash and stores it in the updatedNodes and maybe in liveCache.
func (s *Trie) interiorHash(left, right []byte, batch [][]byte, iBatch uint8, height uint64, oldRoot []byte) []byte {
	var h []byte
	if (len(left) == 0) && (len(right)) == 0 {
		// if a key was deleted, the node becomes default
		batch[2*iBatch+1] = left
		batch[2*iBatch+2] = right
		return nil
	} else if len(left) == 0 {
		h = s.hash(s.defaultHashes[height-1], right[:HashLength])
	} else if len(right) == 0 {
		h = s.hash(left[:HashLength], s.defaultHashes[height-1])
	} else {
		h = s.hash(left[:HashLength], right[:HashLength])
	}
	h = append(h, byte(0))
	batch[2*iBatch+2] = right
	batch[2*iBatch+1] = left
	if height%4 == 0 {
		batch[0] = []byte{0}
		s.storeNode(batch, h, oldRoot, height)
	}
	return h
}

// Commit stores the updated nodes to disk
func (s *Trie) Commit() error {
	if s.db.store == nil {
		return fmt.Errorf("DB not connected to trie")
	}
	// Commit the new nodes to database, clear updatedNodes and store the Root in history for reverts.
	if len(s.pastTries) >= maxPastTries {
		copy(s.pastTries, s.pastTries[1:])
		s.pastTries[len(s.pastTries)-1] = s.Root
	} else {
		s.pastTries = append(s.pastTries, s.Root)
	}
	s.db.commit()
	s.db.updatedNodes = make(map[Hash][][]byte, len(s.db.updatedNodes)*2)
	return nil
}
