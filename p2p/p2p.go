/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pkey"
	"github.com/aergoio/aergo/p2p/raftsupport"
	"github.com/aergoio/aergo/p2p/transport"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/rs/zerolog"
	"strings"
	"sync"
	"time"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/p2p/metric"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/p2p/subproto"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
)

// P2P is actor component for p2p
type P2P struct {
	*component.BaseComponent

	// TODO Which class has role to manager self PeerRole? P2P, PeerManager, or other?
	selfRole p2pcommon.PeerRole
	useRaft bool

	// caching data from genesis block
	chainID *types.ChainID
	nt      p2pcommon.NetworkTransport
	pm      p2pcommon.PeerManager
	vm      p2pcommon.VersionedManager
	sm      p2pcommon.SyncManager
	mm      metric.MetricsManager
	mf      p2pcommon.MoFactory
	signer  p2pcommon.MsgSigner
	ca      types.ChainAccessor
	consacc consensus.ConsensusAccessor
	prm     p2pcommon.PeerRoleManager

	mutex sync.Mutex
}

var (
	_ p2pcommon.ActorService     = (*P2P)(nil)
	_ p2pcommon.HSHandlerFactory = (*P2P)(nil)
)

// NewP2P create a new ActorService for p2p
func NewP2P(cfg *config.Config, chainsvc *chain.ChainService) *P2P {
	p2psvc := &P2P{}
	p2psvc.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2psvc, log.NewLogger("p2p"))
	p2psvc.initP2P(cfg, chainsvc)
	return p2psvc
}

// BeforeStart starts p2p service.
func (p2ps *P2P) BeforeStart() {}

func (p2ps *P2P) AfterStart() {
	p2ps.mutex.Lock()

	p2ps.setSelfRole()
	nt := p2ps.nt
	nt.Start()
	p2ps.mutex.Unlock()

	if err := p2ps.pm.Start(); err != nil {
		panic("Failed to start p2p component")
	}
	p2ps.mm.Start()
}

