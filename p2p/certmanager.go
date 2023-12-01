package p2p

import (
	"fmt"
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
	"github.com/btcsuite/btcd/btcec/v2"
)

var emptyIDArr []types.PeerID
var emptyCertArr []*p2pcommon.AgentCertificateV1

func newCertificateManager(actor p2pcommon.ActorService, is p2pcommon.InternalService, logger *log.Logger) p2pcommon.CertificateManager {
	d := baseCertManager{actor: actor, self: is.SelfMeta(), settings: is.LocalSettings(), logger: logger}
	switch d.self.Role {
	case types.PeerRole_Producer:
		pk := p2putil.ConvertPKToBTCEC(p2pkey.NodePrivKey())
		if pk == nil {
			panic(fmt.Sprintf("invalid pk %v", p2pkey.NodePrivKey()))
		}
		return &bpCertificateManager{baseCertManager: d, key: pk}
	case types.PeerRole_Agent:
		return &agentCertificateManager{baseCertManager: d, certMap: make(map[types.PeerID]*p2pcommon.AgentCertificateV1)}
	case types.PeerRole_Watcher:
		return &watcherCertificateManager{baseCertManager: d}
	default:
		return nil
	}
}

type baseCertManager struct {
	actor    p2pcommon.ActorService
	self     p2pcommon.PeerMeta
	settings p2pcommon.LocalSettings
	logger   *log.Logger
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

func (cm *baseCertManager) CanHandle(bpID types.PeerID) bool {
	return false
}

type bpCertificateManager struct {
	baseCertManager
	key *btcec.PrivateKey
}

func (cm *bpCertificateManager) CreateCertificate(remoteMeta p2pcommon.PeerMeta) (*p2pcommon.AgentCertificateV1, error) {
	if !types.IsSamePeerID(cm.settings.AgentID, remoteMeta.ID) {
		// this agent is not in charge of that bp id.
		cm.logger.Info().Stringer("agentID", types.LogPeerShort(remoteMeta.ID)).Msg("failed to issue certificate, since peer is not registered agent")
		return nil, p2pcommon.ErrInvalidRole
	}

	addrs := make([]string, len(remoteMeta.Addresses))
	for i, ad := range remoteMeta.Addresses {
		addrs[i] = types.AddressFromMultiAddr(ad)
	}
	return p2putil.NewAgentCertV1(cm.self.ID, remoteMeta.ID, cm.key, addrs, p2pcommon.DefaultCertTTL)
}

type agentCertificateManager struct {
	baseCertManager

	ticker *time.Ticker
	mutex  sync.Mutex
	// copy-on-write style slice
	certs []*p2pcommon.AgentCertificateV1
	// not thread-safe map
	certMap map[types.PeerID]*p2pcommon.AgentCertificateV1
}

func (cm *agentCertificateManager) Start() {
	go func() {
		cm.logger.Info().Msg("Starting p2p certificate manager ")
		cm.ticker = time.NewTicker(p2pcommon.LocalCertCheckInterval)
		for range cm.ticker.C {
			cm.mutex.Lock()
			now := time.Now()
			cm.checkCertificates(now)
			cm.mutex.Unlock()
		}
	}()
}

func (cm *agentCertificateManager) checkCertificates(now time.Time) {
	cm.logger.Debug().Int("certCnt", len(cm.certs)).Msg("periodic check for local certificates")

	certs2 := make([]*p2pcommon.AgentCertificateV1, 0, len(cm.certs))
	for _, cert := range cm.certs {
		if cert.IsNeedUpdate(now, p2pcommon.DefaultExpireBufTerm) {
			cm.actor.TellRequest(message.P2PSvc, message.IssueAgentCertificate{ProducerID: cert.BPID})
			if cert.IsValidInTime(now, p2pcommon.TimeErrorTolerance) {
				certs2 = append(certs2, cert)
			} else {
				cm.logger.Debug().Int("certCnt", len(cm.certs)).Msg("removing expired certificates")
				delete(cm.certMap, cert.BPID)
			}
		} else {
			certs2 = append(certs2, cert)
		}
	}
	cm.certs = certs2
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
	if !p2putil.ContainsID(cm.self.ProducerIDs, cert.BPID) {
		// this agent is not in charge of that bp id.
		cm.logger.Info().Stringer("bpID", types.LogPeerShort(cert.BPID)).Msg("drop issued certificate, since issuer is not my managed producer")
		return
	}
	if !types.IsSamePeerID(cm.self.ID, cert.AgentID) {
		// this certificate is not my certificate
		cm.logger.Info().Stringer("bpID", types.LogPeerShort(cert.BPID)).Stringer("agentID", types.LogPeerShort(cert.AgentID)).Msg("drop issued certificate, since agent id is not me")
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
	cm.logger.Info().Object("cert", p2putil.AgentCertMarshaller{AgentCertificateV1: cert}).Msg("issued certificate is added to my certificate list")
	cm.certs = newCerts
	cm.certMap[cert.BPID] = cert
	pCert, err := p2putil.ConvertCertToProto(cert)
	if err != nil {
		return
	}
	cm.actor.TellRequest(message.P2PSvc, message.NotifyCertRenewed{Cert: pCert})
}

func (cm *agentCertificateManager) OnPeerConnect(pid types.PeerID) {
	// check if peer is producer which is managed
	if !p2putil.ContainsID(cm.self.ProducerIDs, pid) {
		return
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	// check if certificate exists and is still valid
	var prevCert *p2pcommon.AgentCertificateV1 = nil
	for _, cert := range cm.certs {
		if types.IsSamePeerID(cert.BPID, pid) {
			prevCert = cert
			break
		}
	}
	// then send issueCert if not
	if prevCert == nil || prevCert.IsNeedUpdate(time.Now(), p2pcommon.DefaultExpireBufTerm) {
		cm.actor.TellRequest(message.P2PSvc, message.IssueAgentCertificate{ProducerID: pid})
	}
}

func (cm *agentCertificateManager) CanHandle(bpID types.PeerID) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	_, found := cm.certMap[bpID]
	return found
}

func (cm *agentCertificateManager) OnPeerDisconnect(peer p2pcommon.RemotePeer) {
}

type watcherCertificateManager struct {
	baseCertManager
}

func (w *watcherCertificateManager) CreateCertificate(remoteMeta p2pcommon.PeerMeta) (*p2pcommon.AgentCertificateV1, error) {
	return nil, p2pcommon.ErrInvalidRole
}
