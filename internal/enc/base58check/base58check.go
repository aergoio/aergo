package base58check

import (
	"github.com/anaskhan96/base58check"
)

func Encode(version string, data string) (string, error) {
	return base58check.Encode(version, data)
}

func EncodeOrNil(version string, data string) string {
	buf, _ := Encode(version, data)
	return buf
}

func Decode(encoded string) (string, error) {
	return base58check.Decode(encoded)
}

func DecodeOrNil(encoded string) string {
	buf, _ := Decode(encoded)
	return buf
}
