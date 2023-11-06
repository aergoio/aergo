package base64

import "encoding/base64"

func Encode(s []byte) string {
	return base64.StdEncoding.EncodeToString(s)
}

func Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// Do not use processing real data, Only use for Logging or Testing.
func DecodeOrNil(s string) []byte {
	buf, _ := Decode(s)
	return buf
}
