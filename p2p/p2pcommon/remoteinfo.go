package p2pcommon

import "net"

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

	Certificates []*AgentCertificateV1
}