func (p2ps *P2P) setSelfRole() {
	selfPID := p2ps.pm.SelfNodeID()
	// set role of self peer
	ccinfo := p2ps.consacc.ConsensusInfo()
	if ccinfo.Type == "raft" {
		// it's raft chain
		p2ps.selfRole = p2pcommon.RaftWatcher
		bps := ccinfo.GetBps()
		for _, bp := range bps {
			if strings.Contains(bp, selfPID.Pretty() ) {
				p2ps.selfRole = p2pcommon.RaftFollower
				break
			}
		}
	} else {
		p2ps.selfRole = p2pcommon.Watcher
		bps := ccinfo.GetBps()
		for _, bp := range bps {
			if strings.Contains(bp, selfPID.Pretty() ) {
				p2ps.selfRole = p2pcommon.BlockProducer
				break
			}
		}
	}
	p2ps.Logger.Debug().Str("role", p2ps.selfRole.String()).Msg("set role of self")
}
// BeforeStop is called before actor hub stops. it finishes underlying peer manager
func (p2ps *P2P) BeforeStop() {
	p2ps.Logger.Debug().Msg("stopping p2p actor.")
	p2ps.mm.Stop()
	if err := p2ps.pm.Stop(); err != nil {
		p2ps.Logger.Warn().Err(err).Msg("Error on stopping peerManager")
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

func (p2ps *P2P) GetNetworkTransport() p2pcommon.NetworkTransport {
	p2ps.mutex.Lock()
	defer p2ps.mutex.Unlock()
	return p2ps.nt
}

func (p2ps *P2P) GetPeerAccessor() p2pcommon.PeerAccessor {
	return p2ps.pm
}

func (p2ps *P2P) SetConsensusAccessor(ca consensus.ConsensusAccessor) {
	p2ps.consacc = ca
}

func (p2ps *P2P) ChainID() *types.ChainID {
	return p2ps.chainID
}

func (p2ps *P2P) initP2P(cfg *config.Config, chainsvc *chain.ChainService) {
	p2ps.ca = chainsvc

	// check genesis block and get meta informations from it
	genesis := chainsvc.CDB().GetGenesisInfo()
	chainIdBytes, err := genesis.ChainID()
	if err != nil {
		panic("genesis block is not set properly: " + err.Error())
	}
	chainID := types.NewChainID()
	err = chainID.Read(chainIdBytes)
	if err != nil {
		panic("invalid chainid: " + err.Error())
	}
	p2ps.chainID = chainID

	useRaft := genesis.ConsensusType() == consensus.ConsensusName[consensus.ConsensusRAFT]
	p2ps.useRaft = useRaft

	netTransport := transport.NewNetworkTransport(cfg.P2P, p2ps.Logger)
	signer := newDefaultMsgSigner(p2pkey.NodePrivKey(), p2pkey.NodePubKey(), p2pkey.NodeID())

	// TODO: it should be refactored to support multi version
	mf := &baseMOFactory{}

	if useRaft {
		p2ps.prm = &RaftRoleManager{p2ps: p2ps, logger:p2ps.Logger, raftBP: make(map[types.PeerID]bool)}
	} else {
		p2ps.prm = &DefaultRoleManager{p2ps: p2ps}
	}
	metricMan := metric.NewMetricManager(10)
	peerMan := NewPeerManager(p2ps, p2ps, p2ps, cfg, p2ps, netTransport, metricMan, p2ps.Logger, mf, useRaft)
	syncMan := newSyncManager(p2ps, peerMan, p2ps.Logger)
	versionMan := newDefaultVersionManager(peerMan, p2ps, p2ps.Logger, p2ps.chainID)

	// connect managers each other

	p2ps.mutex.Lock()
	p2ps.signer = signer
	p2ps.nt = netTransport
	p2ps.mf = mf
	p2ps.pm = peerMan
	p2ps.vm = versionMan
	p2ps.sm = syncMan
	//p2ps.rm = reconMan
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

	case *message.GetSelf:
		context.Respond(p2ps.nt.SelfMeta())
	case *message.GetPeers:
		peers := p2ps.pm.GetPeerAddresses(msg.NoHidden, msg.ShowSelf)
		context.Respond(&message.GetPeersRsp{Peers: peers})
	case *message.GetSyncAncestor:
		p2ps.GetSyncAncestor(context, msg)
	case *message.MapQueryMsg:
		bestBlock, err := p2ps.GetChainAccessor().GetBestBlock()
		if err == nil {
			msg.BestBlock = bestBlock
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
	case *message.GetCluster:
		peers := p2ps.pm.GetPeers()
		clusterReceiver := raftsupport.NewClusterInfoReceiver(p2ps, p2ps.mf, peers, time.Second*5, msg)
		clusterReceiver.StartGet()
	case *message.SendRaft:
		p2ps.SendRaftMessage(context, msg)
	case *message.RaftClusterEvent:
		p2ps.Logger.Debug().Int("added", len(msg.BPAdded)).Int("removed", len(msg.BPRemoved)).Msg("bp changed")
		p2ps.prm.UpdateBP(msg.BPAdded, msg.BPRemoved)
	}
}

// TODO need refactoring. this code is copied from subproto/addrs.go
func (p2ps *P2P) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
	selfPeerID := p2ps.pm.SelfNodeID()
	peerMetas := make([]p2pcommon.PeerMeta, 0, len(peers))
	for _, rPeerAddr := range peers {
		rPeerID := types.PeerID(rPeerAddr.PeerID)
		if selfPeerID == rPeerID {
			continue
		}
		if p2putil.CheckAddressType(rPeerAddr.Address) == p2putil.AddressTypeError {
			continue
		}
		meta := p2pcommon.FromPeerAddress(rPeerAddr)
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
	return p2ps.RequestToFuture(actor, msg, p2pcommon.DefaultActorMsgTTL)
}

// CallRequest implement interface method of ActorService
func (p2ps *P2P) CallRequest(actor string, msg interface{}, timeout time.Duration) (interface{}, error) {
	future := p2ps.RequestToFuture(actor, msg, timeout)
	return future.Result()
}

// CallRequest implement interface method of ActorService
func (p2ps *P2P) CallRequestDefaultTimeout(actor string, msg interface{}) (interface{}, error) {
	future := p2ps.RequestToFuture(actor, msg, p2pcommon.DefaultActorMsgTTL)
	return future.Result()
}

// GetChainAccessor implement interface method of ActorService
func (p2ps *P2P) GetChainAccessor() types.ChainAccessor {
	return p2ps.ca
}

func (p2ps *P2P) InsertHandlers(peer p2pcommon.RemotePeer) {
	logger := p2ps.Logger

	// PingHandlers
	peer.AddMessageHandler(p2pcommon.PingRequest, subproto.NewPingReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.PingResponse, subproto.NewPingRespHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GoAway, subproto.NewGoAwayHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.AddressesRequest, subproto.NewAddressesReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.AddressesResponse, subproto.NewAddressesRespHandler(p2ps.pm, peer, logger, p2ps))

	// BlockHandlers
	peer.AddMessageHandler(p2pcommon.GetBlocksRequest, subproto.NewBlockReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetBlocksResponse, subproto.NewBlockRespHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm))
	peer.AddMessageHandler(p2pcommon.GetBlockHeadersRequest, subproto.NewListBlockHeadersReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetBlockHeadersResponse, subproto.NewListBlockRespHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetAncestorRequest, subproto.NewGetAncestorReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetAncestorResponse, subproto.NewGetAncestorRespHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetHashesRequest, subproto.NewGetHashesReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetHashesResponse, subproto.NewGetHashesRespHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetHashByNoRequest, subproto.NewGetHashByNoReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetHashByNoResponse, subproto.NewGetHashByNoRespHandler(p2ps.pm, peer, logger, p2ps))

	// TxHandlers
	peer.AddMessageHandler(p2pcommon.GetTXsRequest, subproto.WithTimeLog(subproto.NewTxReqHandler(p2ps.pm, peer, logger, p2ps), p2ps.Logger, zerolog.DebugLevel))
	peer.AddMessageHandler(p2pcommon.GetTXsResponse, subproto.WithTimeLog(subproto.NewTxRespHandler(p2ps.pm, peer, logger, p2ps), p2ps.Logger, zerolog.DebugLevel))
	peer.AddMessageHandler(p2pcommon.NewTxNotice, subproto.WithTimeLog(subproto.NewNewTxNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm), p2ps.Logger, zerolog.DebugLevel))

	// block notice handlers
	if p2ps.selfRole == p2pcommon.RaftLeader || p2ps.selfRole == p2pcommon.RaftFollower {
		peer.AddMessageHandler(p2pcommon.BlockProducedNotice, subproto.NewBPNoticeDiscardHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm))
		peer.AddMessageHandler(p2pcommon.NewBlockNotice, subproto.NewBlkNoticeDiscardHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm))
	} else {
		peer.AddMessageHandler(p2pcommon.BlockProducedNotice, subproto.WithTimeLog(subproto.NewBlockProducedNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm), p2ps.Logger, zerolog.DebugLevel ))
		peer.AddMessageHandler(p2pcommon.NewBlockNotice, subproto.NewNewBlockNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm))
	}

	// Raft support
	peer.AddMessageHandler(p2pcommon.GetClusterRequest, subproto.NewGetClusterReqHandler(p2ps.pm, peer, logger, p2ps, p2ps.consacc))
	peer.AddMessageHandler(p2pcommon.GetClusterResponse, subproto.NewGetClusterRespHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.RaftWrapperMessage, subproto.NewRaftWrapperHandler(p2ps.pm, peer, logger, p2ps, p2ps.consacc))

}

