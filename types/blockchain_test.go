package types

import (
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockHashOldNew(t *testing.T) {
	blockHash := func(block *Block) []byte {
		header := block.Header
		digest := sha256.New()
		digest.Write(header.PrevBlockHash)
		binary.Write(digest, binary.LittleEndian, header.BlockNo)
		binary.Write(digest, binary.LittleEndian, header.Timestamp)
		digest.Write(header.TxsRootHash)
		return digest.Sum(nil)
	}

	txIn := make([]*Tx, 0)
	block := NewBlock(nil, txIn)

	h1 := blockHash(block)
	h2 := block.CalculateBlockHash()

	assert.Equal(t, h1, h2)
}
