package util

import (
	"testing"

	"github.com/mr-tron/base58/base58"

	"github.com/aergoio/aergo/types"
)

func TestParseConvBase58Tx(t *testing.T) {
	testjson := "[{\"Hash\":\"525mQMtsWaDLVJbzQZgTFkSG33gtZsho7m4io1HUCeJi\",\"Body\":{\"Nonce\":9,\"Account\":\"AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd\",\"Recipient\":\"AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR\",\"Amount\":100000000,\"Payload\":null,\"Limit\":100,\"Price\":1,\"Type\":0,\"Sign\":\"3tMHYrizQ532D1WJkt5RSs5AcRmq7betw8zvC66Wh3XHUdvNpNzLWh1SkkGYMGJ669nCVuYHrhwfg1HrUUp6KDwzK\"}}]"
	res, err := ParseBase58Tx([]byte(testjson))
	if err != nil {
		t.Errorf("Failed to parse : %s ", err.Error())
	}
	if res == nil {
		t.Error("Return nil")
	}
	if base58.Encode(res[0].Hash) != "525mQMtsWaDLVJbzQZgTFkSG33gtZsho7m4io1HUCeJi" {
		t.Error("Failed to parse hash")
	}
	if base58.Encode(res[0].Body.Sign) != "3tMHYrizQ532D1WJkt5RSs5AcRmq7betw8zvC66Wh3XHUdvNpNzLWh1SkkGYMGJ669nCVuYHrhwfg1HrUUp6KDwzK" {
		t.Error("Failed to parse sign")
	}
	if types.EncodeAddress(res[0].Body.Account) != "AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd" {
		t.Error("Failed to parse account")
	}
	if types.EncodeAddress(res[0].Body.Recipient) != "AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR" {
		t.Error("Failed to parse recipient")
	}
}
func TestParseBase58TxBody(t *testing.T) {
	testjson := "{\"Nonce\":1,\"Account\":\"AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd\",\"Recipient\":\"AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR\",\"Amount\":25000,\"Payload\":\"aergo\",\"Limit\":100,\"Price\":1,\"Type\":0,\"Sign\":\"3roWPzztf5aLLh16vAnd2ugcPux3wJ1oqqvqkWARobjuAC32xftF42nnbTkXUQdkDaFvuUmctrpQSv8FAVUKcywHW\"}"
	res, err := ParseBase58TxBody([]byte(testjson))
	if err != nil {
		t.Errorf("Failed to parse : %s ", err.Error())
	}
	if res == nil {
		t.Error("return nil")
	}
	if base58.Encode(res.Sign) != "3roWPzztf5aLLh16vAnd2ugcPux3wJ1oqqvqkWARobjuAC32xftF42nnbTkXUQdkDaFvuUmctrpQSv8FAVUKcywHW" {
		t.Error("Failed to parse sign")
	}
	if base58.Encode(res.Payload) != "aergo" {
		t.Error("Failed to parse payload")
	}
	if types.EncodeAddress(res.Account) != "AsiFCzSukVNUGufJSzSNLA1nKx39NxKcVBEWvW3riyfixcBjN1Qd" {
		t.Error("Failed to parse account")
	}
	if types.EncodeAddress(res.Recipient) != "AsjHhFbCuULoUVZPiNNV6WEemtEi7Eiy6G4TDaUsMDiedCARbhQR" {
		t.Error("Failed to parse recipient")
	}
}
