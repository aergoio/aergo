/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"github.com/aergoio/aergo/types"
	"github.com/multiformats/go-multiaddr"
	"strconv"
)

// PeerMeta contains non changeable information of peer node during connected state
type PeerMeta struct {
	ID   types.PeerID
	Role types.PeerRole
	// ProducerIDs is a list of block producer IDs produced by this peer if the peer is BP, and if it is Agent, it is a list of block producer IDs that this peer acts as.
	ProducerIDs []types.PeerID
	// Address is advertised address to which other peer can connect.
	Addresses []types.Multiaddr
	// Version is build version of binary
	Version string
	Hidden  bool // Hidden means that meta info of this peer will not be sent to other peers when getting peer list
}

func (m *PeerMeta) GetVersion() string {
	if m.Version == "" {
		return ""
	} else {
		return m.Version
	}
}

// NewMetaWith1Addr make instance of PeerMeta with single address
func NewMetaWith1Addr(id types.PeerID, addr string, port uint32) PeerMeta {
	ma, err := types.ToMultiAddr(addr, port)
	if err != nil {
		return PeerMeta{}
	} else {
		return PeerMeta{ID:id, Addresses:[]types.Multiaddr{ma}}
	}
}
// FromStatusToMeta create peerMeta from Status message
func NewMetaFromStatus(status *types.Status, outbound bool) PeerMeta {
	meta := FromPeerAddressNew(status.Sender)
	//if len(meta.Version) == 0 {
	//	meta.Version = status.Sender.Version
	//}
	return meta
}

// FromPeerAddress convert PeerAddress to PeerMeta
func FromPeerAddress(addr *types.PeerAddress) PeerMeta {
	return FromPeerAddressNew(addr)
}

// ToPeerAddress convert PeerMeta to PeerAddress
func (m PeerMeta) ToPeerAddress() types.PeerAddress {
	addrs := make([]string,len(m.Addresses))
	for i, a := range m.Addresses {
		addrs[i] = a.String()
	}
	pds := make([][]byte,len(m.ProducerIDs))
	for i, a := range m.ProducerIDs {
		pds[i] = []byte(a)
	}

	addr := types.PeerAddress{PeerID: []byte(m.ID), Address:m.PrimaryAddress(), Port:m.PrimaryPort(), Addresses:addrs, ProducerIDs:pds, Version:m.GetVersion(), Role:m.Role}
	return addr
}

// PrimaryAddress is first advertised port of peer
func (m PeerMeta) PrimaryPort() uint32 {
	if len(m.Addresses) > 0 {
		portVal, _ := m.Addresses[0].ValueForProtocol(multiaddr.P_TCP)
		port, _ := strconv.Atoi(portVal)
		return uint32(port)
	} else {
		return 0
	}
}

// PrimaryAddress is first advertised address of peer
func (m PeerMeta) PrimaryAddress() string {
	if len(m.Addresses) > 0 {
		return types.AddressFromMultiAddr(m.Addresses[0])
	} else {
		return ""
	}

}

func FromPeerAddressNew(addr *types.PeerAddress) PeerMeta {
	addrs := make([]types.Multiaddr, 0, len(addr.Addresses))
	for _, a := range addr.Addresses {
		ma, err := multiaddr.NewMultiaddr(a)
		if err != nil {
			continue
		}
		addrs = append(addrs, ma)
	}
	role := types.PeerRole_LegacyVersion
	_, found := types.PeerRole_name[int32(addr.Role)]
	if found {
		role = types.PeerRole(int32(addr.Role))
	}
	producerIds := make([]types.PeerID,len(addr.ProducerIDs))
	for i, id := range addr.ProducerIDs {
		producerIds[i] = types.PeerID(id)
	}
	meta := PeerMeta{ID: types.PeerID(addr.PeerID), Addresses: addrs, Role: role, Version: addr.Version, ProducerIDs:producerIds}
	return meta
}
