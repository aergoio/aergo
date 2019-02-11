/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
)

// msgOrder is abstraction information about the message that will be sent to peer
// some type of msgOrder, such as notice mo, should thread-safe and re-entrant
type msgOrder interface {
	GetMsgID() p2pcommon.MsgID
	// Timestamp is unit time value
	Timestamp() int64
	IsRequest() bool
	IsNeedSign() bool
	GetProtocolID() p2pcommon.SubProtocol

	// SendTo send message to remote peer. it return err if write fails, or nil if write is successful or ignored.
	SendTo(p *remotePeerImpl) error
}

// mf is interface of factory which create mo object
type moFactory interface {
	newMsgRequestOrder(expecteResponse bool, protocolID p2pcommon.SubProtocol, message pbMessage) msgOrder
	newMsgBlockRequestOrder(respReceiver ResponseReceiver, protocolID p2pcommon.SubProtocol, message pbMessage) msgOrder
	newMsgResponseOrder(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message pbMessage) msgOrder
	newMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) msgOrder
	newMsgTxBroadcastOrder(noticeMsg *types.NewTransactionsNotice) msgOrder
	newMsgBPBroadcastOrder(noticeMsg *types.BlockProducedNotice) msgOrder
}
