package hex

import "encoding/hex"

func Encode(b []byte) string {
	return hex.EncodeToString(b)
}

func Decode(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
