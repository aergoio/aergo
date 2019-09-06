/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"fmt"
	"github.com/aergoio/aergo/contract/enterprise"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/libp2p/go-libp2p-core/network"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/polaris/common"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/crypto"
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
func (dntc *dummyNTC) ChainID() *types.ChainID {
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
		{"Tlisten", fields{}},
		{"TNot", fields{}},
		// TODO: Add test cases.
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
	mainnetbytes, err := common.ONEMainNet.Bytes()
	if err != nil {
		t.Error("mainnet var is not set properly", common.ONEMainNet)
	}
	dummyPeerID2, err := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")

	goodPeerMeta := p2pcommon.PeerMeta{ID: dummyPeerID2, IPAddress: "211.34.56.78", Port: 7845}
	good := goodPeerMeta.ToPeerAddress()
	badPeerMeta := p2pcommon.PeerMeta{ID: types.PeerID("bad"), IPAddress: "211.34.56.78", Port: 7845}
	bad := badPeerMeta.ToPeerAddress()
	type args struct {
		status *types.Status
		addme  bool
		size   int32
	}
	tests := []struct {
		name string
		args args

		wantErr bool
		wantMsg bool
	}{
		// check if parameter is bad
		{"TMissingStat", args{nil, true, 9999}, true, false},
		// check if addMe is set or not
		{"TOnlyQuery", args{&types.Status{ChainID: mainnetbytes, Sender: &good}, false, 10}, false, false},
		{"TOnlyQuery", args{&types.Status{ChainID: mainnetbytes, Sender: &bad}, false, 10}, false, false},
		// TODO refator mapservice to run commented test
		//{"TAddWithGood",args{&types.Status{ChainID:mainnetbytes, Sender:&good}, true, 10}, false, false },
		//{"TAddWithBad",args{&types.Status{ChainID:mainnetbytes, Sender:&bad}, true, 10}, false , true },
		// TODO: Add more cases .
		// check if failed to connect back or not
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
			mockNT.EXPECT().GetOrCreateStreamWithTTL(gomock.Any(), common.PolarisPingSub, gomock.Any()).Return(mockStream, nil)

			pms := NewPolarisService(pmapDummyCfg, pmapDummyNTC)
			pms.AfterStart()
			query := &types.MapQuery{Status: tt.args.status, AddMe: tt.args.addme, Size: tt.args.size}

			got, err := pms.handleQuery(nil, query)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerMapService.handleQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (len(got.Message) > 0) != tt.wantMsg {
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
		peerid, _ := types.IDFromPublicKey(pub)
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
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
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
			if err := pms.writeResponse(tt.args.reqContainer, tt.args.meta, tt.args.resp, tt.args.wt); (err != nil) != tt.wantErr {
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
	id2 := p2putil.RandomPeerID()
	id3 := p2putil.RandomPeerID()
	ad10 := "123.45.67.89"
	ad11 := "123.45.67.91" // same C class network
	ad2 := "2001:0db8:0123:4567:89ab:cdef:1234:5678"
	ad3 := "222.8.8.8"
	m10 := p2pcommon.PeerMeta{IPAddress: ad10, ID: id1}
	m11 := p2pcommon.PeerMeta{IPAddress: ad10, ID: p2putil.RandomPeerID()}
	m12 := p2pcommon.PeerMeta{IPAddress: ad11, ID: p2putil.RandomPeerID()}
	m2 := p2pcommon.PeerMeta{IPAddress: ad2, ID: id2}
	m3 := p2pcommon.PeerMeta{IPAddress: ad3, ID: id3}

	type args struct {
		entry enterprise.WhiteListEntry
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

func wle(str string) enterprise.WhiteListEntry {
	ent, err := enterprise.NewWhiteListEntry(str)
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
	entries := []enterprise.WhiteListEntry{}
	for _, d := range []string{d0, d1, d2, d3, d4, d5, d6} {
		e, _ := enterprise.NewWhiteListEntry(d)
		entries = append(entries, e)
	}

	type args struct {
		s types.Stream
	}
	tests := []struct {
		name   string
		args   args
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
