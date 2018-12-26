/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/types"
	"github.com/gofrs/uuid"
)

type V020Wrapper struct {
	*types.P2PMessage
	originalID string
}

func NewV020Wrapper(message *types.P2PMessage, originalID string) *V020Wrapper {
	return &V020Wrapper{message, originalID}
}

func (m *V020Wrapper) Subprotocol() SubProtocol {
	return SubProtocol(m.Header.Subprotocol)
}

func (m *V020Wrapper) Length() uint32 {
	return m.Header.Length

}

func (m *V020Wrapper) Timestamp() int64 {
	return m.Header.Timestamp
}

func (m *V020Wrapper) ID() MsgID {
	return uuidStrToMsgID(m.Header.Id)
}

func (m *V020Wrapper) OriginalID()  MsgID  {
	return uuidStrToMsgID(m.originalID)
}

func (m *V020Wrapper) Payload() []byte {
	return m.Data
}

var _ Message = (*V020Wrapper)(nil)

func uuidStrToMsgID(str string) (id MsgID) {
	uuid, err := uuid.FromString(str)
	if err != nil {
		return
	}
	copy(id[:], uuid[:])
	return
}