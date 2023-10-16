/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"reflect"
	"testing"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

const desigCnt = 10

var (
	desigIDs     []types.PeerID
	desigPeers   []p2pcommon.PeerMeta
	desigPeerMap = make(map[types.PeerID]p2pcommon.PeerMeta)

	unknownIDs   []types.PeerID
	unknownPeers []p2pcommon.PeerMeta
)

func init() {
	desigIDs = make([]types.PeerID, desigCnt)
	desigPeers = make([]p2pcommon.PeerMeta, desigCnt)
	for i := 0; i < desigCnt; i++ {
		pid := types.RandomPeerID()
		desigIDs[i] = pid
		desigPeers[i] = p2pcommon.NewMetaWith1Addr(pid, "192.111.222.33", 7846, "v2.0.0")
		desigPeerMap[desigIDs[i]] = desigPeers[i]
	}
	unknownIDs = make([]types.PeerID, desigCnt)
	unknownPeers = make([]p2pcommon.PeerMeta, desigCnt)
	for i := 0; i < desigCnt; i++ {
		pid := types.RandomPeerID()
		unknownIDs[i] = pid
		unknownPeers[i] = p2pcommon.NewMetaWith1Addr(pid, "192.111.222.33", 7846, "v2.0.0")
	}
}
func createDummyPM() *peerManager {
	dummyPM := &peerManager{designatedPeers: desigPeerMap,
		remotePeers:  make(map[types.PeerID]p2pcommon.RemotePeer),
		waitingPeers: make(map[types.PeerID]*p2pcommon.WaitingPeer, 10),
	}
	return dummyPM
}

func TestNewPeerFinder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		useDiscover bool
		usePolaris  bool
	}
	tests := []struct {
		name string
		args args
		want p2pcommon.PeerFinder
	}{
		{"Tstatic", args{false, false}, &staticPeerFinder{}},
		{"TstaticWPolaris", args{false, true}, &staticPeerFinder{}},
		{"Tdyn", args{true, false}, &dynamicPeerFinder{}},
		{"TdynWPolaris", args{true, true}, &dynamicPeerFinder{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := createDummyPM()
			mockActor := p2pmock.NewMockActorService(ctrl)
			got := NewPeerFinder(logger, dummyPM, mockActor, 10, tt.args.useDiscover, tt.args.usePolaris)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("NewPeerFinder() = %v, want %v", reflect.TypeOf(got), reflect.TypeOf(tt.want))
			}
		})
	}
}

func Test_dynamicPeerFinder_OnPeerDisconnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		preConnected []types.PeerID
		inMeta       p2pcommon.PeerMeta
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{"TDesignatedPeer", args{desigIDs, desigPeers[0]}, 1},
		{"TNonPeer", args{unknownIDs, unknownPeers[0]}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := createDummyPM()
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(tt.args.inMeta.ID).AnyTimes()
			mockPeer.EXPECT().Meta().Return(tt.args.inMeta).AnyTimes()
			mockPeer.EXPECT().Name().Return(p2putil.ShortMetaForm(tt.args.inMeta)).AnyTimes()

			dp := NewPeerFinder(logger, dummyPM, mockActor, 10, true, false).(*dynamicPeerFinder)
			for _, id := range tt.args.preConnected {
				dummyPM.remotePeers[id] = &remotePeerImpl{}
				dp.OnPeerConnect(id)
			}
			statCnt := len(dp.qStats)
			dp.OnPeerDisconnect(mockPeer)

			if statCnt-1 != len(dp.qStats) {
				t.Errorf("count of query peers was not decreaded %v, want %v", len(dp.qStats), statCnt)
			}
		})
	}
}

func Test_dynamicPeerFinder_OnPeerConnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		preConnected []types.PeerID
		inMeta       p2pcommon.PeerMeta
	}
	tests := []struct {
		name          string
		args          args
		wantStatCount int
	}{
		{"TDesigPeer", args{desigIDs, desigPeers[0]}, 1},
		{"TNonPeer", args{unknownIDs, unknownPeers[0]}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := createDummyPM()
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(tt.args.inMeta.ID).AnyTimes()
			mockPeer.EXPECT().Meta().Return(tt.args.inMeta).AnyTimes()
			mockPeer.EXPECT().Name().Return(p2putil.ShortMetaForm(tt.args.inMeta)).AnyTimes()

			dp := NewPeerFinder(logger, dummyPM, mockActor, 10, true, false).(*dynamicPeerFinder)

			dp.OnPeerConnect(tt.args.inMeta.ID)

			if len(dp.qStats) != tt.wantStatCount {
				t.Errorf("count of query peers was not decreaded %v, want %v", len(dp.qStats), tt.wantStatCount)
			} else {
				if _, exist := dp.qStats[tt.args.inMeta.ID]; !exist {
					t.Errorf("peer query for pid %v missing, want exists", p2putil.ShortForm(tt.args.inMeta.ID))
				}
			}
		})
	}
}
