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
func (s *modSMT) MerkleProof(key []byte) ([][]byte, bool, []byte, []byte, error) {
	return s.merkleProof(s.Root, s.TrieHeight, key)
}

// merkleProof generates a Merke proof of inclusion for a given trie root
// The proof of non inclusion is not explicit : it is a proof that
// a leaf node is on the path of the non included key.
// The modified SMT cannot provide a proof that a key is Default (not set in the tree)
// returns the audit path, true (key included), key, value on the path if false (non inclusion), error
func (s *modSMT) merkleProof(root []byte, height uint64, key []byte) ([][]byte, bool, []byte, []byte, error) {
	if height == 0 {
		if bytes.Equal(root, DefaultLeaf) {
			// if we reach DefaultLeaf without running into a leaf, then the key is not included
			return nil, false, nil, nil, nil
		}
		return nil, true, nil, nil, nil
	}
	// Fetch the children of the node
	lnode, rnode, isShortcut, err := s.loadChildren(root)
	if err != nil {
		return nil, true, nil, nil, err
	}
	if isShortcut == 1 {
		// append all default hashes down the tree
		if bytes.Equal(lnode, key) {
			return nil, true, nil, nil, nil
		}
		// Return the proof of the leaf key that is on the path of the non included key
		return nil, false, lnode, rnode, nil
	}

	// append the left or right node to the proof
	if bitIsSet(key, s.TrieHeight-height) {
		mp, included, proofKey, proofValue, err := s.merkleProof(rnode, height-1, key)
		if err != nil {
			return nil, included, proofKey, proofValue, err
		}
		return append(mp, lnode), included, proofKey, proofValue, nil
	}
	mp, included, proofKey, proofValue, err := s.merkleProof(lnode, height-1, key)
	if err != nil {
		return nil, included, proofKey, proofValue, err
	}
	return append(mp, rnode), included, proofKey, proofValue, nil
}

// VerifyMerkleProof verifies that key/value is included in the trie with latest root
func (s *modSMT) VerifyMerkleProof(ap [][]byte, key, value []byte) bool {
	leafHash := s.hash(key, value)
	return bytes.Equal(s.Root, s.verifyMerkleProof(ap, s.TrieHeight, key, leafHash))
}

// VerifyMerkleProofEmpty checks that the proofKey is included in the trie
// and that key and proofKey have the same bits up to len(ap)
func (s *modSMT) VerifyMerkleProofEmpty(ap [][]byte, key, proofKey, proofValue []byte) bool {
	if bytes.Equal(ap[0], DefaultLeaf) {
		// a shortcut node cannot be at height 0 if ap[0] == DefaultLeaf, it would be one level up
		return true
	}
	if !s.VerifyMerkleProof(ap, proofKey, proofValue) {
		// the proof key is not even included in the trie
		return false
	}
	var b uint64
	for b = 0; b < uint64(len(ap)); b++ {
		if bitIsSet(key, b) != bitIsSet(proofKey, b) {
			// the proofKey leaf node is not on the path of the key
			return false
		}
	}
	// this key is not included in the trie
	return true
}

// verifyMerkleProof verifies that a key/value is included in the trie with given root
func (s *modSMT) verifyMerkleProof(ap [][]byte, height uint64, key, leafHash []byte) []byte {
	if height == s.TrieHeight-uint64(len(ap)) {
		return leafHash
	}
	if bitIsSet(key, s.TrieHeight-height) {
		return hash(ap[uint64(len(ap))-(s.TrieHeight-height)-1], s.verifyMerkleProof(ap, height-1, key, leafHash))
	}
	return hash(s.verifyMerkleProof(ap, height-1, key, leafHash), ap[uint64(len(ap))-(s.TrieHeight-height)-1])
}
