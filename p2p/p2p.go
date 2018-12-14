/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/metric"
	"io/ioutil"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
)

type nodeInfo struct {
	id      peer.ID
	sid     string
	pubKey  crypto.PubKey
	privKey crypto.PrivKey
}

// P2P is actor component for p2p
type P2P struct {
	*component.BaseComponent

	nt 	NetworkTransport
	pm     PeerManager
	sm     SyncManager
	rm     ReconnectManager
	mm 	metric.MetricsManager
	mf     moFactory
	signer msgSigner
	ca     types.ChainAccessor
}

type HandlerFactory interface {
	insertHandlers(peer *remotePeerImpl)
}

var (
	_  ActorService = (*P2P)(nil)
	_ HSHandlerFactory = (*P2P)(nil)
	ni *nodeInfo
)

// InitNodeInfo initializes node-specific informations like node id.
// Caution: this must be called before all the goroutines are started.
func InitNodeInfo(cfg *config.P2PConfig, logger *log.Logger) {
	// check Key and address
	var (
		priv crypto.PrivKey
		pub  crypto.PubKey
	)

	if cfg.NPKey != "" {
		dat, err := ioutil.ReadFile(cfg.NPKey)
		if err == nil {
			priv, err = crypto.UnmarshalPrivateKey(dat)
			if err != nil {
				logger.Warn().Str("npkey", cfg.NPKey).Msg("invalid keyfile. It's not private key file")
			}
			pub = priv.GetPublic()
		} else {
			logger.Warn().Str("npkey", cfg.NPKey).Msg("invalid keyfile path")
		}
	}

	if priv == nil {
		logger.Info().Msg("No valid private key file is found. use temporary pk instead")
		priv, pub, _ = crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	}

	id, _ := peer.IDFromPublicKey(pub)

	ni = &nodeInfo{
		id:      id,
		sid:     enc.ToString([]byte(id)),
		pubKey:  pub,
		privKey: priv,
	}
}

// NodeID returns the node id.
func NodeID() peer.ID {
	return ni.id
}

// NodeSID returns the string representation of the node id.
func NodeSID() string {
	if ni == nil {
		return ""
	}
	return ni.sid
}

// NodePrivKey returns the private key of the node.
func NodePrivKey() crypto.PrivKey {
	return ni.privKey
}

// NodePubKey returns the public key of the node.
func NodePubKey() crypto.PubKey {
	return ni.pubKey
}

// NewP2P create a new ActorService for p2p
func NewP2P(cfg *config.Config, chainsvc *chain.ChainService) *P2P {
	p2psvc := &P2P{}
	p2psvc.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2psvc, log.NewLogger("p2p"))
	p2psvc.init(cfg, chainsvc)
	return p2psvc
}

// BeforeStart starts p2p service.
func (p2ps *P2P) BeforeStart() {}

func (p2ps *P2P) AfterStart() {
	p2ps.nt.Start()
	if err := p2ps.pm.Start(); err != nil {
		panic("Failed to start p2p component")
	}
	p2ps.mm.Start()
}

// BeforeStop is called before actor hub stops. it finishes underlying peer manager
func (p2ps *P2P) BeforeStop() {
	p2ps.mm.Stop()
	if err := p2ps.pm.Stop(); err != nil {
		p2ps.Logger.Warn().Err(err).Msg("Erro on stopping peerManager")
	}
	p2ps.nt.Stop()
}

// Statistics show statistic information of p2p module. NOTE: It it not implemented yet
func (p2ps *P2P) Statistics() *map[string]interface{} {
	return nil
}

func (p2ps *P2P) init(cfg *config.Config, chainsvc *chain.ChainService) {
	p2ps.ca = chainsvc

	netTransport := NewNetworkTransport(cfg.P2P, p2ps.Logger)
	signer := newDefaultMsgSigner(ni.privKey, ni.pubKey, ni.id)
	mf := &pbMOFactory{signer: signer}
	reconMan := newReconnectManager(p2ps.Logger)
	metricMan := metric.NewMetricManager(10)
	peerMan := NewPeerManager(p2ps, p2ps, p2ps, cfg, signer, netTransport, reconMan, metricMan, p2ps.Logger, mf)
	syncMan := newSyncManager(p2ps, peerMan, p2ps.Logger)

	// connect managers each other
	reconMan.pm = peerMan

	p2ps.signer = signer
	p2ps.nt = netTransport
	p2ps.mf = mf
	p2ps.pm = peerMan
	p2ps.sm = syncMan
	p2ps.rm = reconMan
	p2ps.mm = metricMan
}

