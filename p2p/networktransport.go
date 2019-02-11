/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"context"
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/libp2p/go-libp2p-protocol"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"

	cfg "github.com/aergoio/aergo/config"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// NTContainer can provide NetworkTransport interface.
type NTContainer interface {
	GetNetworkTransport() NetworkTransport

	// ChainID return id of current chain.
	ChainID() *types.ChainID
}

// NetworkTransport do manager network connection
// TODO need refactoring. it has other role, pk management of self peer
type NetworkTransport interface {
	host.Host
	Start() error
	Stop() error

	PrivateKey() crypto.PrivKey
	PublicKey() crypto.PubKey
	SelfMeta() p2pcommon.PeerMeta
	SelfNodeID() peer.ID

	GetAddressesOfPeer(peerID peer.ID) []string

	// AddStreamHandler wrapper function which call host.SetStreamHandler after transport is initialized, this method is for preventing nil error.
	AddStreamHandler(pid protocol.ID, handler inet.StreamHandler)


	GetOrCreateStream(meta p2pcommon.PeerMeta, protocolID protocol.ID) (inet.Stream, error)
	GetOrCreateStreamWithTTL(meta p2pcommon.PeerMeta, protocolID protocol.ID, ttl time.Duration) (inet.Stream, error)

	FindPeer(peerID peer.ID) bool
	ClosePeerConnection(peerID peer.ID) bool
}

/**
 * networkTransport connect to and listen from other nodes.
 * It implements  Component interface
 */
type networkTransport struct {
	host.Host
	privateKey  crypto.PrivKey
	publicKey   crypto.PubKey

	selfMeta    p2pcommon.PeerMeta
	bindAddress net.IP
	bindPort    uint32


	// hostInited is
	hostInited *sync.WaitGroup

	conf        *cfg.P2PConfig
	logger      *log.Logger
}

var _ NetworkTransport = (*networkTransport)(nil)

func (sl *networkTransport) PrivateKey() crypto.PrivKey {
	return sl.privateKey
}
func (sl *networkTransport) PublicKey() crypto.PubKey {
	return sl.publicKey
}
func (sl *networkTransport) SelfMeta() p2pcommon.PeerMeta {
	return sl.selfMeta
}
func (sl *networkTransport) SelfNodeID() peer.ID {
	return sl.selfMeta.ID
}

func NewNetworkTransport(conf *cfg.P2PConfig, logger *log.Logger) *networkTransport {
	nt := &networkTransport{
		conf:           conf,
		logger:         logger,

		hostInited: &sync.WaitGroup{},
	}
	nt.initNT()

	return nt
}

func (sl *networkTransport) initNT() {
	// check Key and address
	priv := NodePrivKey()
	pub := NodePubKey()
	peerID := NodeID()
	sl.privateKey = priv
	sl.publicKey = pub

	// init address and port
	// if not set, it look up ip addresses of machine and choose suitable one (but not so smart) and default port 7845
	sl.initSelfMeta(peerID)
	sl.initServiceBindAddress()

	sl.hostInited.Add(1)

	// set meta info
	// TODO more survey libp2p NAT configuration
}

func (sl *networkTransport) initSelfMeta(peerID peer.ID) {
	protocolAddr := sl.conf.NetProtocolAddr
	var ipAddress net.IP
	var err error
	var protocolPort int
	if len(sl.conf.NetProtocolAddr) != 0 {
		ipAddress, err = p2putil.GetSingleIPAddress(protocolAddr)
		if err != nil {
			panic("Invalid protocol address "+protocolAddr+" : "+err.Error())
		}
		if ipAddress.IsUnspecified() {
			panic("NetProtocolAddr should be a specified IP address, not 0.0.0.0")
		}
	} else {
		extIP, err := externalIP()
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

	// bind address and port will be overriden if configuration is specified
	sl.bindAddress = ipAddress
	sl.bindPort = sl.selfMeta.Port
}

func (sl *networkTransport) initServiceBindAddress() {
	bindAddr := sl.conf.NPBindAddr
	// if bindAddress or bindPort is not set, it will be same as NetProtocolAddr or NetProtocolPort
	if len(sl.conf.NPBindAddr) > 0 {
		bindIP, err := p2putil.GetSingleIPAddress(bindAddr)
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

func (sl *networkTransport) AddStreamHandler(pid protocol.ID, handler inet.StreamHandler) {
	sl.hostInited.Wait()
	sl.SetStreamHandler(pid, handler)
}

// GetOrCreateStream try to connect and handshake to remote peer. it can be called after peermanager is inited.
// It return true if peer is added or return false if failed to add peer or more suitable connection already exists.
func (sl *networkTransport) GetOrCreateStreamWithTTL(meta p2pcommon.PeerMeta, protocolID  protocol.ID, ttl time.Duration) (inet.Stream, error) {
	var peerAddr, err = PeerMetaToMultiAddr(meta)
	if err != nil {
		sl.logger.Warn().Err(err).Str("addr", meta.IPAddress).Msg("invalid NPAddPeer address")
		return nil,fmt.Errorf("invalid IP address %s:%d",meta.IPAddress,meta.Port)
	}
	var peerID = meta.ID
	sl.Peerstore().AddAddr(peerID, peerAddr,ttl)
	ctx := context.Background()
	s, err := sl.NewStream(ctx, meta.ID, protocolID)
	if err != nil {
		sl.logger.Info().Err(err).Str("addr", meta.IPAddress).Str(LogPeerID, p2putil.ShortForm(meta.ID)).Str(LogProtoID, string(protocolID)).Msg("Error while get stream")
		return nil,err
	}
	return s, nil
}

func (sl *networkTransport) GetOrCreateStream(meta p2pcommon.PeerMeta, protocolID  protocol.ID) (inet.Stream, error) {
	return sl.GetOrCreateStreamWithTTL(meta, protocolID, getTTL(meta))
}

func (sl *networkTransport) FindPeer(peerID peer.ID) bool {
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
func (sl *networkTransport) ClosePeerConnection(peerID peer.ID) bool {
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
	listen, err := ToMultiAddr(sl.bindAddress, sl.bindPort)
	if err != nil {
		panic("Can't estabilish listening address: " + err.Error())
	}
	listens = append(listens, listen)

	peerStore := pstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewPeerMetadata())

	newHost, err := libp2p.New(context.Background(), libp2p.Identity(sl.privateKey), libp2p.Peerstore(peerStore), libp2p.ListenAddrs(listens...))
	if err != nil {
		sl.logger.Fatal().Err(err).Str("addr", listen.String()).Msg("Couldn't listen from")
		panic(err.Error())
	}

	sl.logger.Info().Str(LogFullID, sl.SelfNodeID().Pretty()).Str(LogPeerID, p2putil.ShortForm(sl.SelfNodeID())).Str("addr[0]", listens[0].String()).
		Msg("Set self node's pid, and listening for connections")
	sl.Host = newHost
}


func (sl *networkTransport) Stop() error {
	return sl.Host.Close()
}

func (sl *networkTransport) GetAddressesOfPeer(peerID peer.ID) []string {
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
		return DesignatedNodeTTL
	}
	return DefaultNodeTTL
}
