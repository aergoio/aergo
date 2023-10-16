/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"bytes"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/types"
)

func TestSnapshotReceiver_sendResp(t *testing.T) {
	logger := log.NewLogger("raft.support.test")
	type args struct {
		resp *types.SnapshotResponse
	}
	tests := []struct {
		name string
		args args
	}{
		{"TOK", args{&types.SnapshotResponse{Status: types.ResultStatus_OK}}},
		{"TWrongHead", args{&types.SnapshotResponse{Status: types.ResultStatus_INVALID_ARGUMENT, Message: "wrong type"}}},
		{"TInternal", args{&types.SnapshotResponse{Status: types.ResultStatus_INTERNAL, Message: ""}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &snapshotReceiver{
				logger: logger,
			}
			w := &bytes.Buffer{}
			s.sendResp(w, tt.args.resp)

			if w.Len() < 4 {
				t.Fatalf("snapshotReceiver.sendResp() = written %v bytes, want at least 4 bytes", w.Len())
			}

			resp, err := readWireHSResp(w)
			if err != nil {
				t.Fatalf("readWireHSResp() err %v, want no error ", err.Error())
			}
			if tt.args.resp.Status != resp.Status {
				t.Fatalf("Response status %v, want %v", resp.Status.String(), tt.args.resp.Status.String())
			}
			if tt.args.resp.Message != resp.Message {
				t.Fatalf("Response message %v, want %v", resp.Message, tt.args.resp.Message)
			}
		})
	}
}
