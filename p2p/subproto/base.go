package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
)

// func(msg *types.P2PMessage)
// BaseMsgHandler contains common attributes of MessageHandler
type BaseMsgHandler struct {
	protocol p2pcommon.SubProtocol

	pm p2pcommon.PeerManager
	sm p2pcommon.SyncManager

	peer  p2pcommon.RemotePeer
	actor p2pcommon.ActorService

	logger *log.Logger

	advice  []p2pcommon.HandlerAdvice
	advSize int
}

func (bh *BaseMsgHandler) CheckAuth(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) error {
	// check permissions
	// or etc...

	return nil
}

func (bh *BaseMsgHandler) AddAdvice(advice p2pcommon.HandlerAdvice) {
	bh.advice = append(bh.advice, advice)
	bh.advSize = len(bh.advice)
}

func (bh *BaseMsgHandler) PreHandle() {
	for i := bh.advSize - 1; i >= 0; i-- {
		bh.advice[i].PreHandle()
	}
}

func (bh *BaseMsgHandler) PostHandle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	for i := 0; i < bh.advSize; i++ {
		bh.advice[i].PostHandle(msg, msgBody)
	}
}

// asyncHelper is
type asyncHelper struct {
	w chan int
}

func newAsyncHelper() asyncHelper {
	h := asyncHelper{w: make(chan int, 1)}
	h.w <- 1
	return h
}

func (th *asyncHelper) issue() bool {
	// temporarily disable exclusive locking
	return true
	//select {
	//case <-th.w:
	//	return true
	//default:
	//	return false
	//}
}
func (th *asyncHelper) release() {
	// temporarily disable exclusive locking
	//th.w <- 1
}
