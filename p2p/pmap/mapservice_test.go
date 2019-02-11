/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package pmap

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"reflect"
	"sync"
	"testing"

	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/p2p/mocks"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
)

type dummyNTC struct {
	nt p2p.NetworkTransport
	chainID *types.ChainID
}

func (dntc *dummyNTC) GetNetworkTransport() p2p.NetworkTransport {
	return dntc.nt
}
func (dntc *dummyNTC) ChainID() *types.ChainID {
	return dntc.chainID
}

var (
	pmapDummyCfg = &config.Config{P2P:&config.P2PConfig{},Polaris:&config.PolarisConfig{GenesisFile:"../../examples/genesis.json"}}
	pmapDummyNTC = &dummyNTC{chainID:&types.ChainID{}}
)

func TestPeerMapService_BeforeStop(t *testing.T) {

	type fields struct {
		BaseComponent *component.BaseComponent
		peerRegistry  map[peer.ID]p2pcommon.PeerMeta
	}
	tests := []struct {
		name   string
		fields fields

	}{
		{"Tlisten", fields{}},
		{"TNot", fields{}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockNT := mock_p2p.NewMockNetworkTransport(ctrl)
			pmapDummyNTC.nt = mockNT
			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)

			mockNT.EXPECT().AddStreamHandler(PolarisMapSub, gomock.Any()).Times(1)
			mockNT.EXPECT().RemoveStreamHandler(PolarisMapSub).Times(1)

			pms.AfterStart()

			pms.BeforeStop()

			ctrl.Finish()
		})
	}
}

func TestPeerMapService_readRequest(t *testing.T) {
	dummyMeta := p2pcommon.PeerMeta{ID: ""}
	type args struct {
		meta    p2pcommon.PeerMeta
		readErr error
	}
	tests := []struct {
		name string
		args args

		//want    p2p.Message
		//want1   *types.MapQuery
		wantErr bool
	}{
		{"TNormal", args{meta: dummyMeta, readErr: nil}, false},
		{"TError", args{meta: dummyMeta, readErr: fmt.Errorf("testerr")}, true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockNT := mock_p2p.NewMockNetworkTransport(ctrl)
			pmapDummyNTC.nt = mockNT
			mockNT.EXPECT().AddStreamHandler(PolarisMapSub, gomock.Any()).Times(1)

			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			pms.AfterStart()

			msgStub := &p2p.V030Message{}
			mockRd := mock_p2p.NewMockMsgReader(ctrl)

			mockRd.EXPECT().ReadMsg().Times(1).Return(msgStub, tt.args.readErr)

			got, got1, err := pms.readRequest(tt.args.meta, mockRd)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerMapService.readRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Errorf("PeerMapService.readRequest() got = %v, want %v", got, "not nil")
				}
				if got1 == nil {
					t.Errorf("PeerMapService.readRequest() got = %v, want %v", got, "not nil")
				}
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("PeerMapService.readRequest() got = %v, want %v", got, tt.want)
			//}
			//if !reflect.DeepEqual(got1, tt.want1) {
			//	t.Errorf("PeerMapService.readRequest() got1 = %v, want %v", got1, tt.want1)
			//}
			ctrl.Finish()

		})
	}
}

