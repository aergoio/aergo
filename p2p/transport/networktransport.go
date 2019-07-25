/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package transport

import (
	"context"
	"fmt"
	network2 "github.com/aergoio/aergo/internal/network"
	"github.com/aergoio/aergo/types"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/network"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pkey"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"

	cfg "github.com/aergoio/aergo/config"
	"github.com/libp2p/go-libp2p"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

/**
 * networkTransport connect to and listen from other nodes.
 * It implements  Component interface
 */
type networkTransport struct {
	core.Host
	privateKey crypto.PrivKey
	publicKey  crypto.PubKey

	selfMeta    p2pcommon.PeerMeta
	bindAddress net.IP
	bindPort    uint32

	// hostInited is
	hostInited *sync.WaitGroup

	conf   *cfg.P2PConfig
	logger *log.Logger
}

var _ p2pcommon.NetworkTransport = (*networkTransport)(nil)

func (sl *networkTransport) SelfMeta() p2pcommon.PeerMeta {
	return sl.selfMeta
}

func NewNetworkTransport(conf *cfg.P2PConfig, logger *log.Logger) *networkTransport {
	nt := &networkTransport{
		conf:   conf,
		logger: logger,

		hostInited: &sync.WaitGroup{},
	}
	nt.initNT()

	return nt
}

func (sl *networkTransport) initNT() {
	// check Key and address
	priv := p2pkey.NodePrivKey()
	pub := p2pkey.NodePubKey()
	peerID := p2pkey.NodeID()
	sl.privateKey = priv
	sl.publicKey = pub

	// init address and port
	// if not set, it look up ip addresses of machine and choose suitable one (but not so smart) and default port 7845
	sl.initSelfMeta(peerID, !sl.conf.NPExposeSelf)
	sl.initServiceBindAddress()

	sl.hostInited.Add(1)

	// set meta info
	// TODO more survey libp2p NAT configuration
}

func (sl *networkTransport) initSelfMeta(peerID types.PeerID, noExpose bool) {
	protocolAddr := sl.conf.NetProtocolAddr
	var ipAddress net.IP
	var err error
	var protocolPort int
	if len(sl.conf.NetProtocolAddr) != 0 {
		ipAddress, err = network2.GetSingleIPAddress(protocolAddr)
		if err != nil {
			panic("Invalid protocol address " + protocolAddr + " : " + err.Error())
		}
		if ipAddress.IsUnspecified() {
			panic("NetProtocolAddr should be a specified IP address, not 0.0.0.0")
		}
	} else {
		extIP, err := p2putil.ExternalIP()
		if err != nil {
			panic("error while finding IP address: " + err.Error())
		}
		ipAddress = extIP
		protocolAddr = ipAddress.String()
	}
	protocolPort = sl.conf.NetProtocolPort
	if protocolPort <= 0 {
		panic("invalid NetProtocolPort " + strconv.Itoa(sl.conf.NetProtocolPort))
	}
	sl.selfMeta.IPAddress = protocolAddr
	sl.selfMeta.Port = uint32(protocolPort)
	sl.selfMeta.ID = peerID
	sl.selfMeta.Hidden = noExpose
	sl.selfMeta.Version = p2pkey.NodeVersion()

	// bind address and port will be overridden if configuration is specified
	sl.bindAddress = ipAddress
	sl.bindPort = sl.selfMeta.Port
}

func (sl *networkTransport) initServiceBindAddress() {
	bindAddr := sl.conf.NPBindAddr
	// if bindAddress or bindPort is not set, it will be same as NetProtocolAddr or NetProtocolPort
	if len(sl.conf.NPBindAddr) > 0 {
		bindIP, err := network2.GetSingleIPAddress(bindAddr)
		if err != nil {
			panic("invalid NPBindAddr " + sl.conf.NPBindAddr)
		}
		// check address connectivity
		sl.bindAddress = bindIP
	}
	if sl.conf.NPBindPort > 0 {
		sl.bindPort = uint32(sl.conf.NPBindPort)
	}

}

func (sl *networkTransport) Start() error {
	sl.logger.Debug().Msg("Starting network transport")
	sl.startListener()
	sl.hostInited.Done()
	return nil
}

func (sl *networkTransport) AddStreamHandler(pid core.ProtocolID, handler network.StreamHandler) {
	sl.hostInited.Wait()
	sl.SetStreamHandler(pid, handler)
}

// GetOrCreateStream try to connect and handshake to remote peer. it can be called after peermanager is inited.
// It return true if peer is added or return false if failed to add peer or more suitable connection already exists.
func (sl *networkTransport) GetOrCreateStreamWithTTL(meta p2pcommon.PeerMeta, ttl time.Duration, protocolIDs ...core.ProtocolID) (core.Stream, error) {
	var peerAddr, err = p2putil.PeerMetaToMultiAddr(meta)
	if err != nil {
		sl.logger.Warn().Err(err).Str("addr", meta.IPAddress).Msg("invalid NPAddPeer address")
		return nil, fmt.Errorf("invalid IP address %s:%d", meta.IPAddress, meta.Port)
	}
	var peerID = meta.ID
	sl.Peerstore().AddAddr(peerID, peerAddr, ttl)
	ctx := context.Background()
	s, err := sl.NewStream(ctx, meta.ID, protocolIDs...)
	if err != nil {
		sl.logger.Info().Err(err).Str("addr", meta.IPAddress).Str(p2putil.LogPeerID, p2putil.ShortForm(meta.ID)).Str("p2p_proto", p2putil.ProtocolIDsToString(protocolIDs)).Msg("Error while get stream")
		return nil, err
	}
	return s, nil
}

func (sl *networkTransport) GetOrCreateStream(meta p2pcommon.PeerMeta, protocolIDs ...core.ProtocolID) (core.Stream, error) {
	return sl.GetOrCreateStreamWithTTL(meta, getTTL(meta), protocolIDs...)
}

func (sl *networkTransport) FindPeer(peerID types.PeerID) bool {
	found := false
	for _, existingPeerID := range sl.Peerstore().Peers() {
		if existingPeerID == peerID {
			found = true
			break
		}
	}
	return found
}

// ClosePeerConnection find and Close Peer connection
// It return true if peer is found
func (sl *networkTransport) ClosePeerConnection(peerID types.PeerID) bool {
	// also disconnect connection
	for _, existingPeerID := range sl.Peerstore().Peers() {
		if existingPeerID == peerID {
			sl.Network().ClosePeer(peerID)
			return true
		}
	}
	return false
}

//
//func (sl *networkTransport) Peerstore() pstore.Peerstore {
//	return sl.Host.Peerstore()
//}

func (sl *networkTransport) startListener() {
	var err error
	listens := make([]ma.Multiaddr, 0, 2)
	listen, err := types.ToMultiAddr(sl.bindAddress, sl.bindPort)
	if err != nil {
		panic("Can't establish listening address: " + err.Error())
	}
	listens = append(listens, listen)

	peerStore := pstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewProtoBook(),pstoremem.NewPeerMetadata())

	newHost, err := libp2p.New(context.Background(), libp2p.Identity(sl.privateKey), libp2p.Peerstore(peerStore), libp2p.ListenAddrs(listens...))
	if err != nil {
		sl.logger.Fatal().Err(err).Str("addr", listen.String()).Msg("Couldn't listen from")
		panic(err.Error())
	}
	sl.Host = newHost
	sl.logger.Info().Str(p2putil.LogFullID, sl.ID().Pretty()).Str(p2putil.LogPeerID, p2putil.ShortForm(sl.ID())).Str("addr[0]", listens[0].String()).Msg("Set self node's pid, and listening for connections")
}

func (sl *networkTransport) Stop() error {
	return sl.Host.Close()
}

func (sl *networkTransport) GetAddressesOfPeer(peerID types.PeerID) []string {
	addrs := sl.Peerstore().Addrs(peerID)
	addrStrs := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrs[i] = addr.String()
	}
	return addrStrs
}

// TTL return node's ttl
func getTTL(m p2pcommon.PeerMeta) time.Duration {
	if m.Designated {
		return p2pcommon.DesignatedNodeTTL
	}
	return p2pcommon.DefaultNodeTTL
}
