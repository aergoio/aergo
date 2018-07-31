/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"crypto/sha512"
	"sort"
)

var (
	// Empty is an empty key.
	Empty = []byte{0x0}
	// Set is a set key.
	Set = []byte{0x1}
	// Trie default value : hash of 0x0
	DefaultLeaf = hash(Empty)
)

// TODO make D a type []Hash if Aergo defines a custom Hash type like Ethereum
type DataArray [][]byte
type Hash [HashLength]byte

// for sorting
func (d DataArray) Len() int           { return len(d) }
func (d DataArray) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d DataArray) Less(i, j int) bool { return bytes.Compare(d[i], d[j]) == -1 }

// Split splits d.
func (d DataArray) Split(s []byte) (l, r DataArray) {
	// the smallest index i where d[i] >= s
	i := sort.Search(d.Len(), func(i int) bool {
		return bytes.Compare(d[i], s) >= 0
	})
	return d[:i], d[i:]
}
func bitIsSet(bits []byte, i uint64) bool { return bits[i/8]&(1<<uint(7-i%8)) != 0 }
func bitSet(bits []byte, i uint64)        { bits[i/8] |= 1 << uint(7-i%8) }
func bitSplit(bits []byte, i uint64) (split []byte) {
	split = make([]byte, len(bits))
	copy(split, bits)
	bitSet(split, i)
	return
}

func hash(data ...[]byte) []byte {
	hasher := sha512.New512_256()
	for i := 0; i < len(data); i++ {
		hasher.Write(data[i])
	}
	return hasher.Sum(nil)
}
