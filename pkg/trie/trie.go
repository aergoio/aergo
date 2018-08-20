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

// Update adds a sorted list of keys and their values to the trie
func (s *Trie) Update(keys, values DataArray) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.LoadDbCounter = 0
	s.LoadCacheCounter = 0
	update, _, err := s.update(s.Root, keys, values, s.TrieHeight, nil)
	if err != nil {
		return nil, err
	}
	s.Root = update
	return s.Root, err
}

// mresult is used to contain the result of goroutines and is sent through a channel.
type mresult struct {
	update []byte
	// flag if a node was deleted and a shortcut node maybe has to move up the tree
	deleted bool
	err     error
}

// update adds a sorted list of keys and their values to the trie.
// It returns the root of the updated tree.
func (s *Trie) update(root []byte, keys, values DataArray, height uint64, ch chan<- (mresult)) ([]byte, bool, error) {
	if height == 0 {
		// Delete the key-value from the trie if it is being set to DefaultLeaf
		if bytes.Equal(DefaultLeaf, values[0]) {
			if !bytes.Equal(DefaultLeaf, root) {
				// Delete old liveCache node if it is not default
				s.deleteCacheNode(root)
			}
			if ch != nil {
				ch <- mresult{DefaultLeaf, true, nil}
			}
			return DefaultLeaf, true, nil
		}
		node := s.leafHash(keys[0], values[0], height-1, root)
		if ch != nil {
			ch <- mresult{node, false, nil}
		}
		return node, false, nil
	}

	// Load the node to update
	lnode, rnode, isShortcut, err := s.loadChildren(root)
	if err != nil {
		if ch != nil {
			ch <- mresult{nil, false, err}
		}
		return nil, false, err
	}

	// Check if the keys are updating the shortcut node
	if isShortcut == 1 {
		s.maybeAddShortcutToDataArray(keys, values, lnode, rnode)
		// The shortcut node was added to keys and values so consider this subtree default.
		lnode, rnode = s.defaultHashes[height-1], s.defaultHashes[height-1]
	}

	// Store shortcut node
	if bytes.Equal(s.defaultHashes[height-1], lnode) &&
		bytes.Equal(s.defaultHashes[height-1], rnode) && (len(keys) == 1) {
		// Delete the key-value from the trie if it is being set to DefaultLeaf
		if bytes.Equal(DefaultLeaf, values[0]) {
			if !bytes.Equal(s.defaultHashes[height], root) {
				// Delete old liveCache node if it is not default
				s.deleteCacheNode(root)
			}
			if ch != nil {
				ch <- mresult{s.defaultHashes[height], true, nil}
			}
			return s.defaultHashes[height], true, nil
		}
		// We are adding 1 key to an empty subtree so store it as a shortcut
		node := s.leafHash(keys[0], values[0], height-1, root)
		if ch != nil {
			ch <- mresult{node, false, nil}
		}
		return node, false, nil
	}

	// Split the keys array so each branch can be updated in parallel
	lkeys, rkeys := s.splitKeys(keys, s.TrieHeight-height)
	splitIndex := len(lkeys)
	lvalues, rvalues := values[:splitIndex], values[splitIndex:]

	switch {
	case lkeys.Len() == 0 && rkeys.Len() > 0:
		// all the keys go in the right subtree
		update, deleted, err := s.update(rnode, keys, values, height-1, nil)
		if err != nil {
			if ch != nil {
				ch <- mresult{nil, false, err}
			}
			return nil, false, err
		}
		// Move up a shortcut node if necessary.
		if deleted {
			// If update deleted a subtree, check it's sibling return it if it is a shortcut
			if bytes.Equal(s.defaultHashes[height-1], update) {
				_, _, isShortcut, err := s.loadChildren(lnode)
				if err != nil {
					if ch != nil {
						ch <- mresult{nil, false, err}
					}
					return nil, false, err
				}
				if isShortcut == 1 {
					// root is never default when moving up a shortcut
					s.deleteCacheNode(root)
					// Return the left sibling node to move it up
					if ch != nil {
						ch <- mresult{lnode, true, nil}
					}
					return lnode, true, nil
				}
			}
			// If deleted then update is a shortcut node (because not default),
			// return it if the sibling is default.
			if bytes.Equal(s.defaultHashes[height-1], lnode) {
				// root is never default when moving up a shortcut
				s.deleteCacheNode(root)
				// Return the shortcut node to move it up
				if ch != nil {
					ch <- mresult{update, true, nil}
				}
				return update, true, nil
			}
		}
		node := s.interiorHash(lnode, update, height-1, root)
		if ch != nil {
			ch <- mresult{node, false, nil}
		}
		return node, false, nil
	case lkeys.Len() > 0 && rkeys.Len() == 0:
		// all the keys go in the left subtree
		update, deleted, err := s.update(lnode, keys, values, height-1, nil)
		if err != nil {
			if ch != nil {
				ch <- mresult{nil, false, err}
			}
			return nil, false, err
		}
		// Move up a shortcut node if necessary.
		if deleted {
			// If update deleted a subtree, check it's sibling return it if it is a shortcut
			if bytes.Equal(s.defaultHashes[height-1], update) {
				_, _, isShortcut, err := s.loadChildren(rnode)
				if err != nil {
					if ch != nil {
						ch <- mresult{nil, false, err}
					}
					return nil, false, err
				}
				if isShortcut == 1 {
					// root is never default when moving up a shortcut
					s.deleteCacheNode(root)
					// Return the right sibling node to move it up
					if ch != nil {
						ch <- mresult{rnode, true, nil}
					}
					return rnode, true, nil
				}
			}
			// If deleted then update is a shortcut node (because not default),
			// return it if the sibling is default.
			if bytes.Equal(s.defaultHashes[height-1], rnode) {
				// root is never default when moving up a shortcut
				s.deleteCacheNode(root)
				// Return the shortcut node to move it up
				if ch != nil {
					ch <- mresult{update, true, nil}
				}
				return update, true, nil
			}
		}
		node := s.interiorHash(update, rnode, height-1, root)
		if ch != nil {
			ch <- mresult{node, false, nil}
		}
		return node, false, nil
	default:
		// keys are separated between the left and right branches
		// update the branches in parallel
		lch := make(chan mresult, 1)
		rch := make(chan mresult, 1)
		go s.update(lnode, lkeys, lvalues, height-1, lch)
		go s.update(rnode, rkeys, rvalues, height-1, rch)
		lresult := <-lch
		rresult := <-rch
		if lresult.err != nil {
			if ch != nil {
				ch <- mresult{nil, false, lresult.err}
			}
			return nil, false, lresult.err
		}
		if rresult.err != nil {
			if ch != nil {
				ch <- mresult{nil, false, rresult.err}
			}
			return nil, false, rresult.err
		}

		// Move up a shortcut node if it's sibling is default
		if lresult.deleted && rresult.deleted {
			// root is never default when moving up a shortcut
			s.deleteCacheNode(root)
			if bytes.Equal(s.defaultHashes[height-1], lresult.update) && bytes.Equal(s.defaultHashes[height-1], rresult.update) {
				if ch != nil {
					ch <- mresult{s.defaultHashes[height], false, nil}
				}
				return s.defaultHashes[height], false, nil
			}
			if bytes.Equal(s.defaultHashes[height-1], lresult.update) {
				if ch != nil {
					ch <- mresult{rresult.update, true, nil}
				}
				return rresult.update, true, nil
			}
			if bytes.Equal(s.defaultHashes[height-1], rresult.update) {
				if ch != nil {
					ch <- mresult{lresult.update, true, nil}
				}
				return lresult.update, true, nil
			}
		}
		node := s.interiorHash(lresult.update, rresult.update, height-1, root)
		if ch != nil {
			ch <- mresult{node, false, nil}
		}
		return node, false, nil
	}
}

