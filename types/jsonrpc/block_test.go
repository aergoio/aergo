package jsonrpc

import (
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

const (
	testAccountBase58     = "AmMW2bVcfroiuV4Bvy56op5zzqn42xgrLCwSxMka23K75yTBmudz"
	testRecipientBase58   = "AmMW2bVcfroiuV4Bvy56op5zzqn42xgrLCwSxMka23K75yTBmudz"
	testPayloadBase58     = "525mQMtsWaDLVJbzQZgTFkSG33gtZsho7m4io1HUCeJi"
	testBlockHashBase58   = "5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq"
	testTxHashBase58      = "5v1hmuTmDbS744oHMVJdFtb5LNPn4wbAtcK1HtveUAz"
	testChainIdHashBase58 = "73zLfCHqvPk1oRvV7VeTgXx6XcL3Sat9u" // 73zLfCHqvPk1oRvV7VeTgXx6XcL3Sat9u
	testSignBase58        = "3tMHYrizQ532D1WJkt5RSs5AcRmq7betw8zvC66Wh3XHUdvNpNzL"
)

func TestConvBlock(t *testing.T) {
	account, err := types.DecodeAddress(testAccountBase58)
	assert.NoError(t, err, "should be decode account")

	blockHash, err := base58.Decode(testBlockHashBase58)
	assert.NoError(t, err, "should be decode block hash")

	testBlock := &types.Block{
		Hash: blockHash,
		Header: &types.BlockHeader{
			CoinbaseAccount: account,
		},
		Body: &types.BlockBody{
			Txs: []*types.Tx{},
		},
	}
	result := ConvBlock(nil)
	assert.Empty(t, result, "failed to convert nil")

	result = ConvBlock(testBlock)
	assert.Equal(t, testBlockHashBase58, result.Hash, "failed to convert block hash")
	t.Log(ConvBlock(testBlock))
}

func TestConvBlockHeader(t *testing.T) {
	decodeBlockHash, err := base58.Decode(testBlockHashBase58)
	assert.NoError(t, err, "should be decode block hash")

	for _, test := range []struct {
		types *types.BlockHeader
		inout *InOutBlockHeader
	}{
		{&types.BlockHeader{
			ChainID:       []byte{0x00, 0x00, 0x00, 0x00},
			PrevBlockHash: decodeBlockHash,
			BlockNo:       1,
			Timestamp:     1600000000,
			Confirms:      1,
		}, &InOutBlockHeader{
			ChainID:       base58.Encode([]byte{0x00, 0x00, 0x00, 0x00}),
			Version:       types.DecodeChainIdVersion([]byte{0x00, 0x00, 0x00, 0x00}),
			PrevBlockHash: testBlockHashBase58,
			BlockNo:       1,
			Timestamp:     1600000000,
			Confirms:      1,
		}},
	} {
		result := ConvBlockHeader(test.types)
		assert.Equal(t, test.inout, result, "failed to convert block header")
	}
}

func TestConvBlockBody(t *testing.T) {
	account, err := types.DecodeAddress(testAccountBase58)
	assert.NoError(t, err, "should be decode account")

	recipient, err := types.DecodeAddress(testRecipientBase58)
	assert.NoError(t, err, "should be decode recipient")

	payload, err := base58.Decode(testPayloadBase58)
	assert.NoError(t, err, "should be decode payload")

	for _, test := range []struct {
		types *types.BlockBody
		inout *InOutBlockBody
	}{
		{&types.BlockBody{
			Txs: []*types.Tx{
				{
					Body: &types.TxBody{
						Account:   account,
						Recipient: recipient,
						Payload:   payload,
					},
				},
			},
		}, &InOutBlockBody{
			Txs: []*InOutTx{
				{
					Body: &InOutTxBody{
						Account:   testAccountBase58,
						Recipient: testRecipientBase58,
						Payload:   testPayloadBase58,
					},
				},
			},
		}},
	} {
		result := ConvBlockBody(test.types)
		assert.Equal(t, test.inout, result, "failed to convert block body")
	}
}
