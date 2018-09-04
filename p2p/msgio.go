package p2p

import (
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

type MsgReader interface {
	Read() (*types.MessageData, proto.Message, error)
}
