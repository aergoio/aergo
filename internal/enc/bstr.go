package enc

import "github.com/mr-tron/base58/base58"

// ToString returns human-readable (base58) string from b. Nil parameter return empty string, but not inversable; i.e. calling ToBytes() with empty string parameter will return zero length string error.
func ToString(b []byte) string {
	return base58.Encode(b)
}

// ToBytes returns byte slice from human-readable (base58) string
func ToBytes(s string) ([]byte, error) {
	return base58.Decode(s)
}
