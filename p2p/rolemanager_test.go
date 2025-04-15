/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"net"
	"reflect"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
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
		{"TOverlap", []types.PeerID{p3}, (&EB{}).A(p1, p2).R(p3, p2).C(), 2, []types.PeerID{p1, p2}, []types.PeerID{p3}},
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
				is:     p2ps,
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
					if rm.GetRole(id) != types.PeerRole_Producer {
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

func TestRaftRoleManager_FilterBPNoticeReceiver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pids := []types.PeerID{dummyPeerID, dummyPeerID2, dummyPeerID3}

	tests := []struct {
		name string

		argPeer []rs

		wantSkipped int
		wantSent    int
	}{
		{"TAllBP", []rs{{types.PeerRole_Producer, types.RUNNING}, {types.PeerRole_Producer, types.RUNNING}, {types.PeerRole_Producer, types.RUNNING}}, 3, 0},
		{"TAllWat", []rs{{types.PeerRole_Watcher, types.RUNNING}, {types.PeerRole_Watcher, types.RUNNING}, {types.PeerRole_Watcher, types.RUNNING}}, 0, 3},
		{"TMIX", []rs{{types.PeerRole_Producer, types.RUNNING}, {types.PeerRole_Producer, types.RUNNING}, {types.PeerRole_Watcher, types.RUNNING}}, 2, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false).AnyTimes()
			mockPM.EXPECT().UpdatePeerRole(gomock.Any()).AnyTimes()

			dummyBlock := &types.Block{}

			rm := &RaftRoleManager{
				is:     nil,
				logger: logger,
			}
			mockPeers := make([]p2pcommon.RemotePeer, 0, len(tt.argPeer))
			for i, ap := range tt.argPeer {
				mpeer := p2pmock.NewMockRemotePeer(ctrl)
				mpeer.EXPECT().ID().Return(pids[i]).AnyTimes()
				mpeer.EXPECT().AcceptedRole().Return(ap.r).AnyTimes()
				mpeer.EXPECT().State().Return(ap.s).AnyTimes()
				mpeer.EXPECT().SendMessage(gomock.Any()).MaxTimes(1)

				mockPeers = append(mockPeers, mpeer)
			}
			mockPM.EXPECT().GetPeers().Return(mockPeers).AnyTimes()

			filtered := rm.FilterBPNoticeReceiver(dummyBlock, mockPM, p2pcommon.ExternalZone)
			if len(filtered) != tt.wantSent {
				t.Errorf("RaftRoleManager.FilterBPNoticeReceiver() peers = %v, want %v", len(filtered), tt.wantSent)
			}
		})
	}
}

func TestDPOSRoleManager_updateBP(t *testing.T) {
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
		{"TOverlap", (&EB{}).A(p1, p2).R(p3, p2).C(), 3, []types.PeerID{p1, p2, p3}},
		{"TEmpty", (&EB{}).C(), 0, []types.PeerID{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sentIds []types.PeerID = nil
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false).AnyTimes()
			mockPM.EXPECT().UpdatePeerRole(gomock.Any()).Do(func(cs []p2pcommon.RoleModifier) {
				for _, id := range cs {
					sentIds = append(sentIds, id.ID)
				}
			})

			p2ps := &P2P{pm: mockPM}
			rm := &DPOSRoleManager{
				is: p2ps,
			}
			rm.UpdateBP(tt.args.BPAdded, tt.args.BPRemoved)

			for _, id := range tt.wantExist {
				found := false
				for _, sent := range sentIds {
					if id.String() == sent.String() {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("DPOSRoleManager.UpdateBP() not exist %v, want exist ", id)
				}
			}
		})
	}
}

