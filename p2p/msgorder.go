/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import "github.com/aergoio/aergo/types"

// msgOrder is abstraction information about the message that will be sent to peer
// some type of msgOrder, such as notice mo, should thread-safe and re-entrant
type msgOrder interface {
	GetMsgID() string
	// Timestamp is unit time value
	Timestamp() int64
	IsRequest() bool
	IsGossip() bool
	IsNeedSign() bool
	// ResponseExpected means that remote peer is expected to send response to this request.
	ResponseExpected() bool
	GetProtocolID() SubProtocol

	//
	Skippable() bool

	// send message to remote peer
	SendTo(p *RemotePeer) bool
}

// mf is interface of factory which create mo object
type moFactory interface {
	newMsgRequestOrder(expecteResponse bool, protocolID SubProtocol, message pbMessage) msgOrder
	newMsgResponseOrder(reqID string, protocolID SubProtocol, message pbMessage) msgOrder
	newMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) msgOrder
	newMsgTxBroadcastOrder(message *types.NewTransactionsNotice) msgOrder
}