func TestPeerMapService_handleQuery(t *testing.T) {
	mainnetbytes,err := ONEMainNet.Bytes()
	if err != nil {
		t.Error("mainnet var is not set properly", ONEMainNet)
	}
	dummyPeerID2, err := peer.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")

	goodPeerMeta := p2pcommon.PeerMeta{ID: dummyPeerID2, IPAddress:"211.34.56.78",Port:7845}
	good := goodPeerMeta.ToPeerAddress()
	badPeerMeta := p2pcommon.PeerMeta{ID: peer.ID("bad"), IPAddress:"211.34.56.78",Port:7845}
	bad := badPeerMeta.ToPeerAddress()
	type args struct {
		status *types.Status
		addme bool
		size int32
	}
	tests := []struct {
		name    string
		args    args

		wantErr bool
		wantMsg bool
	}{
		// check if parameter is bad
		{"TMissingStat",args{nil, true, 9999}, true ,false},
		// check if addMe is set or not
		{"TOnlyQuery",args{&types.Status{ChainID:mainnetbytes, Sender:&good}, false, 10}, false,false },
		{"TOnlyQuery",args{&types.Status{ChainID:mainnetbytes, Sender:&bad}, false, 10}, false,false },
		// TODO refator mapservice to run commented test
		//{"TAddWithGood",args{&types.Status{ChainID:mainnetbytes, Sender:&good}, true, 10}, false, false },
		//{"TAddWithBad",args{&types.Status{ChainID:mainnetbytes, Sender:&bad}, true, 10}, false , true },
		// TODO: Add more cases .
		// check if failed to connect back or not
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockNT := mock_p2p.NewMockNetworkTransport(ctrl)
			mockStream := mock_p2p.NewMockStream(ctrl)
			mockStream.EXPECT().Write(gomock.Any()).MaxTimes(1).Return(100, nil)
			mockStream.EXPECT().Close().MaxTimes(1).Return(nil)
			pmapDummyNTC.chainID = &ONEMainNet
			pmapDummyNTC.nt = mockNT
			mockNT.EXPECT().AddStreamHandler(gomock.Any(), gomock.Any())
			mockNT.EXPECT().GetOrCreateStreamWithTTL(gomock.Any(),PolarisPingSub, gomock.Any()).Return(mockStream, nil)

			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			pms.AfterStart()
			query := &types.MapQuery{Status:tt.args.status, AddMe:tt.args.addme, Size:tt.args.size}

			got, err := pms.handleQuery(nil, query)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerMapService.handleQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (len(got.Message)>0) != tt.wantMsg {
				t.Errorf("PeerMapService.handleQuery() msg = %v, wantMsg %v", got.Message, tt.wantMsg)
				return
			}

		})
	}
}

var metas []p2pcommon.PeerMeta

func init() {
	metas = make([]p2pcommon.PeerMeta, 20)
	for i := 0; i < 20; i++ {
		_, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		peerid, _ := peer.IDFromPublicKey(pub)
		metas[i] = p2pcommon.PeerMeta{ID: peerid}
	}
}

func TestPeerMapService_registerPeer(t *testing.T) {
	dupMetas := MakeMetaSlice(metas[2:5], metas[3:7])

	tests := []struct {
		name     string
		args     []p2pcommon.PeerMeta
		wantSize int
		wantErr  bool
	}{
		{"TSingle", metas[:1], 1, false},
		{"TMulti", metas[:5], 5, false},
		{"TDup", dupMetas, 5, false},
		{"TConcurrent", metas, 20, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockNT := mock_p2p.NewMockNetworkTransport(ctrl)
			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			pms.nt = mockNT

			wg := &sync.WaitGroup{}
			finWg := &sync.WaitGroup{}
			wg.Add(1)
			finWg.Add(len(tt.args))
			for _, meta := range tt.args {
				go func(in p2pcommon.PeerMeta) {
					wg.Wait()
					pms.registerPeer(in)
					finWg.Done()
				}(meta)
			}
			wg.Done()
			finWg.Wait()
			assert.Equal(t, tt.wantSize, len(pms.peerRegistry))

			ctrl.Finish()

		})
	}
}


func TestPeerMapService_unregisterPeer(t *testing.T) {
	dupMetas := MakeMetaSlice(metas[2:5], metas[3:7])
	allSize := len(metas)
	tests := []struct {
		name         string
		args         []p2pcommon.PeerMeta
		wantDecrease int
		wantErr      bool
	}{
		{"TSingle", metas[:1], 1, false},
		{"TMulti", metas[:5], 5, false},
		{"TDup", dupMetas, 5, false},
		{"TConcurrent", metas, 20, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockNT := mock_p2p.NewMockNetworkTransport(ctrl)
			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			pms.nt = mockNT
			for _, meta := range metas {
				pms.registerPeer(meta)
			}
			wg := &sync.WaitGroup{}
			finWg := &sync.WaitGroup{}
			wg.Add(1)
			finWg.Add(len(tt.args))
			for _, meta := range tt.args {
				go func(in p2pcommon.PeerMeta) {
					wg.Wait()
					pms.unregisterPeer(in.ID)
					finWg.Done()
				}(meta)
			}
			wg.Done()
			finWg.Wait()
			assert.Equal(t, allSize - tt.wantDecrease, len(pms.peerRegistry))

			ctrl.Finish()

		})
	}
}

func MakeMetaSlice(slis ...[]p2pcommon.PeerMeta) []p2pcommon.PeerMeta {
	result := make([]p2pcommon.PeerMeta, 0, 10)
	for _, sli := range slis {
		result = append(result, sli...)
	}
	return result
}

