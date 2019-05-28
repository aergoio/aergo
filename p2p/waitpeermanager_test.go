/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"errors"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/types"
	"testing"
	"time"

	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/crypto"
)

const (
	OneDay = time.Hour * 24
)

func Test_staticWPManager_OnDiscoveredPeers(t *testing.T) {
	ctrl := gomock.NewController(t)

	type args struct {
		metas []p2pcommon.PeerMeta
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{"TSingleDesign", args{desigPeers[:1]}, 0},
		{"TAllDesign", args{desigPeers}, 0},
		{"TNewID", args{unknowPeers}, 0},
		{"TMixedIDs", args{append(unknowPeers[:5], desigPeers[:5]...)}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := createDummyPM()
			mockActor := p2pmock.NewMockActorService(ctrl)
			dp := NewWaitingPeerManager(logger, dummyPM, mockActor, 10, false, false).(*staticWPManager)

			dp.OnDiscoveredPeers(tt.args.metas)
			if len(dummyPM.waitingPeers) != tt.wantCount {
				t.Errorf("count waitingPeer %v, want %v", len(dummyPM.waitingPeers), tt.wantCount)
			}
		})
	}
}

func Test_dynamicWPManager_OnDiscoveredPeers(t *testing.T) {
	ctrl := gomock.NewController(t)

	type args struct {
		preConnected []types.PeerID
		metas        []p2pcommon.PeerMeta
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{"TAllNew", args{nil, desigPeers[:1]}, 1},
		{"TAllExist", args{desigIDs, desigPeers[:5]}, 0},
		{"TMixedIDs", args{desigIDs, append(unknowPeers[:5], desigPeers[:5]...)}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := createDummyPM()
			mockActor := p2pmock.NewMockActorService(ctrl)
			dp := NewWaitingPeerManager(logger, dummyPM, mockActor, 10, true, false)
			for _, id := range tt.args.preConnected {
				dummyPM.remotePeers[id] = &remotePeerImpl{}
				dp.OnPeerConnect(id)
			}

			dp.OnDiscoveredPeers(tt.args.metas)
			if len(dummyPM.waitingPeers) != tt.wantCount {
				t.Errorf("count waitingPeer %v, want %v", len(dummyPM.waitingPeers), tt.wantCount)
			}
		})
	}
}

func Test_setNextTrial(t *testing.T) {
	dummyDesignated := p2pcommon.PeerMeta{Designated: true}

	type args struct {
		wp     *p2pcommon.WaitingPeer
		setCnt int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TDesig1", args{&p2pcommon.WaitingPeer{Meta: dummyDesignated}, 1}, true},
		{"TDesigSome", args{&p2pcommon.WaitingPeer{Meta: dummyDesignated}, 5}, true},
		{"TDesigMany", args{&p2pcommon.WaitingPeer{Meta: dummyDesignated}, 30}, true},

		{"TUnknown1", args{&p2pcommon.WaitingPeer{Meta: dummyMeta}, 1}, false},
		{"TUnknownSome", args{&p2pcommon.WaitingPeer{Meta: dummyMeta}, 5}, false},
		{"TUnknownMany", args{&p2pcommon.WaitingPeer{Meta: dummyMeta}, 30}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastResult := false
			prevDuration := time.Duration(0)
			for i := 0; i < tt.args.setCnt; i++ {
				now := time.Now()
				lastResult = setNextTrial(tt.args.wp)
				gotDuration := tt.args.wp.NextTrial.Sub(now)
				// nextTrial time will be increated exponetially and clipped when trial count is bigger than internal count
				// the clipped
				if lastResult &&
					(gotDuration < prevDuration && gotDuration < OneDay) {
					t.Errorf("smaller duration %v, want at least %v", gotDuration, prevDuration)
				}
				prevDuration = gotDuration
			}

			if lastResult != tt.want {
				t.Errorf("setNextTrial() = %v, want %v", lastResult, tt.want)
			}
		})
	}
}

