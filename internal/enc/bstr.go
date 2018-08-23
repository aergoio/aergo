package enc

import "encoding/base64"

// ToString returns human-readable (base64) string from b
func ToString(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// ToBytes returns byte slice from human-readable (base64) string
func ToBytes(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
