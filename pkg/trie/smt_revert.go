/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"fmt"
)

// Revert rewinds the state tree to a previous version
// All the nodes (subtree roots and values) reverted are deleted from the database.
func (s *SMT) Revert(toOldRoot []byte) error {
	if bytes.Equal(s.Root, toOldRoot) {
		return fmt.Errorf("Trying to revers to the same root %x", s.Root)
	}
	//check if toOldRoot is in s.pastTries
	canRevert := false
	toIndex := 0
	for i, r := range s.pastTries {
		if bytes.Equal(r, toOldRoot) {
			canRevert = true
			toIndex = i
			break
		}
	}
	if !canRevert {
		return fmt.Errorf("The root is not contained in the cached tries, too old to be reverted : %x", s.Root)
	}

	// For every node of toOldRoot, compare it to the equivalent node in other pasttries between toOldRoot and current s.Root. If a node is different, delete the one from pasttries
	toBeDeleted := make([][]byte, 0, 1e3)
	for i := toIndex + 1; i < len(s.pastTries); i++ {
		err := s.maybeDeleteSubTree(toOldRoot, s.pastTries[i], s.TrieHeight, &toBeDeleted)
		if err != nil {
			return err
		}
	}
	// NOTE The tx interface doesnt handle ErrTxnTooBig
	txn := s.db.store.NewTx(true)
	for _, key := range toBeDeleted {
		txn.Delete(key)
	}
	txn.Commit()

	s.pastTries = s.pastTries[:toIndex+1]
	s.Root = toOldRoot
	// load default hashes in live cache
	s.db.liveCache = make(map[Hash][]byte)
	s.loadDefaultHashes()
	return nil
}

// maybeDeleteSubTree compares the subtree nodes of 2 tries and keeps only the older one
func (s *SMT) maybeDeleteSubTree(original []byte, maybeDelete []byte, height uint64, toBeDeleted *[][]byte) error {
	if bytes.Equal(original, maybeDelete) {
		return nil
	}
	if height == 0 {
		*toBeDeleted = append(*toBeDeleted, maybeDelete)
		return nil
	}

	lnode, rnode, isShortcut, lerr := s.loadChildren(original)
	if lerr != nil {
		return lerr
	}
	lnode2, rnode2, isShortcut2, rerr := s.loadChildren(maybeDelete)
	if rerr != nil {
		return rerr
	}

	if isShortcut != isShortcut2 {
		if isShortcut == 1 {
			return s.deleteSubTree(maybeDelete, height, toBeDeleted)
		}
	} else {
		if isShortcut == 1 {
			// Delete shortcut if not equal
			if !bytes.Equal(lnode, lnode2) || !bytes.Equal(rnode, rnode2) {
				*toBeDeleted = append(*toBeDeleted, maybeDelete)
			}
		} else {
			// Delete subtree if not equal
			*toBeDeleted = append(*toBeDeleted, maybeDelete)
			err := s.maybeDeleteSubTree(lnode, lnode2, height-1, toBeDeleted)
			if err != nil {
				return err
			}
			err = s.maybeDeleteSubTree(rnode, rnode2, height-1, toBeDeleted)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// deleteSubTree deletes all the nodes contained in a tree
func (s *SMT) deleteSubTree(root []byte, height uint64, toBeDeleted *[][]byte) error {
	if height == 0 {
		return nil
	}
	lnode, rnode, isShortcut, err := s.loadChildren(root)
	if err != nil {
		return err
	}
	if isShortcut == 0 {
		lerr := s.deleteSubTree(lnode, height-1, toBeDeleted)
		if lerr != nil {
			return lerr
		}
		rerr := s.deleteSubTree(rnode, height-1, toBeDeleted)
		if rerr != nil {
			return rerr
		}
	}
	*toBeDeleted = append(*toBeDeleted, root)
	return nil
}
