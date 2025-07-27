package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// PrimitiveToByteArray return byte array from a primitive type. It should not be used for non-primitive types.
func PrimitiveToByteArray(data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ToByteArrayOrEmpty(data interface{}) []byte {
	byteArray, err := PrimitiveToByteArray(data)
	if err != nil {
		return make([]byte, 0)
	} else {
		return byteArray
	}
}

func ToUint64(b []byte) (uint64, error) {
	if len(b) < 8 {
		return 0, fmt.Errorf("invalid input")
	}
	return binary.LittleEndian.Uint64(b), nil
}
