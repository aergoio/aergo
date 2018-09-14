package merkle

import (
	"github.com/aergoio/aergo/types"
	"github.com/minio/sha256-simd"
	"hash"
)

var (
	HashSize = 32
	nilHash  = make([]byte, HashSize)
	//logger = log.NewLogger("merkle")
)

func GetMerkleTree(txs []*types.Tx) [][]byte {
	var merkles [][]byte
	txsLen := len(txs)

	if txsLen == 0 {
		merkles = append(merkles, nilHash)
		return merkles
	}

	//leaf count for full binary tree = 2 ^ n > txLen
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

	leafCount := getLeafCount(len(txs))
	totalCount := leafCount*2 - 1

	//logger.Debug().Int("leafcount", leafCount).Int("totCount", totalCount).Msg("start merkling")

	merkles = make([][]byte, totalCount)

	// init leaf hash (0 <= node# < tx len)
	for i, tx := range txs {
		merkles[i] = tx.GetHash()
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
	}

	return merkles
}
