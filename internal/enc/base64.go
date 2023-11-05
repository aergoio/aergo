package enc

import "encoding/base64"

func B64Encode(s []byte) string {
	return base64.StdEncoding.EncodeToString(s)
}

func B64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// Do not use processing real data, Only use for Logging or Testing.
func B64DecodeOrNil(s string) []byte {
	buf, _ := B64Decode(s)
	return buf
}
