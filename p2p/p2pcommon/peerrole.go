/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import "github.com/aergoio/aergo/types"

type PeerRole uint8

const (
	UnknownRole PeerRole = iota
	BlockProducer
	Watcher
	_
	RaftProducer  // node that is ready to produce a block (can be a leader or follower)
	RaftWatcher   // node that is not produce block.
)
//go:generate stringer -type=PeerRole

type PeerRoleManager interface {
	UpdateBP(toAdd []types.PeerID, toRemove []types.PeerID)
	GetRole(pid types.PeerID) PeerRole
	NotifyNewBlockMsg(mo MsgOrder, peers []RemotePeer) (skipped, sent int)
}

type AttrModifier struct {
	ID   types.PeerID
	Role PeerRole
}