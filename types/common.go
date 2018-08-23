package types

import (
	"encoding/base64"

	"github.com/mr-tron/base58/base58"
)

const MAXBLOCKNO BlockNo = 18446744073709551615

func EncodeB64(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}

func DecodeB64(sb string) []byte {
	buf, _ := base64.StdEncoding.DecodeString(sb)
	return buf
}

func EncodeB58(bs []byte) string {
	return base58.Encode(bs)
}

func DecodeB58(sb string) []byte {
	buf, _ := base58.Decode(sb)
	return buf
}
