/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package transport

import (
	"encoding/hex"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pkey"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	sampleKeyFile = "../../test/sample.key"
)

func init() {
	//sampleID := "16Uiu2HAmP2iRDpPumUbKhNnEngoxAUQWBmCyn7FaYUrkaDAMXJPJ"
	baseCfg := &config.BaseConfig{AuthDir: "test"}
	p2pCfg := &config.P2PConfig{NPKey: sampleKeyFile}
	p2pkey.InitNodeInfo(baseCfg, p2pCfg, "0.0.1-test", log.NewLogger("test.transport"))
}

// TODO split this test into two... one is to attempt make connection and the other is test peermanager if same peerid is given
// Ignoring test for now, for lack of abstraction on AergoPeer struct
func IgrenoreTestP2PServiceRunAddPeer(t *testing.T) {
	var sampleBlockHash, _ = hex.DecodeString("4f461d85e869ade8a0544f8313987c33a9c06534e50c4ad941498299579bd7ac")
	var sampleBlockHeight uint64 = 100215

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockActor := p2pmock.NewMockActorService(ctrl)
	dummyBlock := types.Block{Hash: sampleBlockHash, Header: &types.BlockHeader{BlockNo: sampleBlockHeight}}
	mockActor.EXPECT().CallRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(message.GetBlockRsp{Block: &dummyBlock}, nil)
	//mockMF := new(MockMoFactory)
	target := &networkTransport{conf: config.NewServerContext("", "").GetDefaultConfig().(*config.Config).P2P,
		logger: log.NewLogger("test.p2p")}

	//target.Host = &mockHost{peerstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewPeerMetadata())}
	target.Host = p2pmock.NewMockHost(ctrl)
	target.selfMeta.ID = types.PeerID("gwegw")

	sampleAddr1 := p2pcommon.PeerMeta{ID: "ddd", IPAddress: "192.168.0.1", Port: 33888, Outbound: true}
	sampleAddr2 := p2pcommon.PeerMeta{ID: "fff", IPAddress: "192.168.0.2", Port: 33888, Outbound: true}
	target.GetOrCreateStream(sampleAddr1, p2pcommon.LegacyP2PSubAddr)
	target.GetOrCreateStream(sampleAddr1, p2pcommon.LegacyP2PSubAddr)
	time.Sleep(time.Second)
	if len(target.Peerstore().Peers()) != 1 {
		t.Errorf("Peer count : Expected %d, Actually %d", 1, len(target.Peerstore().Peers()))
	}
	target.GetOrCreateStream(sampleAddr2, p2pcommon.LegacyP2PSubAddr)
	time.Sleep(time.Second * 1)
	if len(target.Peerstore().Peers()) != 2 {
		t.Errorf("Peer count : Expected %d, Actually %d", 2, len(target.Peerstore().Peers()))
	}
}

func Test_networkTransport_initSelfMeta(t *testing.T) {
	logger := log.NewLogger("test.transport")
	samplePeerID, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")

	type args struct {
		peerID   types.PeerID
		noExpose bool
	}
	tests := []struct {
		name string
		conf *cfg.P2PConfig

		args args

		wantSameAddr bool
		wantPort     uint32
		wantID       types.PeerID
		wantHidden   bool
	}{
		{"TIP6", &cfg.P2PConfig{NetProtocolAddr: "fe80::dcbf:beff:fe87:e30a", NetProtocolPort: 7845}, args{samplePeerID, false}, true, 7845, samplePeerID, false},
		{"TIP4", &cfg.P2PConfig{NetProtocolAddr: "211.1.1.2", NetProtocolPort: 7845}, args{samplePeerID, false}, true, 7845, samplePeerID, false},
		{"TDN", &cfg.P2PConfig{NetProtocolAddr: "www.aergo.io", NetProtocolPort: 7845}, args{samplePeerID, false}, true, 7845, samplePeerID, false},
		{"TDefault", &cfg.P2PConfig{NetProtocolAddr: "", NetProtocolPort: 7845}, args{samplePeerID, false}, false, 7845, samplePeerID, false},
		{"THidden", &cfg.P2PConfig{NetProtocolAddr: "211.1.1.2", NetProtocolPort: 7845}, args{samplePeerID, true}, true, 7845, samplePeerID, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := &networkTransport{
				conf:   tt.conf,
				logger: logger,
			}

			sl.initSelfMeta(tt.args.peerID, tt.args.noExpose)

			if tt.wantSameAddr {
				assert.Equal(t, tt.conf.NetProtocolAddr, sl.selfMeta.IPAddress)
			} else {
				assert.NotEqual(t, tt.conf.NetProtocolAddr, sl.selfMeta.IPAddress)
			}
			assert.Equal(t, tt.wantPort, sl.selfMeta.Port)
			assert.Equal(t, tt.wantID, sl.selfMeta.ID)
			assert.Equal(t, tt.wantHidden, sl.selfMeta.Hidden)

			assert.NotNil(t, sl.bindAddress)
			fmt.Println("ProtocolAddress: ", sl.selfMeta.IPAddress)
			fmt.Println("bindAddress:     ", sl.bindAddress.String())
		})
	}
}

