/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/list"
	"github.com/aergoio/aergo/v2/p2p/metric"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/p2p/raftsupport"
	"github.com/aergoio/aergo/v2/p2p/subproto"
	"github.com/aergoio/aergo/v2/p2p/transport"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/rs/zerolog"
)

// P2P is actor component for p2p
type P2P struct {
	*component.BaseComponent

	cfg *config.Config

	// inited during construction
	useRaft  bool
	selfMeta p2pcommon.PeerMeta
	// caching data from genesis block
	genesisChainID *types.ChainID
	localSettings  p2pcommon.LocalSettings

	nt     p2pcommon.NetworkTransport
	pm     p2pcommon.PeerManager
	vm     p2pcommon.VersionedManager
	sm     p2pcommon.SyncManager
	mm     metric.MetricsManager
	mf     p2pcommon.MoFactory
	signer p2pcommon.MsgSigner
	ca     types.ChainAccessor
	prm    p2pcommon.PeerRoleManager
	lm     p2pcommon.ListManager
	cm     p2pcommon.CertificateManager
	mutex  sync.Mutex

	// inited between construction and start
	consacc consensus.ConsensusAccessor
}

var (
	_ p2pcommon.ActorService     = (*P2P)(nil)
	_ p2pcommon.HSHandlerFactory = (*P2P)(nil)
)

// NewP2P create a new ActorService for p2p
func NewP2P(cfg *config.Config, chainSvc *chain.ChainService) *P2P {
	p2psvc := &P2P{cfg: cfg}
	p2psvc.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2psvc, log.NewLogger("p2p"))
	p2psvc.initP2P(chainSvc)
	return p2psvc
}

func (p2ps *P2P) initP2P(chainSvc *chain.ChainService) {
	cfg := p2ps.cfg
	p2ps.ca = chainSvc

	// check genesis block and get meta information from it
	genesis := chainSvc.CDB().GetGenesisInfo()
	chainIdBytes, err := genesis.ChainID()
	if err != nil {
		panic("genesis block is not set properly: " + err.Error())
	}
	chainID := types.NewChainID()
	err = chainID.Read(chainIdBytes)
	if err != nil {
		panic("invalid chainid: " + err.Error())
	}
	p2ps.genesisChainID = chainID
	p2ps.useRaft = genesis.ConsensusType() == consensus.ConsensusName[consensus.ConsensusRAFT]

	p2ps.selfMeta = SetupSelfMeta(p2pkey.NodeID(), cfg.P2P, cfg.Consensus.EnableBp)
	p2ps.initLocalSettings(cfg.P2P)
	// set selfMeta.AcceptedRole and init role manager
	p2ps.cm = newCertificateManager(p2ps, p2ps, p2ps.Logger)
	p2ps.prm = p2ps.initRoleManager(p2ps.useRaft, p2ps.selfMeta.Role, p2ps.cm)

	netTransport := transport.NewNetworkTransport(cfg.P2P, p2ps.Logger, p2ps)
	signer := newDefaultMsgSigner(p2pkey.NodePrivKey(), p2pkey.NodePubKey(), p2pkey.NodeID())

	mf := &baseMOFactory{is: p2ps}

	// public network is always disabled white/blacklist in chain
	lm := list.NewListManager(cfg.Auth, cfg.AuthDir, p2ps.ca, p2ps.prm, p2ps.Logger, genesis.PublicNet())
	metricMan := metric.NewMetricManager(10)
	peerMan := NewPeerManager(p2ps, p2ps, p2ps, p2ps, netTransport, metricMan, lm, p2ps.Logger, cfg, p2ps.useRaft)
	syncMan := newSyncManager(p2ps, peerMan, p2ps.Logger)
	versionMan := newDefaultVersionManager(p2ps, p2ps, peerMan, p2ps.ca, p2ps.Logger, p2ps.genesisChainID)

	// connect managers each other
	peerMan.AddPeerEventListener(p2ps.cm)

	p2ps.mutex.Lock()
	p2ps.signer = signer
	p2ps.nt = netTransport
	p2ps.mf = mf
	p2ps.pm = peerMan
	p2ps.vm = versionMan
	p2ps.sm = syncMan
	//p2ps.rm = reconMan
	p2ps.mm = metricMan
	p2ps.lm = lm

	p2ps.mutex.Unlock()
}

