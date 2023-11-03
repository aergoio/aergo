package enc

import "encoding/hex"

func HexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

func HexDecode(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
