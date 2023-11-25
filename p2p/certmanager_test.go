package p2p

import (
	"reflect"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-peerstore/test"
)

func Test_newCertificateManager(t *testing.T) {
	logger := log.NewLogger("p2p.test")
	samplePeerID := types.RandomPeerID()
	type args struct {
		role types.PeerRole
	}
	tests := []struct {
		name     string
		args     args
		wantNil  bool
		wantType reflect.Type
	}{
		{"TProducer", args{types.PeerRole_Producer}, false, reflect.TypeOf(&bpCertificateManager{})},
		{"TAgent", args{types.PeerRole_Agent}, false, reflect.TypeOf(&agentCertificateManager{})},
		{"TWatcher", args{types.PeerRole_Watcher}, false, reflect.TypeOf(&watcherCertificateManager{})},
		{"TWrong", args{99999}, true, reflect.TypeOf(&watcherCertificateManager{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := p2pcommon.NewMetaWith1Addr(samplePeerID, "192.168.0.6", 7846, "v2.0.0")
			meta.Role = tt.args.role
			meta.ProducerIDs = []types.PeerID{types.RandomPeerID()}
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			is := p2pmock.NewMockInternalService(ctrl)
			is.EXPECT().SelfMeta().Return(meta)
			is.EXPECT().LocalSettings().Return(p2pcommon.LocalSettings{}).MaxTimes(1)

			got := newCertificateManager(nil, is, logger)
			if (got == nil) != tt.wantNil {
				t.Errorf("newCertificateManager() = %v, want nil %v", got, tt.wantNil)
			}
			if !tt.wantNil && reflect.TypeOf(got) != tt.wantType {
				t.Errorf("newCertificateManager() = %v, want %v", reflect.TypeOf(got), tt.wantType)
			}
		})
	}
}

func Test_newCertificateManagerAgent(t *testing.T) {
	samplePeerID := types.RandomPeerID()
	type args struct {
		pds []types.PeerID
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{"TSingle", args{[]types.PeerID{types.RandomPeerID()}}, false},
		{"TMulti", args{[]types.PeerID{types.RandomPeerID(), types.RandomPeerID(), types.RandomPeerID()}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := p2pcommon.NewMetaWith1Addr(samplePeerID, "192.168.0.6", 7846, "v2.0.0")
			meta.Role = types.PeerRole_Agent
			meta.ProducerIDs = tt.args.pds
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			is := p2pmock.NewMockInternalService(ctrl)
			is.EXPECT().SelfMeta().Return(meta)
			is.EXPECT().LocalSettings().Return(p2pcommon.LocalSettings{}).MaxTimes(1)

			got := newCertificateManager(nil, is, logger)
			if (got == nil) != tt.wantNil {
				t.Errorf("newCertificateManager() = %v, want nil %v", got, tt.wantNil)
			}
			if !tt.wantNil && len(got.GetProducers()) != len(tt.args.pds) {
				t.Errorf("newCertificateManager() = %v, want %v", got.GetProducers(), tt.args.pds)
			}
		})
	}
}

func Test_agentCertificateManager_AddCertificate(t *testing.T) {
	logger := log.NewLogger("p2p.test")
	agentID := types.RandomPeerID()

	addrs := []string{"192.168.1.2"}
	bpSize := 4
	bpKeys := make([]crypto.PrivKey, bpSize)
	bpIds := make([]types.PeerID, bpSize)
	bpCerts := make([]*p2pcommon.AgentCertificateV1, bpSize)
	for i := 0; i < bpSize; i++ {
		bpKeys[i], _, _ = crypto.GenerateKeyPair(crypto.Secp256k1, 11)
		bpIds[i], _ = types.IDFromPrivateKey(bpKeys[i])
		bpCerts[i], _ = p2putil.NewAgentCertV1(bpIds[i], agentID, p2putil.ConvertPKToBTCEC(bpKeys[i]), addrs, time.Hour)
	}
	wrongCert, _ := p2putil.NewAgentCertV1(bpIds[0], types.RandomPeerID(), p2putil.ConvertPKToBTCEC(bpKeys[0]), addrs, time.Hour)

	type args struct {
		cert *p2pcommon.AgentCertificateV1
	}
	tests := []struct {
		name        string
		producerIds []types.PeerID

		args args

		wantNotify    int
		wantIncreased bool
	}{
		{"TAdded", bpIds, args{bpCerts[3]}, 1, true},
		{"TReplace", bpIds, args{bpCerts[1]}, 1, false},
		// not my cert
		{"TNotMyCert", bpIds, args{wrongCert}, 0, false},
		// not from accounted bp
		{"TUnknownBP", bpIds[2:], args{bpCerts[1]}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sampleMeta := p2pcommon.NewMetaWith1Addr(agentID, addrs[0], 7846, "v2.0.0")
			sampleMeta.Role = types.PeerRole_Agent
			sampleMeta.ProducerIDs = tt.producerIds
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockActor.EXPECT().TellRequest(message.P2PSvc, gomock.Any()).Times(tt.wantNotify)
			cm := &agentCertificateManager{
				baseCertManager: baseCertManager{actor: mockActor, self: sampleMeta, logger: logger},
				certs:           bpCerts[:2],
				certMap:         make(map[types.PeerID]*p2pcommon.AgentCertificateV1),
			}
			for _, c := range cm.certs {
				cm.certMap[c.BPID] = c
			}
			prevSize := len(cm.certs)

			cm.AddCertificate(tt.args.cert)
			if tt.wantIncreased && prevSize == len(cm.certs) {
				t.Errorf("AddCertificate() size = %v, want increase  %v", len(cm.certs), tt.wantIncreased)
			}
			if tt.wantNotify > 0 {
				if len(cm.certMap) != len(cm.certs) {
					t.Fatalf("AddCertificate() size of certificates and map is differ ! %v but %v", len(cm.certMap), len(cm.certs))
				}
				for _, c := range cm.certs {
					if _, found := cm.certMap[c.BPID]; !found {
						t.Errorf("AddCertificate() want exist = %v, but not", p2putil.ShortForm(c.BPID))
					}
				}
			}
		})
	}
}