func TestDPOSRoleManager_FilterBPNoticeReceiver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pids := []types.PeerID{dummyPeerID, dummyPeerID2, dummyPeerID3}

	tests := []struct {
		name string

		argPeer []rs

		wantSkipped int
		wantSent    int
	}{
		{"TAllBP", []rs{{types.PeerRole_Producer, types.RUNNING}, {types.PeerRole_Producer, types.RUNNING}, {types.PeerRole_Producer, types.RUNNING}}, 0, 3},
		{"TAllWat", []rs{{types.PeerRole_Watcher, types.RUNNING}, {types.PeerRole_Watcher, types.RUNNING}, {types.PeerRole_Watcher, types.RUNNING}}, 0, 3},
		{"TMix", []rs{{types.PeerRole_Producer, types.RUNNING}, {types.PeerRole_Producer, types.RUNNING}, {types.PeerRole_Watcher, types.RUNNING}}, 0, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPM.EXPECT().GetPeer(gomock.Any()).Return(nil, false).AnyTimes()
			mockPM.EXPECT().UpdatePeerRole(gomock.Any()).AnyTimes()

			rm := &DPOSRoleManager{
				is: nil,
			}
			mockPeers := make([]p2pcommon.RemotePeer, 0, len(tt.argPeer))
			for i, ap := range tt.argPeer {
				mPeer := p2pmock.NewMockRemotePeer(ctrl)
				mPeer.EXPECT().ID().Return(pids[i]).AnyTimes()
				mPeer.EXPECT().AcceptedRole().Return(ap.r).AnyTimes()
				mPeer.EXPECT().State().Return(ap.s).AnyTimes()
				mPeer.EXPECT().SendMessage(gomock.Any()).MaxTimes(1)

				mockPeers = append(mockPeers, mPeer)
			}
			mockPM.EXPECT().GetPeers().Return(mockPeers).AnyTimes()

			sampleBlock := &types.Block{}
			filtered := rm.FilterBPNoticeReceiver(sampleBlock, mockPM, p2pcommon.ExternalZone)
			if len(filtered) != tt.wantSent {
				t.Errorf("RaftRoleManager.NotifyNewBlockMsg() peers = %v, want %v", len(filtered), tt.wantSent)
			}
		})
	}
}

type rs struct {
	r types.PeerRole
	s types.PeerState
}

func TestDPOSRoleManager_GetRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p1, p2, p3 := dummyPeerID, dummyPeerID2, dummyPeerID3

	tests := []struct {
		name string

		presetIds []types.PeerID
		pid       types.PeerID
		want      types.PeerRole
	}{
		{"TBP", toPIDS(p1, p2), p1, types.PeerRole_Producer},
		{"TWat", toPIDS(p1, p2), p3, types.PeerRole_Watcher},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := make([]types.PeerID, 0, len(tt.presetIds))
			union := make(map[types.PeerID]voteRank)
			for _, id := range tt.presetIds {
				bps = append(bps, id)
				union[id] = BP
			}

			mockCC := p2pmock.NewMockConsensusAccessor(ctrl)
			p2ps := &P2P{consacc: mockCC}
			rm := NewDPOSRoleManager(p2ps, p2ps, nil)
			rm.bps = bps
			rm.unionSet = union

			if got := rm.GetRole(tt.pid); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DPOSRoleManager.GetRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func toPIDS(ids ...types.PeerID) []types.PeerID {
	return ids
}

func TestDPOSRoleManager_reloadVotes(t *testing.T) {
	logger := log.NewLogger("p2p.test")
	initialBPCount := 3
	initialUnion := make(map[types.PeerID]voteRank)
	initialBPs := make([]string, initialBPCount)
	var pids []types.PeerID
	for i := 0; i < 9; i++ {
		pids = append(pids, types.RandomPeerID())
		if i < initialBPCount {
			initialBPs[i] = pids[i].String()
			initialUnion[pids[i]] = BP
		} else if i < initialBPCount*2 {
			initialUnion[pids[i]] = Candidate
		}
	}
	dummyConsensus := &types.ConsensusInfo{Bps: initialBPs}

	tests := []struct {
		name string

		actErr  error
		respIds []types.PeerID
		respErr error

		wantErr  bool
		wantSize int
	}{
		{"TNormal", nil, pids[:5], nil, false, 5},
		{"TActErr", TimeoutError, pids[:6], nil, true, 5},
		{"TChainErr", nil, pids[:6], sampleErr, true, 5},
		{"TLess", nil, pids[:4], nil, false, 4},
		{"TMore", nil, pids[:6], nil, false, 6},
		{"TOver", nil, pids[:8], nil, false, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			votes := &types.VoteList{}
			for _, id := range tt.respIds {
				vt := &types.Vote{Candidate: []byte(id), Amount: []byte{0, 0}}
				votes.Votes = append(votes.Votes, vt)
			}
			aResult := &message.GetVoteRsp{Top: votes, Err: tt.respErr}
			is := p2pmock.NewMockInternalService(ctrl)
			actor := p2pmock.NewMockActorService(ctrl)
			pm := p2pmock.NewMockPeerManager(ctrl)
			cm := p2pmock.NewMockConsensusAccessor(ctrl)
			is.EXPECT().PeerManager().Return(pm).AnyTimes()
			is.EXPECT().ConsensusAccessor().Return(cm).AnyTimes()
			cm.EXPECT().ConsensusInfo().Return(dummyConsensus)
			actor.EXPECT().CallRequest(message.ChainSvc, gomock.AssignableToTypeOf(&message.GetElected{}), getVotesMessageTimeout).Return(aResult, tt.actErr)

			rm := &DPOSRoleManager{
				is:     is,
				actor:  actor,
				logger: logger,
			}
			union, _, err := rm.loadBPVotes()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadBPVotes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(union) != tt.wantSize {
					t.Errorf("loadBPVotes() total = %v, want size %v", union, tt.wantSize)
				}
			}
		})
	}
}

