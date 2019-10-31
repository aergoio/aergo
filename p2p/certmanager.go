package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pkey"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"

	"sync"
	"time"
)

var emptyIDArr []types.PeerID
var emptyCertArr []*p2pcommon.AgentCertificateV1

func newCertificateManager(actor p2pcommon.ActorService, self p2pcommon.PeerMeta, logger *log.Logger) p2pcommon.CertificateManager {
	d := baseCertManager{actor: actor, self: self}
	switch self.Role {
	case types.PeerRole_Producer:
		pk := p2putil.ConvertPKToBTCEC(p2pkey.NodePrivKey())
		if pk == nil {
			panic(fmt.Sprintf("invalid pk %v",p2pkey.NodePrivKey()))
		}
		return &bpCertificateManager{baseCertManager: d, key:pk}
	case types.PeerRole_Agent:
		return &agentCertificateManager{baseCertManager: d, logger:logger}
	case types.PeerRole_Watcher:
		return &watcherCertificateManager{baseCertManager: d}
	default:
		return nil
	}
}

type baseCertManager struct {
	actor p2pcommon.ActorService
	self  p2pcommon.PeerMeta
}

func (cm *baseCertManager) Start() {
}

func (cm *baseCertManager) Stop() {
}

func (cm *baseCertManager) CreateCertificate(remoteMeta p2pcommon.PeerMeta) (*p2pcommon.AgentCertificateV1, error) {
	return nil, p2pcommon.ErrInvalidRole
}

func (cm *baseCertManager) GetProducers() []types.PeerID {
	return emptyIDArr
}

func (cm *baseCertManager) GetCertificates() []*p2pcommon.AgentCertificateV1 {
	return emptyCertArr
}

func (cm *baseCertManager) AddCertificate(cert *p2pcommon.AgentCertificateV1) {
}

func (cm *baseCertManager) OnPeerConnect(pid types.PeerID) {
}

func (cm *baseCertManager) OnPeerDisconnect(peer p2pcommon.RemotePeer) {
}

type bpCertificateManager struct {
	baseCertManager
	key *btcec.PrivateKey
}

func (d *bpCertificateManager) CreateCertificate(remoteMeta p2pcommon.PeerMeta) (*p2pcommon.AgentCertificateV1, error) {
	addrs := make([]string, len(remoteMeta.Addresses))
	for i, ad := range remoteMeta.Addresses {
		addrs[i] = types.AddressFromMultiAddr(ad)
	}
	return p2putil.NewAgentCertV1(d.self.ID, remoteMeta.ID, d.key, addrs, time.Hour*24)
}

type agentCertificateManager struct {
	baseCertManager

	logger *log.Logger
	ticker *time.Ticker
	mutex  sync.Mutex
	certs  []*p2pcommon.AgentCertificateV1
}

func (cm *agentCertificateManager) Start() {
	go func() {
		cm.logger.Info().Msg("Starting p2p certificate manager ")
		cm.ticker = time.NewTicker(time.Hour)
		for range cm.ticker.C {
			cm.mutex.Lock()
			now := time.Now()

			certs2 := cm.certs[:0]
			for _, cert := range cm.certs {
				if cert.IsNeedUpdate(now, p2putil.DefaultExpireBufTerm) {
					cm.actor.TellRequest(message.P2PSvc, message.IssueAgentCertificate{cert.BPID} )
					if cert.IsValidInTime(now, p2putil.TimeErrorTolerance) {
						certs2 = append(certs2, cert)
					}
				}
			}
			cm.certs = certs2
			cm.mutex.Unlock()
		}
	}()
}

func (cm *agentCertificateManager) Stop() {
	cm.logger.Info().Msg("Finishing p2p certificate manager")
	cm.ticker.Stop()

}
func (cm *agentCertificateManager) CreateCertificate(remoteMeta p2pcommon.PeerMeta) (*p2pcommon.AgentCertificateV1, error) {
	return nil, p2pcommon.ErrInvalidRole
}

func (cm *agentCertificateManager) GetProducers() []types.PeerID {
	return cm.self.ProducerIDs
}

func (cm *agentCertificateManager) GetCertificates() []*p2pcommon.AgentCertificateV1 {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	return cm.certs
}

func (cm *agentCertificateManager) AddCertificate(cert *p2pcommon.AgentCertificateV1) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	if !containsID(cm.self.ProducerIDs, cert.BPID) {
		// this agent is not in charge of that bp id.
		cm.logger.Info().Str("bpID",p2putil.ShortForm(cert.BPID)).Msg("drop issued certificate, since issuer is not my managed producer")
		return
	}
	if !types.IsSamePeerID(cm.self.ID, cert.AgentID) {
		// this certificate is not my certificate
		cm.logger.Info().Str("bpID",p2putil.ShortForm(cert.BPID)).Str("agentID",p2putil.ShortForm(cert.AgentID)).Msg("drop issued certificate, since agent id is not me")
		return
	}

	newCerts := make([]*p2pcommon.AgentCertificateV1, 0, len(cm.certs)+1)
	for _, oldCert := range cm.certs {
		if !types.IsSamePeerID(oldCert.BPID, cert.BPID) {
			// replace old certificate if it already exists.
			newCerts = append(newCerts, oldCert)
		}
	}
	newCerts = append(newCerts, cert)
	cm.logger.Info().Str("bpID",p2putil.ShortForm(cert.BPID)).Time("cTime",cert.CreateTime).Time("eTime",cert.ExpireTime).Msg("issued certificate is added to my certificate list")
	cm.certs = newCerts
	pCert, err := p2putil.ConvertCertToProto(cert)
	if err != nil {
		return
	}
	cm.actor.TellRequest(message.P2PSvc, message.NotifyCertRenewed{pCert} )
}

func (cm *agentCertificateManager) OnPeerConnect(pid types.PeerID) {
	// check if peer is producer which is managed
	if !containsID(cm.self.ProducerIDs, pid) {
		return
	}
	// check if certificate exists adn is still valid
	var prevCert *p2pcommon.AgentCertificateV1 = nil
	for _, cert := range cm.certs {
		if types.IsSamePeerID(cert.BPID, pid) {
			prevCert = cert
			break
		}
	}
	// then send issueCert if not.
	// FIXME it still have inefficiency that issue
	if prevCert == nil || prevCert.IsNeedUpdate(time.Now(), p2putil.DefaultExpireBufTerm) {
		cm.actor.TellRequest(message.P2PSvc, message.IssueAgentCertificate{pid} )
	}
}

func (cm *agentCertificateManager) OnPeerDisconnect(peer p2pcommon.RemotePeer) {
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
	baseCertManager
}

func (w *watcherCertificateManager) CreateCertificate(remoteMeta p2pcommon.PeerMeta) (*p2pcommon.AgentCertificateV1, error) {
	return nil, p2pcommon.ErrInvalidRole
}
