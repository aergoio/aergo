/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"time"
)

// V030Message is basic form of p2p message v0.3
type V030Message struct {
	subProtocol p2pcommon.SubProtocol
	// Length is lenght of payload
	length uint32
	// timestamp is unix time (precision of second)
	timestamp int64
	// ID is 16 bytes unique identifier
	id p2pcommon.MsgID
	// OriginalID is message id of request which trigger this message. it will be all zero, if message is request or notice.
	originalID p2pcommon.MsgID

	// marshaled by google protocol buffer v3. object is determined by Subprotocol
	payload []byte
}

// NewV030Message create a new object
func NewV030Message(msgID, originalID p2pcommon.MsgID, timestamp int64, protocol p2pcommon.SubProtocol, payload []byte) *V030Message {
	return &V030Message{id: msgID, originalID:originalID,timestamp:time.Now().UnixNano(), subProtocol:protocol,payload:payload,length:uint32(len(payload))}
}

func (m *V030Message) Subprotocol() p2pcommon.SubProtocol {
	return m.subProtocol
}

func (m *V030Message) Length() uint32 {
	return m.length

}

func (m *V030Message) Timestamp() int64 {
return m.timestamp
}

func (m *V030Message) ID() p2pcommon.MsgID {
	return m.id
}

func (m *V030Message) OriginalID() p2pcommon.MsgID {
	return m.originalID
}

func (m *V030Message) Payload() []byte {
	return m.payload
}

var _ p2pcommon.Message = (*V030Message)(nil)
