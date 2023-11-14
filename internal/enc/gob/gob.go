package gob

import (
	"bytes"
	"encoding/gob"
)

// Encode encodes e by using gob and returns.
func Encode(e interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(e)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode decodes a gob-encoded value v.
func Decode(v []byte, e interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(v))
	return dec.Decode(e)
}
