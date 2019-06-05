/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/subproto"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
)

func TestP2P_CreateHSHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		p2pVersion p2pcommon.P2PVersion
		outbound   bool
	}
	tests := []struct {
		name string

		args     args
		wantType reflect.Type
	}{
		{"TNewIn", args{p2pcommon.P2PVersion031, false}, reflect.TypeOf(&InboundWireHandshaker{})},
		{"TNewOut", args{p2pcommon.P2PVersion031, true}, reflect.TypeOf(&OutboundWireHandshaker{})},
		{"TLegacyIn", args{p2pcommon.P2PVersion030, false}, reflect.TypeOf(&LegacyInboundHSHandler{})},
		{"TLegacyOut", args{p2pcommon.P2PVersion030, true}, reflect.TypeOf(&LegacyOutboundHSHandler{})},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			sampleChainID := types.ChainID{}
			//mockMF := p2pmock.NewMockMoFactory(ctrl)
			//mockPeer := (*p2pmock.MockRemotePeer)(nil)
			//if tt.hasPeer {
			//	mockPeer = p2pmock.NewMockRemotePeer(ctrl)
			//	mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)
			//}
			//p2pmock.NewMockRemotePeer(ctrl)
			//mockPM.EXPECT().GetPeer(dummyPeerID).Return(mockPeer, tt.hasPeer).Times(1)
			//mockPM.EXPECT().SelfMeta().Return(dummyPeerMeta).Times(tt.wantSend).MaxTimes(tt.wantSend)
			//mockMF.EXPECT().NewMsgRequestOrder(true, subproto.AddressesRequest, gomock.AssignableToTypeOf(&types.AddressesRequest{})).Times(tt.wantSend)

			p2ps := &P2P{
				pm: mockPM, chainID: &sampleChainID,
			}
			p2ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2ps, log.NewLogger("p2p.test"))

			got := p2ps.CreateHSHandler(tt.args.p2pVersion, tt.args.outbound, dummyPeerID)
			if !reflect.TypeOf(got).AssignableTo(tt.wantType) {
				t.Errorf("P2P.CreateHSHandler() type = %v, want %v", reflect.TypeOf(got), tt.wantType)
			}
		})
	}
}

func TestP2P_InsertHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name   string
	}{
		{"T1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().AddMessageHandler(gomock.AssignableToTypeOf(subproto.PingResponse), gomock.Any()).MinTimes(1)

			p2ps := &P2P{
				pm: mockPM,
			}
			p2ps.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2ps, log.NewLogger("p2p.test"))

			p2ps.InsertHandlers(mockPeer)
		})
	}
}
