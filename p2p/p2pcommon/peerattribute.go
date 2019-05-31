/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

// PeerRole indicate tye characteristics of peer.
type PeerRole int32

// types for dpos chain
const (
	_ PeerRole = iota
	DPOSProducer
	_
	_
	DPOSWatcher
)

const (
	_ PeerRole = iota
	RaftLeader
	_
	RaftFollower
	RaftWatcher
)