func TestNewNetworkTransport(t *testing.T) {
	logger := log.NewLogger("test.transport")
	svrctx := config.NewServerContext("", "")
	sampleAddr := "211.1.2.3"

	tests := []struct {
		name string

		protocolAddr string
		bindAddr     string
		protocolPort int
		bindPort     int
		wantAddress  net.IP
		wantPort     uint32
	}{
		{"TDefault", "", "", -1, -1, nil, 7846},
		{"TAddrProto", sampleAddr, "", -1, -1, net.ParseIP(sampleAddr), 7846},
		{"TAddrBind", "", sampleAddr,-1, -1, net.ParseIP(sampleAddr), 7846},
		{"TAddrDiffer", "123.45.67.89", sampleAddr,-1, -1, net.ParseIP(sampleAddr), 7846},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := svrctx.GetDefaultP2PConfig()
			if len(tt.protocolAddr) > 0 {
				conf.NetProtocolAddr = tt.protocolAddr
			}
			if len(tt.bindAddr) > 0 {
				conf.NPBindAddr = tt.bindAddr
			}
			if tt.protocolPort > 0 {
				conf.NetProtocolPort = tt.protocolPort
			}
			if tt.bindPort > 0 {
				conf.NPBindPort = tt.bindPort
			}
			got := NewNetworkTransport(conf, logger)

			if got.privateKey == nil {
				t.Errorf("NewNetworkTransport() privkey is nil, want not")
			}
			if got.publicKey == nil {
				t.Errorf("NewNetworkTransport() pubkey is nil, want %v", p2pkey.NodePubKey())
			}
			if got.selfMeta.Version == "" {
				t.Errorf("NewNetworkTransport() = %v, want %v", got.selfMeta.Version, p2pkey.NodeVersion())
			}
			addr := got.bindAddress
			port := got .bindPort
			if tt.wantAddress == nil {
				if addr.IsLoopback() || addr.IsUnspecified() {
					t.Errorf("initServiceBindAddress() addr = %v, want valid addr", addr)
				}
			} else {
				if !addr.Equal(tt.wantAddress) {
					t.Errorf("initServiceBindAddress() addr = %v, want %v", addr, tt.wantAddress)
				}
			}
			if port != tt.wantPort {
				t.Errorf("initServiceBindAddress() port = %v, want %v", port, tt.wantPort)
			}

		})
	}
}

func Test_networkTransport_initServiceBindAddress(t *testing.T) {
	logger := log.NewLogger("test.transport")
	svrctx := config.NewServerContext("", "")
	initialPort := 7846

	tests := []struct {
		name string
		conf *cfg.P2PConfig

		wantAddress net.IP
		wantPort    int
	}{
		{"TEmpty", svrctx.GetDefaultP2PConfig(), nil, initialPort},
		{"TAnywhere", withAddr(svrctx.GetDefaultP2PConfig(), "0.0.0.0") , net.ParseIP("0.0.0.0"), initialPort},
		{"TCustAddr", withAddr(svrctx.GetDefaultP2PConfig(), "211.1.2.3") , net.ParseIP("211.1.2.3"), initialPort},
		{"TCustAddrPort", withPort(withAddr(svrctx.GetDefaultP2PConfig(), "211.1.2.3"),7777) , net.ParseIP("211.1.2.3"), 7777},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := &networkTransport{conf: tt.conf,
				logger: logger,
				bindPort: uint32(initialPort),
			}
			sl.initServiceBindAddress()

			addr := sl.bindAddress
			port := sl.bindPort
			// init result must always bind balid address
			if tt.wantAddress == nil {
				if addr.IsLoopback() || addr.IsUnspecified() {
					t.Errorf("initServiceBindAddress() addr = %v, want valid addr", addr)
				}
			} else {
				if !addr.Equal(tt.wantAddress) {
					t.Errorf("initServiceBindAddress() addr = %v, want %v", addr, tt.wantAddress)
				}
			}
			if int(port) != tt.wantPort {
				t.Errorf("initServiceBindAddress() port = %v, want %v", port, tt.wantPort)
			}
		})
	}
}

func withAddr(conf *cfg.P2PConfig, addr string) *cfg.P2PConfig {
	conf.NPBindAddr = addr
	return conf
}

func withPort(conf *cfg.P2PConfig, port int) *cfg.P2PConfig {
	conf.NPBindPort = port
	return conf
}
