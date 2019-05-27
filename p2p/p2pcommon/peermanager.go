/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

//go:generate mockgen -source=peermanager.go -package=p2pmock -destination=../p2pmock/mock_peermanager.go
package p2pcommon

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

// PeerManager is internal service that provide peer management
type PeerManager interface {
	Start() error
	Stop() error

	//NetworkTransport
	SelfMeta() PeerMeta
	SelfNodeID() types.PeerID

	AddNewPeer(peer PeerMeta)
	// Remove peer from peer list. Peer dispose relative resources and stop itself, and then call RemovePeer to peermanager
	RemovePeer(peer RemotePeer)

	NotifyPeerAddressReceived([]PeerMeta)

	// GetPeer return registered(handshaked) remote peer object
	GetPeer(ID types.PeerID) (RemotePeer, bool)
	GetPeers() []RemotePeer
	GetPeerAddresses(noHidden bool, showSelf bool) []*message.PeerInfo

	GetPeerBlockInfos() []types.PeerBlockInfo
}
