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
func (s *Trie) Revert(toOldRoot []byte) error {
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
	toBeDeleted := make([][]byte, 0)
	for i := toIndex + 1; i < len(s.pastTries); i++ {
		err := s.maybeDeleteSubTree(toOldRoot, s.pastTries[i], s.TrieHeight, &toBeDeleted, nil, 0)
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
	s.db.liveCache = make(map[Hash][][]byte)
	s.loadDefaultHashes()
	return nil
}

// maybeDeleteSubTree compares the subtree nodes of 2 tries and keeps only the older one
func (s *Trie) maybeDeleteSubTree(original []byte, maybeDelete []byte, height uint64, toBeDeleted *[][]byte, batch [][]byte, iBatch uint8) error {
	if height == 0 {
		if !bytes.Equal(original, maybeDelete) && len(maybeDelete) != 0 {
			*toBeDeleted = append(*toBeDeleted, maybeDelete)
		}
		return nil
	}
	if bytes.Equal(original, maybeDelete) || len(maybeDelete) == 0 {
		return nil
	}
	// if this point os reached, then the root of the batch is same
	// so the batch is also same.
	_, _, lnode, rnode, isShortcut, lerr := s.loadChildren(original, height, batch, iBatch)
	if lerr != nil {
		return lerr
	}
	batch, iBatch, lnode2, rnode2, isShortcut2, rerr := s.loadChildren(maybeDelete, height, batch, iBatch)
	if rerr != nil {
		return rerr
	}

	if isShortcut != isShortcut2 {
		if isShortcut {
			return s.deleteSubTree(maybeDelete, height, toBeDeleted, batch, iBatch)
		} else if iBatch == 0 {
			*toBeDeleted = append(*toBeDeleted, maybeDelete)
		}
	} else {
		if isShortcut {
			// Delete shortcut if not equal
			if !bytes.Equal(lnode, lnode2) || !bytes.Equal(rnode, rnode2) {
				if iBatch == 0 {
					*toBeDeleted = append(*toBeDeleted, maybeDelete)
				}
			}
		} else {
			// Delete subtree if not equal
			if iBatch == 0 {
				*toBeDeleted = append(*toBeDeleted, maybeDelete)
			}
			err := s.maybeDeleteSubTree(lnode, lnode2, height-1, toBeDeleted, batch, 2*iBatch+1)
			if err != nil {
				return err
			}
			err = s.maybeDeleteSubTree(rnode, rnode2, height-1, toBeDeleted, batch, 2*iBatch+2)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// deleteSubTree deletes all the nodes contained in a tree
func (s *Trie) deleteSubTree(root []byte, height uint64, toBeDeleted *[][]byte, batch [][]byte, iBatch uint8) error {
	if height == 0 || len(root) == 0 {
		return nil
	}
	batch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(root, height, batch, iBatch)
	if err != nil {
		return err
	}
	if !isShortcut {
		lerr := s.deleteSubTree(lnode, height-1, toBeDeleted, batch, 2*iBatch+1)
		if lerr != nil {
			return lerr
		}
		rerr := s.deleteSubTree(rnode, height-1, toBeDeleted, batch, 2*iBatch+2)
		if rerr != nil {
			return rerr
		}
	}
	if iBatch == 0 {
		*toBeDeleted = append(*toBeDeleted, root)
	}
	return nil
}
