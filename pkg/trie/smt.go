/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

// The Package Trie implements a sparse merkle trie.

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/aergoio/aergo-lib/db"
)

// TODO make a secure trie that hashes keys with a random seed to be sure the trie is sparse.

// SMT is a sparse Merkle tree.
type SMT struct {
	db *CacheDB
	// Root is the current root of the smt.
	Root []byte
	// lock is for the whole struct
	lock sync.RWMutex
	hash func(data ...[]byte) []byte
	// KeySize is the size in bytes corresponding to TrieHeight, is size of an address.
	KeySize       uint64
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
	// CacheHeightLimit is the number of tree levels we want to store in cache
	CacheHeightLimit uint64
	// pastTries stores the past maxPastTries trie roots to revert
	pastTries [][]byte
}

// NewSMT creates a new SMT given a keySize and a hash function.
func NewSMT(keySize uint64, hash func(data ...[]byte) []byte, store db.DB) *SMT {
	s := &SMT{
		hash:             hash,
		TrieHeight:       keySize * 8,
		CacheHeightLimit: 232, //246, //234, // based on the number of nodes we can keep in memory.
		KeySize:          keySize,
	}
	s.db = &CacheDB{
		liveCache:    make(map[Hash][]byte, 5e6),
		updatedNodes: make(map[Hash][]byte, 5e3),
		store:        store,
	}
	s.Root = s.loadDefaultHashes()
	return s
}

// loadDefaultHashes creates the default hashes and stores them in cache
func (s *SMT) loadDefaultHashes() []byte {
	s.defaultHashes = make([][]byte, s.TrieHeight+1)
	s.defaultHashes[0] = DefaultLeaf
	var h []byte
	var node Hash
	for i := 1; i <= int(s.TrieHeight); i++ {
		h = s.hash(s.defaultHashes[i-1], s.defaultHashes[i-1])
		copy(node[:], h)
		s.defaultHashes[i] = h
		// default hashes are always in livecache and don't need to be stored to disk
		s.db.liveCache[node] = append(s.defaultHashes[i-1], append(s.defaultHashes[i-1], byte(0))...)
	}
	return h
}

// Update adds a sorted list of keys and their values to the trie
func (s *SMT) Update(keys, values [][]byte) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.LoadDbCounter = 0
	s.LoadCacheCounter = 0
	ch := make(chan result, 1)
	s.update(s.Root, keys, values, s.TrieHeight, false, true, ch)
	result := <-ch
	if result.err != nil {
		return nil, result.err
	}
	s.Root = result.update
	return s.Root, nil
}

// result is used to contain the result of goroutines and is sent through a channel.
type result struct {
	update []byte
	err    error
}

// update adds a sorted list of keys and their values to the trie.
// It returns the root of the updated tree.
func (s *SMT) update(root []byte, keys, values [][]byte, height uint64, shortcut, store bool, ch chan<- (result)) {
	if height == 0 {
		ch <- result{values[0], nil}
		return
	}
	lnode, rnode, isShortcut, err := s.loadChildren(root)
	if err != nil {
		ch <- result{nil, err}
		return
	}
	if isShortcut == 1 {
		keys, values = s.maybeAddShortcutToKV(keys, values, lnode, rnode)
		// The shortcut node was added to keys and values so consider this subtree default.
		lnode, rnode = s.defaultHashes[height-1], s.defaultHashes[height-1]
	}

	// Split the keys array so each branch can be updated in parallel
	lkeys, rkeys := s.splitKeys(keys, s.TrieHeight-height)
	splitIndex := len(lkeys)
	lvalues, rvalues := values[:splitIndex], values[splitIndex:]

	if shortcut {
		store = false    //stop storing only after the shortcut node.
		shortcut = false // remove shortcut node flag
	}
	if bytes.Equal(s.defaultHashes[height-1], lnode) && bytes.Equal(s.defaultHashes[height-1], rnode) && (len(keys) == 1) && store {
		// if the subtree contains only one key, store the key/value in a shortcut node
		shortcut = true
	}
	switch {
	case len(lkeys) == 0 && len(rkeys) > 0:
		s.updateRight(lnode, rnode, root, keys, values, height, shortcut, store, ch)
	case len(lkeys) > 0 && len(rkeys) == 0:
		s.updateLeft(lnode, rnode, root, keys, values, height, shortcut, store, ch)
	default:
		s.updateParallel(lnode, rnode, root, keys, values, lkeys, rkeys, lvalues, rvalues, height, shortcut, store, ch)
	}
}