func (s *Trie) deleteCacheNode(root []byte) {
	var node Hash
	copy(node[:], root)
	s.db.liveMux.Lock()
	delete(s.db.liveCache, node)
	s.db.liveMux.Unlock()
}

// splitKeys devides the array of keys into 2 so they can update left and right branches in parallel
func (s *Trie) splitKeys(keys DataArray, height uint64) (DataArray, DataArray) {
	for i, key := range keys {
		if bitIsSet(key, height) {
			return keys[:i], keys[i:]
		}
	}
	return keys, nil
}

func (s *Trie) maybeAddShortcutToDataArray(keys, values DataArray, shortcutKey, shortcutVal []byte) (DataArray, DataArray) {
	up := false
	for _, k := range keys {
		if bytes.Equal(k, shortcutKey) {
			up = true
			break
		}
	}
	if !up {
		keys, values = s.addShortcutToDataArray(keys, values, shortcutKey, shortcutVal)
	}
	return keys, values
}

// addShortcutToDataArray adds a shortcut key to the keys array to be updated.
// this is used when a subtree containing a shortcut node is being updated
func (s *Trie) addShortcutToDataArray(keys, values DataArray, shortcutKey, shortcutVal []byte) (DataArray, DataArray) {
	newKeys := make(DataArray, 0, len(keys)+1)
	newVals := make(DataArray, 0, len(keys)+1)

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
				return newKeys, newVals
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
		nodeSize := len(val)
		shortcut := val[nodeSize-1]
		if shortcut == 1 {
			return val[:s.KeySize], val[s.KeySize : nodeSize-1], shortcut, nil
		}
		return val[:HashLength], val[HashLength : nodeSize-1], shortcut, nil
	}

	// checking updated nodes is useful if get() or update() is called twice in a row without db commit
	s.db.updatedMux.RLock()
	val, exists = s.db.updatedNodes[node]
	s.db.updatedMux.RUnlock()
	if exists {
		nodeSize := len(val)
		shortcut := val[nodeSize-1]
		if shortcut == 1 {
			return val[:s.KeySize], val[s.KeySize : nodeSize-1], shortcut, nil
		}
		return val[:HashLength], val[HashLength : nodeSize-1], shortcut, nil
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
		shortcut := val[nodeSize-1]
		if shortcut == 1 {
			return val[:s.KeySize], val[s.KeySize : nodeSize-1], shortcut, nil
		}
		return val[:HashLength], val[HashLength : nodeSize-1], shortcut, nil
	}
	return nil, nil, byte(0), fmt.Errorf("the trie node %x is unavailable in the disk db, db may be corrupted", root)
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
	if height == 0 {
		k, v, isShortcut, err := s.loadChildren(root)
		if err != nil {
			return nil, err
		}
		if isShortcut == 1 && bytes.Equal(k, key) {
			return v, nil
		}
		return nil, fmt.Errorf("the trie leaf node %x did not contain a key-value pair", root)
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
	// Cache the node if it's children are not default and if it's height is over CacheHeightLimit
	if (!bytes.Equal(s.defaultHashes[height], left) ||
		!bytes.Equal(s.defaultHashes[height], right)) &&
		height > s.CacheHeightLimit {
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
