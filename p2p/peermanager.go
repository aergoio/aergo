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
	status         int32
	nt             p2pcommon.NetworkTransport
	hsFactory      p2pcommon.HSHandlerFactory
	handlerFactory p2pcommon.HandlerFactory
	actorService   p2pcommon.ActorService
	signer         p2pcommon.MsgSigner
	mf             p2pcommon.MoFactory
	mm             metric.MetricsManager
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
	peerHandshaked    chan p2pcommon.RemotePeer
	removePeerChannel chan p2pcommon.RemotePeer
	fillPoolChannel   chan []p2pcommon.PeerMeta
	inboundConnChan   chan inboundConnEvent
	workDoneChannel   chan p2pcommon.ConnWorkResult

	finishChannel chan struct{}

	eventListeners []PeerEventListener

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

// PeerEventListener listen peer manage event
type PeerEventListener interface {
	// OnAddPeer is called just after the peer is added.
	OnAddPeer(peerID types.PeerID)

	// OnRemovePeer is called just before the peer is removed
	OnRemovePeer(peerID types.PeerID)
}

// NewPeerManager creates a peer manager object.
func NewPeerManager(handlerFactory p2pcommon.HandlerFactory, hsFactory p2pcommon.HSHandlerFactory, iServ p2pcommon.ActorService, cfg *cfg.Config, signer p2pcommon.MsgSigner, nt p2pcommon.NetworkTransport, mm metric.MetricsManager, logger *log.Logger, mf p2pcommon.MoFactory, skipHandshakeSync bool) p2pcommon.PeerManager {
	p2pConf := cfg.P2P
	//logger.SetLevel("debug")
	pm := &peerManager{
		nt:             nt,
		handlerFactory: handlerFactory,
		hsFactory:      hsFactory,
		actorService:   iServ,
		conf:           p2pConf,
		signer:         signer,
		mf:             mf,
		mm:             mm,
		logger:         logger,
		mutex:          &sync.Mutex{},
		skipHandshakeSync: skipHandshakeSync,

		status:          initial,
		designatedPeers: make(map[types.PeerID]p2pcommon.PeerMeta, len(cfg.P2P.NPAddPeers)),
		hiddenPeerSet:   make(map[types.PeerID]bool, len(cfg.P2P.NPHiddenPeers)),

		remotePeers: make(map[types.PeerID]p2pcommon.RemotePeer, p2pConf.NPMaxPeers),

		waitingPeers: make(map[types.PeerID]*p2pcommon.WaitingPeer, p2pConf.NPPeerPool),

		peerCache: make([]p2pcommon.RemotePeer, 0, p2pConf.NPMaxPeers),

		getPeerChannel:    make(chan getPeerChan),
		peerHandshaked:    make(chan p2pcommon.RemotePeer),
		removePeerChannel: make(chan p2pcommon.RemotePeer),
		fillPoolChannel:   make(chan []p2pcommon.PeerMeta, 2),
		inboundConnChan:   make(chan inboundConnEvent),
		workDoneChannel:   make(chan p2pcommon.ConnWorkResult),
		eventListeners:    make([]PeerEventListener, 0, 4),
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
		case peer := <-pm.peerHandshaked:
			if pm.tryRegister(peer) {
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
		case <-pm.finishChannel:
			finderTimer.Stop()
			connManTimer.Stop()
			break MANLOOP
		}
	}
	// guarrenty no new peer connection will be made
	pm.nt.RemoveStreamHandler(p2pcommon.LegacyP2PSubAddr)
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
func (pm *peerManager) tryRegister(peer p2pcommon.RemotePeer) bool {
	peerID := peer.ID()
	receivedMeta := peer.Meta()
	preExistPeer, ok := pm.remotePeers[peerID]
	if ok {
		pm.logger.Info().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("Peer add collision. Outbound connection of higher hash will survive.")
		iAmLowerOrEqual := p2putil.ComparePeerID(pm.SelfNodeID(), receivedMeta.ID) <= 0
		if iAmLowerOrEqual == receivedMeta.Outbound {
			pm.logger.Info().Str("local_peer_id", p2putil.ShortForm(pm.SelfNodeID())).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Bool("outbound", receivedMeta.Outbound).Msg("Close connection and keep earlier handshake connection.")
			return false
		} else {
			pm.logger.Info().Str("local_peer_id", p2putil.ShortForm(pm.SelfNodeID())).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Bool("outbound", receivedMeta.Outbound).Msg("Keep connection and close earlier handshake connection.")
			// stopping lower valued connection
			preExistPeer.Stop()
		}
	}

	go peer.RunPeer()
	// FIXME type casting is worse
	pm.insertPeer(peerID, peer)
	pm.logger.Info().Bool("outbound", receivedMeta.Outbound).Str(p2putil.LogPeerName, peer.Name()).Str("addr", net.ParseIP(receivedMeta.IPAddress).String()+":"+strconv.Itoa(int(receivedMeta.Port))).Msg("peer is added to peerService")

	// TODO add triggering sync.

	return true
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
	pm.deletePeer(peerID)
	pm.logger.Info().Uint32("manage_num", peer.ManageNumber()).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("removed peer in peermanager")
	for _, listener := range pm.eventListeners {
		listener.OnRemovePeer(peerID)
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
		lastNoti := aPeer.LastStatus()
		pi := &message.PeerInfo{
			&addr, meta.Version, meta.Hidden, lastNoti.CheckTime, lastNoti.BlockHash, lastNoti.BlockNumber, aPeer.State(), false}
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
func (pm *peerManager) deletePeer(ID types.PeerID) {
	pm.mm.Remove(ID)
	delete(pm.remotePeers, ID)
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
