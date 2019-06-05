/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

// MsgReader read byte stream, parse stream with respect to protocol version and return message object used in p2p module
type MsgReader interface {
	// ReadMsg return types.MsgHeader as header, MessageBody as data
	// The header and/or data can be nil if error is not nil
	ReadMsg() (Message, error)
}

// MsgWriter write message to stream
type MsgWriter interface {
	WriteMsg(msg Message) error
}

type MsgReadWriter interface {
	MsgReader
	MsgWriter
}
