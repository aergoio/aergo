/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
)

var (
	// Trie default value : [byte(0)]
	DefaultLeaf = []byte{0}
)

const (
	HashLength   = 32
	maxPastTries = 300
)

// hashData returns the hash prefix of a trie node with cap==len. Subslices of
// mmap-backed DB values inherit spare capacity; uncapped [:HashLength] leaves
// room for append to write into read-only pages.
func hashData(b []byte) []byte {
	if len(b) <= HashLength {
		return b[:len(b):len(b)]
	}
	return b[:HashLength:HashLength]
}

type Hash [HashLength]byte

func bitIsSet(bits []byte, i int) bool {
	return bits[i>>3]&(1<<(7-(i&7))) != 0
}
func bitSet(bits []byte, i int) {
	bits[i>>3] |= 1 << (7 - (i & 7))
}

// for sorting test data
type DataArray [][]byte

func (d DataArray) Len() int {
	return len(d)
}
func (d DataArray) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
func (d DataArray) Less(i, j int) bool {
	return bytes.Compare(d[i], d[j]) == -1
}
