package encoding

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/aergoio/aergo/v2/internal/enc"
)

const CodeVersion = 0xC0

func EncodeCode(code []byte) string {
	encoded, _ := enc.B58CheckEncode(fmt.Sprintf("%x", CodeVersion), hex.EncodeToString(code))
	return encoded
}

func DecodeCode(encodedCode string) ([]byte, error) {
	decodedString, err := enc.B58CheckDecode(encodedCode)
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
