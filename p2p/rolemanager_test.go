/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
)

func TestRaftRoleManager_updateBP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p1, p2, p3 := dummyPeerID, dummyPeerID2, dummyPeerID3

	tests := []struct {
		name   string
		preset []types.PeerID
		args   message.RaftClusterEvent

		wantCnt   int
		wantExist []types.PeerID
		wantNot   []types.PeerID
	}{
		{"TAdd", nil, (&EB{}).A(p1, p2).C(), 2, []types.PeerID{p1, p2}, nil},
		{"TRm", []types.PeerID{p1, p2, p3}, (&EB{}).R(p3, p2).C(), 1, []types.PeerID{p1}, []types.PeerID{p2, p3}},
		{"TOverrap", []types.PeerID{p3}, (&EB{}).A(p1, p2).R(p3, p2).C(), 2, []types.PeerID{p1, p2}, []types.PeerID{p3}},
		{"TEmpty", nil, (&EB{}).C(), 0, []types.PeerID{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presetIDs := make(map[types.PeerID]bool)
			for _, id := range tt.preset {
				presetIDs[id] = true
			}
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false).AnyTimes()
			mockPM.EXPECT().UpdatePeerRole(gomock.Any()).AnyTimes()

			p2ps := &P2P{pm: mockPM}
			rm := &RaftRoleManager{
				p2ps:   p2ps,
				logger: logger,
				raftBP: presetIDs,
			}
			rm.UpdateBP(tt.args.BPAdded, tt.args.BPRemoved)

			if len(rm.raftBP) != tt.wantCnt {
				t.Errorf("P2P.UpdateBP() len = %v, want %v", len(rm.raftBP), tt.wantCnt)
			}
			for _, id := range tt.wantExist {
				if _, found := rm.raftBP[id]; !found {
					t.Errorf("P2P.UpdateBP() not exist %v, want exist ", id)
				} else {
					if rm.GetRole(id) != p2pcommon.RaftFollower {
						t.Errorf("P2P.GetRole(%v) false, want true", id)
					}
				}
			}
			for _, id := range tt.wantNot {
				if _, found := rm.raftBP[id]; found {
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
	e.a = append(e.a, ids...)
	return e
}
func (e *EB) R(ids ...types.PeerID) *EB {
	e.r = append(e.r, ids...)
	return e
}
func (e *EB) C() message.RaftClusterEvent {
	return message.RaftClusterEvent{BPAdded: e.a, BPRemoved: e.r}
}

func TestRaftRoleManager_NotifyNewBlockMsg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pids := []types.PeerID{dummyPeerID, dummyPeerID2, dummyPeerID3}

	tests := []struct {
		name string

		argPeer []rs

		wantSkipped int
		wantSent    int
	}{
		{"TAllBP", []rs{rs{p2pcommon.RaftLeader, types.RUNNING}, rs{p2pcommon.RaftFollower, types.RUNNING}, rs{p2pcommon.RaftFollower, types.RUNNING}}, 3, 0},
		{"TAllWat", []rs{rs{p2pcommon.RaftWatcher, types.RUNNING}, rs{p2pcommon.RaftWatcher, types.RUNNING}, rs{p2pcommon.RaftWatcher, types.RUNNING}}, 0, 3},
		{"TMIX", []rs{rs{p2pcommon.RaftLeader, types.RUNNING}, rs{p2pcommon.RaftFollower, types.RUNNING}, rs{p2pcommon.RaftWatcher, types.RUNNING}}, 2, 1},
		{"TMIXStop", []rs{rs{p2pcommon.RaftLeader, types.RUNNING}, rs{p2pcommon.RaftFollower, types.RUNNING}, rs{p2pcommon.RaftWatcher, types.STOPPING}}, 3, 0},
		{"TMixStop2", []rs{rs{p2pcommon.RaftLeader, types.STOPPING}, rs{p2pcommon.RaftFollower, types.RUNNING}, rs{p2pcommon.RaftWatcher, types.RUNNING}}, 2, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false).AnyTimes()
			mockPM.EXPECT().UpdatePeerRole(gomock.Any()).AnyTimes()

			mockMO := p2pmock.NewMockMsgOrder(ctrl)

			rm := &RaftRoleManager{
				p2ps:   nil,
				logger: logger,
			}
			mockPeers := make([]p2pcommon.RemotePeer, 0, len(tt.argPeer))
			for i, ap := range tt.argPeer {
				mpeer := p2pmock.NewMockRemotePeer(ctrl)
				mpeer.EXPECT().ID().Return(pids[i]).AnyTimes()
				mpeer.EXPECT().Role().Return(ap.r).AnyTimes()
				mpeer.EXPECT().State().Return(ap.s).AnyTimes()
				mpeer.EXPECT().SendMessage(gomock.Any()).MaxTimes(1)

				mockPeers = append(mockPeers, mpeer)
			}
			gotSkipped, gotSent := rm.NotifyNewBlockMsg(mockMO, mockPeers)
			if gotSkipped != tt.wantSkipped {
				t.Errorf("RaftRoleManager.NotifyNewBlockMsg() gotSkipped = %v, want %v", gotSkipped, tt.wantSkipped)
			}
			if gotSent != tt.wantSent {
				t.Errorf("RaftRoleManager.NotifyNewBlockMsg() gotSent = %v, want %v", gotSent, tt.wantSent)
			}
		})
	}
}

