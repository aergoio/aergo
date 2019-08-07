/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

type PeerAttribute struct {
	Chain ChainType
	RaftRole PeerRole
	Producer bool
}

type ChainType uint8
const (
	DPOS ChainType = iota
	RAFT
)

//go:generate stringer -type=ChainType
