/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"github.com/mr-tron/base58"
	"net"
	"strconv"
	"strings"

	"github.com/aergoio/aergo/v2/internal/network"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/test"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

// ProducerID is identifier of block producer. It is same format with the PeerID at version 1.x, but might be changed to other format later.
type ProducerID = core.PeerID

// PeerID is identifier of peer. It is alias of libp2p.PeerID for now
type PeerID = core.PeerID

// NOTE use only in test
func RandomPeerID() PeerID {
	id, _ := test.RandPeerID()
	return id
}

func IDB58Decode(b58 string) (PeerID, error) {
	return peer.Decode(b58)
}
func IDB58Encode(pid PeerID) string {
	return base58.Encode([]byte(pid))
}

func IDFromBytes(b []byte) (PeerID, error) {
	return peer.IDFromBytes(b)
}
func IDFromPublicKey(pubKey crypto.PubKey) (PeerID, error) {
	return peer.IDFromPublicKey(pubKey)
}
func IDFromPrivateKey(sk crypto.PrivKey) (PeerID, error) {
	return peer.IDFromPrivateKey(sk)
}
func IsSamePeerID(pid1, pid2 PeerID) bool {
	return string(pid1) == string(pid2)
}

// Stream is alias of libp2p.Stream
type Stream = core.Stream

// Multiaddr is alias of
type Multiaddr = multiaddr.Multiaddr

var InvalidArgument = errors.New("invalid argument")

const (
	protoDNS4 = "dns4"
	protoIP4  = "ip4"
	protoIP6  = "ip6"
	protoTCP  = "tcp"
	protoP2P  = "p2p"
)

func PeerToMultiAddr(address string, port uint32, pid PeerID) (Multiaddr, error) {
	idport, err := ToMultiAddr(address, port)
	if err != nil {
		return nil, err
	}
	mid, err := multiaddr.NewComponent(protoP2P, pid.Pretty())
	if err != nil {
		return nil, err
	}
	return multiaddr.Join(idport, mid), nil
}

// ToMultiAddr make libp2p compatible Multiaddr object
func ToMultiAddr(address string, port uint32) (Multiaddr, error) {
	var maddr, mport *multiaddr.Component
	var err error
	switch network.CheckAddressType(address) {
	case network.AddressTypeFQDN:
		maddr, err = multiaddr.NewComponent(protoDNS4, address)
	case network.AddressTypeIP:
		ip := net.ParseIP(address)
		if ip.To4() != nil {
			maddr, err = multiaddr.NewComponent(protoIP4, address)
		} else {
			maddr, err = multiaddr.NewComponent(protoIP6, address)
		}
	default:
		return nil, errors.New("invalid address")
	}
	if err != nil {
		return nil, err
	}
	mport, err = multiaddr.NewComponent(protoTCP, strconv.Itoa(int(port)))
	if err != nil {
		return nil, err
	}
	return multiaddr.Join(maddr, mport), nil
}

// ParseMultiaddr parse multiaddr formatted string to Multiaddr with slightly different manner; it automatically auusme that /dns is /dns4
func ParseMultiaddr(str string) (Multiaddr, error) {
	// replace /dns/ to /dns4/ for backward compatibility
	if strings.HasPrefix(str, "/dns/") {
		str = strings.Replace(str, "/dns/", "/dns4/", 1)
	}
	return multiaddr.NewMultiaddr(str)
}

const PortUndefined = uint32(0xffffffff)

var ErrInvalidIPAddress = errors.New("invalid ip address")
var ErrInvalidPort = errors.New("invalid port")

var addrProtos = []int{multiaddr.P_IP4, multiaddr.P_IP6, multiaddr.P_DNS4, multiaddr.P_DNS6}

// AddressFromMultiAddr returns address (ip4, ip6 or full qualified domain name)
func AddressFromMultiAddr(ma Multiaddr) string {
	for _, p := range addrProtos {
		val, err := ma.ValueForProtocol(p)
		if err == nil {
			return val
		}
	}
	return ""
}

func IsAddress(proto multiaddr.Protocol) bool {
	switch proto.Code {
	case multiaddr.P_IP4, multiaddr.P_IP6, multiaddr.P_DNS4, multiaddr.P_DNS6:
		return true
	default:
		return false
	}
}

func GetIPPortFromMultiaddr(ma Multiaddr) (net.IP, uint32, error) {
	ip := GetIPFromMultiaddr(ma)
	if ip == nil {
		return nil, PortUndefined, ErrInvalidIPAddress
	}
	port := GetPortFromMultiaddr(ma)
	if port == PortUndefined {
		return nil, PortUndefined, ErrInvalidPort
	}
	return ip, port, nil
}

func GetIPFromMultiaddr(ma Multiaddr) net.IP {
	val, err := ma.ValueForProtocol(multiaddr.P_IP4)
	if err != nil {
		val, err = ma.ValueForProtocol(multiaddr.P_IP6)
	}
	if err == nil {
		return net.ParseIP(val)
	} else {
		return nil
	}
}

func GetPortFromMultiaddr(ma Multiaddr) uint32 {
	val, err := ma.ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		return PortUndefined
	}
	port, err := strconv.Atoi(val)
	if err != nil {
		return PortUndefined
	}
	return uint32(port)
}
