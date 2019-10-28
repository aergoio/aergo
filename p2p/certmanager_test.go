package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"reflect"
	"testing"
)

func Test_newCertificateManager(t *testing.T) {
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
			got := newCertificateManager(meta)
			if (got==nil) != tt.wantNil {
				t.Errorf("newCertificateManager() = %v, want nil %v", got, tt.wantNil)
			}
			if !tt.wantNil && reflect.TypeOf(got) != tt.wantType {
				t.Errorf("newCertificateManager() = %v, want %v", reflect.TypeOf(got) , tt.wantType)
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
		name     string
		args     args
		wantNil  bool
	}{
		{"TSingle", args{[]types.PeerID{types.RandomPeerID()}}, false, },
		{"TMulti", args{[]types.PeerID{types.RandomPeerID(),types.RandomPeerID(),types.RandomPeerID()}}, false, },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := p2pcommon.NewMetaWith1Addr(samplePeerID, "192.168.0.6", 7846, "v2.0.0")
			meta.Role = types.PeerRole_Agent
			meta.ProducerIDs = tt.args.pds
			got := newCertificateManager(meta)
			if (got==nil) != tt.wantNil {
				t.Errorf("newCertificateManager() = %v, want nil %v", got, tt.wantNil)
			}
			if !tt.wantNil && len(got.GetProducers()) != len(tt.args.pds) {
				t.Errorf("newCertificateManager() = %v, want %v",got.GetProducers() , tt.args.pds)
			}
		})
	}
}
