package jsonrpc

import (
	"strconv"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

type InOutPeer struct {
	Role      string
	Address   InOutPeerAddress
	BestBlock InOutBlockIdx
	LastCheck time.Time
	State     string
	Hidden    bool
	Self      bool
	Version   string
}

func (p *InOutPeer) FromProto(msg *types.Peer) {
	p.Role = msg.AcceptedRole.String()
	if msg.GetAddress() != nil {
		p.Address.FromProto(msg.GetAddress())
	}
	if msg.GetBestblock() != nil {
		p.BestBlock.FromProto(msg.GetBestblock())
	}
	p.LastCheck = time.Unix(0, msg.GetLashCheck())
	p.State = types.PeerState(msg.State).String()
	p.Hidden = msg.Hidden
	p.Self = msg.Selfpeer
	if msg.Version != "" {
		p.Version = msg.Version
	} else {
		p.Version = "(old)"
	}
}

type InOutPeerAddress struct {
	Address string
	Port    string
	PeerId  string
}

func (pa *InOutPeerAddress) FromProto(msg *types.PeerAddress) {
	pa.Address = msg.GetAddress()
	pa.Port = strconv.Itoa(int(msg.GetPort()))
	pa.PeerId = base58.Encode(msg.GetPeerID())
}

type LongInOutPeer struct {
	InOutPeer
	ProducerIDs  []string
	Certificates []*InOutCert
}

func (out *LongInOutPeer) FromProto(p *types.Peer) {
	out.InOutPeer.FromProto(p)

	out.ProducerIDs = make([]string, len(p.Address.ProducerIDs))
	for i, pid := range p.Address.ProducerIDs {
		out.ProducerIDs[i] = base58.Encode(pid)
	}

	if p.Address.Role == types.PeerRole_Agent {
		out.Certificates = make([]*InOutCert, len(p.Certificates))
		for i, cert := range p.Certificates {
			out.Certificates[i] = &InOutCert{}
			out.Certificates[i].FromProto(cert)
		}
	}
}

type InOutCert struct {
	CertVersion uint32
	ProducerID  string
	CreateTime  time.Time
	ExpireTime  time.Time
	AgentID     string
	Addresses   []string
}

func (out *InOutCert) FromProto(msg *types.AgentCertificate) {
	out.CertVersion = msg.CertVersion
	out.ProducerID = base58.Encode(msg.BPID)
	out.AgentID = base58.Encode(msg.AgentID)
	out.CreateTime = time.Unix(0, msg.CreateTime)
	out.ExpireTime = time.Unix(0, msg.ExpireTime)
	out.Addresses = []string{}
	for _, ad := range msg.AgentAddress {
		out.Addresses = append(out.Addresses, string(ad))
	}
}
