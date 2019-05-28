/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"errors"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-core/network"
	"time"
)

const (
	WaitingPeerManagerInterval = time.Minute

	PolarisQueryInterval   = time.Minute * 10
	PeerQueryInterval      = time.Hour
	PeerFirstInterval      = time.Second * 4

	MaxConcurrentHandshake = 5

)

var (
	ErrNoWaitings = errors.New("no waiting peer exists")
)

type PeerEventListener interface {
	OnPeerConnect(pid types.PeerID)
	OnPeerDisconnect(peer RemotePeer)
}

// PeerFinder works for collecting peer candidate.
// It queries to Polaris or other connected peer efficiently.
// It determine if peer is
// NOTE that this object is not thread safe by itself.
type PeerFinder interface {
	PeerEventListener

	// Check if it need to discover more peers and send query request to polaris or other peers if needed.
	CheckAndFill()
}

// WaitingPeerManager manage waiting peer pool and role to connect and handshaking of remote peer.
type WaitingPeerManager interface {
	PeerEventListener
	// OnDiscoveredPeers is called when response of discover query came from polaris or other peer.
	// It returns the count of previously unknown peers.
	OnDiscoveredPeers(metas []PeerMeta) int
	// OnWorkDone
	OnWorkDone(result ConnWorkResult)

	CheckAndConnect()

	OnInboundConn(s network.Stream)

	OnInboundConnLegacy(s network.Stream)
}

type WaitingPeer struct {
	Meta      PeerMeta
	TrialCnt  int
	NextTrial time.Time

	LastResult error
}

type ConnWorkResult struct {
	Inbound bool
	Seq     uint32
	// TargetPeer is nil if Inbound is true
	TargetPeer *WaitingPeer
	Meta       PeerMeta

	P2PVer uint32
	Result error
}