func TestPeerMapService_writeResponse(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		listen        bool
		nt            p2p.NetworkTransport
		mutex         *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	type args struct {
		reqContainer p2pcommon.Message
		meta         p2pcommon.PeerMeta
		resp         *types.MapResponse
		wt           p2p.MsgWriter
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
			pms := &PeerMapService{
				BaseComponent: tt.fields.BaseComponent,
				nt:            tt.fields.nt,
				rwmutex:       tt.fields.mutex,
				peerRegistry:  tt.fields.peerRegistry,
			}
			if err := pms.writeResponse(tt.args.reqContainer, tt.args.meta, tt.args.resp, tt.args.wt); (err != nil) != tt.wantErr {
				t.Errorf("PeerMapService.writeResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMapService(t *testing.T) {
	type args struct {
		cfg    *config.Config
		ntc    p2p.NTContainer
		listen bool
	}
	tests := []struct {
		name string
		args args
		want *PeerMapService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPolarisService(tt.args.cfg, tt.args.ntc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPolarisService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerMapService_BeforeStart(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       *types.ChainID
		PrivateNet    bool
		mapServers    []p2pcommon.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pms := &PeerMapService{
				BaseComponent: tt.fields.BaseComponent,
				PrivateNet:    tt.fields.PrivateNet,
				ntc:           tt.fields.ntc,
				nt:            tt.fields.nt,
				hc:            tt.fields.hc,
				rwmutex:       tt.fields.rwmutex,
				peerRegistry:  tt.fields.peerRegistry,
			}
			pms.BeforeStart()
		})
	}
}

func TestPeerMapService_AfterStart(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       *types.ChainID
		PrivateNet    bool
		mapServers    []p2pcommon.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pms := &PeerMapService{
				BaseComponent: tt.fields.BaseComponent,
				PrivateNet:    tt.fields.PrivateNet,
				ntc:           tt.fields.ntc,
				nt:            tt.fields.nt,
				hc:            tt.fields.hc,
				rwmutex:       tt.fields.rwmutex,
				peerRegistry:  tt.fields.peerRegistry,
			}
			pms.AfterStart()
		})
	}
}

func TestPeerMapService_onConnect(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       *types.ChainID
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
			pms := &PeerMapService{
				BaseComponent: tt.fields.BaseComponent,
				PrivateNet:    tt.fields.PrivateNet,
				ntc:           tt.fields.ntc,
				nt:            tt.fields.nt,
				hc:            tt.fields.hc,
				rwmutex:       tt.fields.rwmutex,
				peerRegistry:  tt.fields.peerRegistry,
			}
			pms.onConnect(tt.args.s)
		})
	}
}

func TestPeerMapService_retrieveList(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       *types.ChainID
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
		maxPeers int
		exclude  peer.ID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*types.PeerAddress
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pms := &PeerMapService{
				BaseComponent: tt.fields.BaseComponent,
				PrivateNet:    tt.fields.PrivateNet,
				ntc:           tt.fields.ntc,
				nt:            tt.fields.nt,
				hc:            tt.fields.hc,
				rwmutex:       tt.fields.rwmutex,
				peerRegistry:  tt.fields.peerRegistry,
			}
			if got := pms.retrieveList(tt.args.maxPeers, tt.args.exclude); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerMapService.retrieveList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createV030Message(t *testing.T) {
	type args struct {
		msgID       p2pcommon.MsgID
		orgID       p2pcommon.MsgID
		subProtocol p2pcommon.SubProtocol
		innerMsg    proto.Message
	}
	tests := []struct {
		name    string
		args    args
		want    *p2p.V030Message
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createV030Message(tt.args.msgID, tt.args.orgID, tt.args.subProtocol, tt.args.innerMsg)
			if (err != nil) != tt.wantErr {
				t.Errorf("createV030Message() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createV030Message() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerMapService_getPeerCheckers(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       *types.ChainID
		PrivateNet    bool
		mapServers    []p2pcommon.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	tests := []struct {
		name   string
		fields fields
		want   []peerChecker
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pms := &PeerMapService{
				BaseComponent: tt.fields.BaseComponent,
				PrivateNet:    tt.fields.PrivateNet,
				ntc:           tt.fields.ntc,
				nt:            tt.fields.nt,
				hc:            tt.fields.hc,
				rwmutex:       tt.fields.rwmutex,
				peerRegistry:  tt.fields.peerRegistry,
			}
			if got := pms.getPeerCheckers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerMapService.getPeerCheckers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeGoAwayMsg(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name    string
		args    args
		want    p2pcommon.Message
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeGoAwayMsg(tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeGoAwayMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeGoAwayMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
