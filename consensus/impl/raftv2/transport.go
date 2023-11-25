/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package raftv2

import (
	"net/http"
	"time"

	"github.com/aergoio/aergo/v2/types"
	rtypes "github.com/aergoio/etcd/pkg/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/aergoio/etcd/rafthttp"
	"github.com/aergoio/etcd/snap"
)

type Transporter interface {
	// Start starts the given Transporter.
	// Start MUST be called before calling other functions in the interface.
	Start() error
	// Handler returns the HTTP handler of the transporter.
	// A transporter HTTP handler handles the HTTP requests
	// from remote peers.
	// The handler MUST be used to handle RaftPrefix(/raft)
	// endpoint.
	Handler() http.Handler
	// Send sends out the given messages to the remote peers.
	// Each message has a To field, which is an id that maps
	// to an existing peer in the transport.
	// If the id cannot be found in the transport, the message
	// will be ignored.
	Send(m []raftpb.Message)
	// SendSnapshot sends out the given snapshot message to a remote peer.
	// The behavior of SendSnapshot is similar to Send.
	SendSnapshot(m snap.Message)

	// AddPeer adds a peer with given peer urls into the transport.
	// It is the caller's responsibility to ensure the urls are all valid,
	// or it panics.
	// Peer urls are used to connect to the remote peer.
	AddPeer(id rtypes.ID, peerID types.PeerID, urls []string)

	// RemovePeer removes the peer with given id.
	RemovePeer(id rtypes.ID)
	// RemoveAllPeers removes all the existing peers in the transport.
	RemoveAllPeers()
	// UpdatePeer updates the peer urls of the peer with the given id.
	// It is the caller's responsibility to ensure the urls are all valid,
	// or it panics.
	UpdatePeer(id rtypes.ID, urls []string)
	// ActiveSince returns the time that the connection with the peer
	// of the given id becomes active.
	// If the connection is active since peer was added, it returns the adding time.
	// If the connection is currently inactive, it returns zero time.
	ActiveSince(id rtypes.ID) time.Time
	// ActivePeers returns the number of active peers.
	ActivePeers() int
	// Stop closes the connections and stops the transporter.
	Stop()
}

type HttpTransportWrapper struct {
	rafthttp.Transport
}

func (t *HttpTransportWrapper) AddPeer(id rtypes.ID, peerID types.PeerID, urls []string) {
	t.Transport.AddPeer(id, urls)
}