func TestDPOSRoleManager_collectAddDel(t *testing.T) {
	logger := log.NewLogger("p2p.test")

	initialBPCount := 3
	initialUnionCnt := 5
	initialUnion := make(map[types.PeerID]voteRank)
	initialBPs := make([]string, initialBPCount)
	var pids []types.PeerID
	for i := 0; i < 9; i++ {
		pids = append(pids, types.RandomPeerID())
		if i < initialBPCount {
			initialBPs[i] = pids[i].String()
			initialUnion[pids[i]] = BP
		} else if i < initialUnionCnt {
			initialUnion[pids[i]] = Candidate
		}
	}

	tests := []struct {
		name string

		newBPcnt int
		newRanks []types.PeerID

		wantAdd int
		wantDel int
	}{
		{"TSame", 3, pids[:5], 0, 0},
		{"TTurned", 3, append(add(pids[2], pids[0], pids[1]), pids[3:5]...), 0, 0},
		{"TAddedTail", 3, pids[:6], 1, 0},
		{"TAddedHead", 3, append(add(pids[8], pids[7]), pids[:4]...), 2, 1},
		{"TShrink", 3, pids[:4], 0, 1},
		{"TMod", 3, pids[2:8], 3, 2},
		{"TBPInc", 4, pids[:5], 0, 0},
		{"TBPIncAdded", 4, pids[:8], 3, 0},
		{"TBPdec", 2, pids[:4], 0, 1},
		{"TBPdecAdded", 2, pids[2:6], 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			is := p2pmock.NewMockInternalService(ctrl)
			actor := p2pmock.NewMockActorService(ctrl)
			pm := p2pmock.NewMockPeerManager(ctrl)
			cm := p2pmock.NewMockConsensusAccessor(ctrl)
			is.EXPECT().PeerManager().Return(pm).AnyTimes()
			is.EXPECT().ConsensusAccessor().Return(cm).AnyTimes()

			rm := NewDPOSAgentRoleManager(is, actor, logger, nil)
			rm.unionSet = initialUnion

			if len(tt.newRanks) > tt.newBPcnt*2 {
				t.Fatalf("Wrong test input %v ", tt.newRanks)
			}
			newUnion := make(map[types.PeerID]voteRank)
			for i, id := range tt.newRanks {
				if i < tt.newBPcnt {
					newUnion[id] = BP
				} else {
					newUnion[id] = Candidate
				}
			}
			gotAdd, gotDel := rm.collectAddDel(newUnion)

			if len(gotAdd) != tt.wantAdd {
				t.Errorf("collectAddDel() gotAdd = %v, want %v", gotAdd, tt.wantAdd)
			}
			if len(gotDel) != tt.wantDel {
				t.Errorf("collectAddDel() gotDel = %v, want %v", gotDel, tt.wantDel)
			}
		})
	}

}

func add(ids ...types.PeerID) []types.PeerID {
	return ids
}

