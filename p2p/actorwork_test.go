/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/golang/mock/gomock"
)

func TestP2P_GetAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ma, _ := types.ParseMultiaddr("/ip4/127.0.0.1/tcp/7846")
	dummyPeerMeta := p2pcommon.PeerMeta{ID: dummyPeerID, Addresses: []types.Multiaddr{ma}}

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
		{"TNormal", args{dummyPeerID, 10}, true, 1, true},
		{"TNoPeer", args{dummyPeerID, 10}, false, 0, false},
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
			//mockPM.EXPECT().SelfMeta().Return(dummyPeerMeta).Times(tt.wantSend).MaxTimes(tt.wantSend)
			mockMF.EXPECT().NewMsgRequestOrder(true, p2pcommon.AddressesRequest, gomock.AssignableToTypeOf(&types.AddressesRequest{})).Times(tt.wantSend)
			p2ps := &P2P{
				pm: mockPM, mf: mockMF, selfMeta: dummyPeerMeta,
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
	mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), p2pcommon.GetBlocksRequest, gomock.Any()).
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
	mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), p2pcommon.GetHashByNoRequest, gomock.Any()).
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
		pid  types.PeerID
		body interface{}
	}
	tests := []struct {
		name string
		args args

		wantErr    bool
		wantReport bool
	}{
		{"TSucc", args{samplePeerID, raftpb.Message{Type: raftpb.MsgVote}}, false, false},
		{"TNoPeer", args{dummyPeerID, raftpb.Message{Type: raftpb.MsgVote}}, true, true},
		{"TWrongBody", args{samplePeerID, types.Status{}}, true, false},
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
			mockConsAcc := p2pmock.NewMockConsensusAccessor(ctrl)
			mockRaftAcc := p2pmock.NewMockAergoRaftAccessor(ctrl)
			if tt.wantReport {
				mockConsAcc.EXPECT().RaftAccessor().Return(mockRaftAcc)
				mockRaftAcc.EXPECT().ReportUnreachable(tt.args.pid)
			}

			mockMF.EXPECT().NewRaftMsgOrder(raftpb.MsgVote, gomock.Any()).Return(dummyMo).MaxTimes(1)
			mockPM.EXPECT().GetPeer(gomock.Eq(samplePeerID)).Return(mockPeer, true).MaxTimes(1)
			mockPM.EXPECT().GetPeer(gomock.Not(samplePeerID)).Return(nil, false).MaxTimes(1)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(sentCnt)
			ps := &P2P{pm: mockPM, mf: mockMF, consacc: mockConsAcc}
			ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, ps, log.NewLogger("p2p"))

			ps.SendRaftMessage(mockCtx, &message.SendRaft{ToWhom: tt.args.pid, Body: tt.args.body})

		})
	}
}

func TestP2P_NotifyBlockProduced(t *testing.T) {
	rp, ra, rw := types.PeerRole_Producer, types.PeerRole_Agent, types.PeerRole_Watcher
	sr, ss := types.RUNNING, types.STOPPING

	tests := []struct {
		name    string
		argPeer []rs

		wantSend int
	}{
		{"TAllRun", []rs{{rp, sr}, {rp, sr}, {rp, sr}}, 3},
		{"TAllStop", []rs{{rw, ss}, {ra, ss}, {rw, ss}}, 0},
		{"TMix", []rs{{rp, sr}, {rp, ss}, {rw, sr}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			selfMeta := p2pcommon.NewMetaWith1Addr(samplePeerID, "192.168.1.2", 7846, "v2.0.0")
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockRM := p2pmock.NewMockPeerRoleManager(ctrl)
			p2ps := &P2P{
				pm: mockPM, mf: mockMF, selfMeta: selfMeta, prm: mockRM,
			}
			p2ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2ps, log.NewLogger("p2p.test"))

			dummyNotice := message.NotifyNewBlock{Block: &types.Block{Hash: []byte(types.RandomPeerID())}}

			sentCnt := 0
			mockPeers := make([]p2pcommon.RemotePeer, 0, 3)
			for _, ap := range tt.argPeer {
				mPeer := p2pmock.NewMockRemotePeer(ctrl)
				mPeer.EXPECT().ID().Return(types.RandomPeerID()).AnyTimes()
				mPeer.EXPECT().AcceptedRole().Return(ap.r).AnyTimes()
				mPeer.EXPECT().State().Return(ap.s).AnyTimes()
				mPeer.EXPECT().SendMessage(gomock.Any()).Do(func(_ interface{}) {
					sentCnt++
				}).MaxTimes(1)

				mockPeers = append(mockPeers, mPeer)
			}
			mockPM.EXPECT().GetPeers().Return(mockPeers)
			mockMF.EXPECT().NewMsgBPBroadcastOrder(gomock.AssignableToTypeOf(&types.BlockProducedNotice{}))

			_ = p2ps.NotifyBlockProduced(dummyNotice)
			if sentCnt != tt.wantSend {
				t.Errorf("P2P.NotifyBlockProduced() sent count = %v, want %v", sentCnt, tt.wantSend)
			}
		})
	}
}
