/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"crypto/sha256"
)

var (
	// Trie default value : [byte(0)]
	DefaultLeaf = []byte{0}
)

const (
	HashLength   = 32
	maxPastTries = 300
)

type Hash [HashLength]byte

func bitIsSet(bits []byte, i int) bool {
	return bits[i/8]&(1<<uint(7-i%8)) != 0
}
func bitSet(bits []byte, i int) {
	bits[i/8] |= 1 << uint(7-i%8)
}
func bitSplit(bits []byte, i int) (split []byte) {
	split = make([]byte, len(bits))
	copy(split, bits)
	bitSet(split, i)
	return
}

func Hasher(data ...[]byte) []byte {
	hasher := sha256.New()
	for i := 0; i < len(data); i++ {
		hasher.Write(data[i])
	}
	return hasher.Sum(nil)
}

// for sorting
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