// Receive got actor message and then handle it.
func (p2ps *P2P) Receive(context actor.Context) {
	rawMsg := context.Message()
	switch msg := rawMsg.(type) {
	case *message.GetAddressesMsg:
		p2ps.GetAddresses(msg.ToWhom, msg.Size)
	case *message.GetMetrics:
		context.Respond(p2ps.mm.Metrics())
	case *message.GetBlockHeaders:
		p2ps.GetBlockHeaders(msg)
	case *message.GetBlockChunks:
		p2ps.GetBlocksChunk(context, msg)
	case *message.GetBlockInfos:
		p2ps.GetBlocks(msg.ToWhom, msg.Hashes)
	case *message.GetHashes:
		p2ps.GetBlockHashes(context, msg)
	case *message.GetHashByNo:
		p2ps.GetBlockHashByNo(context, msg)
	case *message.NotifyNewBlock:
		p2ps.NotifyNewBlock(*msg)
	case *message.GetMissingBlocks:
		p2ps.GetMissingBlocks(msg.ToWhom, msg.Hashes)
	case *message.GetTransactions:
		p2ps.GetTXs(msg.ToWhom, msg.Hashes)
	case *message.NotifyNewTransactions:
		p2ps.NotifyNewTX(*msg)
	case *message.AddBlockRsp:
		// do nothing for now. just for prevent deadletter

	case *message.GetPeers:
		peers, lastBlks, states := p2ps.pm.GetPeerAddresses()
		context.Respond(&message.GetPeersRsp{Peers: peers, LastBlks: lastBlks, States: states})
	case *message.GetSyncAncestor:
		p2ps.GetSyncAncestor(msg.ToWhom, msg.Hashes)
	}
}

// TellRequest implement interface method of ActorService
func (p2ps *P2P) TellRequest(actor string, msg interface{}) {
	p2ps.TellTo(actor, msg)
}

// SendRequest implement interface method of ActorService
func (p2ps *P2P) SendRequest(actor string, msg interface{}) {
	p2ps.RequestTo(actor, msg)
}

// FutureRequest implement interface method of ActorService
func (p2ps *P2P) FutureRequest(actor string, msg interface{}, timeout time.Duration) *actor.Future {
	return p2ps.RequestToFuture(actor, msg, timeout)
}

// FutureRequestDefaultTimeout implement interface method of ActorService
func (p2ps *P2P) FutureRequestDefaultTimeout(actor string, msg interface{}) *actor.Future {
	return p2ps.RequestToFuture(actor, msg, defaultActorMsgTTL)
}

// CallRequest implement interface method of ActorService
func (p2ps *P2P) CallRequest(actor string, msg interface{}, timeout time.Duration) (interface{}, error) {
	future := p2ps.RequestToFuture(actor, msg, timeout)
	return future.Result()
}

// CallRequest implement interface method of ActorService
func (p2ps *P2P) CallRequestDefaultTimeout(actor string, msg interface{}) (interface{}, error) {
	future := p2ps.RequestToFuture(actor, msg, defaultActorMsgTTL)
	return future.Result()
}

// GetChainAccessor implment interface method of ActorService
func (p2ps *P2P) GetChainAccessor() types.ChainAccessor {
	return p2ps.ca
}

func (p2ps *P2P) insertHandlers(peer *remotePeerImpl) {
	logger := p2ps.Logger

	// PingHandlers
	peer.handlers[PingRequest] = newPingReqHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[PingResponse] = newPingRespHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GoAway] = newGoAwayHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[AddressesRequest] = newAddressesReqHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[AddressesResponse] = newAddressesRespHandler(p2ps.pm, peer, logger, p2ps)

	// BlockHandlers
	peer.handlers[GetBlocksRequest] = newBlockReqHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GetBlocksResponse] = newBlockRespHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm)
	peer.handlers[GetBlockHeadersRequest] = newListBlockHeadersReqHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GetBlockHeadersResponse] = newListBlockRespHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GetMissingRequest] = newGetMissingReqHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[NewBlockNotice] = newNewBlockNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm)
	peer.handlers[GetAncestorRequest] = newGetAncestorReqHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GetAncestorResponse] = newGetAncestorRespHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GetHashesRequest] = newGetHashesReqHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GetHashesResponse] = newGetHashesRespHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GetHashByNoRequest] = newGetHashByNoReqHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GetHashByNoResponse] = newGetHashByNoRespHandler(p2ps.pm, peer, logger, p2ps)

	// TxHandlers
	peer.handlers[GetTXsRequest] = newTxReqHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[GetTxsResponse] = newTxRespHandler(p2ps.pm, peer, logger, p2ps)
	peer.handlers[NewTxNotice] = newNewTxNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm)
}

func (p2ps *P2P) CreateHSHandler(outbound bool, pm PeerManager, actor ActorService, log *log.Logger, pid peer.ID) HSHandler {
	handshakeHandler := &PeerHandshaker{pm: pm, actorServ: actor, logger: log, peerID: pid}
	if outbound {
		return &OutboundHSHandler{PeerHandshaker: handshakeHandler}
	} else {
		return &InboundHSHandler{PeerHandshaker: handshakeHandler}
	}
}
