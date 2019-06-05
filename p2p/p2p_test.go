/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/subproto"
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
		name string
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

func TestRaftRoleManager_updateBP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p1,p2,p3 := dummyPeerID, dummyPeerID2, dummyPeerID3


	tests := []struct {
		name   string
		preset []types.PeerID
		args   message.RaftClusterEvent

		wantCnt int
		wantExist []types.PeerID
		wantNot []types.PeerID
	}{
		{"TAdd",nil, (&EB{}).A(p1,p2).C() ,2, []types.PeerID{p1,p2},nil },
		{"TRm",[]types.PeerID{p1,p2,p3}, (&EB{}).R(p3,p2).C() ,1, []types.PeerID{p1},[]types.PeerID{p2,p3} },
		{"TOverrap",[]types.PeerID{p3}, (&EB{}).A(p1,p2).R(p3,p2).C() ,2, []types.PeerID{p1,p2},[]types.PeerID{p3} },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presetIDs := make(map[types.PeerID]bool)
			for _,id := range tt.preset {
				presetIDs[id] = true
			}
			p2ps := &RaftRoleManager{
				raftBP: presetIDs,
			}
			p2ps.UpdateBP(tt.args.BPAdded, tt.args.BPRemoved)

			if len(p2ps.raftBP) != tt.wantCnt {
				t.Errorf("P2P.UpdateBP() len = %v, want %v", len(p2ps.raftBP), tt.wantCnt)
			}
			for _,id := range tt.wantExist {
				if _, found := p2ps.raftBP[id]; !found {
					t.Errorf("P2P.UpdateBP() not exist %v, want exist ", id)
				} else {
					if p2ps.GetRole(id) != p2pcommon.RaftLeader {
						t.Errorf("P2P.GetRole(%v) false, want true", id)
					}
				}
			}
			for _,id := range tt.wantNot {
				if _, found := p2ps.raftBP[id]; found {
					t.Errorf("P2P.UpdateBP() exist %v, want not ", id)
				}
			}
		})
	}
}

type EB struct {
	a, r []types.PeerID
}

func (e *EB) A(ids ...types.PeerID) *EB {
	e.a = append(e.a, ids... )
	return e
}
func (e *EB) R(ids ...types.PeerID) *EB {
	e.r = append(e.r, ids... )
	return e
}
func (e *EB) C() message.RaftClusterEvent {
	return message.RaftClusterEvent{BPAdded:e.a, BPRemoved:e.r}
}
