package jsonrpc

import (
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestBlockConv(t *testing.T) {
	const accountBase58 = "AmMW2bVcfroiuV4Bvy56op5zzqn42xgrLCwSxMka23K75yTBmudz"
	const recipientBase58 = "AmMW2bVcfroiuV4Bvy56op5zzqn42xgrLCwSxMka23K75yTBmudz"
	const payloadBase58 = "525mQMtsWaDLVJbzQZgTFkSG33gtZsho7m4io1HUCeJi"

	account, err := types.DecodeAddress(accountBase58)
	assert.NoError(t, err, "should be decode account")

	testBlock := &types.Block{
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
	assert.Empty(t, result.Body.Txs, "failed to convert txs")
	assert.Equal(t, accountBase58, result.Header.CoinbaseAccount, "failed to convert coinbase account")

	recipient, err := types.DecodeAddress(recipientBase58)
	assert.NoError(t, err, "should be decode recipient")

	payload, err := base58.Decode(payloadBase58)
	assert.NoError(t, err, "should be decode payload")

	testTx := &types.Tx{Body: &types.TxBody{
		Account:   account,
		Recipient: recipient,
		Payload:   payload,
	}}

	testBlock.Body.Txs = append(testBlock.Body.Txs, testTx)
	result = ConvBlock(testBlock)
	assert.Equal(t, accountBase58, result.Body.Txs[0].Body.Account, "failed to convert account")
	assert.Equal(t, recipientBase58, result.Body.Txs[0].Body.Recipient, "failed to convert recipient")
	assert.Equal(t, payloadBase58, result.Body.Txs[0].Body.Payload, "failed to convert payload")
	t.Log(ConvBlock(testBlock))
}
