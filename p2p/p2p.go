/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/metric"
	"github.com/aergoio/aergo/p2p/p2putil"
	"os"
	"os/user"
	"path/filepath"
	"sync"
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

	// caching data from genesis block
	chainID *types.ChainID
	nt 	NetworkTransport
	pm     PeerManager
	sm     SyncManager
	rm     ReconnectManager
	mm 	metric.MetricsManager
	mf     moFactory
	signer msgSigner
	ca     types.ChainAccessor

	mutex sync.Mutex
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
		err  error
	)

	if cfg.NPKey != "" {
		priv, pub, err = LoadKeyFile(cfg.NPKey)
		if err != nil {
			panic("Failed to load Keyfile '" + cfg.NPKey + "' " + err.Error())
		}
	} else {
		logger.Info().Msg("No private key file is configured, so use auto-generated pk file instead.")
		usr, err := user.Current()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to get user's home")
		}

		autogenFilePath := filepath.Join(usr.HomeDir, DefaultPkKeyDirectory, DefaultPkKeyPrefix+DefaultPkKeyExt)
		if _, err := os.Stat(autogenFilePath); os.IsNotExist(err) {
			logger.Info().Str("pk_file", autogenFilePath).Msg("Generate new private key file.")
			priv, pub, err = GenerateKeyFile(filepath.Join(usr.HomeDir, DefaultPkKeyDirectory), DefaultPkKeyPrefix)
			if err != nil {
				panic("Failed to generate new pk file: "+err.Error())
			}
		} else {
			logger.Info().Str("pk_file", autogenFilePath).Msg("Load existing generated private key file.")
			priv, pub, err = LoadKeyFile(autogenFilePath)
			if err != nil {
				panic("Failed to load generated pk file '"+autogenFilePath+"' "+err.Error())
			}
		}
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
	p2ps.mutex.Lock()

	nt := p2ps.nt
	nt.Start()
	p2ps.mutex.Unlock()

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
	p2ps.mutex.Lock()
	nt := p2ps.nt
	p2ps.mutex.Unlock()
	nt.Stop()
}

// Statistics show statistic information of p2p module. NOTE: It it not implemented yet
func (p2ps *P2P) Statistics() *map[string]interface{} {
	stmap := make(map[string]interface{})
	stmap["netstat"] = p2ps.mm.Summary()
	return &stmap
}


func (p2ps *P2P) GetNetworkTransport() NetworkTransport {
	p2ps.mutex.Lock()
	defer p2ps.mutex.Unlock()
	return p2ps.nt
}

func (p2ps *P2P) ChainID() *types.ChainID {
	return p2ps.chainID
}

func (p2ps *P2P) init(cfg *config.Config, chainsvc *chain.ChainService) {
	p2ps.ca = chainsvc

	// check genesis block and get meta informations from it
	genesis := chainsvc.CDB().GetGenesisInfo()
	chainIdBytes, err := genesis.ChainID()
	if err != nil {
		panic("genesis block is not set properly: "+err.Error())
	}
	chainID := types.NewChainID()
	err = chainID.Read(chainIdBytes)
	if err != nil {
		panic("invalid chainid: "+err.Error())
	}
	p2ps.chainID = chainID

	netTransport := NewNetworkTransport(cfg.P2P, p2ps.Logger)
	signer := newDefaultMsgSigner(ni.privKey, ni.pubKey, ni.id)
	mf := &v030MOFactory{}
	reconMan := newReconnectManager(p2ps.Logger)
	metricMan := metric.NewMetricManager(10)
	peerMan := NewPeerManager(p2ps, p2ps, p2ps, cfg, signer, netTransport, reconMan, metricMan, p2ps.Logger, mf)
	syncMan := newSyncManager(p2ps, peerMan, p2ps.Logger)

	// connect managers each other
	reconMan.pm = peerMan

	p2ps.mutex.Lock()
	p2ps.signer = signer
	p2ps.nt = netTransport
	p2ps.mf = mf
	p2ps.pm = peerMan
	p2ps.sm = syncMan
	p2ps.rm = reconMan
	p2ps.mm = metricMan
	p2ps.mutex.Unlock()
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
		if msg.Produced {
			p2ps.NotifyBlockProduced(*msg)
		} else {
			p2ps.NotifyNewBlock(*msg)
		}
	case *message.GetTransactions:
		p2ps.GetTXs(msg.ToWhom, msg.Hashes)
	case *message.NotifyNewTransactions:
		p2ps.NotifyNewTX(*msg)
	case *message.AddBlockRsp:
		// do nothing for now. just for prevent deadletter

	case *message.GetPeers:
		peers, hiddens, lastBlks, states := p2ps.pm.GetPeerAddresses()
		context.Respond(&message.GetPeersRsp{Peers: peers, Hiddens:hiddens, LastBlks: lastBlks, States: states})
	case *message.GetSyncAncestor:
		p2ps.GetSyncAncestor(msg.ToWhom, msg.Hashes)

	case *message.MapQueryMsg:
		bestBlock, err := p2ps.GetChainAccessor().GetBestBlock()
		if err == nil {
			msg.BestBlock=bestBlock
			p2ps.SendRequest(message.MapSvc, msg)
		}
	case *message.MapQueryRsp:
		if msg.Err != nil {
			p2ps.Logger.Info().Err(msg.Err).Msg("polaris returned error")
		} else {
			if len(msg.Peers) > 0 {
				p2ps.checkAndAddPeerAddresses(msg.Peers)
			}
		}
	}
}


// TODO need refactoring. this code is copied from subprotcoladdrs.go
func (p2ps *P2P) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
	selfPeerID := p2ps.pm.SelfNodeID()
	peerMetas := make([]PeerMeta, 0, len(peers))
	for _, rPeerAddr := range peers {
		rPeerID := peer.ID(rPeerAddr.PeerID)
		if selfPeerID == rPeerID {
			continue
		}
		if p2putil.CheckAdddressType(rPeerAddr.Address) == p2putil.AddressTypeError {
			continue
		}
		meta := FromPeerAddress(rPeerAddr)
		peerMetas = append(peerMetas, meta)
	}
	if len(peerMetas) > 0 {
		p2ps.pm.NotifyPeerAddressReceived(peerMetas)
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
	return p2ps.RequestToFuture(actor, msg, DefaultActorMsgTTL)
}

// CallRequest implement interface method of ActorService
func (p2ps *P2P) CallRequest(actor string, msg interface{}, timeout time.Duration) (interface{}, error) {
	future := p2ps.RequestToFuture(actor, msg, timeout)
	return future.Result()
}

// CallRequest implement interface method of ActorService
func (p2ps *P2P) CallRequestDefaultTimeout(actor string, msg interface{}) (interface{}, error) {
	future := p2ps.RequestToFuture(actor, msg, DefaultActorMsgTTL)
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

	// BP protocol handlers
	peer.handlers[BlockProducedNotice] = newBlockProducedNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm)

}

func (p2ps *P2P) CreateHSHandler(outbound bool, pm PeerManager, actor ActorService, log *log.Logger, pid peer.ID) HSHandler {
	handshakeHandler := &PeerHandshaker{pm: pm, actorServ: actor, logger: log, localChainID:p2ps.chainID, peerID: pid}
	if outbound {
		return &OutboundHSHandler{PeerHandshaker: handshakeHandler}
	} else {
		return &InboundHSHandler{PeerHandshaker: handshakeHandler}
	}
}
