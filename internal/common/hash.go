package common

import (
	sha256 "github.com/minio/sha256-simd"
)

// Hasher exports default hash function for trie
var Hasher = func(data ...[]byte) []byte {
	hasher := sha256.New()
	for i := 0; i < len(data); i++ {
		hasher.Write(data[i])
	}
	return hasher.Sum(nil)
}
