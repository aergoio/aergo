package enc

import "github.com/mr-tron/base58/base58"

// ToString returns human-readable (base64) string from b
func ToString(b []byte) string {
	return base58.Encode(b)
}

// ToBytes returns byte slice from human-readable (base64) string
func ToBytes(s string) ([]byte, error) {
	return base58.Decode(s)
}