// updateParallel updates both sides of the trie simultaneously
func (s *SMT) updateParallel(lnode, rnode, root []byte, keys, values, lkeys, rkeys, lvalues, rvalues [][]byte, height uint64, shortcut, store bool, ch chan<- (result)) {
	// keys are separated between the left and right branches
	// update the branches in parallel
	lch := make(chan result, 1)
	rch := make(chan result, 1)
	go s.update(lnode, lkeys, lvalues, height-1, shortcut, store, lch)
	go s.update(rnode, rkeys, rvalues, height-1, shortcut, store, rch)
	lresult := <-lch
	rresult := <-rch
	if lresult.err != nil {
		ch <- result{nil, lresult.err}
		return
	}
	if rresult.err != nil {
		ch <- result{nil, rresult.err}
		return
	}
	ch <- result{s.interiorHash(lresult.update, rresult.update, height-1, root, shortcut, store, keys, values), nil}

}

// updateRight updates the right side of the tree
func (s *SMT) updateRight(lnode, rnode, root []byte, keys, values [][]byte, height uint64, shortcut, store bool, ch chan<- (result)) {
	// all the keys go in the right subtree
	newch := make(chan result, 1)
	s.update(rnode, keys, values, height-1, shortcut, store, newch)
	res := <-newch
	if res.err != nil {
		ch <- result{nil, res.err}
		return
	}
	ch <- result{s.interiorHash(lnode, res.update, height-1, root, shortcut, store, keys, values), nil}
}

// updateLeft updates the left side of the tree
func (s *SMT) updateLeft(lnode, rnode, root []byte, keys, values [][]byte, height uint64, shortcut, store bool, ch chan<- (result)) {
	// all the keys go in the left subtree
	newch := make(chan result, 1)
	s.update(lnode, keys, values, height-1, shortcut, store, newch)
	res := <-newch
	if res.err != nil {
		ch <- result{nil, res.err}
		return
	}
	ch <- result{s.interiorHash(res.update, rnode, height-1, root, shortcut, store, keys, values), nil}
}

// splitKeys devides the array of keys into 2 so they can update left and right branches in parallel
func (s *SMT) splitKeys(keys [][]byte, height uint64) ([][]byte, [][]byte) {
	for i, key := range keys {
		if bitIsSet(key, height) {
			return keys[:i], keys[i:]
		}
	}
	return keys, nil
}

