/* @file @copyright defined in aergo/LICENSE.txt */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pkey"
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
	initial = iota
	running
	stopping
	stopped
)

/**
 * peerManager connect to and listen from other nodes.
 * It implements  Component interface
 */
type peerManager struct {
	status            int32
	is                p2pcommon.InternalService
	nt                p2pcommon.NetworkTransport
	hsFactory         p2pcommon.HSHandlerFactory
	actorService      p2pcommon.ActorService
	peerFactory       p2pcommon.PeerFactory
	mm                metric.MetricsManager
	lm                p2pcommon.ListManager
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

	getPeerChannel    chan getPeerTask
	peerHandshaked    chan connPeerResult
	removePeerChannel chan p2pcommon.RemotePeer
	fillPoolChannel   chan []p2pcommon.PeerMeta
	addPeerChannel    chan p2pcommon.PeerMeta
	inboundConnChan   chan inboundConnEvent
	workDoneChannel   chan p2pcommon.ConnWorkResult
	taskChannel       chan pmTask
	finishChannel     chan struct{}

	eventListeners []p2pcommon.PeerEventListener

	//
	designatedPeers map[types.PeerID]p2pcommon.PeerMeta

	logger *log.Logger
}

// getPeerTask is struct to get peer for concurrent use
type getPeerTask struct {
	id  types.PeerID
	ret chan p2pcommon.RemotePeer
}

var _ p2pcommon.PeerManager = (*peerManager)(nil)

