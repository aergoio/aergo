package enc

import (
	"bytes"
	"encoding/gob"
)

// GobEncode encodes e by using gob and returns.
func GobEncode(e interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(e)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GobDecode decodes a gob-encoded value v.
func GobDecode(v []byte, e interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(v))
	return dec.Decode(e)
}
