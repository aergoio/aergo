/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

// MsgReadWriter read byte stream, parse stream with respect to protocol version and return message object used in p2p module
// It also write Message to stream with serialized form and have Close() to close underlying io stream.
// The implementations should be safe for concurrent read and write, but not concurrent reads or writes.
type MsgReadWriter interface {
	ReadMsg() (Message, error)
	WriteMsg(msg Message) error
	Close() error

	AddIOListener(l MsgIOListener)
}

// MsgIOListener listen read and write of p2p message. The concrete implementations must consume much of times.
type MsgIOListener interface {
	OnRead(protocol SubProtocol, read int)
	OnWrite(protocol SubProtocol, write int)
}

//go:generate mockgen -source=msgio.go -package=p2pmock -destination=../p2pmock/mock_msgio.go
