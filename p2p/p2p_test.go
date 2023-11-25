/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"net"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/list"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

func TestP2P_CreateHSHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		legacy   bool
		outbound bool
	}
	tests := []struct {
		name string

		args     args
		wantType reflect.Type
	}{
		{"TNewIn", args{false, false}, reflect.TypeOf(&InboundWireHandshaker{})},
		{"TNewOut", args{false, true}, reflect.TypeOf(&OutboundWireHandshaker{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			sampleChainID := types.ChainID{}

			p2ps := &P2P{
				pm: mockPM, genesisChainID: &sampleChainID,
			}
			p2ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2ps, log.NewLogger("p2p.test"))

			got := p2ps.CreateHSHandler(tt.args.outbound, dummyPeerID)
			if !reflect.TypeOf(got).AssignableTo(tt.wantType) {
				t.Errorf("P2P.CreateHSHandler() type = %v, want %v", reflect.TypeOf(got), tt.wantType)
			}
		})
	}
}

func TestP2P_InsertHandlers(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"T1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().AddMessageHandler(gomock.AssignableToTypeOf(p2pcommon.PingResponse), gomock.Any()).MinTimes(1)
			mockPeer.EXPECT().ID().AnyTimes()

			p2ps := &P2P{
				pm: mockPM,
			}
			p2ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2ps, log.NewLogger("p2p.test"))

			p2ps.insertHandlers(mockPeer)
		})
	}
}

func TestP2P_banIfFound(t *testing.T) {
	sampleCnt := 5
	addr := "172.21.11.3"

	pids := make([]types.PeerID, sampleCnt)
	for i := 0; i < sampleCnt; i++ {
		pids[i] = types.RandomPeerID()
	}
	tests := []struct {
		name string

		inWhite     []int
		wantStopCnt int
	}{
		{"TAllWhite", []int{1, 1, 1, 1, 1}, 0},
		{"TAllBan", []int{0, 0, 0, 0, 0}, 5},
		{"TMix", []int{0, 1, 1, 0, 1}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockLM := p2pmock.NewMockListManager(ctrl)
			peers := make([]p2pcommon.RemotePeer, sampleCnt)
			for i := 0; i < sampleCnt; i++ {
				meta := p2pcommon.NewMetaWith1Addr(types.RandomPeerID(), addr, 7846, "v2.0.0")
				conn := p2pcommon.RemoteConn{IP: net.ParseIP(addr), Port: 7846, Outbound: false}
				ri := p2pcommon.RemoteInfo{Meta: meta, Connection: conn}
				mPeer := p2pmock.NewMockRemotePeer(ctrl)
				mPeer.EXPECT().ID().Return(pids[i])
				mPeer.EXPECT().RemoteInfo().Return(ri).MaxTimes(2)
				mPeer.EXPECT().Name().Return("peer " + pids[i].ShortString()).AnyTimes()
				if tt.inWhite[i] == 0 {
					mPeer.EXPECT().Stop()
				}

				peers[i] = mPeer
				mockLM.EXPECT().IsBanned(addr, pids[i]).Return(tt.inWhite[i] == 0, list.FarawayFuture)
			}
			mockPM.EXPECT().GetPeers().Return(peers)
			p2ps := &P2P{
				pm: mockPM,
				lm: mockLM,
			}
			p2ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2ps, log.NewLogger("p2p"))

			p2ps.checkAndBanInboundPeers()
		})
	}
}

func checkPanic(t *testing.T, wantPanic bool) {
	if r := recover(); (r != nil) != wantPanic {
		t.Errorf("panic of NewListManager() %v, want %v", r != nil, wantPanic)
	}
}
