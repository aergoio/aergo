package merkle

import (
	"github.com/minio/sha256-simd"
	"hash"
)

type MerkleEntry interface {
	GetHash() []byte
}

var (
	HashSize = 32
	nilHash  = make([]byte, HashSize)
	//logger = log.NewLogger("merkle")
)

func CalculateMerkleRoot(entries []MerkleEntry) []byte {
	merkles := CalculateMerkleTree(entries)

	return merkles[len(merkles)-1]
}

func CalculateMerkleTree(entries []MerkleEntry) [][]byte {
	var merkles [][]byte
	entriesLen := len(entries)

	if entriesLen == 0 {
		merkles = append(merkles, nilHash)
		return merkles
	}

	//leaf count for full binary tree = 2 ^ n > entryLen
	getLeafCount := func(num int) int {
		if (num&num - 1) == 0 {
			return num
		}
		x := 1
		for x < num {
			x = x << 1
		}
		return x
	}

	calcMerkle := func(hasher hash.Hash, lc []byte, rc []byte) []byte {
		hasher.Reset()
		hasher.Write(lc)
		hasher.Write(rc)
		return hasher.Sum(nil)
	}

	hasher := sha256.New()

	leafCount := getLeafCount(len(entries))
	totalCount := leafCount*2 - 1

	//logger.Debug().Int("leafcount", leafCount).Int("totCount", totalCount).Msg("start merkling")

	merkles = make([][]byte, totalCount)

	// init leaf hash (0 <= node# < entry len)
	for i, entry := range entries {
		merkles[i] = entry.GetHash()
	}

	// start from branch height 1 (merkles[leafcount] ~ merkles[totalCount -1])
	var childIdx = 0
	var lc, rc int
	for i := leafCount; i < totalCount; i++ {
		// hash of branch node is zero if all child not exist
		lc = childIdx
		rc = childIdx + 1
		childIdx += 2

		// If all child is nil, merkle is nil
		if merkles[lc] == nil {
			merkles[i] = nil
			continue
		}
		// If only exist left child, copy left child hash to right child
		if merkles[rc] == nil {
			merkles[rc] = merkles[lc]
		}

		merkles[i] = calcMerkle(hasher, merkles[lc], merkles[rc])
		//logger.Debug().Int("i", i).Str("m", EncodeB64(merkles[i])).Msg("merkling")
	}

	return merkles
}
