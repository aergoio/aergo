/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

// MsgReadWriter read byte stream, parse stream with respect to protocol version and return message object used in p2p module
// It also write Message to stream with serialized form and have Close() to close underlying io stream.
type MsgReadWriter interface {
	ReadMsg() (Message, error)
	WriteMsg(msg Message) error
	Close() error
}

//go:generate mockgen -source=msgio.go -package=p2pmock -destination=../p2pmock/mock_msgio.go
