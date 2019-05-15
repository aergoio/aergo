/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"reflect"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/magiconair/properties/assert"
)

func TestParseBytesToMsgID(t *testing.T) {
	sampleUUID := uuid.Must(uuid.NewV4())
	tests := []struct {
		name      string
		in        []byte
		expectErr bool
	}{
		{"TSucc", sampleUUID[:], false},
		{"TWrongSize", sampleUUID[:15], true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseBytesToMsgID(test.in)
			assert.Equal(t, test.expectErr, err != nil, "parse byte")

			got2, gotPanic := checkPanic(test.in)
			assert.Equal(t, test.expectErr, gotPanic, "got panic")
			if !test.expectErr && got != got2 {
				t.Errorf("ParseBytes() and MustParse() were differ: %v , %v", got, got2)
			}

		})
	}
}

func checkPanic(in []byte) (msg MsgID, gotPanic bool) {
	defer func() {
		if r := recover(); r != nil {
			gotPanic = true
		}
	}()

	msg = MustParseBytes(in)
	return
}

func TestNewMsgID(t *testing.T) {
	idMap := make(map[string]MsgID)
	for i := 0; i < 100; i++ {
		gotM := NewMsgID()
		if _, exist := idMap[gotM.String()]; exist {
			t.Errorf("NewMsgID() made duplication = %v", gotM.String())
			t.FailNow()
		}
	}
}

func TestMsgID_UUID(t *testing.T) {
	tests := []struct {
		name string
		id   MsgID
		want uuid.UUID
	}{
		{"TEmpty", EmptyID, uuid.FromBytesOrNil(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.UUID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MsgID.UUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
