package enc

import "github.com/mr-tron/base58/base58"

// B58Encode returns human-readable (base58) string from b. Calling with empty or nil slice returns empty string.
func B58Encode(b []byte) string {
	return base58.Encode(b)
}

// B58Decode returns byte slice from human-readable (base58) string. Calling with empty string returns zero length string error.
func B58Decode(s string) ([]byte, error) {
	return base58.Decode(s)
}
