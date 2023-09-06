package util

import (
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestConvTxEx(t *testing.T) {
	testTx := &types.Tx{Body: &types.TxBody{Payload: []byte("{\"Name\":\"v1createName\",\"Args\":[\"honggildong3\"]}")}}
	result := toString(ConvTxEx(testTx, Base58))
	assert.Equal(t, "{\n \"Body\": {\n  \"Payload\": \"22MZAFWvxtVWehpgwEVxrvoqGL5xmcPmyLBiwraDfxRwKUNrV9tmhuB7Uu6ZeJWvp\"\n }\n}", result, "")
	result = toString(ConvTxEx(testTx, Raw))
	assert.Equal(t, "{\n \"Body\": {\n  \"Payload\": \"{\\\"Name\\\":\\\"v1createName\\\",\\\"Args\\\":[\\\"honggildong3\\\"]}\"\n }\n}", result, "")
}
