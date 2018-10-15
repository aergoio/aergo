/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
)

// MerkleProof generates a Merke proof of inclusion for a given trie root
// The proof of non inclusion is not explicit : it is a proof that
// a leaf node is on the path of the non included key.
// returns the audit path, true (key included), key, value on the path if false (non inclusion), error
func (s *Trie) MerkleProof(key []byte) ([][]byte, bool, []byte, []byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.atomicUpdate = false // so loadChildren doesnt return a copy
	return s.merkleProof(s.Root, key, nil, s.TrieHeight, 0)
}

// MerkleProofCompressed returns a compressed merkle proof
func (s *Trie) MerkleProofCompressed(key []byte) ([]byte, [][]byte, uint64, bool, []byte, []byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	// create a regular merkle proof and then compress it
	mpFull, included, proofKey, proofVal, err := s.merkleProof(s.Root, key, nil, s.TrieHeight, 0)
	if err != nil {
		return nil, nil, 0, true, nil, nil, err
	}
	// the height of the shortcut in the tree will be needed for the proof verification
	length := uint64(len(mpFull))
	var mp [][]byte
	bitmap := make([]byte, len(mpFull)/8+1)
	for i, node := range mpFull {
		if !bytes.Equal(node, s.defaultHashes[i]) {
			bitSet(bitmap, uint64(i))
			mp = append(mp, node)
		}
	}
	return bitmap, mp, length, included, proofKey, proofVal, nil
}

// merkleProof generates a Merke proof of inclusion for a given trie root
// The proof of non inclusion is not explicit : it is a proof that
// a leaf node is on the path of the non included key.
// returns the audit path, true (key included), key, value on the path if false (non inclusion), error
func (s *Trie) merkleProof(root, key []byte, batch [][]byte, height, iBatch uint64) ([][]byte, bool, []byte, []byte, error) {
	if len(root) == 0 {
		rest := make([][]byte, height)
		// copy because the array will be appended to.
		copy(rest, s.defaultHashes[:height])
		return rest, false, nil, nil, nil
	}
	if height == 0 {
		return nil, true, nil, nil, nil
	}
	// Fetch the children of the node
	batch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(root, height, iBatch, batch)
	if err != nil {
		return nil, false, nil, nil, err
	}
	if isShortcut {
		// append all default hashes down the tree
		if bytes.Equal(lnode, key) {
			// return the key-value so a call to trie.Get() is not needed.
			return nil, true, lnode[:HashLength], rnode[:HashLength], nil
		}
		// Return the proof of the leaf key that is on the path of the non included key
		return nil, false, lnode[:HashLength], rnode[:HashLength], nil
	}

	// append the left or right node to the proof
	if bitIsSet(key, s.TrieHeight-height) {
		mp, included, proofKey, proofValue, err := s.merkleProof(rnode, key, batch, height-1, 2*iBatch+2)
		if err != nil {
			return nil, false, nil, nil, err
		}
		if len(lnode) != 0 {
			return append(mp, lnode[:HashLength]), included, proofKey, proofValue, nil
		} else {
			return append(mp, s.defaultHashes[height-1]), included, proofKey, proofValue, nil
		}

	}
	mp, included, proofKey, proofValue, err := s.merkleProof(lnode, key, batch, height-1, 2*iBatch+1)
	if err != nil {
		return nil, false, nil, nil, err
	}
	if len(rnode) != 0 {
		return append(mp, rnode[:HashLength]), included, proofKey, proofValue, nil
	} else {
		return append(mp, s.defaultHashes[height-1]), included, proofKey, proofValue, nil
	}
}

// VerifyMerkleProof verifies that key/value is included in the trie with latest root
func (s *Trie) VerifyMerkleProof(ap [][]byte, key, value []byte) bool {
	leafHash := s.hash(key, value, []byte{byte(int(s.TrieHeight) - len(ap))})
	return bytes.Equal(s.Root, s.verifyMerkleProof(ap, s.TrieHeight, key, leafHash))
}