func Test_bpCertificateManager_CreateCertificate(t *testing.T) {
	sampleVersion := "v2.0.0"
	sampleKey, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 356)
	selfID, _ := types.IDFromPrivateKey(sampleKey)
	agentID := types.RandomPeerID()

	sampleAddrs := []types.Multiaddr{test.Multiaddr("/ip4/192.168.1.2/tcp/7846"), test.Multiaddr("/ip4/172.21.3.4/tcp/7846"), test.Multiaddr("/dns4/test.aergo.io/tcp/7846")}
	samplePids := []types.PeerID{selfID, types.RandomPeerID(), types.RandomPeerID(), types.RandomPeerID()}
	type args struct {
		id    types.PeerID
		addrs []types.Multiaddr
		pids  []types.PeerID
	}
	tests := []struct {
		name string
		args args

		wantErr    bool
		wantNotify int
	}{
		{"TSingleAddr", args{agentID, sampleAddrs[:1], samplePids}, false, 1},
		{"TMultiAddr", args{agentID, sampleAddrs, samplePids}, false, 1},

		{"TNotMyAgent", args{types.RandomPeerID(), sampleAddrs[:0], samplePids}, true, 0},
		{"TNotHisBP", args{types.RandomPeerID(), sampleAddrs[:0], samplePids[1:]}, true, 0},
		{"TNoProducer", args{types.RandomPeerID(), sampleAddrs[:0], samplePids[:0]}, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sampleMeta := p2pcommon.NewMetaWith1Addr(selfID, "172.12.1.1", 7846, "v2.0.0")
			sampleMeta.Role = types.PeerRole_Producer
			mockActor := p2pmock.NewMockActorService(ctrl)
			sampleSetting := p2pcommon.LocalSettings{AgentID: agentID}
			//mockActor.EXPECT().TellRequest(message.P2PSvc, gomock.Any()).Times(tt.wantNotify)

			inMeta := p2pcommon.PeerMeta{ID: tt.args.id, Addresses: tt.args.addrs, Role: types.PeerRole_Agent, ProducerIDs: tt.args.pids, Version: sampleVersion}
			cm := &bpCertificateManager{
				baseCertManager: baseCertManager{actor: mockActor, self: sampleMeta, settings: sampleSetting, logger: logger},
				key:             p2putil.ConvertPKToBTCEC(sampleKey),
			}

			got, err := cm.CreateCertificate(inMeta)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !types.IsSamePeerID(got.AgentID, sampleSetting.AgentID) {
					t.Errorf("CreateCertificate() agentID = %v, want %v", got.AgentID, sampleSetting.AgentID)
					return
				}
			}
		})
	}
}

func Test_agentCertificateManager_checkCertificates(t *testing.T) {
	agentID := types.RandomPeerID()

	addrs := []string{"192.168.1.2"}
	bpSize := 12
	bpKeys := make([]crypto.PrivKey, bpSize)
	bpIds := make([]types.PeerID, bpSize)
	bpCerts := make([]*p2pcommon.AgentCertificateV1, bpSize)
	// make differnent expire times
	for i := 0; i < bpSize; i++ {
		bpKeys[i], _, _ = crypto.GenerateKeyPair(crypto.Secp256k1, 11)
		bpIds[i], _ = types.IDFromPrivateKey(bpKeys[i])
		bpCerts[i], _ = p2putil.NewAgentCertV1(bpIds[i], agentID, p2putil.ConvertPKToBTCEC(bpKeys[i]), addrs, time.Hour*time.Duration(i)+time.Hour*12)
	}
	testTime := time.Now().Add(time.Minute * 2) // considering time error

	type args struct {
		now time.Time
	}
	tests := []struct {
		name string
		arg  time.Time

		wantReqIssue int
		wantDelete   int
	}{
		{"TallYoung", testTime, 0, 0},
		{"TallYoung2", testTime.Add(time.Hour * 5), 0, 0},
		{"TUpdate1", testTime.Add(time.Hour * 6), 1, 0},
		{"TExpire1", testTime.Add(time.Hour * 12), 7, 1},
		{"TAllExpire", testTime.Add(time.Hour * 24), 12, 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sampleMeta := p2pcommon.NewMetaWith1Addr(agentID, "172.12.1.1", 7846, "v2.0.0")
			sampleMeta.Role = types.PeerRole_Agent
			mockActor := p2pmock.NewMockActorService(ctrl)
			sampleSetting := p2pcommon.LocalSettings{AgentID: agentID}
			mockActor.EXPECT().TellRequest(message.P2PSvc, gomock.Any()).Times(tt.wantReqIssue)

			cm := &agentCertificateManager{
				baseCertManager: baseCertManager{actor: mockActor, self: sampleMeta, settings: sampleSetting, logger: logger},
				certs:           []*p2pcommon.AgentCertificateV1{},
				certMap:         make(map[types.PeerID]*p2pcommon.AgentCertificateV1),
			}
			for _, c := range bpCerts {
				cm.certs = append(cm.certs, c)
				cm.certMap[c.BPID] = c
			}
			prevSize := len(cm.certs)

			cm.checkCertificates(tt.arg)
			size := len(cm.certs)
			if size != (prevSize - tt.wantDelete) {
				t.Errorf("checkCertificates() remained certSize %v , want %v", size, (prevSize - tt.wantDelete))
			}
		})
	}
}
