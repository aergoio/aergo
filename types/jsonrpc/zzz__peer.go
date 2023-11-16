package jsonrpc

import (
	"strconv"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvPeer(msg *types.Peer) *InOutPeer {
	p := &InOutPeer{}
	p.Role = msg.AcceptedRole.String()
	if msg.GetAddress() != nil {
		p.Address = *ConvPeerAddress(msg.GetAddress())
	}
	if msg.GetBestblock() != nil {
		p.BestBlock = *ConvBlockIdx(msg.GetBestblock())
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
	return p
}

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

func ConvPeerAddress(msg *types.PeerAddress) *InOutPeerAddress {
	pa := &InOutPeerAddress{}
	pa.Address = msg.GetAddress()
	pa.Port = strconv.Itoa(int(msg.GetPort()))
	pa.PeerId = base58.Encode(msg.GetPeerID())
	return pa
}

type InOutPeerAddress struct {
	Address string
	Port    string
	PeerId  string
}

func ConvLongPeer(msg *types.Peer) *LongInOutPeer {
	p := &LongInOutPeer{}
	p.InOutPeer = *ConvPeer(msg)

	p.ProducerIDs = make([]string, len(msg.Address.ProducerIDs))
	for i, pid := range msg.Address.ProducerIDs {
		p.ProducerIDs[i] = base58.Encode(pid)
	}

	if msg.Address.Role == types.PeerRole_Agent {
		p.Certificates = make([]*InOutCert, len(msg.Certificates))
		for i, cert := range msg.Certificates {
			p.Certificates[i] = &InOutCert{}
			p.Certificates[i] = ConvCert(cert)
		}
	}
	return p
}

type LongInOutPeer struct {
	InOutPeer
	ProducerIDs  []string
	Certificates []*InOutCert
}

func ConvCert(msg *types.AgentCertificate) *InOutCert {
	c := &InOutCert{}
	c.CertVersion = msg.CertVersion
	c.ProducerID = base58.Encode(msg.BPID)
	c.AgentID = base58.Encode(msg.AgentID)
	c.CreateTime = time.Unix(0, msg.CreateTime)
	c.ExpireTime = time.Unix(0, msg.ExpireTime)
	c.Addresses = []string{}
	for _, ad := range msg.AgentAddress {
		c.Addresses = append(c.Addresses, string(ad))
	}
	return c
}

type InOutCert struct {
	CertVersion uint32
	ProducerID  string
	CreateTime  time.Time
	ExpireTime  time.Time
	AgentID     string
	Addresses   []string
}
