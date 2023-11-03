package enc

import (
	"github.com/anaskhan96/base58check"
)

func B58CheckEncode(version string, data string) (string, error) {
	return base58check.Encode(version, data)
}

func B58CheckDecode(encoded string) (string, error) {
	return base58check.Decode(encoded)
}
