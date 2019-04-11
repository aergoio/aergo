/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"errors"
	"github.com/libp2p/go-libp2p-peer"
	"time"
)

const (
	ManagerInterval = time.Minute

	PolarisQueryInterval   = time.Minute * 10
	PeerQueryInterval      = time.Hour
	PeerFirstInterval      = time.Second * 4
	MaxConcurrentHandshake = 5
)

var (
	ErrNoWaitings = errors.New("no waiting peer exists")
)

type PeerEventListener interface {
	OnPeerConnect(pid peer.ID)
	OnPeerDisconnect(peer RemotePeer)
}

// PeerFinder works for collecting peer candidate.
// It queries to Polaris or other connected peer efficiently.
// It determine if peer is
// NOTE that this object is not thread safe by itself.
type PeerFinder interface {
	PeerEventListener
	OnDiscoveredPeers(metas []PeerMeta)

	// Check if it need to discover more peers and send query request to polaris or other peers if needed.
	CheckAndFill()
}

type WaitingPeer struct {
	Meta      PeerMeta
	TrialCnt  int
	NextTrial time.Time

	LastResult error
}
