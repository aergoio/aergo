/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
)

// MerkleProof creates a merkle proof for a key in the latest trie
// A non inclusion proof is a proof to a default value
func (s *SMT) MerkleProof(key []byte) ([][]byte, error) {
	return s.merkleProof(s.Root, s.TrieHeight, key)
}

// MerkleProofCompressed returns a compressed merkle proof.
// The proof contains a bitmap of non default hashes and the non default hashes.
func (s *SMT) MerkleProofCompressed(key []byte) ([]byte, [][]byte, error) {
	bitmap := make([]byte, s.KeySize)
	mp, err := s.merkleProofCompressed(s.Root, s.TrieHeight, key, bitmap)
	return bitmap, mp, err
}

// MerkleProofCompressed2 returns a compressed merkle proof like MerkleProofCompressed
// This version 1st calls MerkleProof and then removes the default nodes.
func (s *SMT) MerkleProofCompressed2(key []byte) ([]byte, [][]byte, error) {
	// create a regular merkle proof and then compress it
	mpFull, err := s.merkleProof(s.Root, s.TrieHeight, key)
	if err != nil {
		return nil, nil, err
	}
	var mp [][]byte
	bitmap := make([]byte, s.KeySize)
	for i, node := range mpFull {
		if !bytes.Equal(node, s.defaultHashes[i]) {
			bitSet(bitmap, uint64(i))
			mp = append(mp, node)
		}
	}
	return bitmap, mp, nil
}

// merkleProof generates a Merke proof of inclusion or non inclusion for a given trie root
func (s *SMT) merkleProof(root []byte, height uint64, key []byte) ([][]byte, error) {
	if height == 0 {
		return nil, nil
	}
	// Fetch the children of the node
	lnode, rnode, isShortcut, err := s.loadChildren(root)
	if err != nil {
		return nil, err
	}
	if isShortcut == 1 {
		// append all default hashes down the tree
		if bytes.Equal(lnode, key) {
			rest := make([][]byte, height)
			copy(rest, s.defaultHashes[:height]) // needed because append will modify underlying array
			return rest, nil
		}
		// if the key is empty, unroll until it diverges from the shortcut key and add the non default node
		return s.unrollShortcutAndKey(lnode, rnode, height, key), nil
	}

	// append the left or right node to the proof
	if bitIsSet(key, s.TrieHeight-height) {
		mp, err := s.merkleProof(rnode, height-1, key)
		if err != nil {
			return nil, err
		}
		return append(mp, lnode), nil
	}
	mp, err := s.merkleProof(lnode, height-1, key)
	if err != nil {
		return nil, err
	}
	return append(mp, rnode), nil
}

// merkleProofCompressed generates a Merke proof of inclusion or non inclusion for a given trie root
// a proof node is only appended if it is non default and the corresponding bit is set in the bitmap
func (s *SMT) merkleProofCompressed(root []byte, height uint64, key []byte, bitmap []byte) ([][]byte, error) {
	if height == 0 {
		return nil, nil
	}
	// Fetch the children of the node
	lnode, rnode, isShortcut, err := s.loadChildren(root)
	if err != nil {
		return nil, err
	}
	if isShortcut == 1 {
		if bytes.Equal(lnode, key) {
			return nil, nil
		}
		// if the key is empty, unroll until it diverges from the shortcut key and add the non default node
		return [][]byte{s.unrollShortcutAndKeyCompressed(lnode, rnode, height, bitmap, key)}, nil
	}

	// append the left or right node to the proof, if it is non default and set bitmap
	if bitIsSet(key, s.TrieHeight-height) {
		if !bytes.Equal(lnode, s.defaultHashes[height-1]) {
			// with validate proof, use a default hash when bit is not set
			bitSet(bitmap, height-1)
			mp, err := s.merkleProofCompressed(rnode, height-1, key, bitmap)
			if err != nil {
				return nil, err
			}
			return append(mp, lnode), nil
		}
		mp, err := s.merkleProofCompressed(rnode, height-1, key, bitmap)
		if err != nil {
			return nil, err
		}
		return mp, nil
	}
	if !bytes.Equal(rnode, s.defaultHashes[height-1]) {
		// with validate proof, use a default hash when bit is not set
		bitSet(bitmap, height-1)
		mp, err := s.merkleProofCompressed(lnode, height-1, key, bitmap)
		if err != nil {
			return nil, err
		}
		return append(mp, rnode), nil
	}
	mp, err := s.merkleProofCompressed(lnode, height-1, key, bitmap)
	if err != nil {
		return nil, err
	}
	return mp, err
}

