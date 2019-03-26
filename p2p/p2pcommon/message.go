/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

//
type Message interface {
	Subprotocol() SubProtocol

	// Length is lenght of payload
	Length() uint32
	Timestamp() int64
	// ID is 16 bytes unique identifier
	ID() MsgID
	// OriginalID is message id of request which trigger this message. it will be all zero, if message is request or notice.
	OriginalID() MsgID

	// marshaled by google protocol buffer v3. object is determined by Subprotocol
	Payload() []byte
}
