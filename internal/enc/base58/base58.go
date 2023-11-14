package base58

import (
	"github.com/mr-tron/base58/base58"
)

// Encode returns human-readable (base58) string from b. Calling with empty or nil slice returns empty string.
func Encode(b []byte) string {
	return base58.Encode(b)
}

// Decode returns byte slice from human-readable (base58) string. Calling with empty string returns zero length string error.
func Decode(s string) ([]byte, error) {
	return base58.Decode(s)
}

func DecodeOrNil(s string) []byte {
	buf, _ := Decode(s)
	return buf
}
