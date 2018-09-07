package dpos

import (
	"testing"

	"github.com/aergoio/aergo/types"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/stretchr/testify/assert"
)

func genKeyPair(assert *assert.Assertions) (crypto.PrivKey, crypto.PubKey) {
	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	assert.Nil(err)

	return privKey, pubKey
}

func genSign(assert *assert.Assertions, block *types.Block, privKey crypto.PrivKey) []byte {
	msg, err := block.Header.BytesForDigest()
	assert.Nil(err)

	sig, err := privKey.Sign(msg)
	assert.Nil(err)

	return sig
}

func TestBlockSignBasic(t *testing.T) {
	signAssert := assert.New(t)

	message := func(block *types.Block) []byte {
		msg, err := block.Header.BytesForDigest()
		signAssert.Nil(err)

		return msg
	}

	sign := func(block *types.Block, privKey crypto.PrivKey) []byte {
		sig, err := privKey.Sign(message(block))
		signAssert.Nil(err)

		return sig
	}

	verify := func(block *types.Block, sig []byte, pubKey crypto.PubKey) bool {
		valid, err := pubKey.Verify(message(block), sig)
		signAssert.Nil(err)

		return valid
	}

	block := types.NewBlock(nil, make([]*types.Tx, 0), 0)

	privKey, pubKey := genKeyPair(signAssert)
	sig := sign(block, privKey)
	valid := verify(block, sig, pubKey)
	signAssert.True(valid)
}

func TestBlockSign(t *testing.T) {
	signAssert := assert.New(t)
	block := types.NewBlock(nil, make([]*types.Tx, 0), 0)

	privKey, _ := genKeyPair(signAssert)
	signAssert.Nil(block.Sign(privKey))

	sig := genSign(signAssert, block, privKey)
	signAssert.NotNil(sig)

	signAssert.Equal(sig, block.Header.Sign)

	valid, err := block.VerifySign()
	signAssert.Nil(err)
	signAssert.True(valid)
}
