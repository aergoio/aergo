/* @file @copyright defined in aergo/LICENSE.txt */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pkey"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/metric"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"

	cfg "github.com/aergoio/aergo/config"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	initial  = iota
	running  = iota
	stopping = iota
	stopped  = iota
)

/**
 * peerManager connect to and listen from other nodes.
 * It implements  Component interface
 */
type peerManager struct {
	status            int32
	nt                p2pcommon.NetworkTransport
	hsFactory         p2pcommon.HSHandlerFactory
	actorService      p2pcommon.ActorService
	peerFactory       p2pcommon.PeerFactory
	mf                p2pcommon.MoFactory
	mm                metric.MetricsManager
	skipHandshakeSync bool

	peerFinder p2pcommon.PeerFinder
	wpManager  p2pcommon.WaitingPeerManager
	// designatedPeers and hiddenPeerSet is set in construction time once and will not be changed
	hiddenPeerSet map[types.PeerID]bool

	mutex        *sync.Mutex
	manageNumber uint32
	remotePeers  map[types.PeerID]p2pcommon.RemotePeer
	waitingPeers map[types.PeerID]*p2pcommon.WaitingPeer

	conf *cfg.P2PConfig
	// peerCache is copy-on-write style
	peerCache []p2pcommon.RemotePeer

	getPeerChannel    chan getPeerChan
	peerHandshaked    chan handshakeResult
	removePeerChannel chan p2pcommon.RemotePeer
	fillPoolChannel   chan []p2pcommon.PeerMeta
	inboundConnChan   chan inboundConnEvent
	workDoneChannel   chan p2pcommon.ConnWorkResult
	taskChannel       chan pmTask
	finishChannel     chan struct{}

	eventListeners []p2pcommon.PeerEventListener

	//
	designatedPeers map[types.PeerID]p2pcommon.PeerMeta

	logger *log.Logger
}

// getPeerChan is struct to get peer for concurrent use
type getPeerChan struct {
	id  types.PeerID
	ret chan p2pcommon.RemotePeer
}

var _ p2pcommon.PeerManager = (*peerManager)(nil)


// NewPeerManager creates a peer manager object.
func NewPeerManager(hsFactory p2pcommon.HSHandlerFactory, actor p2pcommon.ActorService, cfg *cfg.Config, pf p2pcommon.PeerFactory, nt p2pcommon.NetworkTransport, mm metric.MetricsManager, logger *log.Logger, mf p2pcommon.MoFactory, skipHandshakeSync bool) p2pcommon.PeerManager {
	p2pConf := cfg.P2P
	//logger.SetLevel("debug")
	pm := &peerManager{
		nt:                nt,
		hsFactory:         hsFactory,
		actorService:      actor,
		conf:              p2pConf,
		peerFactory:       pf,
		mf:                mf,
		mm:                mm,
		logger:            logger,
		mutex:             &sync.Mutex{},
		skipHandshakeSync: skipHandshakeSync,

		status:          initial,
		designatedPeers: make(map[types.PeerID]p2pcommon.PeerMeta, len(cfg.P2P.NPAddPeers)),
		hiddenPeerSet:   make(map[types.PeerID]bool, len(cfg.P2P.NPHiddenPeers)),

		remotePeers: make(map[types.PeerID]p2pcommon.RemotePeer, p2pConf.NPMaxPeers),

		waitingPeers: make(map[types.PeerID]*p2pcommon.WaitingPeer, p2pConf.NPPeerPool),

		peerCache: make([]p2pcommon.RemotePeer, 0, p2pConf.NPMaxPeers),

		getPeerChannel:    make(chan getPeerChan),
		peerHandshaked:    make(chan handshakeResult),
		removePeerChannel: make(chan p2pcommon.RemotePeer),
		fillPoolChannel:   make(chan []p2pcommon.PeerMeta, 2),
		inboundConnChan:   make(chan inboundConnEvent),
		workDoneChannel:   make(chan p2pcommon.ConnWorkResult),
		eventListeners:    make([]p2pcommon.PeerEventListener, 0, 4),
		taskChannel:       make(chan pmTask, 4),
		finishChannel:     make(chan struct{}),
	}

	// additional initializations
	pm.init()

	return pm
}

func (pm *peerManager) SelfMeta() p2pcommon.PeerMeta {
	return pm.nt.SelfMeta()
}
func (pm *peerManager) SelfNodeID() types.PeerID {
	return p2pkey.NodeID()
}