func Test_basePeerManager_tryAddPeer(t *testing.T) {
	ctrl := gomock.NewController(t)

	// id0 is in both desginated peer and hidden peer
	desigIDs := make([]types.PeerID, 3)
	desigPeers := make(map[types.PeerID]p2pcommon.PeerMeta, 3)

	hiddenIDs := make([]types.PeerID, 3)
	hiddenPeers := make(map[types.PeerID]bool)

	for i := 0; i < 3; i++ {
		pkey, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		pid, _ := types.IDFromPrivateKey(pkey)
		desigIDs[i] = pid
		desigPeers[pid] = p2pcommon.PeerMeta{ID: pid}
	}
	hiddenIDs[0] = desigIDs[0]
	hiddenPeers[desigIDs[0]] = true

	for i := 1; i < 3; i++ {
		pkey, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		pid, _ := types.IDFromPrivateKey(pkey)
		hiddenIDs[i] = pid
		hiddenPeers[pid] = true
	}

	// tests for add peer
	type args struct {
		outbound bool
		meta     p2pcommon.PeerMeta
	}

	tests := []struct {
		name string
		args args

		hsRet *types.Status
		hsErr error

		wantDesign bool
		wantHidden bool
		wantID     types.PeerID
		wantSucc   bool
	}{
		// add inbound peer
		{"TIn", args{false, p2pcommon.PeerMeta{ID: dummyPeerID}},
			dummyStatus(dummyPeerID, false), nil, false, false, dummyPeerID, true},
		// add inbound designated peer
		{"TInDesignated", args{false, p2pcommon.PeerMeta{ID: desigIDs[1]}},
			dummyStatus(desigIDs[1], false), nil, true, false, desigIDs[1], true},
		// add inbound hidden peer
		{"TInHidden", args{false, p2pcommon.PeerMeta{ID: dummyPeerID}},
			dummyStatus(dummyPeerID, true), nil, false, true, dummyPeerID, true},
		// add inbound peer (hidden in node config)
		{"TInHiddenInConf", args{false, p2pcommon.PeerMeta{ID: hiddenIDs[1]}},
			dummyStatus(hiddenIDs[1], false), nil, false, true, hiddenIDs[1], true},
		{"TInH&D", args{false, p2pcommon.PeerMeta{ID: hiddenIDs[0], Hidden: true}},
			dummyStatus(hiddenIDs[0], true), nil, true, true, hiddenIDs[0], true},

		// add outbound peer
		{"TOut", args{true, p2pcommon.PeerMeta{ID: dummyPeerID}},
			dummyStatus(dummyPeerID, false), nil, false, false, dummyPeerID, true},
		// add outbound designated peer
		{"TOutDesignated", args{true, p2pcommon.PeerMeta{ID: desigIDs[1]}},
			dummyStatus(desigIDs[1], false), nil, true, false, desigIDs[1], true},
		// add outbound hidden peer
		{"TOutHidden", args{true, p2pcommon.PeerMeta{ID: dummyPeerID}},
			dummyStatus(dummyPeerID, true), nil, false, true, dummyPeerID, true},
		// add outbound peer (hidden in node config)
		{"TOutHiddenInConf", args{true, p2pcommon.PeerMeta{ID: hiddenIDs[1]}},
			dummyStatus(hiddenIDs[1], false), nil, false, true, hiddenIDs[1], true},
		{"TOutH&D", args{true, p2pcommon.PeerMeta{ID: hiddenIDs[0], Hidden: true}},
			dummyStatus(hiddenIDs[0], true), nil, true, true, hiddenIDs[0], true},

		// failed to handshake
		{"TErrHandshake", args{false, p2pcommon.PeerMeta{ID: dummyPeerID}},
			nil, errors.New("handshake err"), false, false, dummyPeerID, false},
		// invalid status information
		{"TErrDiffPeerID", args{false, p2pcommon.PeerMeta{ID: dummyPeerID}},
			dummyStatus(dummyPeerID2, false), nil, false, false, dummyPeerID, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStream := p2pmock.NewMockStream(ctrl)
			mockHSFactory := p2pmock.NewMockHSHandlerFactory(ctrl)
			mockHSHandler := p2pmock.NewMockHSHandler(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)
			//mockHSFactory.EXPECT().CreateHSHandler(gomock.Any(), tt.args.outbound, tt.args.meta.ID).Return(mockHSHandler)
			mockHSHandler.EXPECT().Handle(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRW, tt.hsRet, tt.hsErr)
			mockHandlerFactory := p2pmock.NewMockHandlerFactory(ctrl)
			mockHandlerFactory.EXPECT().InsertHandlers(gomock.AssignableToTypeOf(&remotePeerImpl{})).MaxTimes(1)

			// in cases of handshake error
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMF.EXPECT().NewMsgRequestOrder(false, subproto.GoAway, gomock.Any()).Return(&pbRequestOrder{}).MaxTimes(1)
			mockRW.EXPECT().WriteMsg(gomock.Any()).MaxTimes(1)

			pm := &peerManager{
				mf:              mockMF,
				hsFactory:       mockHSFactory,
				designatedPeers: desigPeers,
				hiddenPeerSet:   hiddenPeers,
				handlerFactory:  mockHandlerFactory,
				peerHandshaked:  make(chan p2pcommon.RemotePeer, 10),
			}
			dpm := &basePeerManager{
				pm:     pm,
				logger: logger,
			}
			got, got1 := dpm.tryAddPeer(tt.args.outbound, tt.args.meta, mockStream, mockHSHandler)
			if got1 != tt.wantSucc {
				t.Errorf("basePeerManager.tryAddPeer() got1 = %v, want %v", got1, tt.wantSucc)
			}
			if tt.wantSucc {
				if got.ID != tt.wantID {
					t.Errorf("basePeerManager.tryAddPeer() got ID = %v, want %v", got.ID, tt.wantID)
				}
				if got.Outbound != tt.args.outbound {
					t.Errorf("basePeerManager.tryAddPeer() got bound = %v, want %v", got.Outbound, tt.args.outbound)
				}
				if got.Designated != tt.wantDesign {
					t.Errorf("basePeerManager.tryAddPeer() got Designated = %v, want %v", got.Designated, tt.wantDesign)
				}
				if got.Hidden != tt.wantHidden {
					t.Errorf("basePeerManager.tryAddPeer() got Hidden = %v, want %v", got.Hidden, tt.wantHidden)
				}
			}

		})
	}
}

func dummyStatus(id types.PeerID, noexpose bool) *types.Status {
	return &types.Status{Sender: &types.PeerAddress{PeerID: []byte(id)}, NoExpose: noexpose}
}
