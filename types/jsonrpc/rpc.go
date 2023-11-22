package jsonrpc

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

func ConvPeerList(msg *types.PeerList) *InOutPeerList {
	p := &InOutPeerList{}
	p.Peers = make([]*InOutPeer, len(msg.Peers))
	for i, peer := range msg.Peers {
		p.Peers[i] = ConvPeer(peer)
	}
	return p
}

type InOutPeerList struct {
	Peers []*InOutPeer `json:"peers"`
}

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
	p.SelfPeer = msg.Selfpeer
	if msg.Version != "" {
		p.Version = msg.Version
	} else {
		p.Version = "(old)"
	}
	return p
}

type InOutPeer struct {
	Role      string           `json:"peerrole,omitempty"`
	Address   InOutPeerAddress `json:"address,omitempty"`
	BestBlock InOutBlockIdx    `json:"bestblock,omitempty"`
	LastCheck time.Time        `json:"lastCheck,omitempty"`
	State     string           `json:"state,omitempty"`
	Hidden    bool             `json:"hidden,omitempty"`
	SelfPeer  bool             `json:"selfpeer,omitempty"`
	Version   string           `json:"version,omitempty"`
}

func ConvPeerAddress(msg *types.PeerAddress) *InOutPeerAddress {
	return &InOutPeerAddress{
		Address: msg.Address,
		Port:    strconv.Itoa(int(msg.Port)),
		PeerID:  base58.Encode(msg.PeerID),
	}
}

type InOutPeerAddress struct {
	Address string `json:"address,omitempty"`
	Port    string `json:"port,omitempty"`
	PeerID  string `json:"peerID,omitempty"`
}

func ConvShortPeerList(msg *types.PeerList) *InOutShortPeerList {
	p := &InOutShortPeerList{}
	p.Peers = make([]string, len(msg.Peers))
	for i, peer := range msg.Peers {
		pa := peer.Address
		p.Peers[i] = fmt.Sprintf("%s;%s/%d;%s;%d", p2putil.ShortForm(types.PeerID(pa.PeerID)), pa.Address, pa.Port, peer.AcceptedRole.String(), peer.Bestblock.BlockNo)
	}
	return p
}

type InOutShortPeerList struct {
	Peers []string `json:"peers,omitempty"`
}

func ConvLongPeerList(msg *types.PeerList) *InOutLongPeerList {
	p := &InOutLongPeerList{}
	p.Peers = make([]*InOutLongPeer, len(msg.Peers))
	for i, peer := range msg.Peers {
		p.Peers[i] = ConvLongPeer(peer)
	}
	return p
}

type InOutLongPeerList struct {
	Peers []*InOutLongPeer `json:"peers,omitempty"`
}

func ConvLongPeer(msg *types.Peer) *InOutLongPeer {
	p := &InOutLongPeer{}
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

type InOutLongPeer struct {
	InOutPeer    `json:",inline"`
	ProducerIDs  []string     `json:"producerIDs,omitempty"`
	Certificates []*InOutCert `json:"certificates,omitempty"`
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
	CertVersion uint32    `json:"certVersion,omitempty"`
	ProducerID  string    `json:"producerID,omitempty"`
	CreateTime  time.Time `json:"createTime,omitempty"`
	ExpireTime  time.Time `json:"expireTime,omitempty"`
	AgentID     string    `json:"agentID,omitempty"`
	Addresses   []string  `json:"addresses,omitempty"`
}

func ConvMetrics(msg *types.Metrics) *InOutMetrics {
	m := &InOutMetrics{}
	m.Peers = make([]*InOutPeerMetric, len(msg.Peers))
	for i, peer := range msg.Peers {
		m.Peers[i] = ConvPeerMetric(peer)
	}
	return m
}

type InOutMetrics struct {
	Peers []*InOutPeerMetric `json:"peers,omitempty"`
}

func ConvPeerMetric(msg *types.PeerMetric) *InOutPeerMetric {
	return &InOutPeerMetric{
		PeerID: base58.Encode(msg.PeerID),
		SumIn:  msg.SumIn,
		AvrIn:  msg.AvrIn,
		SumOut: msg.SumOut,
		AvrOut: msg.AvrOut,
	}
}

type InOutPeerMetric struct {
	PeerID string `json:"peerID,omitempty"`
	SumIn  int64  `json:"sumIn,omitempty"`
	AvrIn  int64  `json:"avrIn,omitempty"`
	SumOut int64  `json:"sumOut,omitempty"`
	AvrOut int64  `json:"avrOut,omitempty"`
}

func ConvBLConfEntries(msg *types.BLConfEntries) *InOutBLConfEntries {
	return &InOutBLConfEntries{
		Enabled: msg.Enabled,
		Entries: msg.Entries,
	}
}

type InOutBLConfEntries struct {
	Enabled bool     `json:"enabled,omitempty"`
	Entries []string `json:"entries,omitempty"`
}
