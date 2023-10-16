/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/polaris/common"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/stretchr/testify/assert"
)

type dummyNTC struct {
	nt      p2pcommon.NetworkTransport
	self    p2pcommon.PeerMeta
	chainID *types.ChainID
}

func (dntc *dummyNTC) SelfMeta() p2pcommon.PeerMeta {
	return dntc.self
}

func (dntc *dummyNTC) GetNetworkTransport() p2pcommon.NetworkTransport {
	return dntc.nt
}
func (dntc *dummyNTC) GenesisChainID() *types.ChainID {
	return dntc.chainID
}

var (
	pmapDummyCfg = &config.Config{P2P: &config.P2PConfig{}, Polaris: &config.PolarisConfig{GenesisFile: "../../examples/genesis.json"},
		Auth: &config.AuthConfig{EnableLocalConf: false}}
	pmapDummyNTC = &dummyNTC{chainID: &types.ChainID{}}
)

func TestPeerMapService_BeforeStop(t *testing.T) {

	type fields struct {
		BaseComponent *component.BaseComponent
		peerRegistry  map[types.PeerID]p2pcommon.PeerMeta
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"TListen", fields{}},
		{"TNot", fields{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			pmapDummyNTC.nt = mockNT
			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)

			mockNT.EXPECT().AddStreamHandler(common.PolarisMapSub, gomock.Any()).Times(1)
			mockNT.EXPECT().RemoveStreamHandler(common.PolarisMapSub).Times(1)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			pmapDummyNTC.nt = mockNT
			mockNT.EXPECT().AddStreamHandler(common.PolarisMapSub, gomock.Any()).Times(1)

			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			pms.AfterStart()

			msgStub := &p2pcommon.MessageValue{}
			mockRd := p2pmock.NewMockMsgReadWriter(ctrl)

			mockRd.EXPECT().ReadMsg().Times(1).Return(msgStub, tt.args.readErr)
			ri := p2pcommon.RemoteInfo{Meta: tt.args.meta}
			got, got1, err := pms.readRequest(ri, mockRd)
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
	minVersion, _ := p2pcommon.ParseAergoVersion(p2pcommon.MinimumAergoVersion)
	tooNewVersion, _ := p2pcommon.ParseAergoVersion(p2pcommon.MaximumAergoVersion)
	tooOldVersion := minVersion
	tooOldVersion.Patch = tooOldVersion.Patch - 1
	mainnetbytes, err := common.ONEMainNet.Bytes()
	if err != nil {
		t.Error("mainnet var is not set properly", common.ONEMainNet)
	}
	dummyPeerID2, err := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	goodAddr, _ := types.ParseMultiaddr("/ip4/211.34.56.78/tcp/7846")
	goodPeerMeta := p2pcommon.PeerMeta{ID: dummyPeerID2, Addresses: []types.Multiaddr{goodAddr}}
	good := goodPeerMeta.ToPeerAddress()
	sameConn := p2pcommon.RemoteConn{IP: net.ParseIP("211.34.56.78"), Port: 42744, Outbound: false}
	//diffConn := p2pcommon.RemoteConn{net.ParseIP("11.55.56.78"), 42742, false}
	badPeerMeta := p2pcommon.PeerMeta{ID: types.PeerID("bad"), Addresses: []types.Multiaddr{goodAddr}}
	bad := badPeerMeta.ToPeerAddress()

	ok := types.ResultStatus_OK

	type args struct {
		conn   p2pcommon.RemoteConn
		status *types.Status
		addme  bool
		size   int32
	}
	tests := []struct {
		name string
		args args

		wantErr    bool
		wantMsg    bool
		wantStatus types.ResultStatus
	}{
		// check if parameter is bad
		{"TMissingStat", args{sameConn, nil, true, 9999}, true, false, ok},
		// check if addMe is set or not
		{"TOnlyQuery", args{sameConn, &types.Status{ChainID: mainnetbytes, Sender: &good, Version: minVersion.String()}, false, 10}, false, false, ok},
		{"TOnlyQuery2", args{sameConn, &types.Status{ChainID: mainnetbytes, Sender: &bad, Version: minVersion.String()}, false, 10}, false, false, ok},
		// TODO refator mapservice to run commented test
		//{"TAddWithGood",args{&types.Status{ChainID:mainnetbytes, Sender:&good}, true, 10}, false, false, ok },
		//{"TAddWithBad",args{&types.Status{ChainID:mainnetbytes, Sender:&bad}, true, 10}, false , true, ok },
		//{"TDiffConn",args{diffConn,&types.Status{ChainID:mainnetbytes, Sender:&good}, true, 10}, false, false, ok },

		// check if failed to connect back or not
		{"TOldVersion", args{sameConn, &types.Status{ChainID: mainnetbytes, Sender: &good, Version: tooOldVersion.String()}, false, 10}, false, true, types.ResultStatus_FAILED_PRECONDITION},
		{"TNewVersion", args{sameConn, &types.Status{ChainID: mainnetbytes, Sender: &good, Version: tooNewVersion.String()}, false, 10}, false, true, types.ResultStatus_FAILED_PRECONDITION},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			mockStream := p2pmock.NewMockStream(ctrl)
			mockStream.EXPECT().Write(gomock.Any()).MaxTimes(1).Return(100, nil)
			mockStream.EXPECT().Close().MaxTimes(1).Return(nil)
			pmapDummyNTC.chainID = &common.ONEMainNet
			pmapDummyNTC.nt = mockNT
			mockNT.EXPECT().AddStreamHandler(gomock.Any(), gomock.Any())
			mockNT.EXPECT().GetOrCreateStreamWithTTL(gomock.Any(), gomock.Any(), common.PolarisPingSub).Return(mockStream, nil).MinTimes(1)

			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			pms.AfterStart()
			query := &types.MapQuery{Status: tt.args.status, AddMe: tt.args.addme, Size: tt.args.size}
			if query.Status != nil && query.Status.Sender != nil {
				query.Status.Sender.Version = query.Status.Version
			}

			got, err := pms.handleQuery(tt.args.conn, nil, query)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerMapService.handleQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (len(got.Message) > 0) != tt.wantMsg {
				t.Errorf("PeerMapService.handleQuery() msg = %v, wantMsg %v", got.Message, tt.wantMsg)
				return
			}
			if tt.wantMsg && got.Status != tt.wantStatus {
				t.Errorf("PeerMapService.handleQuery() msg status = %v, want %v", got.Status.String(), tt.wantStatus.String())
			}

		})
	}
}

