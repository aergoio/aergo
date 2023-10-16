package p2pcommon

import (
	"net"

	"github.com/aergoio/aergo/v2/types"
)

type PeerZone bool

const (
	ExternalZone PeerZone = false
	InternalZone PeerZone = true
)

func (z PeerZone) String() string {
	if z == InternalZone {
		return "Internal"
	} else {
		return "External"
	}
}

// RemoteConn is information of single peer connection
type RemoteConn struct {
	IP       net.IP
	Port     uint32
	Outbound bool
}

// RemoteInfo is information of connected remote peer
type RemoteInfo struct {
	Meta       PeerMeta
	Connection RemoteConn

	Designated bool // Designated means this peer is designated in config file and connect to in startup phase
	Hidden     bool // Hidden means that meta info of this peer will not be sent to other peers when getting peer list

	// AcceptedRole is role as which the local peer treat the remote peer
	AcceptedRole types.PeerRole
	Certificates []*AgentCertificateV1
	Zone         PeerZone
}