// VerifyMerkleProofEmpty checks that the proofKey is included in the trie
// and that key and proofKey have the same bits up to len(ap)
// InTrie , a merkle proof consists of an audit path + an optional proof node
func (s *Trie) VerifyMerkleProofEmpty(ap [][]byte, key, proofKey, proofValue []byte) bool {
	if uint64(len(ap)) == s.TrieHeight {
		//if bytes.Equal(ap[0], DefaultLeaf) {
		// if the proof goes down to the DefaultLeaf, then there is no shortcut on the way
		return bytes.Equal(s.Root, s.verifyMerkleProof(ap, s.TrieHeight, key, DefaultLeaf))
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
	// this key is not included in the trie : it is default
	return true
}

// VerifyMerkleProofCompressed verifies that key/value is included in the trie with latest root
func (s *Trie) VerifyMerkleProofCompressed(bitmap []byte, ap [][]byte, length uint64, key, value []byte) bool {
	leafHash := s.hash(key, value, []byte{byte(s.TrieHeight - length)})
	return bytes.Equal(s.Root, s.verifyMerkleProofCompressed(bitmap, ap, length, s.TrieHeight, uint64(len(ap)), key, leafHash))
}

// verifyMerkleProof verifies that a key/value is included in the trie with given root
func (s *Trie) verifyMerkleProof(ap [][]byte, height uint64, key, leafHash []byte) []byte {
	if height == s.TrieHeight-uint64(len(ap)) {
		return leafHash
	}
	if bitIsSet(key, s.TrieHeight-height) {
		return s.hash(ap[uint64(len(ap))-(s.TrieHeight-height)-1], s.verifyMerkleProof(ap, height-1, key, leafHash))
	}
	return s.hash(s.verifyMerkleProof(ap, height-1, key, leafHash), ap[uint64(len(ap))-(s.TrieHeight-height)-1])
}

// verifyMerkleProof verifies that a key/value is included in the trie with given root
func (s *Trie) verifyMerkleProofCompressed(bitmap []byte, ap [][]byte, length uint64, height uint64, apIndex uint64, key, leafHash []byte) []byte {
	if height == s.TrieHeight-length {
		return leafHash
	}
	if bitIsSet(key, s.TrieHeight-height) {
		if bitIsSet(bitmap, length-(s.TrieHeight-height)-1) {
			return s.hash(ap[apIndex-1], s.verifyMerkleProofCompressed(bitmap, ap, length, height-1, apIndex-1, key, leafHash))
		}
		return s.hash(s.defaultHashes[height-1], s.verifyMerkleProofCompressed(bitmap, ap, length, height-1, apIndex, key, leafHash))

	}
	if bitIsSet(bitmap, length-(s.TrieHeight-height)-1) {
		return s.hash(s.verifyMerkleProofCompressed(bitmap, ap, length, height-1, apIndex-1, key, leafHash), ap[apIndex-1])
	}
	return s.hash(s.verifyMerkleProofCompressed(bitmap, ap, length, height-1, apIndex, key, leafHash), s.defaultHashes[height-1])
}

// VerifyMerkleProofCompressedEmpty verifies that a key is not included in the tree.
// if the proof didnt run into a shortcut, it verifies as usual, otherwise
// it checks that the proofKey is included in the trie
// and that key and proofKey have the same bits up to the proofKey shortcut (length)
// A merkle proof consists of an audit path + an optional proof node
func (s *Trie) VerifyMerkleProofCompressedEmpty(bitmap []byte, ap [][]byte, length uint64, key, proofKey, proofValue []byte) bool {
	if length == s.TrieHeight {
		// if the proof goes down to the DefaultLeaf, then there is no shortcut on the way
		return bytes.Equal(s.Root, s.verifyMerkleProofCompressed(bitmap, ap, length, s.TrieHeight, uint64(len(ap)), key, DefaultLeaf))
	}
	if !s.VerifyMerkleProofCompressed(bitmap, ap, length, proofKey, proofValue) {
		// the proof key is not even included in the trie
		return false
	}
	var b uint64
	for b = 0; b < length; b++ {
		if bitIsSet(key, b) != bitIsSet(proofKey, b) {
			// the proofKey leaf node is not on the path of the key
			return false
		}
	}
	// this key is not included in the trie
	return true
}
