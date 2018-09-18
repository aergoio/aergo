package p2p

import (
	"github.com/aergoio/aergo/types"
)

// MsgReader read stream and return message object
type MsgReader interface {
	// ReadMsg return types.MsgHeader as header, proto.Message as data
	// The header and/or data can be nil if error is not nil
	ReadMsg() (*types.P2PMessage, error)
}

// MsgWriter write message to stream
type MsgWriter interface {
	WriteMsg(header *types.P2PMessage) error
}

type MsgReadWriter interface {
	MsgReader
	MsgWriter
}
