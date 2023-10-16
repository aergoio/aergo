package util

import (
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/assert"
)

func TestParseConvBase58Tx(t *testing.T) {
	testjson := "[{\"Hash\":\"525mQMtsWaDLVJbzQZgTFkSG33gtZsho7m4io1HUCeJi\",\"Body\":{\"Nonce\":9,\"Account\":\"AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd\",\"Recipient\":\"AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR\",\"Amount\":\"100000000\",\"Payload\":null,\"Limit\":100,\"Price\":\"1\",\"Type\":0,\"Sign\":\"3tMHYrizQ532D1WJkt5RSs5AcRmq7betw8zvC66Wh3XHUdvNpNzLWh1SkkGYMGJ669nCVuYHrhwfg1HrUUp6KDwzK\"}}]"
	res, err := ParseBase58Tx([]byte(testjson))
	assert.NoError(t, err, "should be success")
	assert.NotEmpty(t, res, "failed to parse json")
	assert.Equal(t, "525mQMtsWaDLVJbzQZgTFkSG33gtZsho7m4io1HUCeJi", base58.Encode(res[0].Hash), "wrong hash")
	assert.Equal(t, "3tMHYrizQ532D1WJkt5RSs5AcRmq7betw8zvC66Wh3XHUdvNpNzLWh1SkkGYMGJ669nCVuYHrhwfg1HrUUp6KDwzK", base58.Encode(res[0].Body.Sign), "wrong sign")

	account, err := types.DecodeAddress("AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd")
	assert.NoError(t, err, "should be success")
	assert.Equal(t, account, res[0].Body.Account, "wrong account")

	recipient, err := types.DecodeAddress("AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR")
	assert.NoError(t, err, "should be success")
	assert.Equal(t, recipient, res[0].Body.Recipient, "wrong recipient")
}

func TestParseBase58TxBody(t *testing.T) {
	testjson := "{\"Nonce\":1,\"Account\":\"AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd\",\"Recipient\":\"AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR\",\"Amount\":\"25000\",\"Payload\":\"aergo\",\"Limit\":100,\"Price\":\"1\",\"Type\":0,\"Sign\":\"3roWPzztf5aLLh16vAnd2ugcPux3wJ1oqqvqkWARobjuAC32xftF42nnbTkXUQdkDaFvuUmctrpQSv8FAVUKcywHW\"}"
	res, err := ParseBase58TxBody([]byte(testjson))
	assert.NoError(t, err, "should be success")
	assert.NotEmpty(t, res, "failed to parse json")

	assert.Equal(t, "3roWPzztf5aLLh16vAnd2ugcPux3wJ1oqqvqkWARobjuAC32xftF42nnbTkXUQdkDaFvuUmctrpQSv8FAVUKcywHW", base58.Encode(res.Sign), "wrong sign")
	assert.Equal(t, "aergo", base58.Encode(res.Payload), "wrong payload")
	account, err := types.DecodeAddress("AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd")
	assert.NoError(t, err, "should be success")
	assert.Equal(t, account, res.Account, "wrong account")

	recipient, err := types.DecodeAddress("AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR")
	assert.NoError(t, err, "should be success")
	assert.Equal(t, recipient, res.Recipient, "wrong recipient")
}

func TestBlockConvBase58(t *testing.T) {
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
	t.Log(BlockConvBase58Addr(testBlock))
}