func (pm *peerManager) init() {
	// set designated peers
	pm.initDesignatedPeerList()
	// init hidden peers
	for _, pidStr := range pm.conf.NPHiddenPeers {
		pid, err := types.IDB58Decode(pidStr)
		if err != nil {
			panic("Invalid pid in NPHiddenPeers : " + pidStr + " err " + err.Error())
		}
		pm.hiddenPeerSet[pid] = true
	}

	pm.peerFinder = NewPeerFinder(pm.logger, pm, pm.actorService, pm.conf.NPPeerPool, pm.conf.NPDiscoverPeers, pm.conf.NPUsePolaris)
	pm.wpManager = NewWaitingPeerManager(pm.logger, pm, pm.actorService, pm.conf.NPPeerPool, pm.conf.NPDiscoverPeers, pm.conf.NPUsePolaris)
	// add designated peers to waiting pool at initial time.
	for _, meta := range pm.designatedPeers {
		if _, foundInWait := pm.waitingPeers[meta.ID]; !foundInWait {
			pm.waitingPeers[meta.ID] = &p2pcommon.WaitingPeer{Meta: meta, NextTrial: time.Now()}
		}
	}
}

func (pm *peerManager) AddPeerEventListener(l p2pcommon.PeerEventListener) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.eventListeners = append(pm.eventListeners, l)
}

func (pm *peerManager) Start() error {
	go pm.runManagePeers()

	return nil
}

func (pm *peerManager) Stop() error {
	if atomic.CompareAndSwapInt32(&pm.status, running, stopping) {
		pm.finishChannel <- struct{}{}
	} else {
		// leave stopped if already stopped
		if atomic.SwapInt32(&pm.status, stopping) == stopped {
			atomic.StoreInt32(&pm.status, stopped)
		}
	}
	return nil
}

func (pm *peerManager) initDesignatedPeerList() {
	// add remote node from config
	for _, target := range pm.conf.NPAddPeers {
		peerMeta, err := p2putil.ParseMultiAddrString(target)
		if err != nil {
			pm.logger.Warn().Err(err).Str("str", target).Msg("invalid NPAddPeer address")
			continue
		}
		peerMeta.Designated = true
		peerMeta.Outbound = true
		pm.logger.Info().Str(p2putil.LogFullID, peerMeta.ID.Pretty()).Str(p2putil.LogPeerID, p2putil.ShortForm(peerMeta.ID)).Str("addr", peerMeta.IPAddress).Uint32("port", peerMeta.Port).Msg("Adding Designated peer")
		pm.designatedPeers[peerMeta.ID] = peerMeta
	}
}

func (pm *peerManager) runManagePeers() {

	pm.logger.Info().Str("p2p_proto", p2putil.ProtocolIDsToString([]protocol.ID{p2pcommon.P2PSubAddr, p2pcommon.LegacyP2PSubAddr})).Msg("Starting p2p listening")
	pm.nt.AddStreamHandler(p2pcommon.LegacyP2PSubAddr, pm.wpManager.OnInboundConnLegacy)
	pm.nt.AddStreamHandler(p2pcommon.P2PSubAddr, pm.wpManager.OnInboundConn)

	if !atomic.CompareAndSwapInt32(&pm.status, initial, running) {
		panic("wrong internal status")
	}
	instantStart := time.Millisecond << 4
	initialAddrDelay := time.Second * 2
	finderTimer := time.NewTimer(initialAddrDelay)
	connManTimer := time.NewTimer(initialAddrDelay << 1)

MANLOOP:
	for {
		select {
		case req := <-pm.getPeerChannel:
			peer, exist := pm.remotePeers[req.id]
			if exist {
				req.ret <- peer
			} else {
				req.ret <- nil
			}
		case hsreslt := <-pm.peerHandshaked:
			if peer := pm.tryRegister(hsreslt); peer != nil {
				pm.peerFinder.OnPeerConnect(peer.ID())
				pm.wpManager.OnPeerConnect(peer.ID())

				pm.checkSync(peer)

				// query other peers
				if !finderTimer.Stop() {
					<-finderTimer.C
				}
				finderTimer.Reset(instantStart)
			}
		case peer := <-pm.removePeerChannel:
			if pm.removePeer(peer) {
				pm.peerFinder.OnPeerDisconnect(peer)
				pm.wpManager.OnPeerDisconnect(peer)
			}
			if !connManTimer.Stop() {
				<-connManTimer.C
			}
			connManTimer.Reset(instantStart)
		case inInfo := <-pm.inboundConnChan:
			id := inInfo.meta.ID
			if _, found := pm.remotePeers[id]; found {
				inInfo.foundC <- true
			} else {
				inInfo.foundC <- false
			}
		case workResult := <-pm.workDoneChannel:
			pm.wpManager.OnWorkDone(workResult)
			// Retry
			if !connManTimer.Stop() {
				<-connManTimer.C
			}
			connManTimer.Reset(instantStart)
		case <-finderTimer.C:
			pm.peerFinder.CheckAndFill()
			finderTimer.Reset(DiscoveryQueryInterval)
		case <-connManTimer.C:
			pm.wpManager.CheckAndConnect()
			// fire at next interval
			connManTimer.Reset(p2pcommon.WaitingPeerManagerInterval)
		case peerMetas := <-pm.fillPoolChannel:
			if pm.wpManager.OnDiscoveredPeers(peerMetas) > 0 {
				if !connManTimer.Stop() {
					<-connManTimer.C
				}
				connManTimer.Reset(instantStart)
			}
		case task := <-pm.taskChannel:
			task.task()
		case <-pm.finishChannel:
			finderTimer.Stop()
			connManTimer.Stop()
			break MANLOOP
		}
	}
	// guaranty no new peer connection will be made
	pm.nt.RemoveStreamHandler(p2pcommon.LegacyP2PSubAddr)
	pm.nt.RemoveStreamHandler(p2pcommon.P2PSubAddr)

	pm.logger.Info().Msg("Finishing peerManager")

	go func() {
		// closing all peer connections
		for _, peer := range pm.peerCache {
			peer.Stop()
		}
	}()
	timer := time.NewTimer(time.Second * 30)
	finishPoll := time.NewTicker(time.Millisecond << 6)
CLEANUPLOOP:
	for {
		select {
		case req := <-pm.getPeerChannel:
			req.ret <- nil
		case peer := <-pm.removePeerChannel:
			pm.removePeer(peer)
		case <-finishPoll.C:
			if len(pm.remotePeers) == 0 {
				pm.logger.Debug().Msg("All peers were finished peerManager")
				break CLEANUPLOOP
			}
		case <-timer.C:
			pm.logger.Warn().Int("remained", len(pm.peerCache)).Msg("peermanager stop timeout. some peers were not finished.")
			break CLEANUPLOOP
		}
	}
	atomic.StoreInt32(&pm.status, stopped)
}

