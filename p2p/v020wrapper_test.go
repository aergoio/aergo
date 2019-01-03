/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/gofrs/uuid"
	"reflect"
	"testing"
)

func Test_uuidStrToMsgID(t *testing.T) {
	sampleMsgID := NewMsgID()
	tests := []struct {
		name   string
		str string
		wantId MsgID
	}{
		{"TSucc", sampleMsgID.String(), sampleMsgID},
		{"TEmpty", "00000000-0000-0000-0000-000000000000", MsgID(uuid.Nil) },
		{"TFail1",  "AL000000-0000-0000-0000-000000000000", MsgID(uuid.Nil) },
		{"TFail2",  "It's not an uuid", MsgID(uuid.Nil) },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotId := uuidStrToMsgID(tt.str); !reflect.DeepEqual(gotId, tt.wantId) {
				t.Errorf("uuidStrToMsgID() = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}
