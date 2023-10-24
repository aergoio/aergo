/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package transport

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	cfg "github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

const (
	sampleKeyFile = "../test/sample/sample.key"
)

func init() {
	//sampleID := "16Uiu2HAmP2iRDpPumUbKhNnEngoxAUQWBmCyn7FaYUrkaDAMXJPJ"
	baseCfg := &config.BaseConfig{AuthDir: "test"}
	p2pCfg := &config.P2PConfig{NPKey: sampleKeyFile}
	p2pkey.InitNodeInfo(baseCfg, p2pCfg, "0.0.1-test", log.NewLogger("test.transport"))
}

// TODO split this test into two... one is to attempt make connection and the other is test peermanager if same peerid is given
// Ignoring test for now, for lack of abstraction on AergoPeer struct
func IgnoredTestP2PServiceRunAddPeer(t *testing.T) {
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

	sampleAddr1 := p2pcommon.NewMetaWith1Addr(types.RandomPeerID(), "192.168.0.1", 33888, "v2.0.0")
	sampleAddr2 := p2pcommon.NewMetaWith1Addr(types.RandomPeerID(), "192.168.0.2", 33888, "v2.0.0")
	target.GetOrCreateStream(sampleAddr1, p2pcommon.P2PSubAddr)
	target.GetOrCreateStream(sampleAddr1, p2pcommon.P2PSubAddr)
	time.Sleep(time.Second)
	if len(target.Peerstore().Peers()) != 1 {
		t.Errorf("Peer count : Expected %d, Actually %d", 1, len(target.Peerstore().Peers()))
	}
	target.GetOrCreateStream(sampleAddr2, p2pcommon.P2PSubAddr)
	time.Sleep(time.Second * 1)
	if len(target.Peerstore().Peers()) != 2 {
		t.Errorf("Peer count : Expected %d, Actually %d", 2, len(target.Peerstore().Peers()))
	}
}

func TestNewNetworkTransport(t *testing.T) {
	logger := log.NewLogger("test.transport")
	svrctx := config.NewServerContext("", "")
	localIP := "192.168.11.3"
	sampleIP := "211.1.2.3"
	acceptAll := "0.0.0.0"
	tests := []struct {
		name string

		protocolAddr string
		bindAddr     string
		protocolPort int
		bindPort     int
		wantAddress  string
		wantPort     uint32
	}{
		{"TDefault", "", "", -1, -1, acceptAll, 7846},
		{"TAddrProto", sampleIP, "", -1, -1, acceptAll, 7846},
		{"TAddrBind", "", sampleIP, -1, -1, sampleIP, 7846},
		{"TAddrSame", sampleIP, sampleIP, -1, -1, sampleIP, 7846},

		{"TAddrDiffer", "123.45.67.89", sampleIP, -1, -1, sampleIP, 7846},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			port := uint32(7846)
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
			sampleMeta := p2pcommon.NewMetaWith1Addr(p2pkey.NodeID(), localIP, port, "v2.0.0")

			mockIS := p2pmock.NewMockInternalService(ctrl)
			mockIS.EXPECT().SelfMeta().Return(sampleMeta)
			got := NewNetworkTransport(conf, logger, mockIS)

			if got.privateKey == nil {
				t.Errorf("NewNetworkTransport() privkey is nil, want not")
			}

			actualAddr := got.bindAddress
			actualPort := got.bindPort
			if actualAddr != tt.wantAddress {
				t.Errorf("initServiceBindAddress() addr = %v, want %v", actualAddr, tt.wantAddress)
			}
			if actualPort != tt.wantPort {
				t.Errorf("initServiceBindAddress() port = %v, want %v", actualPort, tt.wantPort)
			}

		})
	}
}

func Test_networkTransport_initServiceBindAddress(t *testing.T) {
	logger := log.NewLogger("test.transport")
	svrctx := config.NewServerContext("", "")
	initialPort := uint32(7846)
	acceptAll := "0.0.0.0"
	sampleDomain := "unittest.aergo.io"
	ipMeta := p2pcommon.NewMetaWith1Addr(types.RandomPeerID(), acceptAll, initialPort, "v2.0.0")
	dnMeta := p2pcommon.NewMetaWith1Addr(types.RandomPeerID(), sampleDomain, initialPort, "v2.0.0")
	tests := []struct {
		name string
		meta p2pcommon.PeerMeta
		conf *cfg.P2PConfig

		wantAddress string
		wantPort    uint32
	}{
		{"TEmpty", ipMeta, svrctx.GetDefaultP2PConfig(), acceptAll, initialPort},
		{"TAnywhere", ipMeta, withAddr(svrctx.GetDefaultP2PConfig(), "0.0.0.0"), "0.0.0.0", initialPort},
		{"TAnywhereDN", dnMeta, withAddr(svrctx.GetDefaultP2PConfig(), "0.0.0.0"), "0.0.0.0", initialPort},
		{"TLoopbackDN", dnMeta, withAddr(svrctx.GetDefaultP2PConfig(), "127.0.0.1"), "127.0.0.1", initialPort},
		{"TCustAddr", ipMeta, withAddr(svrctx.GetDefaultP2PConfig(), "211.1.2.3"), "211.1.2.3", initialPort},
		{"TCustAddrPort", ipMeta, withPort(withAddr(svrctx.GetDefaultP2PConfig(), "211.1.2.3"), 7777), "211.1.2.3", 7777},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := &networkTransport{conf: tt.conf,
				logger:   logger,
				bindPort: uint32(initialPort),
				selfMeta: tt.meta,
			}
			sl.initServiceBindAddress()

			addr := sl.bindAddress
			port := sl.bindPort
			// init result must always bind valid address
			if addr != tt.wantAddress {
				t.Errorf("initServiceBindAddress() addr = %v, want %v", addr, tt.wantAddress)
			}
			if port != tt.wantPort {
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