// maybeAddShortcutToKV adds a shortcut key to the keys array to be updated.
// this is used when a subtree containing a shortcut node is being updated
func (s *SMT) maybeAddShortcutToKV(keys, values [][]byte, shortcutKey, shortcutVal []byte) ([][]byte, [][]byte) {
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
func (s *SMT) loadChildren(root []byte) ([]byte, []byte, byte, error) {
	var node Hash
	copy(node[:], root)

	s.db.liveMux.RLock()
	val, exists := s.db.liveCache[node]
	s.db.liveMux.RUnlock()
	if exists {
		s.liveCountMux.Lock()
		s.LoadCacheCounter++
		s.liveCountMux.Unlock()
		return s.parseValue(val)
	}

	// checking updated nodes is useful if get() or update() is called twice in a row without db commit
	s.db.updatedMux.RLock()
	val, exists = s.db.updatedNodes[node]
	s.db.updatedMux.RUnlock()
	if exists {
		return s.parseValue(val)
	}
	//Fetch node in disk database
	if s.db.store == nil {
		return nil, nil, byte(0), fmt.Errorf("DB not connected to trie")
	}
	s.loadDbMux.Lock()
	s.LoadDbCounter++
	s.loadDbMux.Unlock()
	s.db.lock.Lock()
	val = s.db.store.Get(root)
	s.db.lock.Unlock()
	nodeSize := len(val)
	if nodeSize != 0 {
		return s.parseValue(val)
	}
	return nil, nil, byte(0), fmt.Errorf("the trie node %x is unavailable in the disk db, db may be corrupted", root)
}

// parseValue returns a subtree roots or a shortcut node
func (s *SMT) parseValue(val []byte) ([]byte, []byte, byte, error) {
	nodeSize := len(val)
	shortcut := val[nodeSize-1]
	if shortcut == 1 {
		return val[:s.KeySize], val[s.KeySize : nodeSize-1], shortcut, nil
	}
	return val[:HashLength], val[HashLength : nodeSize-1], shortcut, nil
}

// Get fetches the value of a key by going down the current trie root.
func (s *SMT) Get(key []byte) ([]byte, error) {
	return s.get(s.Root, key, s.TrieHeight)
}

// get fetches the value of a key given a trie root
func (s *SMT) get(root []byte, key []byte, height uint64) ([]byte, error) {
	if height == 0 {
		if bytes.Equal(root, DefaultLeaf) {
			return nil, nil
		}
		return root, nil
	}
	// Fetch the children of the node
	lnode, rnode, isShortcut, err := s.loadChildren(root)
	if err != nil {
		return nil, err
	}
	if isShortcut == 1 {
		if bytes.Equal(lnode, key) {
			return rnode, nil
		}
		return nil, nil
	}
	if bitIsSet(key, s.TrieHeight-height) {
		return s.get(rnode, key, height-1)
	}
	return s.get(lnode, key, height-1)
}

// DefaultHash is a getter for the defaultHashes array
func (s *SMT) DefaultHash(height uint64) []byte {
	return s.defaultHashes[height]
}

// interiorHash hashes 2 children to get the parent hash and stores it in the updatedNodes and maybe in liveCache.
// the key is the hash and the value is the appended child nodes or the appended key/value in case of a shortcut.
// keys of go mappings cannot be byte slices so the hash is copied to a byte array
func (s *SMT) interiorHash(left, right []byte, height uint64, oldRoot []byte, shortcut, store bool, keys, values [][]byte) []byte {
	h := s.hash(left, right)
	var node Hash
	copy(node[:], h)
	if store {
		if !shortcut {
			children := make([]byte, 0, HashLength*2+1)
			children = append(children, left...)
			children = append(children, right...)
			children = append(children, byte(0))
			// Cache the node if it's children are not default and if it's height is over CacheHeightLimit
			if height > s.CacheHeightLimit {
				s.db.liveMux.Lock()
				s.db.liveCache[node] = children
				s.db.liveMux.Unlock()
			}
			// store new node in db
			s.db.updatedMux.Lock()
			s.db.updatedNodes[node] = children
			s.db.updatedMux.Unlock()

		} else {
			// shortcut is only true if len(keys)==1
			kv := make([]byte, 0, s.KeySize+HashLength+1)
			kv = append(kv, keys[0]...)
			kv = append(kv, values[0]...)
			kv = append(kv, byte(1))
			if !bytes.Equal(s.defaultHashes[height+1], h) &&
				height > s.CacheHeightLimit {
				// When deleting, the shortcut node for the newly default key should not be created.
				s.db.liveMux.Lock()
				s.db.liveCache[node] = kv
				s.db.liveMux.Unlock()
			}
			// store new node in db
			if !bytes.Equal(s.defaultHashes[height+1], h) {
				// When deleting, don't rewrite a default hash in db
				s.db.updatedMux.Lock()
				s.db.updatedNodes[node] = kv
				s.db.updatedMux.Unlock()
			}
		}
		if !bytes.Equal(s.defaultHashes[height+1], oldRoot) && !bytes.Equal(h, oldRoot) {
			// Delete old liveCache node if it has been updated and is not default
			var node Hash
			copy(node[:], oldRoot)
			s.db.liveMux.Lock()
			delete(s.db.liveCache, node)
			s.db.liveMux.Unlock()
			//NOTE this could delete a node used by another part of the tree if some values are equal.
		}
	}
	return h
}

// Commit stores the updated nodes to disk
func (s *SMT) Commit() error {
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
	s.db.updatedNodes = make(map[Hash][]byte, len(s.db.updatedNodes)*2)
	return nil
}
