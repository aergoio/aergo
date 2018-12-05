package common

import (
	"bytes"
	"encoding/gob"
)

// IsZero returns true if argument is empty or zero
func IsZero(argv []byte) bool {
	if len(argv) == 0 {
		return true
	}
	for i := range argv {
		if argv[i] != 0x00 {
			return false
		}
	}
	return true
}

// Compactz returns nil if argument is empty or zero
func Compactz(argv []byte) []byte {
	if IsZero(argv) {
		return nil
	}
	return argv
}

// GobEncode encodes e by using gob and returns.
func GobEncode(e interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(e)

	return buf.Bytes(), err
}

// GobDecode decodes a gob-encoded value v.
func GobDecode(v []byte, e interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(v))
	return dec.Decode(e)
}
