package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"sync"
	"time"
)

var emptyIDArr []types.PeerID
var emptyCertArr []*p2pcommon.AgentCertificateV1

func newCertificateManager(self p2pcommon.PeerMeta) p2pcommon.CertificateManager {
	switch self.Role {
	case types.PeerRole_Producer:
		return &bpCertificateManager{self: self}
	case types.PeerRole_Agent:
		return &agentCertificateManager{self: self}
	case types.PeerRole_Watcher:
		return &watcherCertificateManager{self: self}
	default:
		return nil
	}
}

type bpCertificateManager struct {
	self p2pcommon.PeerMeta
	key  *btcec.PrivateKey
}

func (d *bpCertificateManager) CreateCertificate(remoteMeta p2pcommon.PeerMeta) (*p2pcommon.AgentCertificateV1, error) {
	addrs := make([]string, len(remoteMeta.Addresses))
	for i, ad := range remoteMeta.Addresses {
		addrs[i] = types.AddressFromMultiAddr(ad)
	}
	return p2putil.NewAgentCertV1(d.self.ID, remoteMeta.ID, d.key, addrs, time.Hour*24)
}

func (d *bpCertificateManager) GetProducers() []types.PeerID {
	return emptyIDArr
}

func (d *bpCertificateManager) GetCertificates() []*p2pcommon.AgentCertificateV1 {
	return emptyCertArr
}

func (d *bpCertificateManager) AddCertificate(cert *p2pcommon.AgentCertificateV1) {
}

type agentCertificateManager struct {
	self p2pcommon.PeerMeta

	mutex sync.Mutex
	certs []*p2pcommon.AgentCertificateV1
}

func (a *agentCertificateManager) CreateCertificate(remoteMeta p2pcommon.PeerMeta) (*p2pcommon.AgentCertificateV1, error) {
	return nil, p2pcommon.ErrInvalidRole
}

func (a *agentCertificateManager) GetProducers() []types.PeerID {
	return a.self.ProducerIDs
}

func (a *agentCertificateManager) GetCertificates() []*p2pcommon.AgentCertificateV1 {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.certs
}

func (a *agentCertificateManager) AddCertificate(cert *p2pcommon.AgentCertificateV1) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if !containsID(a.self.ProducerIDs, cert.BPID) {
		// this agent is not in charge of that bp id.
		return
	}
	if !types.IsSamePeerID(a.self.ID, cert.AgentID) {
		// this certificate is not my certificate
		return
	}

	newCerts := make([]*p2pcommon.AgentCertificateV1, len(a.certs)+1)
	for _, oldCert := range a.certs {
		if !types.IsSamePeerID(oldCert.BPID, cert.BPID) {
			// replace old certificate if it already exists.
			newCerts = append(newCerts, oldCert)
		}
	}
	newCerts = append(newCerts, cert)
	a.certs = newCerts
}

func containsID(pool []types.PeerID, id types.PeerID) bool {
	for _, pid := range pool {
		if types.IsSamePeerID(pid, id) {
			return true
		}
	}
	return false
}

type watcherCertificateManager struct {
	self p2pcommon.PeerMeta
}

func (w *watcherCertificateManager) CreateCertificate(remoteMeta p2pcommon.PeerMeta) (*p2pcommon.AgentCertificateV1, error) {
	return nil, p2pcommon.ErrInvalidRole
}

func (w *watcherCertificateManager) GetProducers() []types.PeerID {
	return emptyIDArr
}

func (w *watcherCertificateManager) GetCertificates() []*p2pcommon.AgentCertificateV1 {
	return emptyCertArr
}

func (w *watcherCertificateManager) AddCertificate(cert *p2pcommon.AgentCertificateV1) {
	return
}
