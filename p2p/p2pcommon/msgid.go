/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"fmt"
	"github.com/gofrs/uuid"
)

// MsgID is
type MsgID [IDLength]byte

// NewMsgID return random id
func NewMsgID() (m MsgID) {
	uid := uuid.Must(uuid.NewV4())
	return MsgID(uid)
}

var (
	EmptyID = MsgID(uuid.Nil)
)

func ParseBytesToMsgID(b []byte) (MsgID, error) {
	var m MsgID
	if b == nil || len(b) != IDLength {
		return m, fmt.Errorf("wrong format")
	}
	copy(m[:], b)
	return m, nil
}

// MustParseBytes return msgid from byte slice
func MustParseBytes(b []byte) MsgID {
	if m, err := ParseBytesToMsgID(b); err == nil {
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