// NewPeerManager creates a peer manager object.
func NewPeerManager(is p2pcommon.InternalService, hsFactory p2pcommon.HSHandlerFactory, actor p2pcommon.ActorService, pf p2pcommon.PeerFactory, nt p2pcommon.NetworkTransport, mm metric.MetricsManager, lm p2pcommon.ListManager, logger *log.Logger, cfg *cfg.Config, skipHandshakeSync bool) p2pcommon.PeerManager {
	p2pConf := cfg.P2P
	//logger.SetLevel("debug")
	pm := &peerManager{
		is:                is,
		nt:                nt,
		hsFactory:         hsFactory,
		actorService:      actor,
		conf:              p2pConf,
		peerFactory:       pf,
		mm:                mm,
		lm:                lm,
		logger:            logger,
		mutex:             &sync.Mutex{},
		skipHandshakeSync: skipHandshakeSync,

		status:          initial,
		designatedPeers: make(map[types.PeerID]p2pcommon.PeerMeta, len(cfg.P2P.NPAddPeers)),
		hiddenPeerSet:   make(map[types.PeerID]bool, len(cfg.P2P.NPHiddenPeers)),

		remotePeers: make(map[types.PeerID]p2pcommon.RemotePeer, p2pConf.NPMaxPeers),

		waitingPeers: make(map[types.PeerID]*p2pcommon.WaitingPeer, p2pConf.NPPeerPool),

		peerCache: make([]p2pcommon.RemotePeer, 0, p2pConf.NPMaxPeers),

		getPeerChannel:    make(chan getPeerTask),
		peerHandshaked:    make(chan connPeerResult),
		removePeerChannel: make(chan p2pcommon.RemotePeer),
		fillPoolChannel:   make(chan []p2pcommon.PeerMeta, 2),
		addPeerChannel:    make(chan p2pcommon.PeerMeta),
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
	pm.wpManager = NewWaitingPeerManager(pm.logger, pm, pm.lm, pm.conf.NPPeerPool, pm.conf.NPDiscoverPeers)
	// add designated peers to waiting pool at initial time.
	for _, meta := range pm.designatedPeers {
		if _, foundInWait := pm.waitingPeers[meta.ID]; !foundInWait {
			pm.waitingPeers[meta.ID] = &p2pcommon.WaitingPeer{Meta: meta, Designated: true, NextTrial: time.Now()}
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
	for _, addrStr := range pm.conf.NPAddPeers {
		ma, err := types.ParseMultiaddr(addrStr)
		if err != nil {
			pm.logger.Warn().Err(err).Str("str", addrStr).Msg("invalid NPAddPeer address")
			continue
		}
		peerMeta, err := p2putil.FromMultiAddrToPeerInfo(ma)
		if err != nil {
			pm.logger.Warn().Err(err).Str("str", addrStr).Msg("invalid NPAddPeer address")
			continue
		}
		pm.logger.Info().Str(p2putil.LogFullID, peerMeta.ID.Pretty()).Str(p2putil.LogPeerID, p2putil.ShortForm(peerMeta.ID)).Str("addr", peerMeta.Addresses[0].String()).Msg("Adding Designated peer")
		pm.designatedPeers[peerMeta.ID] = peerMeta
	}
}

func (pm *peerManager) runManagePeers() {

	pm.logger.Info().Str("p2p_proto", p2putil.ProtocolIDsToString([]protocol.ID{p2pcommon.P2PSubAddr})).Msg("Starting p2p listening")
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
			//connManTimer.Reset(time.Second*5)
		case peerMeta := <-pm.addPeerChannel:
			pm.wpManager.InstantConnect(peerMeta)
		case peerMetas := <-pm.fillPoolChannel:
			if pm.wpManager.OnDiscoveredPeers(peerMetas) > 0 {
				if !connManTimer.Stop() {
					<-connManTimer.C
				}
				connManTimer.Reset(instantStart)
			}
		case task := <-pm.taskChannel:
			task()
		case <-pm.finishChannel:
			finderTimer.Stop()
			connManTimer.Stop()
			break MANLOOP
		}
	}
	// guaranty no new peer connection will be made
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
func (pm *peerManager) tryRegister(hsResult connPeerResult) p2pcommon.RemotePeer {
	remote := hsResult.remote
	meta := remote.Meta
	peerID := meta.ID
	preExistPeer, ok := pm.remotePeers[peerID]
	if ok {
		pm.logger.Info().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("Peer add collision. Outbound connection of higher hash will survive.")
		iAmLowerOrEqual := p2putil.ComparePeerID(pm.is.SelfNodeID(), meta.ID) <= 0
		if iAmLowerOrEqual == remote.Connection.Outbound {
			pm.logger.Info().Str("local_peer_id", p2putil.ShortForm(pm.is.SelfNodeID())).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Bool("outbound", remote.Connection.Outbound).Msg("Close connection and keep earlier handshake connection.")
			return nil
		} else {
			pm.logger.Info().Str("local_peer_id", p2putil.ShortForm(pm.is.SelfNodeID())).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Bool("outbound", remote.Connection.Outbound).Msg("Keep connection and close earlier handshake connection.")
			// stopping lower valued connection
			preExistPeer.Stop()
		}
	}

	remote = pm.changePeerAttributes(remote, peerID)
	newPeer := pm.peerFactory.CreateRemotePeer(remote, pm.GetNextManageNum(), hsResult.msgRW)
	newPeer.UpdateBlkCache(hsResult.bestHash, hsResult.bestNo)

	go newPeer.RunPeer()

	pm.insertPeer(peerID, newPeer)
	pm.logger.Info().Str("role", newPeer.Role().String()).Bool("outbound", remote.Connection.Outbound).Str(p2putil.LogPeerName, newPeer.Name()).Str("addr", remote.Connection.IP.String()+":"+strconv.Itoa(int(remote.Connection.Port))).Msg("peer is added to peerService")

	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	for _, listener := range pm.eventListeners {
		listener.OnPeerConnect(peerID)
	}

	return newPeer
}

func (pm *peerManager) changePeerAttributes(remote p2pcommon.RemoteInfo, peerID types.PeerID) (p2pcommon.RemoteInfo) {
	// override options by configurations of node
	_, remote.Designated = pm.designatedPeers[peerID]
	// hidden is set by either remote peer's asking or local node's config
	if _, exist := pm.hiddenPeerSet[peerID]; exist {
		remote.Hidden = true
	}

	return remote
}

func (pm *peerManager) GetNextManageNum() uint32 {
	return atomic.AddUint32(&pm.manageNumber, 1)
}

func (pm *peerManager) AddNewPeer(meta p2pcommon.PeerMeta) {
	pm.addPeerChannel <- meta
}

func (pm *peerManager) RemovePeer(peer p2pcommon.RemotePeer) {
	pm.removePeerChannel <- peer
}

func (pm *peerManager) NotifyPeerAddressReceived(metas []p2pcommon.PeerMeta) {
	pm.fillPoolChannel <- metas
}

func (pm *peerManager) UpdatePeerRole(changes []p2pcommon.AttrModifier) {
	pm.taskChannel <- func() {
		pm.logger.Debug().Int("size", len(changes)).Msg("changing roles of peers")
		for _, ch := range changes {
			if peer, found := pm.remotePeers[ch.ID]; found {
				pm.logger.Debug().Str(p2putil.LogPeerName, peer.Name()).Str("from", peer.Role().String()).Str("to", ch.Role.String()).Msg("changing role of peer")
				peer.ChangeRole(ch.Role)
			}
		}
	}
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

	gc := getPeerTask{id: ID, ret: make(chan p2pcommon.RemotePeer)}
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
		meta := pm.is.SelfMeta()
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

func (pm *peerManager) AddDesignatedPeer(meta p2pcommon.PeerMeta) {
	finished := make(chan interface{})
	pm.taskChannel <- func() {
		pm.designatedPeers[meta.ID] = meta
		finished <- struct{}{}
	}
	<-finished
}

func (pm *peerManager) RemoveDesignatedPeer(peerID types.PeerID) {
	finished := make(chan interface{})
	pm.taskChannel <- func() {
		delete(pm.designatedPeers, peerID)
		finished <- struct{}{}
	}
	<-finished
}

func (pm *peerManager) ListDesignatedPeers() []p2pcommon.PeerMeta {
	retChan := make(chan []p2pcommon.PeerMeta)
	pm.taskChannel <- func() {
		arr := make([]p2pcommon.PeerMeta, 0, len(pm.designatedPeers))
		for _, m := range pm.designatedPeers {
			arr = append(arr, m)
		}
		retChan <- arr
	}
	return <-retChan
}

func (pm *peerManager) GetProducers() []types.PeerID {
	// TODO
	return nil
}
func (pm *peerManager) GetCertificates() []p2pcommon.AgentCertificateV1 {
	// TODO
	return nil
}

// pmTask should not consume lots of time to process.
type pmTask func()
