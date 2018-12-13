/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/libp2p/go-libp2p-protocol"
	"net"
	"strconv"

	"github.com/aergoio/aergo-lib/log"

	cfg "github.com/aergoio/aergo/config"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// NetworkTransport do manager network connection
// TODO need refactoring. it has other role, pk management of self peer
type NetworkTransport interface {
	host.Host
	Start() error
	Stop() error

	PrivateKey() crypto.PrivKey
	PublicKey() crypto.PubKey
	SelfMeta() PeerMeta
	SelfNodeID() peer.ID

	GetAddressesOfPeer(peerID peer.ID) []string

	GetOrCreateStream(meta PeerMeta, protocolID protocol.ID) (inet.Stream, error)

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
	bindAddress net.IP
	bindPort    int
	selfMeta    PeerMeta

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
func (sl *networkTransport) SelfMeta() PeerMeta {
	return sl.selfMeta
}
func (sl *networkTransport) SelfNodeID() peer.ID {
	return sl.selfMeta.ID
}

func NewNetworkTransport(conf *cfg.P2PConfig, logger *log.Logger) *networkTransport {
	nt := &networkTransport{
		conf:           conf,
		logger:         logger,
	}
	nt.init()

	return nt
}

func (sl *networkTransport) init() {
	// check Key and address
	priv := NodePrivKey()
	pub := NodePubKey()
	pid := NodeID()

	sl.privateKey = priv
	sl.publicKey = pub
	// init address and port
	// if not set, it look up ip addresses of machine and choose suitable one (but not so smart) and default port 7845
	peerAddr, peerPort := sl.getProtocolAddrs()
	sl.selfMeta.IPAddress = peerAddr.String()
	sl.selfMeta.Port = uint32(peerPort)
	sl.selfMeta.ID = pid

	// if bindAddress or bindPort is not set, it will be same as NetProtocolAddr or NetProtocolPort
	if len(sl.conf.NPBindAddr) > 0 {
		bindAddr := net.ParseIP(sl.conf.NPBindAddr)
		if bindAddr == nil {
			panic("invalid NPBindAddr " + sl.conf.NPBindAddr)
		}
		sl.bindAddress = bindAddr
	} else {
		sl.bindAddress = peerAddr
	}
	if sl.conf.NPBindPort > 0 {
		sl.bindPort = sl.conf.NPBindPort
	} else {
		sl.bindPort = peerPort
	}

	// set meta info
	// TODO more survey libp2p NAT configuration
}

func (sl *networkTransport) getProtocolAddrs() (protocolAddr net.IP, protocolPort int) {
	if len(sl.conf.NetProtocolAddr) != 0 {
		protocolAddr = net.ParseIP(sl.conf.NetProtocolAddr)
		if protocolAddr == nil {
			panic("invalid NetProtocolAddr " + sl.conf.NetProtocolAddr)
		}
		if protocolAddr.IsUnspecified() {
			panic("NetProtocolAddr should be a specified IP address, not 0.0.0.0")
		}
	} else {
		extIP, err := externalIP()
		if err != nil {
			panic("error while finding IP address: " + err.Error())
		}
		protocolAddr = extIP
	}
	protocolPort = sl.conf.NetProtocolPort
	if protocolPort <= 0 {
		panic("invalid NetProtocolPort " + strconv.Itoa(sl.conf.NetProtocolPort))
	}
	return
}

func (sl *networkTransport) Start() error {
	sl.startListener()
	return nil
}


// GetOrCreateStream try to connect and handshake to remote peer. it can be called after peermanager is inited.
// It return true if peer is added or return false if failed to add peer or more suitable connection already exists.
func (sl *networkTransport) GetOrCreateStream(meta PeerMeta, protocolID  protocol.ID) (inet.Stream, error) {
	addrString := fmt.Sprintf("/ip4/%s/tcp/%d", meta.IPAddress, meta.Port)
	var peerAddr, err = ma.NewMultiaddr(addrString)
	if err != nil {
		sl.logger.Warn().Err(err).Str("addr", addrString).Msg("invalid NPAddPeer address")
		return nil,fmt.Errorf("invalid IP address %s:%d",meta.IPAddress,meta.Port)
	}
	var peerID = meta.ID
	sl.Peerstore().AddAddr(peerID, peerAddr, meta.TTL())
	ctx := context.Background()
	s, err := sl.NewStream(ctx, meta.ID, protocolID)
	if err != nil {
		sl.logger.Info().Err(err).Str("addr", addrString).Str(LogPeerID, meta.ID.Pretty()).Str(LogProtoID, string(protocolID)).Msg("Error while get stream")
		return nil,err
	}
	return s, nil
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
	// FIXME: should also support ip6 later
	listen, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", sl.bindAddress, sl.bindPort))
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

	sl.logger.Info().Str("pid", sl.SelfNodeID().Pretty()).Str("addr[0]", listens[0].String()).
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