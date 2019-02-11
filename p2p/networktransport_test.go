/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/stretchr/testify/mock"
)

// TODO split this test into two... one is to attempt make connection and the other is test peermanager if same peerid is given
// Ignoring test for now, for lack of abstraction on AergoPeer struct
func IgrenoreTestP2PServiceRunAddPeer(t *testing.T) {
	mockActor := new(MockActorService)
	dummyBlock := types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActor.On("CallRequest", mock.Anything, mock.Anything).Return(message.GetBlockRsp{Block: &dummyBlock}, nil)
	//mockMF := new(MockMoFactory)
	target := &networkTransport{conf: config.NewServerContext("", "").GetDefaultConfig().(*config.Config).P2P,
		logger: log.NewLogger("test.p2p")}

	target.Host = &mockHost{peerstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewPeerMetadata())}
	target.selfMeta.ID = peer.ID("gwegw")

	sampleAddr1 := p2pcommon.PeerMeta{ID: "ddd", IPAddress: "192.168.0.1", Port: 33888, Outbound: true}
	sampleAddr2 := p2pcommon.PeerMeta{ID: "fff", IPAddress: "192.168.0.2", Port: 33888, Outbound: true}
	target.GetOrCreateStream(sampleAddr1, aergoP2PSub)
	target.GetOrCreateStream(sampleAddr1, aergoP2PSub)
	time.Sleep(time.Second)
	if len(target.Peerstore().Peers()) != 1 {
		t.Errorf("Peer count : Expected %d, Actually %d", 1, len(target.Peerstore().Peers()))
	}
	target.GetOrCreateStream(sampleAddr2, aergoP2PSub)
	time.Sleep(time.Second * 1)
	if len(target.Peerstore().Peers()) != 2 {
		t.Errorf("Peer count : Expected %d, Actually %d", 2, len(target.Peerstore().Peers()))
	}
}

func Test_networkTransport_initSelfMeta(t *testing.T) {
	type args struct {
		peerID peer.ID
	}
	tests := []struct {
		name   string
		conf        *cfg.P2PConfig

		args   args

		wantSameAddr bool
		wantPort uint32
		wantID peer.ID

	}{
		{"TIP6", &cfg.P2PConfig{NetProtocolAddr:"fe80::dcbf:beff:fe87:e30a",NetProtocolPort:7845}, args{dummyPeerID},true ,7845, dummyPeerID},
		{"TIP4", &cfg.P2PConfig{NetProtocolAddr:"211.1.1.2",NetProtocolPort:7845}, args{dummyPeerID},			true, 7845, dummyPeerID},
		{"TDN", &cfg.P2PConfig{NetProtocolAddr:"www.aergo.io",NetProtocolPort:7845}, args{dummyPeerID},			true, 7845, dummyPeerID},
		{"TDefault", &cfg.P2PConfig{NetProtocolAddr:"",NetProtocolPort:7845}, args{dummyPeerID},			false, 7845, dummyPeerID},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := &networkTransport{
				conf:        tt.conf,
				logger:      logger,
			}

			sl.initSelfMeta(tt.args.peerID)

			if tt.wantSameAddr {
				assert.Equal(t, tt.conf.NetProtocolAddr, sl.selfMeta.IPAddress)
			} else {
				assert.NotEqual(t,  tt.conf.NetProtocolAddr, sl.selfMeta.IPAddress)
			}
			assert.Equal(t, tt.wantPort, sl.selfMeta.Port)
			assert.Equal(t, tt.wantID, sl.selfMeta.ID)

			assert.NotNil(t, sl.bindAddress)
			fmt.Println("ProtocolAddress: ", sl.selfMeta.IPAddress)
			fmt.Println("bindAddress:     ", sl.bindAddress.String())
		})
	}
}
