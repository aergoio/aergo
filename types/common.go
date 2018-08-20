package types

import "encoding/base64"

const MAXBLOCKNO BlockNo = 18446744073709551615

func EncodeB64(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}

func DecodeB64(sb string) []byte {
	buf, _ := base64.StdEncoding.DecodeString(sb)
	return buf
}
