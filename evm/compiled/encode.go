package compiled

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type jsonCompiledContract struct {
	ABI string
	Bin HexString
}

// MarshalJSON serializes ByteArray to hex
func (s CompiledContract) MarshalJSON() ([]byte, error) {
	abi1, err := json.Marshal(s.ABI)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jsonCompiledContract{ABI: string(abi1), Bin: s.Bin})
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *CompiledContract) UnmarshalJSON(data []byte) error {
	var x jsonCompiledContract
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}

	s.Bin = x.Bin
	if err := json.Unmarshal([]byte(x.ABI), &s.ABI); err != nil {
		return fmt.Errorf("failed to unmarshal ABI: %w", err)
	}

	return nil
}

// HexString is a byte array that serializes to hex
type HexString []byte

// MarshalJSON serializes ByteArray to hex
func (s HexString) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%x", string(s)))
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *HexString) UnmarshalJSON(data []byte) error {
	var x string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	str, err := hex.DecodeString(x)
	if err != nil {
		return err
	}
	*s = str
	return nil
}
