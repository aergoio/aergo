package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
)

// MessageHandler handle incoming subprotocol message
type MessageHandler func(msg *types.P2PMessage)

// BaseMsgHandler contains common attributes of MessageHandler
type BaseMsgHandler struct {
	protocol SubProtocol

	pm    PeerManager
	peer  *RemotePeer
	actor ActorService

	logger *log.Logger
}