func (p2ps *P2P) initRoleManager(useRaft bool, role types.PeerRole, cm p2pcommon.CertificateManager) p2pcommon.PeerRoleManager {
	var prm p2pcommon.PeerRoleManager
	if useRaft {
		prm = NewRaftRoleManager(p2ps, p2ps, p2ps.Logger)
	} else {
		if role == types.PeerRole_Agent {
			if len(p2ps.cfg.P2P.Producers) == 0 {
				panic("agent must have one or more producers ")
			}
			var pds = make(map[types.PeerID]bool)
			for _, pidStr := range p2ps.cfg.P2P.Producers {
				pid, err := types.IDB58Decode(pidStr)
				if err != nil {
					panic("invalid producer id " + pidStr)
				}
				pds[pid] = true
			}
			prm = NewDPOSAgentRoleManager(p2ps, p2ps, p2ps.Logger, pds)
		} else {
			prm = NewDPOSRoleManager(p2ps, p2ps, p2ps.Logger)
		}
	}
	return prm
}

// BeforeStart starts p2p service.
func (p2ps *P2P) BeforeStart() {}

func (p2ps *P2P) AfterStart() {
	versions := make([]fmt.Stringer, len(p2pcommon.AcceptedInboundVersions))
	for i, ver := range p2pcommon.AcceptedInboundVersions {
		versions[i] = ver
	}
	p2ps.lm.Start()
	p2ps.mutex.Lock()
	p2ps.checkConsensus()
	p2ps.Logger.Info().Array("supportedVersions", p2putil.NewLogStringersMarshaller(versions, 10)).Str("info", p2putil.ShortMetaForm(p2ps.selfMeta)).Str("role", p2ps.selfMeta.Role.String()).Msg("Starting p2p component")

	nt := p2ps.nt
	nt.Start()
	p2ps.mutex.Unlock()

	if err := p2ps.pm.Start(); err != nil {
		panic("Failed to start p2p component")
	}
	p2ps.mm.Start()
	p2ps.cm.Start()
	p2ps.prm.Start()
	p2ps.sm.Start()
}

func (p2ps *P2P) checkConsensus() {
	// set role of self peer
	ccinfo := p2ps.consacc.ConsensusInfo()
	if ccinfo.Type == "raft" {
		if !p2ps.useRaft {
			panic("configuration failure. consensus type of genesis block and consensus accessor are differ")
		}
	}
}

// BeforeStop is called before actor hub stops. it finishes underlying peer manager
func (p2ps *P2P) BeforeStop() {
	p2ps.Logger.Debug().Msg("stopping p2p actor.")
	p2ps.prm.Stop()
	p2ps.cm.Stop()
	p2ps.mm.Stop()
	if err := p2ps.pm.Stop(); err != nil {
		p2ps.Logger.Warn().Err(err).Msg("Error on stopping peerManager")
	}
	p2ps.mutex.Lock()
	nt := p2ps.nt
	p2ps.mutex.Unlock()
	nt.Stop()
	p2ps.lm.Stop()
}

// Statistics show statistic information of p2p module. NOTE: It it not implemented yet
func (p2ps *P2P) Statistics() *map[string]interface{} {
	stmap := make(map[string]interface{})
	stmap["netstat"] = p2ps.mm.Summary()
	stmap["config"] = p2ps.cfg.P2P
	stmap["status"] = p2ps.selfMeta
	wlSummary := p2ps.lm.Summary()
	stmap["whitelist"] = wlSummary["whitelist"]
	stmap["whitelist_on"] = wlSummary["whitelist_on"]
	stmap["syncman"] = p2ps.sm.Summary()

	return &stmap
}

func (p2ps *P2P) GetNetworkTransport() p2pcommon.NetworkTransport {
	p2ps.mutex.Lock()
	defer p2ps.mutex.Unlock()
	return p2ps.nt
}

