/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
		{"TDN", &config.P2PConfig{NetProtocolAddr: "www.aergo.io", NetProtocolPort: 7845, NPExposeSelf:true}, args{samplePeerID, false}, true, 7845, samplePeerID, false},
		{"TDefault", &config.P2PConfig{NetProtocolAddr: "", NetProtocolPort: 7845, NPExposeSelf:true}, args{samplePeerID, false}, false, 7845, samplePeerID, false},
		{"THidden", &config.P2PConfig{NetProtocolAddr: "211.1.1.2", NetProtocolPort: 7845, NPExposeSelf:false}, args{samplePeerID, true}, true, 7845, samplePeerID, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetupSelfMeta(tt.args.peerID, tt.conf)

			if tt.wantSameAddr {
				assert.Equal(t, tt.conf.NetProtocolAddr, got.IPAddress)
			} else {
				assert.NotEqual(t, tt.conf.NetProtocolAddr, got.IPAddress)
			}
			assert.Equal(t, tt.wantPort, got.Port)
			assert.Equal(t, tt.wantID, got.ID)
			assert.Equal(t, tt.wantHidden, got.Hidden)

			//assert.NotNil(t, sl.bindAddress)
			//fmt.Println("ProtocolAddress: ", sl.selfMeta.IPAddress)
			//fmt.Println("bindAddress:     ", sl.bindAddress.String())
		})
	}
}
