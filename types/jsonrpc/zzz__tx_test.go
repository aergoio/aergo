package jsonrpc

import (
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
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

func TestConvTx(t *testing.T) {
	testTx := &types.Tx{Body: &types.TxBody{Payload: []byte("{\"Name\":\"v1createName\",\"Args\":[\"honggildong3\"]}")}}
	result := MarshalJSON(ConvTx(testTx, Base58))
	assert.Equal(t, "{\n \"Body\": {\n  \"Payload\": \"22MZAFWvxtVWehpgwEVxrvoqGL5xmcPmyLBiwraDfxRwKUNrV9tmhuB7Uu6ZeJWvp\"\n }\n}", result, "")
	result = MarshalJSON(ConvTx(testTx, Raw))
	assert.Equal(t, "{\n \"Body\": {\n  \"Payload\": \"{\\\"Name\\\":\\\"v1createName\\\",\\\"Args\\\":[\\\"honggildong3\\\"]}\"\n }\n}", result, "")
}