var metas []p2pcommon.PeerMeta

func init() {
	metas = make([]p2pcommon.PeerMeta, 20)
	for i := 0; i < 20; i++ {
		peerid := types.RandomPeerID()
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
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			pms.nt = mockNT

			conn := p2pcommon.RemoteConn{IP: net.ParseIP("192.168.1.2"), Port: 7846}
			wg := &sync.WaitGroup{}
			finWg := &sync.WaitGroup{}
			wg.Add(1)
			finWg.Add(len(tt.args))
			for _, meta := range tt.args {
				go func(in p2pcommon.PeerMeta) {
					wg.Wait()
					pms.registerPeer(in, conn)
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
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			conn := p2pcommon.RemoteConn{IP: net.ParseIP("192.168.1.2"), Port: 7846}

			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			pms.nt = mockNT
			for _, meta := range metas {
				pms.registerPeer(meta, conn)
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
			assert.Equal(t, allSize-tt.wantDecrease, len(pms.peerRegistry))

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
		nt            p2pcommon.NetworkTransport
		mutex         *sync.RWMutex
		peerRegistry  map[types.PeerID]*peerState
	}
	type args struct {
		reqContainer p2pcommon.Message
		meta         p2pcommon.PeerMeta
		resp         *types.MapResponse
		wt           p2pcommon.MsgReadWriter
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
			ri := p2pcommon.RemoteInfo{Meta: tt.args.meta}
			if err := pms.writeResponse(tt.args.reqContainer, ri, tt.args.resp, tt.args.wt); (err != nil) != tt.wantErr {
				t.Errorf("PeerMapService.writeResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMapService(t *testing.T) {
	type args struct {
		cfg    *config.Config
		ntc    p2pcommon.NTContainer
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
		ntc           p2pcommon.NTContainer
		listen        bool
		nt            p2pcommon.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[types.PeerID]*peerState
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
		ntc           p2pcommon.NTContainer
		listen        bool
		nt            p2pcommon.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[types.PeerID]*peerState
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
		ntc           p2pcommon.NTContainer
		listen        bool
		nt            p2pcommon.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[types.PeerID]*peerState
	}
	type args struct {
		s network.Stream
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
		ntc           p2pcommon.NTContainer
		listen        bool
		nt            p2pcommon.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[types.PeerID]*peerState
	}
	type args struct {
		maxPeers int
		exclude  types.PeerID
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
		want    *p2pcommon.MessageValue
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
		ntc           p2pcommon.NTContainer
		listen        bool
		nt            p2pcommon.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[types.PeerID]*peerState
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

func TestPeerMapService_applyNewBLEntry(t *testing.T) {
	id1, _ := types.IDB58Decode("16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt")
	id2 := types.RandomPeerID()
	id3 := types.RandomPeerID()
	ad10, _ := types.ParseMultiaddr("/ip4/123.45.67.89/tcp/7846")
	ad11, _ := types.ParseMultiaddr("/ip4/123.45.67.91/tcp/7846") // same C class network
	ad2, _ := types.ParseMultiaddr("/ip6/2001:0db8:0123:4567:89ab:cdef:1234:5678/tcp/7846")
	ad3, _ := types.ParseMultiaddr("/ip4/222.8.8.8/tcp/7846")
	m10 := p2pcommon.PeerMeta{Addresses: []types.Multiaddr{ad10}, ID: id1}
	m11 := p2pcommon.PeerMeta{Addresses: []types.Multiaddr{ad10}, ID: types.RandomPeerID()}
	m12 := p2pcommon.PeerMeta{Addresses: []types.Multiaddr{ad11}, ID: types.RandomPeerID()}
	m2 := p2pcommon.PeerMeta{Addresses: []types.Multiaddr{ad2}, ID: id2}
	m3 := p2pcommon.PeerMeta{Addresses: []types.Multiaddr{ad3}, ID: id3}

	type args struct {
		entry types.WhiteListEntry
	}
	tests := []struct {
		name string
		args args

		wantDeleted []p2pcommon.PeerMeta
	}{
		{"IDOnly", args{wle(`{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt"}`)}, mm(m10)},
		{"IDOnlyNotFound", args{wle(`{"peerid":"16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR"}`)}, mm()},
		{"AddrOnly", args{wle(`{"address":"123.45.67.89"}`)}, mm(m10, m11)},
		{"AddrRange", args{wle(`{"peerid":"", "cidr":"123.45.67.89/24"}`)}, mm(m10, m11, m12)},
		{"IDAddr", args{wle(`{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"123.45.67.89"}`)}, mm(m10)},
		{"IDAddrNotFound", args{wle(`{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"122.1.3.4"}`)}, mm()},
		{"IDAdRange24", args{wle(`{"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "cidr":"123.45.67.1/24"}`)}, mm(m10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			pmapDummyNTC.nt = mockNT
			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			// add sample peers
			for _, m := range mm(m10, m11, m12, m2, m3) {
				pms.peerRegistry[m.ID] = &peerState{PeerMapService: pms, meta: m, addr: m.ToPeerAddress(), lCheckTime: time.Now(), temporary: false}
			}
			prevLen := len(pms.peerRegistry)
			pms.applyNewBLEntry(tt.args.entry)

			if len(pms.peerRegistry) != (prevLen - len(tt.wantDeleted)) {
				t.Errorf("applyNewBLEntry() %v remains, want %v", len(pms.peerRegistry), prevLen-len(tt.wantDeleted))
			}
			for _, m := range tt.wantDeleted {
				if _, found := pms.peerRegistry[m.ID]; found {
					t.Errorf("applyNewBLEntry() peer %v exist, want deleted", p2putil.ShortForm(m.ID))
				}
			}
		})
	}
}

func wle(str string) types.WhiteListEntry {
	ent, err := types.ParseListEntry(str)
	if err != nil {
		panic("invalid input : " + str + " : " + err.Error())
	}
	return ent
}

func mm(metas ...p2pcommon.PeerMeta) []p2pcommon.PeerMeta {
	return metas
}

func TestPeerMapService_onConnectWithBlacklist(t *testing.T) {
	d0 := "{\"address\":\"172.21.3.35\",\"cidr\":\"\",\"peerid\":\"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt\"}"
	d1 := "{\"address\":\"\",\"cidr\":\"172.21.0.0/16\",\"peerid\":\"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt\"}"
	d2 := "{\"address\":\"0.0.0.0\",\"cidr\":\"\",\"peerid\":\"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt\"}"
	d3 := "{\"address\":\"\",\"cidr\":\"172.21.0.0/16\",\"peerid\":\"\"}"
	d4 := "{\"address\":\"2001:db8:123:4567:89ab:cdef:1234:5678\",\"cidr\":\"\",\"peerid\":\"\"}"
	d5 := "{\"address\":\"0.0.0.0\",\"cidr\":\"\",\"peerid\":\"\"}"
	d6 := "{\"address\":\"\",\"cidr\":\"2001:db8:123:4567:89ab:cdef::/96\",\"peerid\":\"\"}"
	entries := []types.WhiteListEntry{}
	for _, d := range []string{d0, d1, d2, d3, d4, d5, d6} {
		e, _ := types.ParseListEntry(d)
		entries = append(entries, e)
	}

	type args struct {
		s types.Stream
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockNT := p2pmock.NewMockNetworkTransport(ctrl)

			pmapDummyNTC.nt = mockNT
			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			for _, e := range entries {
				pms.lm.AddEntry(e)
			}

			pms.onConnect(tt.args.s)
		})
	}
}

func Test_isEqualMeta(t *testing.T) {
	pid1, pid2 := types.RandomPeerID(), types.RandomPeerID()
	v1, v2 := "v1.3.3", "v2.0.0"
	a1, _ := types.ParseMultiaddr("/ip4/192.168.0.58/tcp/11002/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh")
	a2, _ := types.ParseMultiaddr("/ip6/FE80::0202:B3FF:FE1E:8329/tcp/11003/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh")
	a3, _ := types.ParseMultiaddr("/dns4/test.aergo.io/tcp/11002/p2p/16Uiu2HAmHuBgtnisgPLbujFvxPNZw3Qvpk3VLUwTzh5C67LAZSFh")
	addrs := []types.Multiaddr{a1, a2, a3}
	type args struct {
		m1 p2pcommon.PeerMeta
		m2 p2pcommon.PeerMeta
	}
	tests := []struct {
		name   string
		args   args
		wantEq bool
	}{
		{"TEq", args{
			p2pcommon.PeerMeta{ID: pid1, Addresses: addrs, Version: v1, Role: types.PeerRole_Watcher},
			p2pcommon.PeerMeta{ID: pid1, Addresses: addrs, Version: v1, Role: types.PeerRole_Watcher}}, true},
		{"TDiffID", args{
			p2pcommon.PeerMeta{ID: pid1, Addresses: addrs, Version: v1, Role: types.PeerRole_Watcher},
			p2pcommon.PeerMeta{ID: pid2, Addresses: addrs, Version: v1, Role: types.PeerRole_Watcher}}, false},
		{"TDiffVer", args{
			p2pcommon.PeerMeta{ID: pid1, Addresses: addrs, Version: v1, Role: types.PeerRole_Watcher},
			p2pcommon.PeerMeta{ID: pid1, Addresses: addrs, Version: v2, Role: types.PeerRole_Watcher}}, false},
		{"TDiffAddr", args{
			p2pcommon.PeerMeta{ID: pid1, Addresses: addrs, Version: v1, Role: types.PeerRole_Watcher},
			p2pcommon.PeerMeta{ID: pid1, Addresses: addrs[:2], Version: v1, Role: types.PeerRole_Watcher}}, false},
		{"TDiffRole", args{
			p2pcommon.PeerMeta{ID: pid1, Addresses: addrs, Version: v1, Role: types.PeerRole_Watcher},
			p2pcommon.PeerMeta{ID: pid1, Addresses: addrs, Version: v1, Role: types.PeerRole_Producer}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotEq := isEqualMeta(tt.args.m1, tt.args.m2); gotEq != tt.wantEq {
				t.Errorf("isEqualMeta() = %v, want %v", gotEq, tt.wantEq)
			}
		})
	}
}
