/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"fmt"

	"github.com/aergoio/aergo/v2/types/dbkey"
)

// Revert rewinds the state tree to a previous version
// All the nodes (subtree roots and values) reverted are deleted from the database.
func (s *Trie) Revert(toOldRoot []byte) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// safety precaution if reverting to a shortcut batch that might have been deleted
	s.atomicUpdate = false // so loadChildren doesnt return a copy
	batch, _, _, _, isShortcut, err := s.loadChildren(toOldRoot, s.TrieHeight, 0, nil)
	if err != nil {
		return err
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
	if !canRevert || bytes.Equal(s.Root, toOldRoot) {
		return fmt.Errorf("The root cannot be reverted, because already latest of not in pastTries : current : %x, target : %x", s.Root, toOldRoot)
	}

	// For every node of toOldRoot, compare it to the equivalent node in other pasttries between toOldRoot and current s.Root. If a node is different, delete the one from pasttries
	s.db.nodesToRevert = make([][]byte, 0)
	for i := toIndex + 1; i < len(s.pastTries); i++ {
		ch := make(chan error, 1)
		s.maybeDeleteSubTree(toOldRoot, s.pastTries[i], s.TrieHeight, 0, nil, nil, ch)
		err := <-ch
		if err != nil {
			return err
		}
	}
	// NOTE The tx interface doesnt handle ErrTxnTooBig
	txn := s.db.Store.NewTx()
	for _, key := range s.db.nodesToRevert {
		txn.Delete(dbkey.Trie(key[:HashLength]))
	}
	txn.Commit()

	s.pastTries = s.pastTries[:toIndex+1]
	s.Root = toOldRoot
	s.db.liveCache = make(map[Hash][][]byte)
	s.db.updatedNodes = make(map[Hash][][]byte)
	if isShortcut {
		// If toOldRoot is a shortcut batch, it is possible that
		// revert has deleted it if the key was ever stored at height0
		// because in leafHash byte(0) = byte(256)
		s.db.Store.Set(dbkey.Trie(toOldRoot), s.db.serializeBatch(batch))
	}
	return nil
}

// maybeDeleteSubTree compares the subtree nodes of 2 tries and keeps only the older one
func (s *Trie) maybeDeleteSubTree(original, maybeDelete []byte, height, iBatch int, batch, batch2 [][]byte, ch chan<- (error)) {
	if height == 0 {
		if !bytes.Equal(original, maybeDelete) && len(maybeDelete) != 0 {
			s.maybeDeleteRevertedNode(maybeDelete, 0)
		}
		ch <- nil
		return
	}
	if bytes.Equal(original, maybeDelete) || len(maybeDelete) == 0 {
		ch <- nil
		return
	}
	// if this point os reached, then the root of the batch is same
	// so the batch is also same.
	batch, iBatch, lnode, rnode, isShortcut, lerr := s.loadChildren(original, height, iBatch, batch)
	if lerr != nil {
		ch <- lerr
		return
	}
	batch2, _, lnode2, rnode2, isShortcut2, rerr := s.loadChildren(maybeDelete, height, iBatch, batch2)
	if rerr != nil {
		ch <- rerr
		return
	}

	if isShortcut != isShortcut2 {
		if isShortcut {
			ch1 := make(chan error, 1)
			s.deleteSubTree(maybeDelete, height, iBatch, batch2, ch1)
			err := <-ch1
			if err != nil {
				ch <- err
				return
			}
		} else {
			s.maybeDeleteRevertedNode(maybeDelete, iBatch)
		}
	} else {
		if isShortcut {
			// Delete shortcut if not equal
			if !bytes.Equal(lnode, lnode2) || !bytes.Equal(rnode, rnode2) {
				s.maybeDeleteRevertedNode(maybeDelete, iBatch)
			}
		} else {
			// Delete subtree if not equal
			s.maybeDeleteRevertedNode(maybeDelete, iBatch)
			ch1 := make(chan error, 1)
			ch2 := make(chan error, 1)
			go s.maybeDeleteSubTree(lnode, lnode2, height-1, 2*iBatch+1, batch, batch2, ch1)
			go s.maybeDeleteSubTree(rnode, rnode2, height-1, 2*iBatch+2, batch, batch2, ch2)
			err1, err2 := <-ch1, <-ch2
			if err1 != nil {
				ch <- err1
				return
			}
			if err2 != nil {
				ch <- err2
				return
			}
		}
	}
	ch <- nil
}

// deleteSubTree deletes all the nodes contained in a tree
func (s *Trie) deleteSubTree(root []byte, height, iBatch int, batch [][]byte, ch chan<- (error)) {
	if len(root) == 0 || height == 0 {
		if height == 0 {
			s.maybeDeleteRevertedNode(root, 0)
		}
		ch <- nil
		return
	}
	batch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(root, height, iBatch, batch)
	if err != nil {
		ch <- err
		return
	}
	if !isShortcut {
		ch1 := make(chan error, 1)
		ch2 := make(chan error, 1)
		go s.deleteSubTree(lnode, height-1, 2*iBatch+1, batch, ch1)
		go s.deleteSubTree(rnode, height-1, 2*iBatch+2, batch, ch2)
		lerr := <-ch1
		rerr := <-ch2

		if lerr != nil {
			ch <- lerr
			return
		}
		if rerr != nil {
			ch <- rerr
			return
		}
	}
	s.maybeDeleteRevertedNode(root, iBatch)
	ch <- nil
}

// maybeDeleteRevertedNode adds the node to updatedNodes to be reverted
// if it is a batch node at height%4 == 0
func (s *Trie) maybeDeleteRevertedNode(root []byte, iBatch int) {
	if iBatch == 0 {
		s.db.revertMux.Lock()
		s.db.nodesToRevert = append(s.db.nodesToRevert, root)
		s.db.revertMux.Unlock()
	}
}