func TestDefaultRoleManager_updateBP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p1, p2, p3 := dummyPeerID, dummyPeerID2, dummyPeerID3

	tests := []struct {
		name string
		args message.RaftClusterEvent

		wantCnt   int
		wantExist []types.PeerID
	}{
		{"TAdd", (&EB{}).A(p1, p2).C(), 2, []types.PeerID{p1, p2}},
		{"TRemove", (&EB{}).R(p3, p2).C(), 2, []types.PeerID{p3, p2}},
		{"TOverrap", (&EB{}).A(p1, p2).R(p3, p2).C(), 3, []types.PeerID{p1, p2, p3}},
		{"TEmpty", (&EB{}).C(), 0, []types.PeerID{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sentIds []types.PeerID = nil
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false).AnyTimes()
			mockPM.EXPECT().UpdatePeerRole(gomock.Any()).Do(func(cs []p2pcommon.AttrModifier) {
				for _, id := range cs {
					sentIds = append(sentIds, id.ID)
				}
			})

			p2ps := &P2P{pm: mockPM}
			rm := &DefaultRoleManager{
				p2ps: p2ps,
			}
			rm.UpdateBP(tt.args.BPAdded, tt.args.BPRemoved)

			for _, id := range tt.wantExist {
				found := false
				for _, sent := range sentIds {
					if id.Pretty() == sent.Pretty() {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("DefaultRoleManager.UpdateBP() not exist %v, want exist ", id)
				}
			}
		})
	}
}

func TestDefaultRoleManager_NotifyNewBlockMsg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pids := []types.PeerID{dummyPeerID, dummyPeerID2, dummyPeerID3}

	tests := []struct {
		name string

		argPeer []rs

		wantSkipped int
		wantSent    int
	}{
		{"TAllBP", []rs{rs{p2pcommon.BlockProducer, types.RUNNING}, rs{p2pcommon.BlockProducer, types.RUNNING}, rs{p2pcommon.BlockProducer, types.RUNNING}}, 0, 3},
		{"TAllWat", []rs{rs{p2pcommon.Watcher, types.RUNNING}, rs{p2pcommon.Watcher, types.RUNNING}, rs{p2pcommon.Watcher, types.RUNNING}}, 0, 3},
		{"TMIX", []rs{rs{p2pcommon.BlockProducer, types.RUNNING}, rs{p2pcommon.BlockProducer, types.RUNNING}, rs{p2pcommon.Watcher, types.RUNNING}}, 0, 3},
		{"TMIXStop", []rs{rs{p2pcommon.BlockProducer, types.RUNNING}, rs{p2pcommon.BlockProducer, types.RUNNING}, rs{p2pcommon.Watcher, types.STOPPING}}, 1, 2},
		{"TMixStop2", []rs{rs{p2pcommon.BlockProducer, types.STOPPING}, rs{p2pcommon.BlockProducer, types.RUNNING}, rs{p2pcommon.Watcher, types.RUNNING}}, 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false).AnyTimes()
			mockPM.EXPECT().UpdatePeerRole(gomock.Any()).AnyTimes()

			mockMO := p2pmock.NewMockMsgOrder(ctrl)

			rm := &DefaultRoleManager{
				p2ps: nil,
			}
			mockPeers := make([]p2pcommon.RemotePeer, 0, len(tt.argPeer))
			for i, ap := range tt.argPeer {
				mpeer := p2pmock.NewMockRemotePeer(ctrl)
				mpeer.EXPECT().ID().Return(pids[i]).AnyTimes()
				mpeer.EXPECT().Role().Return(ap.r).AnyTimes()
				mpeer.EXPECT().State().Return(ap.s).AnyTimes()
				mpeer.EXPECT().SendMessage(gomock.Any()).MaxTimes(1)

				mockPeers = append(mockPeers, mpeer)
			}
			gotSkipped, gotSent := rm.NotifyNewBlockMsg(mockMO, mockPeers)
			if gotSkipped != tt.wantSkipped {
				t.Errorf("RaftRoleManager.NotifyNewBlockMsg() gotSkipped = %v, want %v", gotSkipped, tt.wantSkipped)
			}
			if gotSent != tt.wantSent {
				t.Errorf("RaftRoleManager.NotifyNewBlockMsg() gotSent = %v, want %v", gotSent, tt.wantSent)
			}
		})
	}
}

type rs struct {
	r p2pcommon.PeerRole
	s types.PeerState
}

func TestDefaultRoleManager_GetRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p1, p2, p3 := dummyPeerID, dummyPeerID2, dummyPeerID3

	tests := []struct {
		name string

		presetIds []types.PeerID
		pid       types.PeerID
		want      p2pcommon.PeerRole
	}{
		{"TBP", toPIDS(p1,p2), p1, p2pcommon.BlockProducer},
		{"TWat", toPIDS(p1,p2), p3, p2pcommon.Watcher},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := make([]string,0,len(tt.presetIds))
			for _, id := range tt.presetIds {
				bps = append(bps, fmt.Sprintf("{\"%d\":\"%s\"}",i,id.Pretty()))
			}
			dummyConsensus := &types.ConsensusInfo{Bps:bps}
			mockCC := p2pmock.NewMockConsensusAccessor(ctrl)
			mockCC.EXPECT().ConsensusInfo().Return(dummyConsensus)
			p2ps := &P2P{consacc: mockCC}
			rm := &DefaultRoleManager{
				p2ps: p2ps,
			}
			if got := rm.GetRole(tt.pid); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultRoleManager.GetRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func toPIDS(ids ...types.PeerID) []types.PeerID {
	return ids
}