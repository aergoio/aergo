package p2p

import (
	"fmt"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

// MessageHandler handle incoming subprotocol message
type MessageHandler interface {
	parsePayload([]byte) (proto.Message, error)
	checkAuth(msgHeader *types.MsgHeader, msgBody proto.Message) error
	handle(msgHeader *types.MsgHeader, msgBody proto.Message)
}

// func(msg *types.P2PMessage)

// BaseMsgHandler contains common attributes of MessageHandler
type BaseMsgHandler struct {
	protocol SubProtocol

	pm    PeerManager
	peer  *RemotePeer
	actor ActorService

	logger *log.Logger

	prototype proto.Message
}

func (bh *BaseMsgHandler) checkAuth(msgHeader *types.MsgHeader, msgBody proto.Message) error {
	valid := bh.pm.AuthenticateMessage(msgBody, msgHeader)
	if !valid {
		return fmt.Errorf("Failed to authenticate message")
	}
	return nil
}
