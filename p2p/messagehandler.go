package p2p

import (
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/golang/protobuf/proto"
)

// MessageHandler handle incoming subprotocol message
type MessageHandler interface {
	parsePayload([]byte) (proto.Message, error)
	checkAuth(msgHeader Message, msgBody proto.Message) error
	handle(msgHeader Message, msgBody proto.Message)
	preHandle()
	postHandle(msgHeader Message, msgBody proto.Message)
}

// func(msg *types.P2PMessage)

// BaseMsgHandler contains common attributes of MessageHandler
type BaseMsgHandler struct {
	protocol SubProtocol

	pm PeerManager
	sm SyncManager

	peer  RemotePeer
	actor ActorService

	logger    *log.Logger
	timestamp time.Time
	prototype proto.Message
}

func (bh *BaseMsgHandler) checkAuth(msg Message, msgBody proto.Message) error {
	// check permissions
	// or etc...

	return nil
}

func (bh *BaseMsgHandler) preHandle() {
	bh.timestamp = time.Now()
}

func (bh *BaseMsgHandler) postHandle(msg Message, msgBody proto.Message) {
	bh.logger.Debug().
		Str("elapsed", time.Since(bh.timestamp).String()).
		Str("protocol", msg.Subprotocol().String()).
		Str("msgid", msg.ID().String()).
		Msg("handle takes")
}