func TestDPOSAgentRoleManager_FilterBPNoticeReceiver(t *testing.T) {
	logger := log.NewLogger("p2p.test")

	intIp, internalNet, _ := net.ParseCIDR("192.168.1.1/24")
	sampleSetting := p2pcommon.LocalSettings{InternalZones: []*net.IPNet{internalNet}}
	sampleMeta := p2pcommon.NewMetaWith1Addr(types.RandomPeerID(), intIp.String(), 7846, "v2.0.0")
	//  0,1 are my producer, 2 is other internal producer,
	//  3 is other internal agent, 4 is internal watcher,
	//  5,6 are external bp ,
	//  7,8 are external agent
	//  9 is external watcher
	p, a, w := types.PeerRole_Producer, types.PeerRole_Agent, types.PeerRole_Watcher
	i, e := p2pcommon.InternalZone, p2pcommon.ExternalZone
	roles := []types.PeerRole{p, p, p, a, w, p, p, a, a, w}
	zones := []p2pcommon.PeerZone{i, i, i, i, i, e, e, e, e, e}
	var pids []types.PeerID
	for i := 0; i < len(roles); i++ {
		pids = append(pids, types.RandomPeerID())
	}
	myPds := make(map[types.PeerID]bool)
	myPds[pids[0]] = true
	myPds[pids[1]] = true

	// test tointernal (to my producers only)
	// test toexternal
	tests := []struct {
		name string

		argZone p2pcommon.PeerZone

		want []int
	}{
		{"TToInt", p2pcommon.InternalZone, []int{0, 1}},
		{"TToExt", p2pcommon.ExternalZone, []int{5, 6, 7, 8}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			is := p2pmock.NewMockInternalService(ctrl)
			actor := p2pmock.NewMockActorService(ctrl)
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			is.EXPECT().LocalSettings().Return(sampleSetting).AnyTimes()
			is.EXPECT().SelfMeta().Return(sampleMeta).AnyTimes()
			is.EXPECT().PeerManager().Return(mockPM).AnyTimes()

			mockPeers := make([]p2pcommon.RemotePeer, 0, len(pids))
			mockBPs := make([]p2pcommon.RemotePeer, 0, len(pids))
			mockWatches := make([]p2pcommon.RemotePeer, 0, len(pids))
			for i, pid := range pids {
				mPeer := p2pmock.NewMockRemotePeer(ctrl)
				ri := p2pcommon.RemoteInfo{Meta: p2pcommon.NewMetaWith1Addr(pid, "1.1.1.1", 7846, "v2.0.0"), AcceptedRole: roles[i], Zone: zones[i]}
				mPeer.EXPECT().ID().Return(pid).AnyTimes()
				mPeer.EXPECT().RemoteInfo().Return(ri).AnyTimes()
				mPeer.EXPECT().AcceptedRole().Return(roles[i]).AnyTimes()

				mockPeers = append(mockPeers, mPeer)
				if roles[i] == w {
					mockWatches = append(mockWatches, mPeer)
				} else {
					mockBPs = append(mockBPs, mPeer)
				}
			}
			mockPM.EXPECT().GetPeers().Return(mockPeers).AnyTimes()
			mockPM.EXPECT().GetProducerClassPeers().Return(mockBPs).AnyTimes()
			mockPM.EXPECT().GetWatcherClassPeers().Return(mockWatches).AnyTimes()

			rm := NewDPOSAgentRoleManager(is, actor, logger, myPds)

			sampleBlock := &types.Block{}
			filtered := rm.FilterBPNoticeReceiver(sampleBlock, mockPM, tt.argZone)
			if len(filtered) != len(tt.want) {
				t.Fatalf("RaftRoleManager.NotifyNewBlockMsg() peers = %v, want %v", len(filtered), len(tt.want))
			}
			for i, idx := range tt.want {
				fp := filtered[i]
				wp := mockPeers[idx]
				if !types.IsSamePeerID(fp.ID(), wp.ID()) {
					t.Errorf("RaftRoleManager.NotifyNewBlockMsg() peerID = %v, want %v", fp.ID(), wp.ID())
				}

			}
		})
	}
}
