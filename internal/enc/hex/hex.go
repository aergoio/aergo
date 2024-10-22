package hex

import "encoding/hex"

func Encode(b []byte) string {
	return hex.EncodeToString(b)
}

func Decode(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func IsHexString(s string) bool {
	// check is the input has even number of characters
	if len(s)%2 != 0 {
		return false
	}
	// check if the input contains only hex characters
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
