/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"github.com/aergoio/aergo/internal/network"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/test"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"net"
	"strconv"
)

// PeerID is identier of peer. it is alias of libp2p.PeerID for now
type PeerID = core.PeerID

// NOTE use only in test
func RandomPeerID() PeerID {
	id, _ := test.RandPeerID()
	return id
}

func IDB58Decode(b58 string) (PeerID, error) {
	return peer.IDB58Decode(b58)
}
func IDB58Encode(pid PeerID) string {
	return peer.IDB58Encode(pid)
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

func PeerToMultiAddr(address string, port uint32, pid PeerID) (multiaddr.Multiaddr, error) {
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
func ToMultiAddr(address string, port uint32) (multiaddr.Multiaddr, error) {
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

func ParseMultiaddr(str string) (multiaddr.Multiaddr, error) {
	return multiaddr.NewMultiaddr(str)
}
// Deprecated: official multiaddr now support /dns (and also /dns4 and /dns6 )
// ParseMultiaddrWithResolve parse string to multiaddr, but doesn't check if domain name is valid.
func ParseMultiaddrWithResolve(str string) (multiaddr.Multiaddr, error) {
	return multiaddr.NewMultiaddr(str)
	//if err != nil {
	//	// multiaddr is not support domain name yet. change domain name to ip address manually
	//	split := strings.Split(str, "/")
	//	if len(split) < 3 || !strings.HasPrefix(split[1], "dns") {
	//		return nil, err
	//	}
	//	domainName := split[2]
	//	ips, err := network.ResolveHostDomain(domainName)
	//	if err != nil {
	//		return nil, fmt.Errorf("Could not get IPs: %v\n", err)
	//	}
	//	// use ipv4 as possible.
	//	ipSelected := false
	//	for _, ip := range ips {
	//		if ip.To4() != nil {
	//			split[1] = "ip4"
	//			split[2] = ip.To4().String()
	//			ipSelected = true
	//			break
	//		}
	//	}
	//	if !ipSelected {
	//		for _, ip := range ips {
	//			if ip.To16() != nil {
	//				split[1] = "ip6"
	//				split[2] = ip.To16().String()
	//				ipSelected = true
	//				break
	//			}
	//		}
	//	}
	//	if !ipSelected {
	//		return nil, err
	//	}
	//	rejoin := strings.Join(split, "/")
	//	return multiaddr.NewMultiaddr(rejoin)
	//}
	//return ma, nil
}

