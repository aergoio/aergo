package types

type PeerState int32

// indicating status of remote peer
const (
	DISCONNECTED PeerState = iota
	// STARTING means connection is just estabished.
	STARTING
	// HANDSHAKING means that local host sent status message but not receive status message from remote
	HANDSHAKING
	// RUNNING means complete handshake (i.e. exchanged status message) and can communicate each other
	RUNNING
	STOPPED // totally finished peer.
)

//go:generate stringer -type=PeerState
