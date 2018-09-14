/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

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

	//
	SendTo(p *RemotePeer) bool
	// Deprecated
	SendOver(w MsgWriter) error
}
