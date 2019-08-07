/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"fmt"
	"github.com/aergoio/aergo/internal/network"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"net"
	"strings"
)

// PeerID is identier of peer. it is alias of libp2p.PeerID for now
type PeerID = core.PeerID

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

// ToMultiAddr make libp2p compatible Multiaddr object
func ToMultiAddr(ipAddr net.IP, port uint32) (multiaddr.Multiaddr, error) {
	var addrString string
	if ipAddr.To4() != nil {
		addrString = fmt.Sprintf("/ip4/%s/tcp/%d", ipAddr.String(), port)
	} else if ipAddr.To16() != nil {
		addrString = fmt.Sprintf("/ip6/%s/tcp/%d", ipAddr.String(), port)
	} else {
		return nil, InvalidArgument
	}
	peerAddr, err := multiaddr.NewMultiaddr(addrString)
	if err != nil {
		return nil, err
	}
	return peerAddr, nil
}

// ParseMultiaddrWithResolve parse string to multiaddr, additionally accept domain name with protocol /dns
// NOTE: this function is temporarily use until go-multiaddr start to support dns.
func ParseMultiaddrWithResolve(str string) (multiaddr.Multiaddr, error) {
	ma, err := multiaddr.NewMultiaddr(str)
	if err != nil {
		// multiaddr is not support domain name yet. change domain name to ip address manually
		split := strings.Split(str, "/")
		if len(split) < 3 || !strings.HasPrefix(split[1], "dns") {
			return nil, err
		}
		domainName := split[2]
		ips, err := network.ResolveHostDomain(domainName)
		if err != nil {
			return nil, fmt.Errorf("Could not get IPs: %v\n", err)
		}
		// use ipv4 as possible.
		ipSelected := false
		for _, ip := range ips {
			if ip.To4() != nil {
				split[1] = "ip4"
				split[2] = ip.To4().String()
				ipSelected = true
				break
			}
		}
		if !ipSelected {
			for _, ip := range ips {
				if ip.To16() != nil {
					split[1] = "ip6"
					split[2] = ip.To16().String()
					ipSelected = true
					break
				}
			}
		}
		if !ipSelected {
			return nil, err
		}
		rejoin := strings.Join(split, "/")
		return multiaddr.NewMultiaddr(rejoin)
	}
	return ma, nil
}
