package types

import (
	sha256 "github.com/minio/sha256-simd"

	"testing"

	crypto "github.com/libp2p/go-libp2p-crypto"
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
	block := NewBlock(nil, nil, txIn, 0)

	h1 := blockHash(block)
	h2 := block.calculateBlockHash()

	assert.Equal(t, h1, h2)
}

func genKeyPair(assert *assert.Assertions) (crypto.PrivKey, crypto.PubKey) {
	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	assert.Nil(err)

	return privKey, pubKey
}

func genSign(assert *assert.Assertions, block *Block, privKey crypto.PrivKey) []byte {
	msg, err := block.Header.bytesForDigest()
	assert.Nil(err)

	sig, err := privKey.Sign(msg)
	assert.Nil(err)

	return sig
}

func TestBlockSignBasic(t *testing.T) {
	signAssert := assert.New(t)

	message := func(block *Block) []byte {
		msg, err := block.Header.bytesForDigest()
		signAssert.Nil(err)

		return msg
	}

	sign := func(block *Block, privKey crypto.PrivKey) []byte {
		sig, err := privKey.Sign(message(block))
		signAssert.Nil(err)

		return sig
	}

	verify := func(block *Block, sig []byte, pubKey crypto.PubKey) bool {
		valid, err := pubKey.Verify(message(block), sig)
		signAssert.Nil(err)

		return valid
	}

	block := NewBlock(nil, nil, make([]*Tx, 0), 0)

	privKey, pubKey := genKeyPair(signAssert)
	sig := sign(block, privKey)
	valid := verify(block, sig, pubKey)
	signAssert.True(valid)
}

func TestBlockSign(t *testing.T) {
	signAssert := assert.New(t)
	block := NewBlock(nil, nil, make([]*Tx, 0), 0)

	privKey, _ := genKeyPair(signAssert)
	signAssert.Nil(block.Sign(privKey))

	sig := genSign(signAssert, block, privKey)
	signAssert.NotNil(sig)

	signAssert.Equal(sig, block.Header.Sign)

	valid, err := block.VerifySign()
	signAssert.Nil(err)
	signAssert.True(valid)
}
