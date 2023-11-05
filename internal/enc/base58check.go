package enc

import (
	"github.com/anaskhan96/base58check"
)

func B58CheckEncode(version string, data string) (string, error) {
	return base58check.Encode(version, data)
}

func B58CheckEncodeOrNil(version string, data string) string {
	buf, _ := B58CheckEncode(version, data)
	return buf
}

func B58CheckDecode(encoded string) (string, error) {
	return base58check.Decode(encoded)
}

func B58CheckDecodeOrNil(encoded string) string {
	buf, _ := B58CheckDecode(encoded)
	return buf
}
