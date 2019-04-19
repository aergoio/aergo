/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"reflect"
	"testing"
	"time"

	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-peer"
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
		preConnected []peer.ID
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

	// tests for add peer.
	type args struct {
		outbound bool
		meta     p2pcommon.PeerMeta
	}

	tests := []struct {
		name   string
		args   args

		wantHidden bool
		wantMeta   p2pcommon.PeerMeta
		wantSucc   bool
	}{
		// add inbound peer
		// add inbound hidden peer
		// add inbound peer (hidden in node config)
		// add outbound peer
		// add outbound hidden peer
		// add outbound peer (hidden in node config)

		// failed to handshake
		// invalid status information

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHSFactory := p2pmock.NewMockHSHandlerFactory(ctrl)
			mockHandlerFactory := p2pmock.NewMockHandlerFactory(ctrl)
			mockStream := p2pmock.NewMockStream(ctrl)

			pm := &peerManager{
				hsFactory: mockHSFactory,
				designatedPeers: nil,
				hiddenPeerSet: nil,
				handlerFactory:mockHandlerFactory,
				peerHandshaked: make(chan p2pcommon.RemotePeer, 10),
			}
			dpm := &basePeerManager{
				pm:          pm,
				logger:      logger,
			}
			got, got1 := dpm.tryAddPeer(tt.args.outbound, tt.args.meta, mockStream)
			if got1 != tt.wantSucc {
				t.Errorf("basePeerManager.tryAddPeer() got1 = %v, want %v", got1, tt.wantSucc)
			}
			if tt.wantSucc && !reflect.DeepEqual(got, tt.wantMeta) {
				t.Errorf("basePeerManager.tryAddPeer() got = %v, want %v", got, tt.wantMeta)
			}
		})
	}
}
