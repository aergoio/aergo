package subproto

import (
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
)

// func(msg *types.P2PMessage)
// BaseMsgHandler contains common attributes of MessageHandler
type BaseMsgHandler struct {
	protocol p2pcommon.SubProtocol

	pm p2pcommon.PeerManager
	sm p2pcommon.SyncManager

	peer  p2pcommon.RemotePeer
	actor p2pcommon.ActorService

	logger    *log.Logger
	timestamp time.Time
}

func (bh *BaseMsgHandler) CheckAuth(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) error {
	// check permissions
	// or etc...

	return nil
}

func (bh *BaseMsgHandler) PreHandle() {
	bh.timestamp = time.Now()
}

func (bh *BaseMsgHandler) PostHandle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	bh.logger.Debug().
		Str("elapsed", time.Since(bh.timestamp).String()).
		Str("protocol", msg.Subprotocol().String()).
		Str("msgid", msg.ID().String()).
		Msg("handle takes")
}
