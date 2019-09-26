/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/p2p/p2pkey"
	"github.com/aergoio/aergo/types"
	"testing"
)


const (
	sampleKeyFile = "../test/sample.key"
)

func init() {
	//sampleID := "16Uiu2HAmP2iRDpPumUbKhNnEngoxAUQWBmCyn7FaYUrkaDAMXJPJ"
	baseCfg := &config.BaseConfig{AuthDir: "test"}
	p2pCfg := &config.P2PConfig{NPKey: sampleKeyFile}
	p2pkey.InitNodeInfo(baseCfg, p2pCfg, "0.0.1-test", log.NewLogger("test.p2p"))
}

func Test_initMeta(t *testing.T) {
	samplePeerID, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")

	type args struct {
		peerID   types.PeerID
		noExpose bool
	}
	tests := []struct {
		name string
		conf *config.P2PConfig

		args args

		wantSameAddr bool
		wantPort     uint32
		wantID       types.PeerID
		wantHidden   bool
	}{
		{"TIP6", &config.P2PConfig{NetProtocolAddr: "fe80::dcbf:beff:fe87:e30a", NetProtocolPort: 7845, NPExposeSelf:true}, args{samplePeerID, false}, true, 7845, samplePeerID, false},
		{"TIP4", &config.P2PConfig{NetProtocolAddr: "211.1.1.2", NetProtocolPort: 7845, NPExposeSelf:true}, args{samplePeerID, false}, true, 7845, samplePeerID, false},
		{"TDNS", &config.P2PConfig{NetProtocolAddr: "www.aergo.io", NetProtocolPort: 7845, NPExposeSelf:true}, args{samplePeerID, false}, false, 7845, samplePeerID, false},
		{"TDefault", &config.P2PConfig{NetProtocolAddr: "", NetProtocolPort: 7845, NPExposeSelf:true}, args{samplePeerID, false}, false, 7845, samplePeerID, false},
		{"THidden", &config.P2PConfig{NetProtocolAddr: "211.1.1.2", NetProtocolPort: 7845, NPExposeSelf:false}, args{samplePeerID, true}, true, 7845, samplePeerID, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetupSelfMeta(tt.args.peerID, tt.conf, false)

			if tt.wantSameAddr && got.PrimaryAddress() != tt.conf.NetProtocolAddr {
				t.Errorf("SetupSelfMeta() addr = %v , want %v",got.PrimaryAddress() , tt.conf.NetProtocolAddr )
			}
			if got.PrimaryPort() != tt.wantPort {
				t.Errorf("SetupSelfMeta() port = %v , want %v",got.PrimaryPort() , tt.wantPort )

			}
			if !types.IsSamePeerID(got.ID,tt.wantID) {
				t.Errorf("SetupSelfMeta() ID = %v , want %v",got.ID,tt.wantID )

			}
			if got.Hidden != tt.wantHidden {
				t.Errorf("SetupSelfMeta() hidden = %v , want %v",got.Hidden , tt.wantHidden )

			}

		})
	}
}
