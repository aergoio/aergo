/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import "github.com/aergoio/aergo/types"

type PeerRoleManager interface {
	UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID)

	// SelfRole returns role of this peer itself
	SelfRole() types.PeerRole
	// GetRole returns role of remote peer
	GetRole(pid types.PeerID) types.PeerRole
	// NotifyNewBlockMsg selects target peers with the appropriate role and sends them a NewBlockNotice
	NotifyNewBlockMsg(mo MsgOrder, peers []RemotePeer) (skipped, sent int)
}
//go:generate mockgen -source=peerrole.go -package=p2pmock -destination=../p2pmock/mock_peerrole.go

type AttrModifier struct {
	ID   types.PeerID
	Role types.PeerRole
}