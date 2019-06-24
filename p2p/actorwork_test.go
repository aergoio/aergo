/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/golang/mock/gomock"
)

func TestP2P_GetAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dummyPeerMeta := p2pcommon.PeerMeta{ID:dummyPeerID, IPAddress:"127.0.0.1", Port:7846}

	type args struct {
		peerID types.PeerID
		size   uint32
	}
	tests := []struct {
		name    string
		args    args
		hasPeer bool

		wantSend int
		want     bool
	}{
		{"TNormal", args{dummyPeerID, 10}, true, 1,true},
		{"TNoPeer", args{dummyPeerID, 10}, false, 0,false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockPeer := (*p2pmock.MockRemotePeer)(nil)
			if tt.hasPeer {
				mockPeer = p2pmock.NewMockRemotePeer(ctrl)
				mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)
			}
			p2pmock.NewMockRemotePeer(ctrl)
			mockPM.EXPECT().GetPeer(dummyPeerID).Return(mockPeer, tt.hasPeer).Times(1)
			mockPM.EXPECT().SelfMeta().Return(dummyPeerMeta).Times(tt.wantSend).MaxTimes(tt.wantSend)
			mockMF.EXPECT().NewMsgRequestOrder(true, p2pcommon.AddressesRequest, gomock.AssignableToTypeOf(&types.AddressesRequest{})).Times(tt.wantSend)
			p2ps := &P2P{
				pm:mockPM, mf:mockMF,
			}
			p2ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2ps, log.NewLogger("p2p.test"))

			if got := p2ps.GetAddresses(tt.args.peerID, tt.args.size); got != tt.want {
				t.Errorf("P2P.GetAddresses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestP2P_GetBlocksChunk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleMsg := &message.GetBlockChunks{GetBlockInfos: message.GetBlockInfos{ToWhom: samplePeerID}, TTL: time.Minute}

	// fail: cancel create receiver and return fail instantly
	mockPM := p2pmock.NewMockPeerManager(ctrl)
	mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false)
	mockCtx := p2pmock.NewMockContext(ctrl)
	mockCtx.EXPECT().Respond(gomock.Any()).Times(1)
	ps := &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM

	ps.GetBlocksChunk(mockCtx, sampleMsg)

	// success case
	mockPeer := p2pmock.NewMockRemotePeer(ctrl)
	mockMF := p2pmock.NewMockMoFactory(ctrl)

	mockPM.EXPECT().GetPeer(gomock.Any()).Return(mockPeer, true)
	mockCtx.EXPECT().Respond(gomock.Any()).Times(0)
	mockPeer.EXPECT().MF().Return(mockMF)
	mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)

	dummyMo = createDummyMo(ctrl)
	mockMF.EXPECT().NewMsgBlockRequestOrder(gomock.Any(), p2pcommon.GetBlocksRequest, gomock.Any()).
		Return(dummyMo).Times(1)

	ps = &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM
	ps.GetBlocksChunk(mockCtx, sampleMsg)

	//mockCtx.AssertNotCalled(t, "Respond", mock.Anything)
	// verify that receiver start working.
	//mockMF.AssertNumberOfCalls(t, "newMsgBlockRequestOrder", 1)
	//mockPeer.AssertNumberOfCalls(t, "sendMessage", 1)
}

func TestP2P_GetBlockHashByNo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()


	sampleMsg := &message.GetHashByNo{ToWhom: samplePeerID, BlockNo: uint64(111111)}

	// fail: cancel create receiver and return fail instantly
	mockPM := p2pmock.NewMockPeerManager(ctrl)
	mockCtx := p2pmock.NewMockContext(ctrl)
	mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false)
	mockCtx.EXPECT().Respond(gomock.Any()).Times(1)
	ps := &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM

	ps.GetBlockHashByNo(mockCtx, sampleMsg)

	// success case
	mockPM = p2pmock.NewMockPeerManager(ctrl)
	mockCtx = p2pmock.NewMockContext(ctrl)
	mockPeer := p2pmock.NewMockRemotePeer(ctrl)
	mockMF := p2pmock.NewMockMoFactory(ctrl)

	mockCtx.EXPECT().Respond(gomock.Any()).Times(0)
	mockPM.EXPECT().GetPeer(gomock.Any()).Return(mockPeer, true)
	mockPeer.EXPECT().MF().Return(mockMF)
	mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)

	dummyMo = createDummyMo(ctrl)
	mockMF.EXPECT().NewMsgBlockRequestOrder(gomock.Any(), p2pcommon.GetHashByNoRequest, gomock.Any()).
		Return(dummyMo).Times(1)

	ps = &P2P{}
	ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))
	ps.pm = mockPM
	ps.GetBlockHashByNo(mockCtx, sampleMsg)
}

func TestP2P_SendRaftMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		pid 	types.PeerID
		body    interface{}
	}
	tests := []struct {
		name   string
		args   args

		wantErr bool
	}{
		{"TSucc", args{samplePeerID, raftpb.Message{Type:raftpb.MsgVote}}, false },
		{"TNoPeer", args{dummyPeerID, raftpb.Message{Type:raftpb.MsgVote}}, true },
		{"TWrongBody", args{samplePeerID, types.Status{}}, true },
		//{"TNilBody", args{samplePeerID, &types.Status{}}, true },

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentCnt := 1
			if tt.wantErr {
				sentCnt = 0
			}
			dummyMo = createDummyMo(ctrl)

			mockCtx := p2pmock.NewMockContext(ctrl)
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMF.EXPECT().NewRaftMsgOrder(raftpb.MsgVote,gomock.Any() ).Return(dummyMo).MaxTimes(1)
			mockPM.EXPECT().GetPeer(gomock.Eq(samplePeerID)).Return(mockPeer, true).MaxTimes(1)
			mockPM.EXPECT().GetPeer(gomock.Not(samplePeerID)).Return(nil, false).MaxTimes(1)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(sentCnt)
			mockCtx.EXPECT().Respond(gomock.Any()).Times(1)
			ps := &P2P{pm:mockPM, mf:mockMF}
			ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))

			ps.SendRaftMessage(mockCtx, &message.SendRaft{ToWhom:tt.args.pid, Body:tt.args.body})

		})
	}
}
