/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftsupport

import (
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/consensus/impl/raftv2"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/aergoio/etcd/snap"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func TestAergoRaftTransport_SendSnapshot(t *testing.T) {
	// SendSnapshot acquire snap.Message, which must be closed after using.

	logger := log.NewLogger("raft.support.test")
	dummyChainID := make([]byte, 32)
	dummyPeerID := types.RandomPeerID()
	dummyPeerMeta := p2pcommon.PeerMeta{ID: dummyPeerID}
	dummyMemID := uint64(112345531252)
	dummyMember := &consensus.Member{MemberAttr: types.MemberAttr{ID: dummyMemID, PeerID: []byte(dummyPeerID)}}

	type mockOPs struct {
		toID      uint64
		raMem     *consensus.Member
		pmGetPeer bool
		sr        error // snapshotSender result
	}
	tests := []struct {
		name string

		mOP mockOPs

		wantResult bool
	}{
		{"TSucc", mockOPs{toID: dummyMemID, raMem: dummyMember, pmGetPeer: true}, true},
		{"TWrongID", mockOPs{toID: 0, raMem: dummyMember, pmGetPeer: true}, false},
		{"TInvalidMemberID", mockOPs{toID: dummyMemID, raMem: nil, pmGetPeer: true}, false},
		{"TNotConnPeer", mockOPs{toID: dummyMemID, raMem: dummyMember, pmGetPeer: false}, false},
		{"TTransportErr", mockOPs{toID: dummyMemID, raMem: dummyMember, pmGetPeer: true, sr: errors.New("transport error")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockCA := p2pmock.NewMockConsensusAccessor(ctrl)
			mockRA := p2pmock.NewMockAergoRaftAccessor(ctrl)
			mockRC := p2pmock.NewMockReadCloser(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			dummyCl := raftv2.NewCluster(dummyChainID, nil, "test", dummyPeerID, 0, nil)

			// not checked mock operations
			mockCA.EXPECT().RaftAccessor().Return(mockRA).AnyTimes()
			mockPM.EXPECT().AddPeerEventListener(gomock.Any()).AnyTimes()
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return(dummyPeerID.ShortString()).AnyTimes()
			mockPeer.EXPECT().Meta().Return(dummyPeerMeta).AnyTimes()
			mockRC.EXPECT().Read(gomock.Any()).DoAndReturn(func(buf []byte) (int, error) {
				return len(buf), nil
			}).AnyTimes()

			// checked mock operations
			// close must be called in any cases
			mockRC.EXPECT().Close().Times(1)
			if tt.mOP.toID != 0 {
				mockRA.EXPECT().GetMemberByID(tt.mOP.toID).Return(tt.mOP.raMem).Times(1)
				if tt.mOP.raMem != nil {
					if tt.mOP.pmGetPeer {
						mockPM.EXPECT().GetPeer(dummyPeerID).Return(mockPeer, true).Times(1)
					} else {
						mockPM.EXPECT().GetPeer(dummyPeerID).Return(nil, false).Times(1)
						mockRA.EXPECT().ReportUnreachable(dummyPeerID)
						mockRA.EXPECT().ReportSnapshot(dummyPeerID, raft.SnapshotFailure)
					}
				}
			}
			// emulate snapshotSender
			rs := raftpb.Message{To: tt.mOP.toID}
			msg := snap.NewMessage(rs, mockRC, 1000)
			ssf := &snapStubFactory{serr: tt.mOP.sr, rsize: rs.Size() + 1000}

			target := NewAergoRaftTransport(logger, mockNT, mockPM, mockMF, mockCA, dummyCl)
			target.snapF = ssf

			target.SendSnapshot(*msg)

			timer := time.NewTimer(time.Millisecond * 100)
			select {
			case r := <-msg.CloseNotify():
				if r != tt.wantResult {
					t.Errorf("close result %v , want %v", r, tt.wantResult)
				}
			case <-timer.C:
				t.Error("unexpected timeout")
			}
		})
	}
}

func TestAergoRaftTransport_NewSnapshotSender(t *testing.T) {
	logger := log.NewLogger("raft.support.test")
	dummyChainID := make([]byte, 32)
	dummyPeerID := types.RandomPeerID()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNT := p2pmock.NewMockNetworkTransport(ctrl)
	mockPM := p2pmock.NewMockPeerManager(ctrl)
	mockMF := p2pmock.NewMockMoFactory(ctrl)
	mockCA := p2pmock.NewMockConsensusAccessor(ctrl)
	mockRA := p2pmock.NewMockAergoRaftAccessor(ctrl)
	mockRWC := p2pmock.NewMockReadWriteCloser(ctrl)
	mockPeer := p2pmock.NewMockRemotePeer(ctrl)
	dummyCl := raftv2.NewCluster(dummyChainID, nil, "test", dummyPeerID, 0, nil)
	// not checked mock operations
	mockCA.EXPECT().RaftAccessor().Return(mockRA).AnyTimes()
	mockPM.EXPECT().AddPeerEventListener(gomock.Any()).AnyTimes()

	target := NewAergoRaftTransport(logger, mockNT, mockPM, mockMF, mockCA, dummyCl)
	got := target.NewSnapshotSender(mockPeer)
	if _, ok := got.(*snapshotSender); !ok {
		t.Errorf("AergoRaftTransport.NewSnapshotSender() type is differ: %v", reflect.TypeOf(got).Name())
	} else {
		if got.(*snapshotSender).peer != mockPeer {
			t.Errorf("AergoRaftTransport.NewSnapshotSender() assign failed")
		}
	}

	got2 := target.NewSnapshotReceiver(mockPeer, mockRWC)
	if _, ok := got2.(*snapshotReceiver); !ok {
		t.Errorf("AergoRaftTransport.NewSnapshotSender() type is differ: %v", reflect.TypeOf(got).Name())
	} else {
		if got2.(*snapshotReceiver).peer != mockPeer {
			t.Errorf("AergoRaftTransport.NewSnapshotSender() assign failed")
		}
	}
}

type snapStubFactory struct {
	rsize int
	serr  error
}

func (f snapStubFactory) NewSnapshotSender(peer p2pcommon.RemotePeer) SnapshotSender {
	return &snapSenderStub{f.rsize, f.serr}
}

func (snapStubFactory) NewSnapshotReceiver(peer p2pcommon.RemotePeer, rwc io.ReadWriteCloser) SnapshotReceiver {
	return &snapReceiverStub{}
}

type snapSenderStub struct {
	rsize int
	err   error
}

func (s snapSenderStub) Send(snapMsg *snap.Message) {
	if s.rsize > 0 {
		buf := make([]byte, s.rsize)
		snapMsg.ReadCloser.Read(buf)
	}
	snapMsg.CloseWithError(s.err)
}

type snapReceiverStub struct {
}

func (r snapReceiverStub) Receive() {
}

func TestAergoRaftTransport_Send(t *testing.T) {
	// SendSnapshot acquire snap.Message, which must be closed after using.

	logger := log.NewLogger("raft.support.test")
	dummyChainID := make([]byte, 32)
	dummyPeerID := types.RandomPeerID()
	dummyPeerMeta := p2pcommon.PeerMeta{ID: dummyPeerID}
	dummyMemID := uint64(11111)
	dummyMember := &consensus.Member{MemberAttr: types.MemberAttr{ID: dummyMemID, PeerID: []byte(dummyPeerID)}}
	unreachableMemID := uint64(33333)
	unreachablePeerID := types.RandomPeerID()
	unreachableMember := &consensus.Member{MemberAttr: types.MemberAttr{ID: unreachableMemID, PeerID: []byte(unreachablePeerID)}}

	zeroM := raftpb.Message{To: 0, Type: raftpb.MsgApp}
	notM := raftpb.Message{To: 98767, Type: raftpb.MsgApp}
	memM := raftpb.Message{To: dummyMemID, Type: raftpb.MsgApp}
	unM := raftpb.Message{To: unreachableMemID, Type: raftpb.MsgApp}
	type args struct {
		toID      uint64
		raMem     *consensus.Member
		pmGetPeer bool
		sr        error // snapshotSender result
	}
	tests := []struct {
		name string
		msgs []raftpb.Message

		wantChkMemCnt  int
		wantGetPeerCnt int
		wantSendCnt    int
		wantUnreachCnt int
		wantResult     bool
	}{
		{"TSingle", ToM(memM), 1, 1, 1, 0, true},
		{"TMulti", ToM(memM, memM, memM), 3, 3, 3, 0, true},
		{"TWZero", ToM(memM, zeroM, zeroM, memM), 2, 2, 2, 0, true},
		{"TInvalidM", ToM(notM, notM), 2, 0, 0, 0, true},
		{"TUnreachable", ToM(unM, unM, memM), 3, 3, 1, 2, true},

		//{"TWrongID", args{toID: 0, raMem: dummyMember, pmGetPeer: true}, false},
		//{"TInvalidMemberID", args{toID: dummyMemID, raMem: nil, pmGetPeer: true}, false},
		//{"TNotConnPeer", args{toID: dummyMemID, raMem: dummyMember, pmGetPeer: false}, false},
		//{"TTransportErr", args{toID: dummyMemID, raMem: dummyMember, pmGetPeer: true, sr: errors.New("transport error")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockCA := p2pmock.NewMockConsensusAccessor(ctrl)
			mockRA := p2pmock.NewMockAergoRaftAccessor(ctrl)
			mockRC := p2pmock.NewMockReadCloser(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			dummyCl := raftv2.NewCluster(dummyChainID, nil, "test", dummyPeerID, 0, nil)
			dummyMO := p2pmock.NewMockMsgOrder(ctrl)

			// not checked mock operations
			mockCA.EXPECT().RaftAccessor().Return(mockRA).AnyTimes()
			mockPM.EXPECT().AddPeerEventListener(gomock.Any()).AnyTimes()
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return(dummyPeerID.ShortString()).AnyTimes()
			mockPeer.EXPECT().Meta().Return(dummyPeerMeta).AnyTimes()
			mockRC.EXPECT().Read(gomock.Any()).DoAndReturn(func(buf []byte) (int, error) {
				return len(buf), nil
			}).AnyTimes()

			// checked mock operations
			mockRA.EXPECT().GetMemberByID(gomock.Any()).DoAndReturn(func(id uint64) *consensus.Member {
				if id == dummyMemID {
					return dummyMember
				} else if id == unreachableMemID {
					return unreachableMember
				} else {
					return nil
				}
			}).Times(tt.wantChkMemCnt)
			mockPM.EXPECT().GetPeer(gomock.Any()).DoAndReturn(func(pid types.PeerID) (p2pcommon.RemotePeer, bool) {
				if pid == dummyPeerID {
					return mockPeer, true
				} else {
					return nil, false
				}
			}).MaxTimes(tt.wantGetPeerCnt)
			mockMF.EXPECT().NewRaftMsgOrder(gomock.Any(), gomock.Any()).Return(dummyMO).Times(tt.wantSendCnt)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(tt.wantSendCnt)
			mockRA.EXPECT().ReportUnreachable(unreachablePeerID).Times(tt.wantUnreachCnt)

			target := NewAergoRaftTransport(logger, mockNT, mockPM, mockMF, mockCA, dummyCl)

			target.Send(tt.msgs)

		})
	}
}

func ToM(ms ...raftpb.Message) []raftpb.Message {
	return ms
}
