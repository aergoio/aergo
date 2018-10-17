package p2p

import (
	"fmt"
	"github.com/google/uuid"
)

// MsgID is
type MsgID [IDLength]byte

// NewMsgID return random id
func NewMsgID() (m MsgID) {
	uid := uuid.Must(uuid.NewRandom())
	return MsgID(uid)
}

func ParseBytesToMsgID(b []byte) (MsgID, error) {
	var m MsgID
	if b == nil || len(b) != IDLength {
		return m, fmt.Errorf("wrong format")
	}
	copy(m[:],b)
	return m, nil
}

// MustParseBytes return msgid from byte slice
func MustParseBytes(b []byte) MsgID {
	if m, err := ParseBytesToMsgID(b) ; err == nil {
		return m
	} else {
		panic(err)
	}
}

func (id MsgID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id MsgID) String() string {
	return uuid.Must(uuid.FromBytes(id[:])).String()
}
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
