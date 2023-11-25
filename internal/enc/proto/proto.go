package proto

import (
	"github.com/golang/protobuf/proto"
)

type Message = proto.Message

// Encode encodes e by using gob and returns.
func Encode(m proto.Message) ([]byte, error) {
	return proto.Marshal(m)
}

// Decode decodes a gob-encoded value v.
func Decode(v []byte, m proto.Message) error {
	return proto.Unmarshal(v, m)
}

func Size(m proto.Message) int {
	return proto.Size(m)
}

func Equal(m1, m2 proto.Message) bool {
	return proto.Equal(m1, m2)
}

func Clone(m proto.Message) proto.Message {
	return proto.Clone(m)
}
