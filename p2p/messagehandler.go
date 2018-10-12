package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/golang/protobuf/proto"
)

// MessageHandler handle incoming subprotocol message
type MessageHandler interface {
	parsePayload([]byte) (proto.Message, error)
	checkAuth(msgHeader Message, msgBody proto.Message) error
	handle(msgHeader Message, msgBody proto.Message)
}

// func(msg *types.P2PMessage)

// BaseMsgHandler contains common attributes of MessageHandler
type BaseMsgHandler struct {
	protocol SubProtocol

	pm     PeerManager
	sm     SyncManager

	peer   RemotePeer
	actor  ActorService

	logger *log.Logger

	prototype proto.Message
}

func (bh *BaseMsgHandler) checkAuth(msg Message, msgBody proto.Message) error {
	// check permissions
	// or etc...

	return nil
}
