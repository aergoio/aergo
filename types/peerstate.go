package types

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

//go:generate stringer -type=PeerState
