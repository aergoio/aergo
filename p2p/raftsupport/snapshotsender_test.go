/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"bytes"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/aergoio/etcd/snap"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/pkg/errors"
)

func Test_snapshotSender_Send(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleSnaps := make([]byte, 10000)
	logger := log.NewLogger("raft.support.test")
	pid, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	sampleMeta := p2pcommon.PeerMeta{ID: pid}
	tests := []struct {
		name string

		ntErr    error
		wantSucc bool
	}{
		{"TRemoteDown", errors.New("conn fail"), false},
		{"TLaterFail", nil, false},
		// TODO : add success cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			mockRaft := p2pmock.NewMockAergoRaftAccessor(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)

			rc := &testStream{in: sampleSnaps, out: nil}
			dummyStream := &testStream{out: bytes.NewBuffer(nil)}
			mockPeer.EXPECT().ID().Return(pid).AnyTimes()
			mockPeer.EXPECT().Meta().Return(sampleMeta).AnyTimes()
			mockPeer.EXPECT().Name().Return("tester").AnyTimes()
			mockNT.EXPECT().GetOrCreateStream(sampleMeta, p2pcommon.RaftSnapSubAddr).Return(dummyStream, tt.ntErr)
			if !tt.wantSucc {
				mockRaft.EXPECT().ReportUnreachable(gomock.Any())
			}
			mockRaft.EXPECT().ReportSnapshot(gomock.Any(), gomock.Any())

			rs := raftpb.Message{}
			msg := snap.NewMessage(rs, rc, 1000)

			s := snapshotSender{nt: mockNT, logger: logger, rAcc: mockRaft, stopChan: make(chan interface{}), peer: mockPeer}

			s.Send(msg)

			if tt.ntErr != nil {
				return
			}

			// Wait for send function finished
			tick := time.NewTicker(time.Millisecond * 100)
			select {
			case r := <-msg.CloseNotify():
				if r != tt.wantSucc {
					t.Errorf("send result %v , want %v", r, tt.wantSucc)
				}
			case <-tick.C:
				t.Errorf("unexpected timeout in send")
			}

		})
	}
}

func Test_readWireHSResp(t *testing.T) {
	sampleBuf := bytes.NewBuffer(nil)
	sampleResp := types.SnapshotResponse{Status: types.ResultStatus_INVALID_ARGUMENT, Message: "wrong type"}
	(&snapshotReceiver{}).sendResp(sampleBuf, &sampleResp)
	sample := sampleBuf.Bytes()
	currupted := CopyOf(sample)
	lastidx := len(currupted) - 1
	currupted[lastidx] = currupted[lastidx] ^ 0xff

	tests := []struct {
		name string
		in   []byte

		wantErr bool
	}{
		{"TNormal", CopyOf(sample), false},
		{"TLongBody", append(CopyOf(sample), []byte("dummies")...), false},
		{"TShortBody", CopyOf(sample)[:len(sample)-1], true},
		{"TWrongHead", CopyOf(sample)[:3], true},
		//		{"TInvalidByte", CopyOf(currupted), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tt.in)

			gotResp, err := readWireHSResp(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("readWireHSResp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotResp.Status != sampleResp.Status || gotResp.Message != sampleResp.Message {
					t.Errorf("readWireHSResp() = %v, want %v", gotResp, sampleResp)
				}
			}
		})
	}
}

func CopyOf(org []byte) []byte {
	dst := make([]byte, len(org))
	copy(dst, org)
	return dst
}

type testStream struct {
	in     []byte
	out    *bytes.Buffer
	closed bool
}

func (s *testStream) CloseWrite() error {
	//TODO implement me
	panic("implement me")
}

func (s *testStream) CloseRead() error {
	//TODO implement me
	panic("implement me")
}

func (s *testStream) ID() string {
	//TODO implement me
	panic("implement me")
}

func (s *testStream) Scope() network.StreamScope {
	//TODO implement me
	panic("implement me")
}

func (s *testStream) SetProtocol(id protocol.ID) error {
	//TODO implement me
	panic("implement me")
}

func (s *testStream) Read(p []byte) (n int, err error) {
	size := copy(p, s.in)
	return size, nil
}

func (s *testStream) Write(p []byte) (n int, err error) {
	return s.out.Write(p)
}

func (s *testStream) Close() error {
	s.closed = true
	return nil
}

func (*testStream) Reset() error {
	panic("implement me")
}

func (*testStream) SetDeadline(time.Time) error {
	panic("implement me")
}

func (*testStream) SetReadDeadline(time.Time) error {
	panic("implement me")
}

func (*testStream) SetWriteDeadline(time.Time) error {
	panic("implement me")
}

func (*testStream) Protocol() protocol.ID {
	panic("implement me")
}

func (*testStream) Stat() network.Stats {
	panic("implement me")
}

func (*testStream) Conn() network.Conn {
	panic("implement me")
}