func (p2ps *P2P) CreateHSHandler(p2pVersion p2pcommon.P2PVersion, outbound bool, pid types.PeerID) p2pcommon.HSHandler {
	if p2pVersion == p2pcommon.P2PVersion030 {
		handshakeHandler := newHandshaker(p2ps.pm, p2ps, p2ps.Logger, p2ps.chainID, pid)
		if outbound {
			return &LegacyOutboundHSHandler{LegacyWireHandshaker: handshakeHandler}
		} else {
			return &LegacyInboundHSHandler{LegacyWireHandshaker: handshakeHandler}
		}
	} else {
		if outbound {
			return NewOutboundHSHandler(p2ps.pm, p2ps, p2ps.vm, p2ps.Logger, p2ps.chainID, pid)
		} else {
			return NewInboundHSHandler(p2ps.pm, p2ps, p2ps.vm, p2ps.Logger, p2ps.chainID, pid)
		}
	}
}

func (p2ps *P2P) CreateRemotePeer(meta p2pcommon.PeerMeta, seq uint32, status *types.Status, stream network.Stream, rw p2pcommon.MsgReadWriter) p2pcommon.RemotePeer {
	newPeer := newRemotePeer(meta, seq, p2ps.pm, p2ps, p2ps.Logger, p2ps.mf, p2ps.signer, rw)
	newPeer.UpdateBlkCache(status.GetBestBlockHash(), status.GetBestHeight())

	// TODO tune to set prefer role
	newPeer.role = p2ps.prm.GetRole(meta.ID)
	// insert Handlers
	p2ps.InsertHandlers(newPeer)

	return newPeer
}

