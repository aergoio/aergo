/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"strconv"

	"github.com/aergoio/aergo/v2/types"
	"github.com/multiformats/go-multiaddr"
)

// PeerMeta contains non changeable information of peer node during connected state
type PeerMeta struct {
	ID types.PeerID
	// AcceptedRole is the role that the remote peer claims: the local peer may not admit it, and only admits it when there is a proper proof, such as vote result in chain or AgentCertificate.
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
func NewMetaWith1Addr(id types.PeerID, addr string, port uint32, version string) PeerMeta {
	ma, err := types.ToMultiAddr(addr, port)
	if err != nil {
		return PeerMeta{}
	} else {
		return PeerMeta{ID: id, Addresses: []types.Multiaddr{ma}, Version: version}
	}
}

// FromStatusToMeta create peerMeta from Status message
func NewMetaFromStatus(status *types.Status) PeerMeta {
	meta := FromPeerAddressNew(status.Sender)
	// hidden field of remote peer should be got from Status struct
	meta.Hidden = status.NoExpose
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
	addrs := make([]string, len(m.Addresses))
	for i, a := range m.Addresses {
		addrs[i] = a.String()
	}
	pds := make([][]byte, len(m.ProducerIDs))
	for i, a := range m.ProducerIDs {
		pds[i] = []byte(a)
	}

	addr := types.PeerAddress{PeerID: []byte(m.ID), Address: m.PrimaryAddress(), Port: m.PrimaryPort(), Addresses: addrs, ProducerIDs: pds, Version: m.GetVersion(), Role: m.Role}
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

func (m PeerMeta) Equals(o PeerMeta) bool {
	if !types.IsSamePeerID(m.ID, o.ID) {
		return false
	}
	if m.Role != o.Role {
		return false
	}
	if len(m.ProducerIDs) != len(o.ProducerIDs) {
		return false
	}
	for i, id := range m.ProducerIDs {
		if !types.IsSamePeerID(id, o.ProducerIDs[i]) {
			return false
		}
	}
	if len(m.Addresses) != len(o.Addresses) {
		return false
	}
	for i, ad := range m.Addresses {
		if !ad.Equal(o.Addresses[i]) {
			return false
		}
	}
	if m.Version != o.Version {
		return false
	}
	if m.Hidden != o.Hidden {
		return false
	}
	return true
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
	producerIds := make([]types.PeerID, len(addr.ProducerIDs))
	for i, id := range addr.ProducerIDs {
		producerIds[i] = types.PeerID(id)
	}
	meta := PeerMeta{ID: types.PeerID(addr.PeerID), Addresses: addrs, Role: role, Version: addr.Version, ProducerIDs: producerIds}
	return meta
}
