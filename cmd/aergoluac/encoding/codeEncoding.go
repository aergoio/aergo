package encoding

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/anaskhan96/base58check"
)

const CodeVersion = 0xC0

func EncodeCode(code []byte) string {
	encoded, _ := base58check.Encode(fmt.Sprintf("%x", CodeVersion), hex.EncodeToString(code))
	return encoded
}

func DecodeCode(encodedCode string) ([]byte, error) {
	decodedString, err := base58check.Decode(encodedCode)
	if err != nil {
		return nil, err
	}
	decodedBytes, err := hex.DecodeString(decodedString)
	if err != nil {
		return nil, err
	}
	version := decodedBytes[0]
	if version != CodeVersion {
		return nil, errors.New("Invalid code version")
	}
	decoded := decodedBytes[1:]
	return decoded, nil
}
