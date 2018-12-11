/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/multiformats/go-multiaddr"
	"net"
	"strings"
	"time"

	"strconv"

	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

// PeerMeta contains non changeable information of peer node during connected state
// TODO: PeerMeta is almost same as PeerAddress, so TODO to unify them.
type PeerMeta struct {
	ID peer.ID
	// IPAddress is human readable form of ip address such as "192.168.0.1" or "2001:0db8:0a0b:12f0:33:1"
	IPAddress  string
	Port       uint32
	Designated bool // Designated means this peer is designated in config file and connect to in startup phase

	Hidden    bool // Hidden means that meta info of this peer will not be sent to other peers when getting peer list
	Outbound   bool
}

func (m PeerMeta) String() string {
	return m.ID.Pretty() + "/" + m.IPAddress + ":" + strconv.Itoa(int(m.Port))
}

// FromPeerAddress convert PeerAddress to PeerMeta
func FromPeerAddress(addr *types.PeerAddress) PeerMeta {
	meta := PeerMeta{IPAddress: net.IP(addr.Address).String(),
		Port: addr.Port, ID: peer.ID(addr.PeerID)}
	return meta
}

// ToPeerAddress convert PeerMeta to PeerAddress
func (m PeerMeta) ToPeerAddress() types.PeerAddress {
	addr := types.PeerAddress{Address: []byte(net.ParseIP(m.IPAddress)), Port: m.Port,
		PeerID: []byte(m.ID)}
	return addr
}

func FromMultiAddrString(str string)  (PeerMeta, error) {
	ma, err := ParseMultiaddrWithResolve(str)
	if err != nil {
		return PeerMeta{}, err
	}
	return FromMultiAddr(ma)
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
		ID:         peerID,
		Port:       uint32(peerPort),
		IPAddress:  peerAddrString,
	}
	return meta, nil
}

// TTL return node's ttl
func (m PeerMeta) TTL() time.Duration {
	if m.Designated {
		return DesignatedNodeTTL
	}
	return DefaultNodeTTL
}
