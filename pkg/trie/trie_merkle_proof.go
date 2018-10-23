/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
)

// MerkleProof generates a Merke proof of inclusion or non-inclusion
// for the current trie root
// returns the audit path, bool (key included), key, value, error
// (key,value) can be 1- the kv of the included key, 2- the kv of a LeafNode
// on the path of the non-included key, 3- (nil, nil) for a non-included key
// with a DefaultLeaf on the path
func (s *Trie) MerkleProof(key []byte) ([][]byte, bool, []byte, []byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.atomicUpdate = false // so loadChildren doesnt return a copy
	return s.merkleProof(s.Root, key, nil, s.TrieHeight, 0)
}

// MerkleProofPast generates a Merke proof of inclusion or non-inclusion
// for a given past trie root
// returns the audit path, bool (key included), key, value, error
// (key,value) can be 1- the kv of the included key, 2- the kv of a LeafNode
// on the path of the non-included key, 3- (nil, nil) for a non-included key
// with a DefaultLeaf on the path
func (s *Trie) MerkleProofPast(key []byte, root []byte) ([][]byte, bool, []byte, []byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.atomicUpdate = false // so loadChildren doesnt return a copy
	return s.merkleProof(root, key, nil, s.TrieHeight, 0)
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
	height := uint64(len(mpFull))
	var mp [][]byte
	bitmap := make([]byte, len(mpFull)/8+1)
	for i, node := range mpFull {
		if !bytes.Equal(node, DefaultLeaf) {
			bitSet(bitmap, uint64(i))
			mp = append(mp, node)
		}
	}
	return bitmap, mp, height, included, proofKey, proofVal, nil
}

// merkleProof generates a Merke proof of inclusion or non-inclusion
// for a given trie root.
// returns the audit path, bool (key included), key, value, error
// (key,value) can be 1- the kv of the included key, 2- the kv of a LeafNode
// on the path of the non-included key, 3- (nil, nil) for a non-included key
// with a DefaultLeaf on the path
func (s *Trie) merkleProof(root, key []byte, batch [][]byte, height, iBatch uint64) ([][]byte, bool, []byte, []byte, error) {
	if len(root) == 0 {
		// proove that an empty subtree is on the path of the key
		return nil, false, nil, nil, nil
	}
	// Fetch the children of the node
	batch, iBatch, lnode, rnode, isShortcut, err := s.loadChildren(root, height, iBatch, batch)
	if err != nil {
		return nil, false, nil, nil, err
	}
	if isShortcut || height == 0 {
		if bytes.Equal(lnode[:HashLength], key) {
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
			return append(mp, DefaultLeaf), included, proofKey, proofValue, nil
		}

	}
	mp, included, proofKey, proofValue, err := s.merkleProof(lnode, key, batch, height-1, 2*iBatch+1)
	if err != nil {
		return nil, false, nil, nil, err
	}
	if len(rnode) != 0 {
		return append(mp, rnode[:HashLength]), included, proofKey, proofValue, nil
	} else {
		return append(mp, DefaultLeaf), included, proofKey, proofValue, nil
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
	//if uint64(len(ap)) == s.TrieHeight {
	if len(proofValue) == 0 {
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

// verifyMerkleProofCompressed verifies that a key/value is included in the trie with given root
func (s *Trie) verifyMerkleProofCompressed(bitmap []byte, ap [][]byte, length uint64, height uint64, apIndex uint64, key, leafHash []byte) []byte {
	if height == s.TrieHeight-length {
		return leafHash
	}
	if bitIsSet(key, s.TrieHeight-height) {
		if bitIsSet(bitmap, length-(s.TrieHeight-height)-1) {
			return s.hash(ap[apIndex-1], s.verifyMerkleProofCompressed(bitmap, ap, length, height-1, apIndex-1, key, leafHash))
		}
		return s.hash(DefaultLeaf, s.verifyMerkleProofCompressed(bitmap, ap, length, height-1, apIndex, key, leafHash))

	}
	if bitIsSet(bitmap, length-(s.TrieHeight-height)-1) {
		return s.hash(s.verifyMerkleProofCompressed(bitmap, ap, length, height-1, apIndex-1, key, leafHash), ap[apIndex-1])
	}
	return s.hash(s.verifyMerkleProofCompressed(bitmap, ap, length, height-1, apIndex, key, leafHash), DefaultLeaf)
}

// VerifyMerkleProofCompressedEmpty verifies that a key is not included in the tree.
// if the proof didnt run into a shortcut, it verifies as usual, otherwise
// it checks that the proofKey is included in the trie
// and that key and proofKey have the same bits up to the proofKey shortcut (length)
// A merkle proof consists of an audit path + an optional proof node
func (s *Trie) VerifyMerkleProofCompressedEmpty(bitmap []byte, ap [][]byte, length uint64, key, proofKey, proofValue []byte) bool {
	//if length == s.TrieHeight {
	if len(proofValue) == 0 {
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
