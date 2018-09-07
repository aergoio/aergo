package types

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockHash(t *testing.T) {
	blockHash := func(block *Block) []byte {
		header := block.Header
		digest := sha256.New()
		writeBlockHeaderOld(digest, header)
		return digest.Sum(nil)
	}

	txIn := make([]*Tx, 0)
	block := NewBlock(nil, txIn, 0)

	h1 := blockHash(block)
	h2 := block.calculateBlockHash()

	assert.Equal(t, h1, h2)
}