func (p2ps *P2P) GetPeerAccessor() p2pcommon.PeerAccessor {
	return p2ps
}

func (p2ps *P2P) SetConsensusAccessor(ca consensus.ConsensusAccessor) {
	p2ps.consacc = ca
}

func (p2ps *P2P) GenesisChainID() *types.ChainID {
	return p2ps.genesisChainID
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
		p2ps.NotifyNewTX(msg)
	case *message.AddBlockRsp:
		// do nothing for now. just for prevent deadletter

	case *message.GetSelf:
		context.Respond(p2ps.selfMeta)
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
		//clusterReceiver := raftsupport.NewClusterInfoReceiver(p2ps, p2ps.mf, peers, time.Second*5, msg)
		clusterReceiver := raftsupport.NewConcClusterInfoReceiver(p2ps, p2ps.mf, peers, time.Second*5, msg, p2ps.Logger)
		clusterReceiver.StartGet()
	case *message.SendRaft:
		p2ps.SendRaftMessage(context, msg)
	case *message.RaftClusterEvent:
		p2ps.Logger.Debug().Array("added", p2putil.NewLogPeerIdsMarshaller(msg.BPAdded, 10)).Array("removed", p2putil.NewLogPeerIdsMarshaller(msg.BPRemoved, 10)).Msg("bp changed")
		p2ps.prm.UpdateBP(msg.BPAdded, msg.BPRemoved)
	case message.GetRaftTransport:
		context.Respond(raftsupport.NewAergoRaftTransport(p2ps.Logger, p2ps.nt, p2ps.pm, p2ps.mf, p2ps.consacc, msg.Cluster))
	case message.P2PWhiteListConfEnableEvent:
		p2ps.Logger.Debug().Bool("enabled", msg.On).Msg("p2p whitelist on/off changed")
		// TODO do more fine grained work
		p2ps.lm.RefineList()
		// disconnect newly blacklisted peer.
		p2ps.checkAndBanInboundPeers()
	case message.P2PWhiteListConfSetEvent:
		p2ps.Logger.Debug().Array("entries", p2putil.NewLogStringsMarshaller(msg.Values, 10)).Msg("p2p whitelist entries changed")
		// TODO do more fine grained work
		p2ps.lm.RefineList()
		// disconnect newly blacklisted peer.
		p2ps.checkAndBanInboundPeers()
	case message.IssueAgentCertificate:
		p2ps.SendIssueCertMessage(context, msg)
	case message.NotifyCertRenewed:
		p2ps.NotifyCertRenewed(context, msg)
	case message.TossBPNotice:
		p2ps.TossBPNotice(msg)
	}
}

func (p2ps *P2P) checkAndBanInboundPeers() {
	for _, peer := range p2ps.pm.GetPeers() {
		ip := peer.RemoteInfo().Connection.IP
		// TODO temporal treatment. need more works.
		// just inbound peers will be disconnected
		if peer.RemoteInfo().Connection.Outbound {
			p2ps.Debug().Str(p2putil.LogPeerName, peer.Name()).Msg("outbound peer is not banned")
			continue
		}
		if banned, _ := p2ps.lm.IsBanned(ip.String(), peer.ID()); banned {
			p2ps.Info().Str(p2putil.LogPeerName, peer.Name()).Msg("peer is banned by list manager")
			peer.Stop()
		}
	}
}

