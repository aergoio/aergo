package util

import (
	"testing"
)

func TestParseConvBase58Tx(t *testing.T) {
	testjson := "[{\"Hash\":\"VnYaUY1Y+b3yXilhaY45Jomv974QGk8CV8jtAVUAH8E=\",\"Body\":{\"Nonce\":1,\"Account\":\"2t4pDa4qcb4rX86AJbdq3qkTY9UZ\",\"Recipient\":\"gkdhQDvLi23xxgpiLbmzodcayx3\",\"Amount\":256,\"Payload\":null,\"Sign\":\"IGhZSQU+QRXTTpFTS5ibIMYeSdy+XXKd5w7roDi17+LfFKwSB3VxSLjp0R1XFH5J5ACFSdBAqd5V06/hn4uHNLo=\"}}]"
	res, err := ParseBase58Tx([]byte(testjson))
	if err != nil {
		t.Errorf("Failed to parse : %s ", err.Error())
	}
	if res == nil {
		t.Error("Return nil")
	}
	if string(res[0].Body.Recipient) != "12345678901234567890" {
		t.Error("Failed to parse recipient")
	}
}
func TestParseBase58TxBody(t *testing.T) {
	testjson := "{\"Nonce\":1,\"Account\":\"2t4pDa4qcb4rX86AJbdq3qkTY9UZ\",\"Recipient\":\"gkdhQDvLi23xxgpiLbmzodcayx3\",\"Amount\":256,\"Payload\":null,\"Sign\":\"IGhZSQU+QRXTTpFTS5ibIMYeSdy+XXKd5w7roDi17+LfFKwSB3VxSLjp0R1XFH5J5ACFSdBAqd5V06/hn4uHNLo=\"}"
	res, err := ParseBase58TxBody([]byte(testjson))
	if err != nil {
		t.Errorf("Failed to parse : %s ", err.Error())
	}
	if res == nil {
		t.Error("return nil")
	}
	if string(res.Recipient) != "12345678901234567890" {
		t.Error("Failed to parse recipient")
	}
}
