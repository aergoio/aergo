/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"github.com/aergoio/aergo/types"
)

type PeerRoleManager interface {
	UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID)

	// SelfRole returns role of this peer itself
	SelfRole() types.PeerRole
	// GetRole returns role of remote peer
	GetRole(pid types.PeerID) types.PeerRole
	// FilterBPNoticeReceiver selects target peers with the appropriate role and sends them a BlockProducedNotice
	FilterBPNoticeReceiver(block *types.Block, pm PeerManager, targetZone PeerZone) []RemotePeer

	// FilterNewBlockNoticeReceiver selects target peers with the appropriate role and sends them a NewBlockNotice
	FilterNewBlockNoticeReceiver(block *types.Block, pm PeerManager) []RemotePeer

	//ListBlockManagePeers(exclude types.PeerID) map[types.PeerID]bool
}

//go:generate mockgen -source=peerrole.go -package=p2pmock -destination=../p2pmock/mock_peerrole.go

type AttrModifier struct {
	ID   types.PeerID
	Role types.PeerRole
}

