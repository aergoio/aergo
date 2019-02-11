/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2putil"
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

	Hidden   bool // Hidden means that meta info of this peer will not be sent to other peers when getting peer list
	Outbound bool
}

func (m PeerMeta) String() string {
	return fmt.Sprintf("%s:%d/%s", m.IPAddress, m.Port, p2putil.ShortForm(m.ID))
}

// FromPeerAddress convert PeerAddress to PeerMeta
func FromPeerAddress(addr *types.PeerAddress) PeerMeta {
	meta := PeerMeta{IPAddress: addr.Address,
		Port: addr.Port, ID: peer.ID(addr.PeerID)}
	return meta
}

// ToPeerAddress convert PeerMeta to PeerAddress
func (m PeerMeta) ToPeerAddress() types.PeerAddress {
	addr := types.PeerAddress{Address: m.IPAddress, Port: m.Port,
		PeerID: []byte(m.ID)}
	return addr
}
