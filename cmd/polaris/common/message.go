/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package common

import (
	"time"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
)

// PolarisMessage is data struct for transferring between polaris server and client.
// as of 2019.04.23, this is copy of MessageValue.
type PolarisMessage struct {
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

// NewPolarisMessage create a new object
func NewPolarisMessage(msgID p2pcommon.MsgID, protocol p2pcommon.SubProtocol, payload []byte) *PolarisMessage {
	return &PolarisMessage{id: msgID, timestamp: time.Now().UnixNano(), subProtocol: protocol, payload: payload, length: uint32(len(payload))}
}
func NewPolarisRespMessage(msgID, orgReqID p2pcommon.MsgID, protocol p2pcommon.SubProtocol, payload []byte) *PolarisMessage {
	return &PolarisMessage{id: msgID, originalID: orgReqID, timestamp: time.Now().UnixNano(), subProtocol: protocol, payload: payload, length: uint32(len(payload))}
}

func (m *PolarisMessage) Subprotocol() p2pcommon.SubProtocol {
	return m.subProtocol
}

func (m *PolarisMessage) Length() uint32 {
	return m.length

}

func (m *PolarisMessage) Timestamp() int64 {
	return m.timestamp
}

func (m *PolarisMessage) ID() p2pcommon.MsgID {
	return m.id
}

func (m *PolarisMessage) OriginalID() p2pcommon.MsgID {
	return m.originalID
}

func (m *PolarisMessage) Payload() []byte {
	return m.payload
}

var _ p2pcommon.Message = (*PolarisMessage)(nil)
