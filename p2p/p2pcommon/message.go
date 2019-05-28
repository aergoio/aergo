/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

//go:generate mockgen -source=message.go  -package=p2pmock -destination=../p2pmock/mock_message.go
package p2pcommon

import (
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

// Message is unit structure transferred from a peer to another peer.
type Message interface {
	Subprotocol() SubProtocol

	// Length is lentgh of payload
	Length() uint32

	// Timestamp is when this message was created with unixnano format
	Timestamp() int64

	// ID is 16 bytes unique identifier
	ID() MsgID

	// OriginalID is message id of request which trigger this message. it will be all zero, if message is request or notice.
	OriginalID() MsgID

	// Payload is MessageBody struct, marshaled by google protocol buffer v3. object is determined by Subprotocol
	Payload() []byte
}

// MessageBody is content of p2p message.
// The actual data types are varied by subprotocol, so
// For version 0.3.x, it is just wrapper of proto.Message
type MessageBody interface {
	proto.Message
}

type HandlerFactory interface {
	InsertHandlers(peer RemotePeer)
}

// MessageHandler handle incoming message
type MessageHandler interface {
	ParsePayload([]byte) (MessageBody, error)
	CheckAuth(msg Message, msgBody MessageBody) error
	Handle(msg Message, msgBody MessageBody)
	PreHandle()
	PostHandle(msg Message, msgBody MessageBody)
}

// MsgSigner sign or verify p2p message
// this is not used since v0.3, but interface is not removed for future version.
type MsgSigner interface {
	// signMsg calulate signature and fill related fields in msg(peerid, pubkey, signature or etc)
	SignMsg(msg *types.P2PMessage) error
	// verifyMsg check signature is valid
	VerifyMsg(msg *types.P2PMessage, senderID types.PeerID) error
}

// ResponseReceiver is handler function for the corresponding response message.
// It returns true when receiver handled it, or false if this receiver is not the expected handler.
type ResponseReceiver func(Message, MessageBody) bool