// VerifyMerkleProof verifies that key/value is included in the trie with latest root
func (s *SMT) VerifyMerkleProof(ap [][]byte, key, value []byte) bool {
	return bytes.Equal(s.Root, s.verifyMerkleProof(ap, s.TrieHeight, key, value))
}

// VerifyMerkleProofCompressed verifies that key/value is included in the trie with latest root
func (s *SMT) VerifyMerkleProofCompressed(bitmap []byte, ap [][]byte, key, value []byte) bool {
	return bytes.Equal(s.Root, s.verifyMerkleProofCompressed(bitmap, ap, s.TrieHeight, uint64(len(ap)), key, value))
}

// verifyMerkleProof verifies that a key/value is included in the trie with given root
func (s *SMT) verifyMerkleProof(ap [][]byte, height uint64, key, value []byte) []byte {
	if height == 0 {
		return value
	}
	if bitIsSet(key, s.TrieHeight-height) {
		return hash(ap[height-1], s.verifyMerkleProof(ap, height-1, key, value))
	}
	return hash(s.verifyMerkleProof(ap, height-1, key, value), ap[height-1])
}

// verifyMerkleProof verifies that a key/value is included in the trie with given root
func (s *SMT) verifyMerkleProofCompressed(bitmap []byte, ap [][]byte, height uint64, apIndex uint64, key, value []byte) []byte {
	if height == 0 {
		return value
	}
	if bitIsSet(key, s.TrieHeight-height) {
		if bitIsSet(bitmap, height-1) {
			return hash(ap[apIndex-1], s.verifyMerkleProofCompressed(bitmap, ap, height-1, apIndex-1, key, value))
		}
		return hash(s.defaultHashes[height-1], s.verifyMerkleProofCompressed(bitmap, ap, height-1, apIndex, key, value))

	}
	if bitIsSet(bitmap, height-1) {
		return hash(s.verifyMerkleProofCompressed(bitmap, ap, height-1, apIndex-1, key, value), ap[apIndex-1])
	}
	return hash(s.verifyMerkleProofCompressed(bitmap, ap, height-1, apIndex, key, value), s.defaultHashes[height-1])
}

// shortcutToSubTreeRoot computes the subroot at height of a subtree containing one key
func (s *SMT) shortcutToSubTreeRoot(key, value []byte, height uint64) []byte {
	if height == 0 {
		return value
	}
	if bitIsSet(key, s.TrieHeight-height) {
		return s.hash(s.defaultHashes[height-1], s.shortcutToSubTreeRoot(key, value, height-1))
	}
	return s.hash(s.shortcutToSubTreeRoot(key, value, height-1), s.defaultHashes[height-1])
}

// unrollShortcutAndKey returns the merkle proof nodes of an empty key in a subtree that contains another key
// the key we are proving is not in the tree, it's value is default
func (s *SMT) unrollShortcutAndKey(key, value []byte, height uint64, emptyKey []byte) [][]byte {
	// if the keys have the same bits add a default hash to the proof
	if bitIsSet(key, s.TrieHeight-height) == bitIsSet(emptyKey, s.TrieHeight-height) {
		return append(s.unrollShortcutAndKey(key, value, height-1, emptyKey), s.defaultHashes[height-1])
	}
	// if the keys diverge, calculate the non default subroot and add default hashes until the leaf
	rest := make([][]byte, height-1)
	copy(rest, s.defaultHashes[:height-1])
	return append(rest, s.shortcutToSubTreeRoot(key, value, height-1))
}

// unrollShortcutAndKeyCompressed returns the merkle proof nodes of an empty key in a subtree that contains another key
// the key we are proving is not in the tree, it's value is default
func (s *SMT) unrollShortcutAndKeyCompressed(key, value []byte, height uint64, bitmap []byte, emptyKey []byte) []byte {
	// this version of unroll for compressed proofs simply sets the bitmap for non default nodes
	if bitIsSet(key, s.TrieHeight-height) == bitIsSet(emptyKey, s.TrieHeight-height) {
		return s.unrollShortcutAndKeyCompressed(key, value, height-1, bitmap, emptyKey)
	}
	bitSet(bitmap, height-1)
	return s.shortcutToSubTreeRoot(key, value, height-1)
}