// TODO need refactoring. this code is copied from subproto/addrs.go
func (p2ps *P2P) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
	selfPeerID := p2ps.SelfNodeID()
	peerMetas := make([]p2pcommon.PeerMeta, 0, len(peers))
	for _, rPeerAddr := range peers {
		rPeerID := types.PeerID(rPeerAddr.PeerID)
		if selfPeerID == rPeerID {
			continue
		}
		if network.CheckAddressType(rPeerAddr.Address) == network.AddressTypeError {
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

// GetChainAccessor implement interface method of InternalService
func (p2ps *P2P) GetChainAccessor() types.ChainAccessor {
	return p2ps.ca
}

// ConsensusAccessor implement interface method of InternalService
func (p2ps *P2P) ConsensusAccessor() consensus.ConsensusAccessor {
	return p2ps.consacc
}

func (p2ps *P2P) insertHandlers(peer p2pcommon.RemotePeer) {
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
	peer.AddMessageHandler(p2pcommon.GetBlockHeadersRequest, subproto.NewGetBlockHeadersReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetBlockHeadersResponse, subproto.NewGetBlockHeaderRespHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetAncestorRequest, subproto.NewGetAncestorReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetAncestorResponse, subproto.NewGetAncestorRespHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetHashesRequest, subproto.NewGetHashesReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetHashesResponse, subproto.NewGetHashesRespHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetHashByNoRequest, subproto.NewGetHashByNoReqHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.GetHashByNoResponse, subproto.NewGetHashByNoRespHandler(p2ps.pm, peer, logger, p2ps))

	// TxHandlers
	peer.AddMessageHandler(p2pcommon.GetTXsRequest, subproto.WithTimeLog(subproto.NewTxReqHandler(p2ps.pm, p2ps.sm, peer, logger, p2ps), p2ps.Logger, zerolog.DebugLevel))
	peer.AddMessageHandler(p2pcommon.GetTXsResponse, subproto.WithTimeLog(subproto.NewTxRespHandler(p2ps.pm, peer, logger, p2ps), p2ps.Logger, zerolog.DebugLevel))
	peer.AddMessageHandler(p2pcommon.NewTxNotice, subproto.WithTimeLog(subproto.NewNewTxNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm), p2ps.Logger, zerolog.DebugLevel))

	// block notice handlers
	if p2ps.useRaft && p2ps.selfMeta.Role == types.PeerRole_Producer {
		peer.AddMessageHandler(p2pcommon.BlockProducedNotice, subproto.NewBPNoticeDiscardHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm))
		peer.AddMessageHandler(p2pcommon.NewBlockNotice, subproto.NewBlkNoticeDiscardHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm))
	} else if p2ps.selfMeta.Role == types.PeerRole_Agent {
		peer.AddMessageHandler(p2pcommon.BlockProducedNotice, subproto.WithTimeLog(subproto.NewAgentBlockProducedNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm, p2ps.cm), p2ps.Logger, zerolog.DebugLevel))
		peer.AddMessageHandler(p2pcommon.NewBlockNotice, subproto.NewNewBlockNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm))
	} else {
		peer.AddMessageHandler(p2pcommon.BlockProducedNotice, subproto.WithTimeLog(subproto.NewBlockProducedNoticeHandler(p2ps, p2ps.pm, peer, logger, p2ps, p2ps.sm), p2ps.Logger, zerolog.DebugLevel))
		peer.AddMessageHandler(p2pcommon.NewBlockNotice, subproto.NewNewBlockNoticeHandler(p2ps.pm, peer, logger, p2ps, p2ps.sm))
	}

	// Raft support
	peer.AddMessageHandler(p2pcommon.GetClusterRequest, subproto.NewGetClusterReqHandler(p2ps.pm, peer, logger, p2ps, p2ps.consacc))
	peer.AddMessageHandler(p2pcommon.GetClusterResponse, subproto.NewGetClusterRespHandler(p2ps.pm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.RaftWrapperMessage, subproto.NewRaftWrapperHandler(p2ps.pm, peer, logger, p2ps, p2ps.consacc))

	// certificate
	peer.AddMessageHandler(p2pcommon.IssueCertificateRequest, subproto.NewIssueCertReqHandler(p2ps.pm, p2ps.cm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.IssueCertificateResponse, subproto.NewIssueCertRespHandler(p2ps.pm, p2ps.cm, peer, logger, p2ps))
	peer.AddMessageHandler(p2pcommon.CertificateRenewedNotice, subproto.NewCertRenewedNoticeHandler(p2ps.pm, p2ps.cm, peer, logger, p2ps))
}

func (p2ps *P2P) CreateHSHandler(outbound bool, pid types.PeerID) p2pcommon.HSHandler {
	if outbound {
		return NewOutboundHSHandler(p2ps.pm, p2ps, p2ps.vm, p2ps.Logger, p2ps.genesisChainID, pid)
	} else {
		return NewInboundHSHandler(p2ps.pm, p2ps, p2ps.vm, p2ps.Logger, p2ps.genesisChainID, pid)
	}
}

func (p2ps *P2P) CreateRemotePeer(remoteInfo p2pcommon.RemoteInfo, seq uint32, rw p2pcommon.MsgReadWriter) p2pcommon.RemotePeer {
	// local peer can refuse to accept claimed role by consensus
	if p2ps.prm.CheckRole(remoteInfo, remoteInfo.Meta.Role) {
		remoteInfo.AcceptedRole = remoteInfo.Meta.Role
	} else {
		remoteInfo.AcceptedRole = types.PeerRole_Watcher
	}

	newPeer := newRemotePeer(remoteInfo, seq, p2ps.pm, p2ps, p2ps.Logger, p2ps.mf, p2ps.signer, rw)
	rw.AddIOListener(p2ps.mm.NewMetric(newPeer.ID(), newPeer.ManageNumber()))

	// insert Handlers
	p2ps.insertHandlers(newPeer)

	return newPeer
}

type notifyNewTXs struct {
	ids         []types.TxID
	alreadySent []types.PeerID
}

func (p2ps *P2P) SelfMeta() p2pcommon.PeerMeta {
	return p2ps.selfMeta
}

func (p2ps *P2P) SelfNodeID() types.PeerID {
	return p2ps.selfMeta.ID
}

func (p2ps *P2P) LocalSettings() p2pcommon.LocalSettings {
	return p2ps.localSettings
}

func (p2ps *P2P) GetPeerBlockInfos() []types.PeerBlockInfo {
	return p2ps.pm.GetPeerBlockInfos()
}

func (p2ps *P2P) GetPeer(ID types.PeerID) (p2pcommon.RemotePeer, bool) {
	return p2ps.pm.GetPeer(ID)
}

func (p2ps *P2P) PeerManager() p2pcommon.PeerManager {
	return p2ps.pm
}

func (p2ps *P2P) CertificateManager() p2pcommon.CertificateManager {
	return p2ps.cm
}

func (p2ps *P2P) RoleManager() p2pcommon.PeerRoleManager {
	return p2ps.prm
}

func (p2ps *P2P) initLocalSettings(conf *config.P2PConfig) {
	meta := p2ps.selfMeta
	switch meta.Role {
	case types.PeerRole_Producer:
		// set agent id
		if len(conf.Agent) > 0 {
			pid, err := types.IDB58Decode(conf.Agent)
			if err != nil {
				panic("invalid agentID " + conf.Agent + " : " + err.Error())
			}
			p2ps.Logger.Info().Str("fullID", pid.String()).Str("agentID", p2putil.ShortForm(pid)).Msg("found agent setting. use peer as agent if connected")
			p2ps.localSettings.AgentID = pid
		} else {
			p2ps.Logger.Debug().Msg("no agent was set. local peer is standalone producer.")
		}
	case types.PeerRole_Agent:
		// set internal zone for agent
		if len(conf.InternalZones) > 0 {
			nets := make([]*net.IPNet, len(conf.InternalZones))
			for i, z := range conf.InternalZones {
				_, ipnet, err := net.ParseCIDR(z)
				if err != nil {
					panic("invalid address range " + z + " : " + err.Error())
				}
				nets[i] = ipnet
			}
			p2ps.Logger.Info().Array("producerIDs", p2putil.NewLogPeerIdsMarshaller(meta.ProducerIDs, 25)).Array("internalZones", p2putil.NewLogIPNetMarshaller(nets, 10)).Msg("init agent setting. use peer as agent if connected")
			p2ps.localSettings.InternalZones = nets
		} else {
			panic("agent must configure one or more internalzones ")
		}
	default:
		// do nothing for now
	}

}