// tryRegister register peer to peer manager, if peer with same peer
func (pm *peerManager) tryRegister(hsresult handshakeResult) p2pcommon.RemotePeer {
	meta := hsresult.meta
	peerID := meta.ID
	preExistPeer, ok := pm.remotePeers[peerID]
	if ok {
		pm.logger.Info().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("Peer add collision. Outbound connection of higher hash will survive.")
		iAmLowerOrEqual := p2putil.ComparePeerID(pm.SelfNodeID(), meta.ID) <= 0
		if iAmLowerOrEqual == meta.Outbound {
			pm.logger.Info().Str("local_peer_id", p2putil.ShortForm(pm.SelfNodeID())).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Bool("outbound", meta.Outbound).Msg("Close connection and keep earlier handshake connection.")
			return nil
		} else {
			pm.logger.Info().Str("local_peer_id", p2putil.ShortForm(pm.SelfNodeID())).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Bool("outbound", meta.Outbound).Msg("Keep connection and close earlier handshake connection.")
			// stopping lower valued connection
			preExistPeer.Stop()
		}
	}

	meta = pm.changePeerAttributes(meta, peerID)
	newPeer := pm.peerFactory.CreateRemotePeer(meta, pm.GetNextManageNum(), hsresult.status, hsresult.s, hsresult.msgRW)

	go newPeer.RunPeer()

	pm.insertPeer(peerID, newPeer)
	pm.logger.Info().Str("role", newPeer.Role().String()).Bool("outbound", meta.Outbound).Str(p2putil.LogPeerName, newPeer.Name()).Str("addr", net.ParseIP(meta.IPAddress).String()+":"+strconv.Itoa(int(meta.Port))).Msg("peer is added to peerService")

	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	for _, listener := range pm.eventListeners {
		listener.OnPeerConnect(peerID)
	}

	return newPeer
}

func (pm *peerManager) changePeerAttributes(meta p2pcommon.PeerMeta, peerID types.PeerID) (p2pcommon.PeerMeta) {
	// override options by configurations of node
	_, meta.Designated = pm.designatedPeers[peerID]
	// hidden is set by either remote peer's asking or local node's config
	if _, exist := pm.hiddenPeerSet[peerID]; exist {
		meta.Hidden = true
	}

	return meta
}

func (pm *peerManager) GetNextManageNum() uint32 {
	return atomic.AddUint32(&pm.manageNumber, 1)
}

func (pm *peerManager) AddNewPeer(meta p2pcommon.PeerMeta) {
	sli := []p2pcommon.PeerMeta{meta}
	pm.fillPoolChannel <- sli
}

func (pm *peerManager) RemovePeer(peer p2pcommon.RemotePeer) {
	pm.removePeerChannel <- peer
}

func (pm *peerManager) NotifyPeerAddressReceived(metas []p2pcommon.PeerMeta) {
	pm.fillPoolChannel <- metas
}

