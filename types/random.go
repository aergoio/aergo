package types

import (
	"math/big"
	"math/rand"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
)

func GetRandomSeed(isQuery bool, block *BlockHeaderInfo, txHash []byte) *rand.Rand {
	var randSrc rand.Source
	if isQuery {
		randSrc = rand.NewSource(block.Ts)
	} else {
		b, _ := new(big.Int).SetString(base58.Encode(block.PrevBlockHash[:7]), 62)
		t, _ := new(big.Int).SetString(base58.Encode(txHash[:7]), 62)
		b.Add(b, t)
		randSrc = rand.NewSource(b.Int64())
	}

	return rand.New(randSrc)
}
