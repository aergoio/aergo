package jsonrpc

import (
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestParseConvBase58Tx(t *testing.T) {
	for _, test := range []struct {
		txJson string
	}{
		{`[{"Hash":"525mQMtsWaDLVJbzQZgTFkSG33gtZsho7m4io1HUCeJi","Body":{"Nonce":9,"Account":"AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd","Recipient":"AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR","Amount":"100000000","Payload":null,"Limit":100,"Price":"1","Type":0,"Sign":"3tMHYrizQ532D1WJkt5RSs5AcRmq7betw8zvC66Wh3XHUdvNpNzLWh1SkkGYMGJ669nCVuYHrhwfg1HrUUp6KDwzK"}}]`},
		{`[{"hash":"525mQMtsWaDLVJbzQZgTFkSG33gtZsho7m4io1HUCeJi","body":{"nonce":9,"account":"AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd","recipient":"AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR","amount":"100000000","payload":null,"limit":100,"price":"1","type":0,"sign":"3tMHYrizQ532D1WJkt5RSs5AcRmq7betw8zvC66Wh3XHUdvNpNzLWh1SkkGYMGJ669nCVuYHrhwfg1HrUUp6KDwzK"}}]`},
	} {
		res, err := ParseBase58Tx([]byte(test.txJson))
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
}

func TestParseBase58TxBody(t *testing.T) {
	for _, test := range []struct {
		txJson string
	}{
		{`{"Nonce":1,"Account":"AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd","Recipient":"AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR","Amount":"25000","Payload":"aergo","Limit":100,"Price":"1","Type":0,"Sign":"3roWPzztf5aLLh16vAnd2ugcPux3wJ1oqqvqkWARobjuAC32xftF42nnbTkXUQdkDaFvuUmctrpQSv8FAVUKcywHW"}`},
		{`{"nonce":1,"account":"AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd","recipient":"AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR","amount":"25000","payload":"aergo","limit":100,"price":"1","type":0,"sign":"3roWPzztf5aLLh16vAnd2ugcPux3wJ1oqqvqkWARobjuAC32xftF42nnbTkXUQdkDaFvuUmctrpQSv8FAVUKcywHW"}`},
	} {
		res, err := ParseBase58TxBody([]byte(test.txJson))
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
}

func TestParseTx(t *testing.T) {
	testTx := &types.Tx{Body: &types.TxBody{Payload: []byte(`{"Name":"v1createName","Args":["honggildong3"]}`)}}
	result := MarshalJSON(ConvTx(testTx, Base58))
	assert.Equal(t, "{\n \"body\": {\n  \"payload\": \"22MZAFWvxtVWehpgwEVxrvoqGL5xmcPmyLBiwraDfxRwKUNrV9tmhuB7Uu6ZeJWvp\"\n }\n}", result, "")
	result = MarshalJSON(ConvTx(testTx, Raw))
	assert.Equal(t, "{\n \"body\": {\n  \"payload\": \"{\\\"Name\\\":\\\"v1createName\\\",\\\"Args\\\":[\\\"honggildong3\\\"]}\"\n }\n}", result, "")
}

func TestConvTxBody(t *testing.T) {
	account, err := types.DecodeAddress(testAccountBase58)
	assert.NoError(t, err, "should be success")
	recipient, err := types.DecodeAddress(testRecipientBase58)
	assert.NoError(t, err, "should be decode recipient")
	payload, err := base58.Decode(testPayloadBase58)
	assert.NoError(t, err, "should be decode payload")
	chainIdHash, err := base58.Decode(testChainIdHashBase58)
	assert.NoError(t, err, "should be decode chainIdHash")
	sign, err := base58.Decode(testSignBase58)
	assert.NoError(t, err, "should be decode sign")

	for _, test := range []struct {
		encType EncodingType
		types   *types.TxBody
		inout   *InOutTxBody
	}{
		{Base58, &types.TxBody{
			Nonce:       1,
			Account:     account,
			Recipient:   recipient,
			Payload:     payload,
			GasLimit:    100000,
			GasPrice:    types.NewAmount(5, types.Aergo).Bytes(),
			Type:        types.TxType_NORMAL,
			ChainIdHash: chainIdHash,
			Sign:        sign,
		}, &InOutTxBody{
			Nonce:       1,
			Account:     testAccountBase58,
			Recipient:   testRecipientBase58,
			Payload:     testPayloadBase58,
			GasLimit:    100000,
			GasPrice:    types.NewAmount(5, types.Aergo).String(),
			Type:        types.TxType_NORMAL,
			ChainIdHash: testChainIdHashBase58,
			Sign:        testSignBase58,
		}},
		{Raw, &types.TxBody{
			Nonce:       1,
			Account:     account,
			Recipient:   recipient,
			Payload:     payload,
			GasLimit:    100000,
			GasPrice:    types.NewAmount(5, types.Aergo).Bytes(),
			Type:        types.TxType_NORMAL,
			ChainIdHash: chainIdHash,
			Sign:        sign,
		}, &InOutTxBody{
			Nonce:       1,
			Account:     testAccountBase58,
			Recipient:   testRecipientBase58,
			Payload:     string(payload),
			GasLimit:    100000,
			GasPrice:    types.NewAmount(5, types.Aergo).String(),
			Type:        types.TxType_NORMAL,
			ChainIdHash: testChainIdHashBase58,
			Sign:        testSignBase58,
		}},
	} {
		assert.Equal(t, test.inout, ConvTxBody(test.types, test.encType))
	}
}

func TestConvTxInBlock(t *testing.T) {
	for _, test := range []struct {
		types *types.TxInBlock
		inout *InOutTxInBlock
	}{
		{&types.TxInBlock{}, &InOutTxInBlock{}},
	} {
		assert.EqualValues(t, test.inout, ConvTxInBlock(test.types, Base58))
	}
}

func TestConvTxIdx(t *testing.T) {
	blockHash, err := base58.Decode(testBlockHashBase58)
	assert.NoError(t, err, "should be decode blockHash")

	for _, test := range []struct {
		types *types.TxIdx
		inout *InOutTxIdx
	}{
		{&types.TxIdx{
			BlockHash: blockHash,
			Idx:       1,
		}, &InOutTxIdx{
			BlockHash: testBlockHashBase58,
			Idx:       1,
		}},
	} {
		assert.Equal(t, test.inout, ConvTxIdx(test.types))
	}
}
