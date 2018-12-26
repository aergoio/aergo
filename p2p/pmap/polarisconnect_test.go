/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package pmap

import (
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/p2p/mocks"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"reflect"
	"sync"
	"testing"
)

func TestNewPolarisConnectSvc(t *testing.T) {
	type args struct {
		cfg *config.P2PConfig
		ntc p2p.NTContainer
	}
	tests := []struct {
		name string
		args args
		want *PolarisConnectSvc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPolarisConnectSvc(tt.args.cfg, tt.args.ntc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPolarisConnectSvc() = %v, want %v", got, tt.want)
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
			pms := NewPolarisConnectSvc(pmapDummyCfg, pmapDummyNTC)

			mockNT.EXPECT().AddStreamHandler(PolarisPingSub, gomock.Any()).Times(1)
			mockNT.EXPECT().RemoveStreamHandler(PolarisPingSub).Times(1)

			pms.AfterStart()

			pms.BeforeStop()

			ctrl.Finish()
		})
	}
}


func TestPeerMapService_onPing(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       []byte
		PrivateNet    bool
		mapServers    []p2p.PeerMeta
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
				ChainID:       tt.fields.ChainID,
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
		mapServers    []p2p.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	type args struct {
		mapServerMeta p2p.PeerMeta
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
				ChainID:       tt.fields.ChainID,
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

func TestPeerMapService_sendRequest(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       []byte
		PrivateNet    bool
		mapServers    []p2p.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	type args struct {
		status        *types.Status
		mapServerMeta p2p.PeerMeta
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
				ChainID:       tt.fields.ChainID,
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

func TestPeerMapService_readResponse(t *testing.T) {
	type fields struct {
		BaseComponent *component.BaseComponent
		ChainID       []byte
		PrivateNet    bool
		mapServers    []p2p.PeerMeta
		ntc           p2p.NTContainer
		listen        bool
		nt            p2p.NetworkTransport
		hc            HealthCheckManager
		rwmutex       *sync.RWMutex
		peerRegistry  map[peer.ID]*peerState
	}
	type args struct {
		mapServerMeta p2p.PeerMeta
		rd            p2p.MsgReader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    p2p.Message
		want1   *types.MapResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pms := &PolarisConnectSvc{
				BaseComponent: tt.fields.BaseComponent,
				ChainID:       tt.fields.ChainID,
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
