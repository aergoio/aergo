/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
	"net"
	"strconv"
	"strings"
)

var InvalidArgument = fmt.Errorf("invalid argument")

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

// PeerMetaToMultiAddr make libp2p compatible Multiaddr object from peermeta
func PeerMetaToMultiAddr(m PeerMeta) (multiaddr.Multiaddr, error) {
	ipAddr, err := p2putil.GetSingleIPAddress(m.IPAddress)
	if err != nil {
		return nil, err
	}
	return ToMultiAddr(ipAddr, m.Port)
}

func FromMultiAddr(targetAddr multiaddr.Multiaddr) (PeerMeta, error) {
	meta := PeerMeta{}
	splitted := strings.Split(targetAddr.String(), "/")
	if len(splitted) != 7 {
		return meta, fmt.Errorf("invalid NPAddPeer addr format %s", targetAddr.String())
	}
	addrType := splitted[1]
	if addrType != "ip4" && addrType != "ip6" {
		return meta, fmt.Errorf("invalid NPAddPeer addr type %s", addrType)
	}
	peerAddrString := splitted[2]
	peerPortString := splitted[4]
	peerPort, err := strconv.Atoi(peerPortString)
	if err != nil {
		return meta, fmt.Errorf("invalid Peer port %s", peerPortString)
	}
	peerIDString := splitted[6]
	peerID, err := peer.IDB58Decode(peerIDString)
	if err != nil {
		return meta, fmt.Errorf("invalid PeerID %s", peerIDString)
	}
	meta = PeerMeta{
		ID:        peerID,
		Port:      uint32(peerPort),
		IPAddress: peerAddrString,
	}
	return meta, nil
}

func ParseMultiAddrString(str string) (PeerMeta, error) {
	ma, err := ParseMultiaddrWithResolve(str)
	if err != nil {
		return PeerMeta{}, err
	}
	return FromMultiAddr(ma)
}

// ParseMultiaddrWithResolve parse string to multiaddr, additionally accept domain name with protocol /dns
// NOTE: this function is temporarilly use until go-multiaddr start to support dns.
func ParseMultiaddrWithResolve(str string) (multiaddr.Multiaddr, error) {
	ma, err := multiaddr.NewMultiaddr(str)
	if err != nil {
		// multiaddr is not support domain name yet. change domain name to ip address manually
		splitted := strings.Split(str, "/")
		if len(splitted) < 3 || !strings.HasPrefix(splitted[1], "dns") {
			return nil, err
		}
		domainName := splitted[2]
		ips, err := p2putil.ResolveHostDomain(domainName)
		if err != nil {
			return nil, fmt.Errorf("Could not get IPs: %v\n", err)
		}
		// use ipv4 as possible.
		ipSelected := false
		for _, ip := range ips {
			if ip.To4() != nil {
				splitted[1] = "ip4"
				splitted[2] = ip.To4().String()
				ipSelected = true
				break
			}
		}
		if !ipSelected {
			for _, ip := range ips {
				if ip.To16() != nil {
					splitted[1] = "ip6"
					splitted[2] = ip.To16().String()
					ipSelected = true
					break
				}
			}
		}
		if !ipSelected {
			return nil, err
		}
		rejoin := strings.Join(splitted, "/")
		return multiaddr.NewMultiaddr(rejoin)
	}
	return ma, nil
}

