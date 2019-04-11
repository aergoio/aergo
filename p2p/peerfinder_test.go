/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"reflect"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/golang/mock/gomock"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
)

const desigCnt = 10

var (
	desigIDs     []peer.ID
	desigPeers   []p2pcommon.PeerMeta
	desigPeerMap = make(map[peer.ID]p2pcommon.PeerMeta)

	unknowIDs   []peer.ID
	unknowPeers []p2pcommon.PeerMeta
)

func init() {
	desigIDs = make([]peer.ID, desigCnt)
	desigPeers = make([]p2pcommon.PeerMeta, desigCnt)
	for i := 0; i < desigCnt; i++ {
		priv, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		pid, _ := peer.IDFromPrivateKey(priv)
		desigIDs[i] = pid
		desigPeers[i] = p2pcommon.PeerMeta{ID: pid, Designated: true}
		desigPeerMap[desigIDs[i]] = desigPeers[i]
	}
	unknowIDs = make([]peer.ID, desigCnt)
	unknowPeers = make([]p2pcommon.PeerMeta, desigCnt)
	for i := 0; i < desigCnt; i++ {
		priv, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		pid, _ := peer.IDFromPrivateKey(priv)
		unknowIDs[i] = pid
		unknowPeers[i] = p2pcommon.PeerMeta{ID: pid, Designated: false}
	}
}
func createDummyPM() *peerManager {
	dummyPM := &peerManager{designatedPeers: desigPeerMap,
		remotePeers:  make(map[peer.ID]*remotePeerImpl),
		awaitPeers:   make(map[peer.ID]*reconnectJob, 10),
		waitingPeers: make(map[peer.ID]*p2pcommon.WaitingPeer, 10),
	}
	return dummyPM
}

func TestNewPeerFinder(t *testing.T) {
	ctrl := gomock.NewController(t)

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
		// TODO: Add test cases.
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

func Test_staticPeerFinder_OnDiscoveredPeers(t *testing.T) {
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
			dp := NewPeerFinder(logger, dummyPM, mockActor, 10, false, false)

			dp.OnDiscoveredPeers(tt.args.metas)
			if len(dummyPM.waitingPeers) != tt.wantCount {
				t.Errorf("count waitingPeer %v, want %v", len(dummyPM.waitingPeers), tt.wantCount)
			}
		})
	}
}

func Test_dynamicPeerFinder_OnDiscoveredPeers(t *testing.T) {
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
			dp := NewPeerFinder(logger, dummyPM, mockActor, 10, true, false)
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

func Test_staticPeerFinder_OnPeerDisconnect(t *testing.T) {
	ctrl := gomock.NewController(t)

	type args struct {
		inMeta p2pcommon.PeerMeta
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{"TDesgintedPeer", args{desigPeers[0]}, 1},
		// it should not occur, though.
		{"TNonPeer", args{unknowPeers[0]}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := createDummyPM()
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(tt.args.inMeta.ID).AnyTimes()
			mockPeer.EXPECT().Meta().Return(tt.args.inMeta).AnyTimes()

			dp := NewPeerFinder(logger, dummyPM, mockActor, 10, false, false)
			dp.OnPeerDisconnect(mockPeer)

			if len(dummyPM.waitingPeers) != tt.wantCount {
				t.Errorf("count waitingPeer %v, want %v", len(dummyPM.waitingPeers), tt.wantCount)
			}
		})
	}
}

func Test_dynamicPeerFinder_OnPeerDisconnect(t *testing.T) {
	ctrl := gomock.NewController(t)

	type args struct {
		preConnected []peer.ID
		inMeta       p2pcommon.PeerMeta
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{"TDesgintedPeer", args{desigIDs, desigPeers[0]}, 1},
		// it should not occur, though.
		{"TNonPeer", args{unknowIDs, unknowPeers[0]}, 0},
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
			if len(dummyPM.waitingPeers) != tt.wantCount {
				t.Errorf("count waitingPeer %v, want %v", len(dummyPM.waitingPeers), tt.wantCount)
			}
			if statCnt-1 != len(dp.qStats) {
				t.Errorf("count of query peers was not decreaded %v, want %v", len(dp.qStats), statCnt)
			}
		})
	}
}

func Test_staticPeerFinder_OnPeerConnect(t *testing.T) {
	type fields struct {
		pm     *peerManager
		logger *log.Logger
	}
	type args struct {
		pid peer.ID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := &staticPeerFinder{
				pm:     tt.fields.pm,
				logger: tt.fields.logger,
			}
			dp.OnPeerConnect(tt.args.pid)
		})
	}
}

func Test_dynamicPeerFinder_OnPeerConnect(t *testing.T) {
	type fields struct {
		logger       *log.Logger
		pm           *peerManager
		actorService p2pcommon.ActorService
		usePolaris   bool
		qStats       map[peer.ID]*queryStat
		maxCap       int
		polarisTurn  time.Time
	}
	type args struct {
		pid peer.ID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := &dynamicPeerFinder{
				logger:       tt.fields.logger,
				pm:           tt.fields.pm,
				actorService: tt.fields.actorService,
				usePolaris:   tt.fields.usePolaris,
				qStats:       tt.fields.qStats,
				maxCap:       tt.fields.maxCap,
				polarisTurn:  tt.fields.polarisTurn,
			}
			dp.OnPeerConnect(tt.args.pid)
		})
	}
}
