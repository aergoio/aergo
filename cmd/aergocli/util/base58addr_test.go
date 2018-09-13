package util

import (
	"testing"

	"github.com/mr-tron/base58/base58"
)

func TestParseConvBase58Tx(t *testing.T) {
	testjson := "[{\"Hash\":\"4yCbNsb5MTb3ZTuZ7RvkpFmCbDc1uNTnCuUv4HkAHeKW\",\"Body\":{\"Nonce\":1,\"Account\":\"s7RA1sZhAzec5WhPdthJcLen63f\",\"Recipient\":\"2TjCzArubVYD6tSzsaF2HZWNWMcz\",\"Amount\":25000,\"Payload\":null,\"Limit\":100,\"Price\":1,\"Type\":0,\"Sign\":\"3roWPzztf5aLLh16vAnd2ugcPux3wJ1oqqvqkWARobjuAC32xftF42nnbTkXUQdkDaFvuUmctrpQSv8FAVUKcywHW\"}}]"
	res, err := ParseBase58Tx([]byte(testjson))
	if err != nil {
		t.Errorf("Failed to parse : %s ", err.Error())
	}
	if res == nil {
		t.Error("Return nil")
	}
	if base58.Encode(res[0].Body.Recipient) != "2TjCzArubVYD6tSzsaF2HZWNWMcz" {
		t.Error("Failed to parse recipient")
	}
}
func TestParseBase58TxBody(t *testing.T) {
	testjson := "{\"Nonce\":1,\"Account\":\"s7RA1sZhAzec5WhPdthJcLen63f\",\"Recipient\":\"2TjCzArubVYD6tSzsaF2HZWNWMcz\",\"Amount\":25000,\"Payload\":null,\"Limit\":100,\"Price\":1,\"Type\":0,\"Sign\":\"3roWPzztf5aLLh16vAnd2ugcPux3wJ1oqqvqkWARobjuAC32xftF42nnbTkXUQdkDaFvuUmctrpQSv8FAVUKcywHW\"}"
	res, err := ParseBase58TxBody([]byte(testjson))
	if err != nil {
		t.Errorf("Failed to parse : %s ", err.Error())
	}
	if res == nil {
		t.Error("return nil")
	}
	if base58.Encode(res.Recipient) != "2TjCzArubVYD6tSzsaF2HZWNWMcz" {
		t.Error("Failed to parse recipient")
	}
}
