package enc

import "github.com/mr-tron/base58/base58"

// ToString returns human-readable (base58) string from b. Calling with empty or nil slice returns empty string.
func ToString(b []byte) string {
	return base58.Encode(b)
}

// ToBytes returns byte slice from human-readable (base58) string. Calling with empty string returns zero length string error.
func ToBytes(s string) ([]byte, error) {
	return base58.Decode(s)
}
