/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import "github.com/aergoio/aergo/p2p/p2pcommon"

// MsgReader read stream and return message object
type MsgReader interface {
	// ReadMsg return types.MsgHeader as header, proto.Message as data
	// The header and/or data can be nil if error is not nil
	ReadMsg() (p2pcommon.Message, error)
}

// MsgWriter write message to stream
type MsgWriter interface {
	WriteMsg(msg p2pcommon.Message) error
}

type MsgReadWriter interface {
	MsgReader
	MsgWriter
}
