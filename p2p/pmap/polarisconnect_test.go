/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package pmap

import (
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/p2p/mocks"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"reflect"
	"sync"
	"testing"
)

// initSvc select Polarises to connect, or disable polaris
func TestPolarisConnectSvc_initSvc(t *testing.T) {
	polarisIDMain,_ := peer.IDB58Decode("16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF")
	polarisIDTest,_ := peer.IDB58Decode("16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF")
	dummyPeerID2, _ := peer.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	polar2 := "/ip4/172.21.1.2/tcp/8915/p2p/16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm"
	dummyPeerID3, _ := peer.IDB58Decode("16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD")
	polar3 := "/ip4/172.22.2.3/tcp/8915/p2p/16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD"

	customChainID := types.ChainID{Magic:"unittest.blocko.io"}
	type args struct {
		use bool
		polarises []string

		chainID *types.ChainID
	}
	tests := []struct {
		name string
		args args

		wantCnt int
		peerIDs []peer.ID
	}{
		//
		{"TAergoNoPolaris", args{false, nil, &ONEMainNet}, 0, []peer.ID{}},
		{"TAergoMainDefault", args{true, nil, &ONEMainNet}, 1, []peer.ID{polarisIDMain}},
		{"TAergoMainPlusCfg", args{true, []string{polar2,polar3},&ONEMainNet}, 3, []peer.ID{polarisIDMain,dummyPeerID2,dummyPeerID3}},
		{"TAergoTestDefault", args{true, nil, &ONETestNet}, 1, []peer.ID{polarisIDTest}},
		{"TAergoTestPlusCfg", args{true, []string{polar2,polar3}, &ONETestNet}, 3, []peer.ID{polarisIDTest,dummyPeerID2,dummyPeerID3}},
		{"TCustom", args{true, nil,&customChainID}, 0, []peer.ID{}},
		{"TCustomPlusCfg", args{true, []string{polar2,polar3},&customChainID}, 2, []peer.ID{dummyPeerID2,dummyPeerID3}},
		{"TWrongPolarisAddr", args{true, []string{"/ip4/256.256.1.1/tcp/8915/p2p/16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD"},&customChainID}, 0, []peer.ID{}},
		{"TWrongPolarisAddr2", args{true, []string{"/egwgew5/p2p/16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD"},&customChainID}, 0, []peer.ID{}},
		{"TWrongPolarisAddr3", args{true, []string{"/dns/nowhere1234.aergo.io/tcp/8915/p2p/16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD"},&customChainID}, 0, []peer.ID{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockNT := mock_p2p.NewMockNetworkTransport(ctrl)
			pmapDummyNTC.nt = mockNT
			pmapDummyNTC.chainID = tt.args.chainID

			cfg:= config.NewServerContext("",""	).GetDefaultP2PConfig()
			cfg.NPUsePolaris = tt.args.use
			cfg.NPAddPolarises = tt.args.polarises

			pcs :=  NewPolarisConnectSvc(cfg, pmapDummyNTC)

			if len(pcs.mapServers) != tt.wantCnt {
				t.Errorf("NewPolarisConnectSvc() = %v, want %v", len(pcs.mapServers), tt.wantCnt)
			}
			for _, wantPeerID := range tt.peerIDs {
				found := false
				for _, polarisMeta := range pcs.mapServers {
					if wantPeerID == polarisMeta.ID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("initSvc() want exist %v but not ", wantPeerID)
				}
			}
		})
	}
}

func TestPolarisConnectSvc_BeforeStop(t *testing.T) {

	type fields struct {
		BaseComponent *component.BaseComponent
	}
	tests := []struct {
		name   string
		fields fields

		calledStreamHandler bool
	}{
		{"TNot", fields{}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockNT := mock_p2p.NewMockNetworkTransport(ctrl)
			pmapDummyNTC.nt = mockNT
			pms := NewPolarisConnectSvc(pmapDummyCfg.P2P, pmapDummyNTC)

			mockNT.EXPECT().AddStreamHandler(PolarisPingSub, gomock.Any()).Times(1)
			mockNT.EXPECT().RemoveStreamHandler(PolarisPingSub).Times(1)

			pms.AfterStart()

			pms.BeforeStop()

			ctrl.Finish()
		})
	}
}


func TestPolarisConnectSvc_onPing(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       []byte
		PrivateNet    bool
		mapServers    []p2pcommon.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	type args struct {
		s net.Stream
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
			pms := &PolarisConnectSvc{
				BaseComponent: tt.fields.BaseComponent,
				PrivateChain:  tt.fields.PrivateNet,
				mapServers:    tt.fields.mapServers,
				ntc:           tt.fields.ntc,
				nt:            tt.fields.nt,
				rwmutex:       tt.fields.rwmutex,
			}
			pms.onPing(tt.args.s)
		})
	}
}


func TestPeerMapService_connectAndQuery(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       []byte
		PrivateNet    bool
		mapServers    []p2pcommon.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	type args struct {
		mapServerMeta p2pcommon.PeerMeta
		bestHash      []byte
		bestHeight    uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*types.PeerAddress
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pms := &PolarisConnectSvc{
				BaseComponent: tt.fields.BaseComponent,
				PrivateChain:  tt.fields.PrivateNet,
				mapServers:    tt.fields.mapServers,
				ntc:           tt.fields.ntc,
				nt:            tt.fields.nt,
				rwmutex:       tt.fields.rwmutex,
			}
			got, err := pms.connectAndQuery(tt.args.mapServerMeta, tt.args.bestHash, tt.args.bestHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("PolarisConnectSvc.connectAndQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PolarisConnectSvc.connectAndQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolarisConnectSvc_sendRequest(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       []byte
		PrivateNet    bool
		mapServers    []p2pcommon.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	type args struct {
		status        *types.Status
		mapServerMeta p2pcommon.PeerMeta
		register      bool
		size          int
		wt            p2p.MsgWriter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pms := &PolarisConnectSvc{
				BaseComponent: tt.fields.BaseComponent,
				PrivateChain:  tt.fields.PrivateNet,
				mapServers:    tt.fields.mapServers,
				ntc:           tt.fields.ntc,
				nt:            tt.fields.nt,
				rwmutex:       tt.fields.rwmutex,
			}
			if err := pms.sendRequest(tt.args.status, tt.args.mapServerMeta, tt.args.register, tt.args.size, tt.args.wt); (err != nil) != tt.wantErr {
				t.Errorf("PolarisConnectSvc.sendRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPolarisConnectSvc_readResponse(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       []byte
		PrivateNet    bool
		mapServers    []p2pcommon.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	type args struct {
		mapServerMeta p2pcommon.PeerMeta
		rd            p2p.MsgReader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    p2pcommon.Message
		want1   *types.MapResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pms := &PolarisConnectSvc{
				BaseComponent: tt.fields.BaseComponent,
				PrivateChain:  tt.fields.PrivateNet,
				mapServers:    tt.fields.mapServers,
				ntc:           tt.fields.ntc,
				nt:            tt.fields.nt,
				rwmutex:       tt.fields.rwmutex,
			}
			got, got1, err := pms.readResponse(tt.args.mapServerMeta, tt.args.rd)
			if (err != nil) != tt.wantErr {
				t.Errorf("PolarisConnectSvc.readResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PolarisConnectSvc.readResponse() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("PolarisConnectSvc.readResponse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
