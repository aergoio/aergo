/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/aergoio/aergo/pkg/db"
)

// TODO make a secure trie that hashes keys with a random seed to be sure the trie is sparse.

// Trie is a modified sparse Merkle tree.
// Instead of storing values at the leaves of the tree,
// the values are stored at the subtree root that contains only that value.
// If the tree is sparse, this requires fewer hashing operations.
// SMT optimises storage by creating shortcutnodes so subtree nodes wouldn't have to be stored
// Trie also optimises execution by not requiering subtrees to be hashed.
type Trie struct {
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
func NewTrie(keySize uint64, hash func(data ...[]byte) []byte, store db.DB) *Trie {
	s := &Trie{
		hash:             hash,
		TrieHeight:       keySize * 8,
		CacheHeightLimit: 233, // 246, //234, // based on the number of nodes we can keep in memory.
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
func (s *Trie) loadDefaultHashes() []byte {
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

// LoadCache loads the first layers of the merkle tree given a root
// This is called after a node restarts so that it doesnt become slow with db reads
func (s *Trie) LoadCache(root []byte) error {
	s.loadDefaultHashes()
	ch := make(chan error, 1)
	s.loadCache(root, s.TrieHeight, ch)
	return <-ch
}

// loadCache loads the first layers of the merkle tree given a root
func (s *Trie) loadCache(root []byte, height uint64, ch chan<- (error)) {
	if height <= s.CacheHeightLimit+1 || bytes.Equal(root, s.defaultHashes[height]) {
		ch <- nil
		return
	}
	// Load the node from db
	s.db.lock.Lock()
	val := s.db.store.Get(root)
	s.db.lock.Unlock()
	nodeSize := len(val)
	if nodeSize == 0 {
		ch <- fmt.Errorf("the trie node %x is unavailable in the disk db, db may be corrupted", root)
		return
	}
	//Store node in cache.
	var node Hash
	copy(node[:], root)
	s.db.liveMux.Lock()
	s.db.liveCache[node] = val
	s.db.liveMux.Unlock()
	isShortcut := val[nodeSize-1]
	if isShortcut == 1 {
		ch <- nil
	} else {
		// Load subtree
		lnode, rnode := val[:HashLength], val[HashLength:nodeSize-1]

		lch := make(chan error, 1)
		rch := make(chan error, 1)
		go s.loadCache(lnode, height-1, lch)
		go s.loadCache(rnode, height-1, rch)
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
	s.update(s.Root, keys, values, s.TrieHeight, ch)
	result := <-ch
	if result.err != nil {
		return nil, result.err
	}
	s.Root = result.update
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
// A DefaultLeaf cannot be updated to a DefaultLeaf as shortcut nodes
// could be moved down the tree resulting in an invalid root
func (s *Trie) update(root []byte, keys, values [][]byte, height uint64, ch chan<- (mresult)) {
	if height == 0 {
		if bytes.Equal(DefaultLeaf, values[0]) {
			// Delete the key-value from the trie if it is being set to DefaultLeaf
			if !bytes.Equal(DefaultLeaf, root) {
				// Delete old liveCache node if it is not default
				s.deleteCacheNode(root)
			}
			ch <- mresult{DefaultLeaf, true, nil}
		} else {
			// Set the value
			node := s.leafHash(keys[0], values[0], height-1, root)
			ch <- mresult{node, false, nil}
		}
		return
	}

	// Load the node to update
	lnode, rnode, isShortcut, err := s.loadChildren(root)
	if err != nil {
		ch <- mresult{nil, false, err}
		return
	}

	// Check if the keys are updating the shortcut node
	if isShortcut == 1 {
		keys, values = s.maybeAddShortcutToKV(keys, values, lnode, rnode)
		if len(keys) == 0 {
			// The shortcut is being deleted
			s.deleteCacheNode(root)
			ch <- mresult{s.defaultHashes[height], true, nil}
			return
		}
		// The shortcut node was added to keys and values so consider this subtree default.
		lnode, rnode = s.defaultHashes[height-1], s.defaultHashes[height-1]
	}
	// Store shortcut node
	if bytes.Equal(s.defaultHashes[height-1], lnode) &&
		bytes.Equal(s.defaultHashes[height-1], rnode) && (len(keys) == 1) {
		// We are adding 1 key to an empty subtree so store it as a shortcut
		node := s.leafHash(keys[0], values[0], height-1, root)
		ch <- mresult{node, false, nil}
		return
	}

	// Split the keys array so each branch can be updated in parallel
	lkeys, rkeys := s.splitKeys(keys, s.TrieHeight-height)
	splitIndex := len(lkeys)
	lvalues, rvalues := values[:splitIndex], values[splitIndex:]

	switch {
	case len(lkeys) == 0 && len(rkeys) > 0:
		s.updateRight(lnode, rnode, root, keys, values, height, ch)
	case len(lkeys) > 0 && len(rkeys) == 0:
		s.updateLeft(lnode, rnode, root, keys, values, height, ch)
	default:
		s.updateParallel(lnode, rnode, root, lkeys, rkeys, lvalues, rvalues, height, ch)
	}
}

// updateRight updates the right side of the tree
func (s *Trie) updateRight(lnode, rnode, root []byte, keys, values [][]byte, height uint64, ch chan<- (mresult)) {
	// all the keys go in the right subtree
	newch := make(chan mresult, 1)
	s.update(rnode, keys, values, height-1, newch)
	result := <-newch
	if result.err != nil {
		ch <- mresult{nil, false, result.err}
		return
	}
	// Move up a shortcut node if necessary.
	if result.deleted {
		if s.maybeMoveUpShortcut(result.update, lnode, root, height, ch) {
			return
		}
	}
	node := s.interiorHash(lnode, result.update, height-1, root)
	ch <- mresult{node, false, nil}
}

// updateLeft updates the left side of the tree
func (s *Trie) updateLeft(lnode, rnode, root []byte, keys, values [][]byte, height uint64, ch chan<- (mresult)) {
	// all the keys go in the left subtree
	newch := make(chan mresult, 1)
	s.update(lnode, keys, values, height-1, newch)
	result := <-newch
	if result.err != nil {
		ch <- mresult{nil, false, result.err}
		return
	}
	// Move up a shortcut node if necessary.
	if result.deleted {
		if s.maybeMoveUpShortcut(result.update, rnode, root, height, ch) {
			return
		}
	}
	node := s.interiorHash(result.update, rnode, height-1, root)
	ch <- mresult{node, false, nil}
}

// updateParallel updates both sides of the trie simultaneously
func (s *Trie) updateParallel(lnode, rnode, root []byte, lkeys, rkeys, lvalues, rvalues [][]byte, height uint64, ch chan<- (mresult)) {
	lch := make(chan mresult, 1)
	rch := make(chan mresult, 1)
	go s.update(lnode, lkeys, lvalues, height-1, lch)
	go s.update(rnode, rkeys, rvalues, height-1, rch)
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
	if lresult.deleted && rresult.deleted {
		// root is never default when moving up a shortcut
		s.deleteCacheNode(root)
		if bytes.Equal(s.defaultHashes[height-1], lresult.update) && bytes.Equal(s.defaultHashes[height-1], rresult.update) {
			ch <- mresult{s.defaultHashes[height], false, nil}
			return
		}
		// Move up one of the shortcut nodes
		if bytes.Equal(s.defaultHashes[height-1], lresult.update) {
			ch <- mresult{rresult.update, true, nil}
		} else if bytes.Equal(s.defaultHashes[height-1], rresult.update) {
			ch <- mresult{lresult.update, true, nil}
		}
	}
	node := s.interiorHash(lresult.update, rresult.update, height-1, root)
	ch <- mresult{node, false, nil}
}

// deleteCacheNode deletes the node from liveCache
func (s *Trie) deleteCacheNode(root []byte) {
	var node Hash
	copy(node[:], root)
	s.db.liveMux.Lock()
	delete(s.db.liveCache, node)
	s.db.liveMux.Unlock()
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

// maybeMoveUpShortcut moves up a shortcut after a deletion if it is no more at
// the highest root of an empty subtree.
func (s *Trie) maybeMoveUpShortcut(update, sibling, root []byte, height uint64, ch chan<- (mresult)) bool {
	if bytes.Equal(s.defaultHashes[height-1], update) {
		// If update deleted a subtree, check it's sibling and
		// return it if it is a shortcut
		_, _, isShortcut, err := s.loadChildren(sibling)
		if err != nil {
			ch <- mresult{nil, false, err}
			return true
		}
		if isShortcut == 1 {
			// root is never default when moving up a shortcut
			s.deleteCacheNode(root)
			// Return the left sibling node to move it up
			ch <- mresult{sibling, true, nil}
			return true
		}
	} else if bytes.Equal(s.defaultHashes[height-1], sibling) {
		// If update deleted something and returned a shortcut, return that
		// shortcut if the sibling is default
		// root is never default when moving up a shortcut
		s.deleteCacheNode(root)
		// Return the shortcut node to move it up
		ch <- mresult{update, true, nil}
		return true
	}
	return false
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
func (s *Trie) loadChildren(root []byte) ([]byte, []byte, byte, error) {
	var node Hash
	copy(node[:], root)

	s.db.liveMux.RLock()
	val, exists := s.db.liveCache[node]
	s.db.liveMux.RUnlock()
	if exists {
		s.liveCountMux.Lock()
		s.LoadCacheCounter++
		s.liveCountMux.Unlock()
		return s.parseValue(val, len(val))
	}

	// checking updated nodes is useful if get() or update() is called twice in a row without db commit
	s.db.updatedMux.RLock()
	val, exists = s.db.updatedNodes[node]
	s.db.updatedMux.RUnlock()
	if exists {
		return s.parseValue(val, len(val))
	}
	//Fetch node in disk database
	s.loadDbMux.Lock()
	s.LoadDbCounter++
	s.loadDbMux.Unlock()
	s.db.lock.Lock()
	val = s.db.store.Get(root)
	s.db.lock.Unlock()
	nodeSize := len(val)
	if nodeSize != 0 {
		return s.parseValue(val, nodeSize)
	}
	return nil, nil, byte(0), fmt.Errorf("the trie node %x is unavailable in the disk db, db may be corrupted", root)
}

// parseValue returns a subtree roots or a shortcut node
func (s *Trie) parseValue(val []byte, nodeSize int) ([]byte, []byte, byte, error) {
	shortcut := val[nodeSize-1]
	if shortcut == 1 {
		return val[:s.KeySize], val[s.KeySize : nodeSize-1], shortcut, nil
	}
	return val[:HashLength], val[HashLength : nodeSize-1], shortcut, nil
}

// Get fetches the value of a key by going down the current trie root.
func (s *Trie) Get(key []byte) ([]byte, error) {
	return s.get(s.Root, key, s.TrieHeight)
}

// get fetches the value of a key given a trie root
func (s *Trie) get(root []byte, key []byte, height uint64) ([]byte, error) {
	if bytes.Equal(root, s.defaultHashes[height]) {
		// the trie does not contain the key
		return nil, nil
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
func (s *Trie) DefaultHash(height uint64) []byte {
	return s.defaultHashes[height]
}

// leafHash returns the hash of key-value concatenated stores it in the updatedNodes and maybe in liveCache.
// keys of go mappings cannot be byte slices so the hash is copied to a byte array
func (s *Trie) leafHash(key, value []byte, height uint64, oldRoot []byte) []byte {
	h := s.hash(key, value, []byte{1})
	var node Hash
	copy(node[:], h)
	kv := make([]byte, 0, s.KeySize+HashLength+1)
	kv = append(kv, key...)
	kv = append(kv, value...)
	kv = append(kv, byte(1))
	// Cache the shortcut node if it's height is over CacheHeightLimit
	if height > s.CacheHeightLimit {
		s.db.liveMux.Lock()
		s.db.liveCache[node] = kv
		s.db.liveMux.Unlock()
	}
	// store new node in db
	s.db.updatedMux.Lock()
	s.db.updatedNodes[node] = kv
	s.db.updatedMux.Unlock()
	if !bytes.Equal(s.defaultHashes[height+1], oldRoot) && !bytes.Equal(h, oldRoot) {
		// Delete old liveCache node if it has been updated and is not default
		s.deleteCacheNode(oldRoot)
	}
	return h
}

// interiorHash hashes 2 children to get the parent hash and stores it in the updatedNodes and maybe in liveCache.
// the key is the hash and the value is the appended child nodes.
// keys of go mappings cannot be byte slices so the hash is copied to a byte array
func (s *Trie) interiorHash(left, right []byte, height uint64, oldRoot []byte) []byte {
	h := s.hash(left, right)
	var node Hash
	copy(node[:], h)
	children := make([]byte, 0, HashLength*2+1)
	children = append(children, left...)
	children = append(children, right...)
	children = append(children, byte(0))
	// TODO test if it is possible to use a caching stratergy instead of a fixed CacheHeightLimit
	// a caching stratergy also requires modifying loadCache()
	// stratergy : cache if shortcut or both children are not default
	// if !bytes.Equal(s.defaultHashes[height], left) && !bytes.Equal(s.defaultHashes[height], right)) {
	if height > s.CacheHeightLimit {
		s.db.liveMux.Lock()
		s.db.liveCache[node] = children
		s.db.liveMux.Unlock()
	}
	// store new node in db
	s.db.updatedMux.Lock()
	s.db.updatedNodes[node] = children
	s.db.updatedMux.Unlock()

	if !bytes.Equal(s.defaultHashes[height+1], oldRoot) && !bytes.Equal(h, oldRoot) {
		// Delete old liveCache node if it has been updated and is not default
		s.deleteCacheNode(oldRoot)
	}
	return h
}

// Commit stores the updated nodes to disk
func (s *Trie) Commit() {
	// Commit the new nodes to database, clear updatedNodes and store the Root in history for reverts.
	if len(s.pastTries) >= maxPastTries {
		copy(s.pastTries, s.pastTries[1:])
		s.pastTries[len(s.pastTries)-1] = s.Root
	} else {
		s.pastTries = append(s.pastTries, s.Root)
	}
	s.db.commit()
	s.db.updatedNodes = make(map[Hash][]byte, len(s.db.updatedNodes)*2)
}
