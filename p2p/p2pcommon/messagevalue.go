/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import "time"

// MessageValue is basic implementation of Message. It is used since p2p v0.3
type MessageValue struct {
	subProtocol SubProtocol
	// Length is length of payload
	length uint32
	// timestamp is unix time (precision of second)
	timestamp int64
	// ID is 16 bytes unique identifier
	id MsgID
	// OriginalID is message id of request which trigger this message. it will be all zero, if message is request or notice.
	originalID MsgID

	// marshaled by google protocol buffer v3. object is determined by Subprotocol
	payload []byte
}

// NewLiteMessageValue create MessageValue object which payload is empty
func NewLiteMessageValue(protocol SubProtocol, msgID, originalID MsgID, timestamp int64,) *MessageValue {
	return &MessageValue{id: msgID, originalID: originalID, timestamp: timestamp, subProtocol: protocol}
}

// NewMessageValue create a new object
func NewMessageValue(protocol SubProtocol, msgID, originalID MsgID, timestamp int64, payload []byte) *MessageValue {
	msg := NewLiteMessageValue(protocol, msgID, originalID, timestamp)
	msg.SetPayload(payload)
	return msg
}

func NewSimpleMsgVal(protocol SubProtocol, msgID MsgID) *MessageValue {
	return NewLiteMessageValue(protocol, msgID, EmptyID, time.Now().UnixNano())
}

func NewSimpleRespMsgVal(protocol SubProtocol, msgID MsgID, originalID MsgID) *MessageValue {
	return NewLiteMessageValue(protocol, msgID, originalID, time.Now().UnixNano())
}

func (m *MessageValue) Subprotocol() SubProtocol {
	return m.subProtocol
}

func (m *MessageValue) Length() uint32 {
	return m.length

}

func (m *MessageValue) Timestamp() int64 {
	return m.timestamp
}

func (m *MessageValue) ID() MsgID {
	return m.id
}

func (m *MessageValue) OriginalID() MsgID {
	return m.originalID
}

func (m *MessageValue) Payload() []byte {
	return m.payload
}

func (m *MessageValue) SetPayload(payload []byte) {
	m.payload = payload
	m.length = uint32(len(payload))
}

var _ Message = (*MessageValue)(nil)

