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
// (key,value) can be 1- (nil, value), value of the included key, 2- the kv of a LeafNode
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
// (key,value) can be 1- (nil, value), value of the included key, 2- the kv of a LeafNode
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
// (key,value) can be 1- (nil, value), value of the included key, 2- the kv of a LeafNode
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
			// return the value so a call to trie.Get() is not needed.
			return nil, true, nil, rnode[:HashLength], nil
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

// VerifyInclusion verifies that key/value is included in the trie with latest root
func (s *Trie) VerifyInclusion(ap [][]byte, key, value []byte) bool {
	leafHash := s.hash(key, value, []byte{byte(int(s.TrieHeight) - len(ap))})
	return bytes.Equal(s.Root, s.verifyInclusion(ap, 0, key, leafHash))
}

// verifyInclusion returns the merkle root by hashing the merkle proof items
func (s *Trie) verifyInclusion(ap [][]byte, keyIndex uint64, key, leafHash []byte) []byte {
	if keyIndex == uint64(len(ap)) {
		return leafHash
	}
	if bitIsSet(key, keyIndex) {
		return s.hash(ap[uint64(len(ap))-keyIndex-1], s.verifyInclusion(ap, keyIndex+1, key, leafHash))
	}
	return s.hash(s.verifyInclusion(ap, keyIndex+1, key, leafHash), ap[uint64(len(ap))-keyIndex-1])
}

// VerifyNonInclusion verifies a proof of non inclusion,
// Returns true if the non-inclusion is verified
func (s *Trie) VerifyNonInclusion(ap [][]byte, key, value, proofKey []byte) bool {
	// Check if an empty subtree is on the key path
	if len(proofKey) == 0 {
		// return true if a DefaultLeaf in the key path is included in the trie
		return bytes.Equal(s.Root, s.verifyInclusion(ap, 0, key, DefaultLeaf))
	}
	// Check if another kv leaf is on the key path in 2 steps
	// 1- Check the proof leaf exists
	if !s.VerifyInclusion(ap, proofKey, value) {
		// the proof leaf is not included in the trie
		return false
	}
	// 2- Check the proof leaf is on the key path
	var b uint64
	for b = 0; b < uint64(len(ap)); b++ {
		if bitIsSet(key, b) != bitIsSet(proofKey, b) {
			// the proofKey leaf node is not on the path of the key
			return false
		}
	}
	// return true because we verified another leaf is on the key path
	return true
}

// VerifyInclusionC verifies that key/value is included in the trie with latest root
func (s *Trie) VerifyInclusionC(bitmap, key, value []byte, ap [][]byte, length uint64) bool {
	leafHash := s.hash(key, value, []byte{byte(s.TrieHeight - length)})
	return bytes.Equal(s.Root, s.verifyInclusionC(bitmap, key, leafHash, ap, length, 0, 0))
}

// verifyInclusionC returns the merkle root by hashing the merkle proof items
func (s *Trie) verifyInclusionC(bitmap, key, leafHash []byte, ap [][]byte, length, keyIndex, apIndex uint64) []byte {
	if keyIndex == length {
		return leafHash
	}
	if bitIsSet(key, keyIndex) {
		if bitIsSet(bitmap, length-keyIndex-1) {
			return s.hash(ap[uint64(len(ap))-apIndex-1], s.verifyInclusionC(bitmap, key, leafHash, ap, length, keyIndex+1, apIndex+1))
		}
		return s.hash(DefaultLeaf, s.verifyInclusionC(bitmap, key, leafHash, ap, length, keyIndex+1, apIndex))

	}
	if bitIsSet(bitmap, length-keyIndex-1) {
		return s.hash(s.verifyInclusionC(bitmap, key, leafHash, ap, length, keyIndex+1, apIndex+1), ap[uint64(len(ap))-apIndex-1])
	}
	return s.hash(s.verifyInclusionC(bitmap, key, leafHash, ap, length, keyIndex+1, apIndex), DefaultLeaf)
}

// VerifyNonInclusionC verifies a proof of non inclusion,
// Returns true if the non-inclusion is verified
func (s *Trie) VerifyNonInclusionC(ap [][]byte, length uint64, bitmap, key, value, proofKey []byte) bool {
	// Check if an empty subtree is on the key path
	if len(proofKey) == 0 {
		// return true if a DefaultLeaf in the key path is included in the trie
		return bytes.Equal(s.Root, s.verifyInclusionC(bitmap, key, DefaultLeaf, ap, length, 0, 0))
	}
	// Check if another kv leaf is on the key path in 2 steps
	// 1- Check the proof leaf exists
	if !s.VerifyInclusionC(bitmap, proofKey, value, ap, length) {
		// the proof leaf is not included in the trie
		return false
	}
	// 2- Check the proof leaf is on the key path
	var b uint64
	for b = 0; b < length; b++ {
		if bitIsSet(key, b) != bitIsSet(proofKey, b) {
			// the proofKey leaf node is not on the path of the key
			return false
		}
	}
	// return true because we verified another leaf is on the key path
	return true
}
