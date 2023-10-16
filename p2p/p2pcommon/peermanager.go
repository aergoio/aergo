/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

//go:generate mockgen -source=peermanager.go -package=p2pmock -destination=../p2pmock/mock_peermanager.go
package p2pcommon

import (
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/types"
)

// PeerManager is internal service that provide peer management
type PeerManager interface {
	AddPeerEventListener(l PeerEventListener)

	Start() error
	Stop() error

	//NetworkTransport
	SelfMeta() PeerMeta
	SelfNodeID() types.PeerID

	// AddNewPeer connect to peer. It will reset reconnect schedule and try to connect immediately if this peer is in reconnect cooltime.
	AddNewPeer(meta PeerMeta)
	// Remove peer from peer list. Peer dispose relative resources and stop itself, and then call PeerManager.RemovePeer
	RemovePeer(peer RemotePeer)
	UpdatePeerRole(changes []RoleModifier)

	NotifyPeerAddressReceived([]PeerMeta)

	// GetPeer return registered(handshaked) remote peer object. It is thread safe
	GetPeer(ID types.PeerID) (RemotePeer, bool)
	// GetPeers return all registered(handshaked) remote peers. It is thread safe
	GetPeers() []RemotePeer
	GetProducerClassPeers() []RemotePeer
	GetWatcherClassPeers() []RemotePeer

	GetPeerAddresses(noHidden bool, showSelf bool) []*message.PeerInfo

	GetPeerBlockInfos() []types.PeerBlockInfo

	AddDesignatedPeer(meta PeerMeta)
	RemoveDesignatedPeer(peerID types.PeerID)
	ListDesignatedPeers() []PeerMeta
}
