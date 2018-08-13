package types

import "sync/atomic"

// PeerState indicated current state of peer, but
type PeerState int32

// indicating status of remote peer
const (
	// STARTING means connection is just estabished.
	STARTING PeerState = iota
	// HANDSHAKING means that local host sent status message but not receive status message from remote
	HANDSHAKING
	// RUNNING means complete handshake (i.e. exchanged status message) and can communicate each other
	RUNNING
	// DOWN means server can't communicate to remote peer. peer will be delete after TTL or
	DOWN
	// STOPPED is totally finished peer, and maybe local server is shutting down.
	STOPPED
)

// Get returns current state with concurrent manner
func (s *PeerState) Get() PeerState {
	return PeerState(atomic.LoadInt32((*int32)(s)))
}

// SetAndGet change state in atomic manner
func (s *PeerState) SetAndGet(ns PeerState) PeerState {
	return PeerState(atomic.SwapInt32((*int32)(s), int32(ns)))
}

func (s *PeerState) IncreaseAndGet() PeerState {
	return PeerState(atomic.AddInt32((*int32)(s), 1))
}

//go:generate stringer -type=PeerState