func (pm *peerManager) UpdatePeerRole(changes []p2pcommon.AttrModifier) {
	pm.taskChannel <- pmTask{pm: pm, task: func() {
		pm.logger.Debug().Int("size", len(changes)).Msg("changing roles of peers")
		for _, ch := range changes {
			if peer, found := pm.remotePeers[ch.ID]; found {
				pm.logger.Debug().Str(p2putil.LogPeerName, peer.Name()).Str("from", peer.Role().String()).Str("to", ch.Role.String()).Msg("changing role of peer")
				peer.ChangeRole(ch.Role)
			}
		}
	}}
}

// removePeer unregister managed remote peer connection
// It return true if peer is exist and managed by peermanager
// it must called in peermanager goroutine
func (pm *peerManager) removePeer(peer p2pcommon.RemotePeer) bool {
	peerID := peer.ID()
	target, ok := pm.remotePeers[peerID]
	if !ok {
		return false
	}
	if target.ManageNumber() != peer.ManageNumber() {
		pm.logger.Debug().Uint32("remove_num", peer.ManageNumber()).Uint32("exist_num", target.ManageNumber()).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("remove peer is requested but already removed and other instance is on")
		return false
	}
	if target.State() == types.RUNNING {
		pm.logger.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("remove peer is requested but peer is still running")
	}
	pm.deletePeer(peer)
	pm.logger.Info().Uint32("manage_num", peer.ManageNumber()).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("removed peer in peermanager")

	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	for _, listener := range pm.eventListeners {
		listener.OnPeerDisconnect(peer)
	}

	return true
}

func (pm *peerManager) GetPeer(ID types.PeerID) (p2pcommon.RemotePeer, bool) {

	gc := getPeerChan{id: ID, ret: make(chan p2pcommon.RemotePeer)}
	// vs code's lint does not allow direct return of map operation
	pm.getPeerChannel <- gc
	ptr := <-gc.ret
	if ptr == nil {
		return nil, false
	}
	return ptr, true
}

func (pm *peerManager) GetPeers() []p2pcommon.RemotePeer {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	return pm.peerCache
}

func (pm *peerManager) GetPeerBlockInfos() []types.PeerBlockInfo {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	infos := make([]types.PeerBlockInfo, len(pm.peerCache))
	for i, peer := range pm.peerCache {
		infos[i] = peer
	}
	return infos
}

func (pm *peerManager) GetPeerAddresses(noHidden bool, showSelf bool) []*message.PeerInfo {
	peers := make([]*message.PeerInfo, 0, len(pm.peerCache))
	if showSelf {
		meta := pm.SelfMeta()
		addr := meta.ToPeerAddress()
		bestBlk, err := pm.actorService.GetChainAccessor().GetBestBlock()
		if err != nil {
			return nil
		}
		selfpi := &message.PeerInfo{
			&addr, meta.Version, meta.Hidden, time.Now(), bestBlk.BlockHash(), bestBlk.Header.BlockNo, types.RUNNING, true}
		peers = append(peers, selfpi)
	}
	for _, aPeer := range pm.peerCache {
		meta := aPeer.Meta()
		if noHidden && meta.Hidden {
			continue
		}
		addr := meta.ToPeerAddress()
		lastStatus := aPeer.LastStatus()
		pi := &message.PeerInfo{
			&addr, meta.Version, meta.Hidden, lastStatus.CheckTime, lastStatus.BlockHash, lastStatus.BlockNumber, aPeer.State(), false}
		peers = append(peers, pi)
	}
	return peers
}

// this method should be called inside pm.mutex
func (pm *peerManager) insertPeer(ID types.PeerID, peer p2pcommon.RemotePeer) {
	pm.remotePeers[ID] = peer
	pm.updatePeerCache()
}

// this method should be called inside pm.mutex
func (pm *peerManager) deletePeer(peer p2pcommon.RemotePeer) {
	pm.mm.Remove(peer.ID(), peer.ManageNumber())
	delete(pm.remotePeers, peer.ID())
	pm.updatePeerCache()
}

func (pm *peerManager) updatePeerCache() {
	newSlice := make([]p2pcommon.RemotePeer, 0, len(pm.remotePeers))
	for _, rPeer := range pm.remotePeers {
		newSlice = append(newSlice, rPeer)
	}
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.peerCache = newSlice
}

func (pm *peerManager) checkSync(peer p2pcommon.RemotePeer) {
	if pm.skipHandshakeSync {
		return
	}

	pm.logger.Debug().Uint64("target", peer.LastStatus().BlockNumber).Msg("request new syncer")
	pm.actorService.SendRequest(message.SyncerSvc, &message.SyncStart{PeerID: peer.ID(), TargetNo: peer.LastStatus().BlockNumber})
}

type pmTask struct {
	pm   *peerManager
	task func()
}
