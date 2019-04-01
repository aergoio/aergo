/* @file @copyright defined in aergo/LICENSE.txt */

package p2p

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo/p2p/metric"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/p2p/subproto"

	inet "github.com/libp2p/go-libp2p-net"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"

	cfg "github.com/aergoio/aergo/config"
	peer "github.com/libp2p/go-libp2p-peer"
)

const (
	initial  = iota
	running  = iota
	stopping = iota
)

/**
 * peerManager connect to and listen from other nodes.
 * It implements  Component interface
 */
type peerManager struct {
	status         int32
	nt             p2pcommon.NetworkTransport
	hsFactory      HSHandlerFactory
	handlerFactory HandlerFactory
	actorService   p2pcommon.ActorService
	signer         p2pcommon.MsgSigner
	mf             p2pcommon.MoFactory
	mm             metric.MetricsManager

	// designatedPeers and hiddenPeerSet is set in construction time once and will not be changed
	hiddenPeerSet map[peer.ID]bool

	mutex        *sync.Mutex
	manageNumber uint32
	remotePeers  map[peer.ID]*remotePeerImpl
	peerPool     map[peer.ID]p2pcommon.PeerMeta
	conf         *cfg.P2PConfig
	// peerCache is copy-on-write style
	peerCache []p2pcommon.RemotePeer

	addPeerChannel  chan p2pcommon.PeerMeta
	fillPoolChannel chan []p2pcommon.PeerMeta
	finishChannel   chan struct{}
	eventListeners  []PeerEventListener

	//
	designatedPeers map[peer.ID]p2pcommon.PeerMeta
	awaitMutex      sync.RWMutex
	awaitPeers      map[peer.ID]*reconnectJob
	awaitDone       chan struct{}

	logger *log.Logger
}

var _ p2pcommon.PeerManager = (*peerManager)(nil)

// PeerEventListener listen peer manage event
type PeerEventListener interface {
	// OnAddPeer is called just after the peer is added.
	OnAddPeer(peerID peer.ID)

	// OnRemovePeer is called just before the peer is removed
	OnRemovePeer(peerID peer.ID)
}

// NewPeerManager creates a peer manager object.
func NewPeerManager(handlerFactory HandlerFactory, hsFactory HSHandlerFactory, iServ p2pcommon.ActorService, cfg *cfg.Config, signer p2pcommon.MsgSigner, nt p2pcommon.NetworkTransport, mm metric.MetricsManager, logger *log.Logger, mf p2pcommon.MoFactory) p2pcommon.PeerManager {
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

		status:          initial,
		designatedPeers: make(map[peer.ID]p2pcommon.PeerMeta, len(cfg.P2P.NPAddPeers)),
		hiddenPeerSet:   make(map[peer.ID]bool, len(cfg.P2P.NPHiddenPeers)),

		remotePeers: make(map[peer.ID]*remotePeerImpl, p2pConf.NPMaxPeers),

		awaitPeers: make(map[peer.ID]*reconnectJob, p2pConf.NPPeerPool),
		peerPool:   make(map[peer.ID]p2pcommon.PeerMeta, p2pConf.NPPeerPool),
		peerCache:  make([]p2pcommon.RemotePeer, 0, p2pConf.NPMaxPeers),
		awaitDone:  make(chan struct{}),

		addPeerChannel:  make(chan p2pcommon.PeerMeta, 2),
		fillPoolChannel: make(chan []p2pcommon.PeerMeta, 2),
		eventListeners:  make([]PeerEventListener, 0, 4),
		finishChannel:   make(chan struct{}),
	}

	// additional initializations
	pm.init()

	return pm
}

func (pm *peerManager) SelfMeta() p2pcommon.PeerMeta {
	return pm.nt.SelfMeta()
}
func (pm *peerManager) SelfNodeID() peer.ID {
	return pm.nt.ID()
}

func (pm *peerManager) RegisterEventListener(listener PeerEventListener) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.eventListeners = append(pm.eventListeners, listener)
}

func (pm *peerManager) init() {
	// set designated peers
	pm.initDesignatedPeerList()
	// init hidden peers
	for _, pidStr := range pm.conf.NPHiddenPeers {
		pid, err := peer.IDB58Decode(pidStr)
		if err != nil {
			panic("Invalid pid in NPHiddenPeers : " + pidStr + " err " + err.Error())
		}
		pm.hiddenPeerSet[pid] = true
	}
}

func (pm *peerManager) Start() error {

	go pm.runManagePeers()
	// need to start listen after chainservice is read to init
	// FIXME: adhoc code
	go func() {
		//time.Sleep(time.Second * 3)
		pm.nt.AddStreamHandler(p2pcommon.AergoP2PSub, pm.onConnect)
		pm.logger.Info().Str("version", string(p2pcommon.AergoP2PSub)).Msg("Starting p2p listening")

		// addition should start after all modules are started
		go func() {
			time.Sleep(time.Second * 2)
			for _, meta := range pm.designatedPeers {
				pm.addPeerChannel <- meta
			}
		}()
	}()

	if !atomic.CompareAndSwapInt32(&pm.status, initial, running) {
		panic("wrong internal status")
	}
	return nil
}

func (pm *peerManager) Stop() error {
	if !atomic.CompareAndSwapInt32(&pm.status, running, stopping) {
		pm.finishChannel <- struct{}{}
	}
	return nil
}

func (pm *peerManager) initDesignatedPeerList() {
	// add remote node from config
	for _, target := range pm.conf.NPAddPeers {
		peerMeta, err := ParseMultiAddrString(target)
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
	initialAddrDelay := time.Second * 20
	initialTimer := time.NewTimer(initialAddrDelay)
	addrTicker := time.NewTicker(DiscoveryQueryInterval)
MANLOOP:
	for {
		select {
		case meta := <-pm.addPeerChannel:
			if pm.addOutboundPeer(meta) {
				pm.cancelAwait(meta.ID)
			}
		case <-initialTimer.C:
			initialTimer.Stop()
			pm.checkAndCollectPeerListFromAll()
		case <-addrTicker.C:
			pm.checkAndCollectPeerListFromAll()
		case peerMetas := <-pm.fillPoolChannel:
			pm.tryFillPool(&peerMetas)
		case <-pm.finishChannel:
			addrTicker.Stop()
			break MANLOOP
		}
	}
	// guarrenty no new peer connection will be made
	pm.cancelAllAwait()
	pm.nt.RemoveStreamHandler(p2pcommon.AergoP2PSub)
	pm.logger.Info().Msg("Finishing peerManager")

	go func() {
		// closing all peer connections
		for _, peer := range pm.peerCache {
			peer.Stop()
		}
	}()
	timer := time.NewTimer(time.Second * 30)
	finishPoll := time.NewTicker(time.Second)
CLEANUPLOOP:
	for {
		select {
		case <-finishPoll.C:
			pm.mutex.Lock()
			if len(pm.remotePeers) == 0 {
				pm.mutex.Unlock()
				pm.logger.Debug().Msg("All peers were finished peerManager")
				break CLEANUPLOOP
			}
			pm.mutex.Unlock()
		case <-timer.C:
			pm.logger.Warn().Int("remained", len(pm.peerCache)).Msg("peermanager stop timeout. some peers were not finished.")
			break CLEANUPLOOP
		}
	}
}

// addOutboundPeer try to connect and handshake to remote peer. it can be called after peermanager is inited.
// It return true if peer is added or return false if failed to add peer or more suitable connection already exists.
func (pm *peerManager) addOutboundPeer(meta p2pcommon.PeerMeta) bool {
	s, err := pm.nt.GetOrCreateStream(meta, p2pcommon.AergoP2PSub)
	if err != nil {
		pm.logger.Info().Err(err).Str(p2putil.LogPeerID, p2putil.ShortForm(meta.ID)).Msg("Failed to get stream.")
		return false
	}

	completeMeta, added := pm.tryAddPeer(true, meta, s)
	if !added {
		s.Close()
		return false
	} else {
		if meta.IPAddress != completeMeta.IPAddress {
			pm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(completeMeta.ID)).Str("before", meta.IPAddress).Str("after", completeMeta.IPAddress).Msg("IP address of remote peer is changed to ")
		}
	}
	return true
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer. stream s will be owned to remotePeer if succeed to add perr.
func (pm *peerManager) tryAddPeer(outbound bool, meta p2pcommon.PeerMeta, s inet.Stream) (p2pcommon.PeerMeta, bool) {
	var peerID = meta.ID
	rd := metric.NewReader(s)
	wt := metric.NewWriter(s)
	h := pm.hsFactory.CreateHSHandler(outbound, pm, pm.actorService, pm.logger, peerID)
	rw, remoteStatus, err := h.Handle(rd, wt, defaultHandshakeTTL)
	if err != nil {
		pm.logger.Debug().Err(err).Str(p2putil.LogPeerID, p2putil.ShortForm(meta.ID)).Msg("Failed to handshake")
		if rw != nil {
			pm.sendGoAway(rw, err.Error())
		}
		return meta, false
	}
	// update peer meta info using sent information from remote peer
	receivedMeta := p2pcommon.NewMetaFromStatus(remoteStatus, outbound)
	if receivedMeta.ID != peerID {
		pm.logger.Debug().Str("received_peer_id", receivedMeta.ID.Pretty()).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("Inconsistent peerID")
		pm.sendGoAway(rw, "Inconsistent peerID")
		return meta, false
	}
	_, receivedMeta.Designated = pm.designatedPeers[peerID]

	// adding peer to peer list
	newPeer, err := pm.registerPeer(peerID, receivedMeta, remoteStatus, s, rw)
	if err != nil {
		pm.sendGoAway(rw, err.Error())
		return meta, false
	}
	newPeer.metric = pm.mm.Add(peerID, rd, wt)

	if pm.logger.IsDebugEnabled() {
		addrStrs := pm.nt.GetAddressesOfPeer(peerID)
		pm.logger.Debug().Strs("addrs", addrStrs).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("addresses of peer")
	}

	pm.doPostHandshake(peerID, remoteStatus)
	// notice to p2pmanager that handshaking is finished
	pm.NotifyPeerHandshake(peerID)

	return receivedMeta, true
}

func (pm *peerManager) registerPeer(peerID peer.ID, receivedMeta p2pcommon.PeerMeta, status *types.Status, s inet.Stream, rw p2pcommon.MsgReadWriter) (*remotePeerImpl, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	preExistPeer, ok := pm.remotePeers[peerID]
	if ok {
		pm.logger.Info().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("Peer add collision. Outbound connection of higher hash will survive.")
		iAmLower := p2putil.ComparePeerID(pm.SelfNodeID(), receivedMeta.ID) <= 0
		if iAmLower == receivedMeta.Outbound {
			pm.logger.Info().Str("local_peer_id", p2putil.ShortForm(pm.SelfNodeID())).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Bool("outbound", receivedMeta.Outbound).Msg("Close connection and keep earlier handshake connection.")
			return nil, fmt.Errorf("Already handshake peer %s ", p2putil.ShortForm(peerID))
		} else {
			pm.logger.Info().Str("local_peer_id", p2putil.ShortForm(pm.SelfNodeID())).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Bool("outbound", receivedMeta.Outbound).Msg("Keep connection and close earlier handshake connection.")
			// stopping lower valued connection
			preExistPeer.Stop()
		}
	}

	outboundPeer := newRemotePeer(receivedMeta, pm.GetNextManageNum(), pm, pm.actorService, pm.logger, pm.mf, pm.signer, s, rw)
	outboundPeer.UpdateBlkCache(status.GetBestBlockHash(), status.GetBestHeight())

	// insert Handlers
	pm.handlerFactory.insertHandlers(outboundPeer)

	go outboundPeer.RunPeer()
	pm.insertPeer(peerID, outboundPeer)
	pm.logger.Info().Bool("outbound", receivedMeta.Outbound).Str(p2putil.LogPeerName, outboundPeer.Name()).Str("addr", net.ParseIP(receivedMeta.IPAddress).String()+":"+strconv.Itoa(int(receivedMeta.Port))).Msg("peer is added to peerService")

	return outboundPeer, nil
}

// doPostHandshake is additional work after peer is added.
func (pm *peerManager) doPostHandshake(peerID peer.ID, remoteStatus *types.Status) {

	pm.logger.Debug().Uint64("target", remoteStatus.BestHeight).Msg("request new syncer")
	pm.actorService.SendRequest(message.SyncerSvc, &message.SyncStart{PeerID: peerID, TargetNo: remoteStatus.BestHeight})

	// sync mempool tx infos
	// TODO add tx handling
}

func (pm *peerManager) GetNextManageNum() uint32 {
	return atomic.AddUint32(&pm.manageNumber, 1)
}
func (pm *peerManager) sendGoAway(rw p2pcommon.MsgReadWriter, msg string) {
	goMsg := &types.GoAwayNotice{Message: msg}
	// TODO code smell. non safe casting.
	mo := pm.mf.NewMsgRequestOrder(false, subproto.GoAway, goMsg).(*pbRequestOrder)
	container := mo.message

	rw.WriteMsg(container)
}

func (pm *peerManager) AddNewPeer(peer p2pcommon.PeerMeta) {
	pm.addPeerChannel <- peer
}

func (pm *peerManager) RemovePeer(peer p2pcommon.RemotePeer) {
	pm.removePeer(peer)
}

func (pm *peerManager) NotifyPeerHandshake(peerID peer.ID) {
	// TODO code smell.
	if pm.conf.NPDiscoverPeers {
		pm.checkAndCollectPeerList(peerID)
	}
}

func (pm *peerManager) NotifyPeerAddressReceived(metas []p2pcommon.PeerMeta) {
	pm.fillPoolChannel <- metas
}

// removePeer unregister managed remote peer connection
// It return true if peer is exist and managed by peermanager
func (pm *peerManager) removePeer(peer p2pcommon.RemotePeer) bool {
	peerID := peer.ID()
	pm.mutex.Lock()
	target, ok := pm.remotePeers[peerID]
	if !ok {
		pm.mutex.Unlock()
		return false
	}
	if target.manageNum != peer.ManageNumber() {
		pm.logger.Debug().Uint32("remove_num", peer.ManageNumber()).Uint32("exist_num", target.ManageNumber()).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("remove peer is requested but already removed and other instance is on")
		pm.mutex.Unlock()
		return false
	}
	if target.State() == types.RUNNING {
		pm.logger.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("remove peer is requested but peer is still running")
	}
	pm.deletePeer(peerID)
	pm.logger.Info().Uint32("manage_num", peer.ManageNumber()).Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Msg("removed peer in peermanager")
	pm.mutex.Unlock()
	for _, listener := range pm.eventListeners {
		listener.OnRemovePeer(peerID)
	}

	if meta, found := pm.designatedPeers[peer.ID()]; found {
		pm.addAwait(meta)
	}
	return true
}

func (pm *peerManager) onConnect(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	tempMeta := p2pcommon.PeerMeta{ID: peerID}
	addr := s.Conn().RemoteMultiaddr()

	pm.logger.Debug().Str(p2putil.LogFullID, peerID.Pretty()).Str("multiaddr", addr.String()).Msg("new inbound peer arrived")
	completeMeta, added := pm.tryAddPeer(false, tempMeta, s)
	if !added {
		s.Close()
	} else {
		if tempMeta.IPAddress != completeMeta.IPAddress {
			pm.logger.Debug().Str("after", completeMeta.IPAddress).Msg("Update IP address of inbound remote peer")
		}
	}
}

func (pm *peerManager) checkAndCollectPeerListFromAll() {
	if pm.hasEnoughPeers() {
		return
	}
	if pm.conf.NPUsePolaris {
		pm.logger.Debug().Msg("Sending map query to polaris")
		pm.actorService.SendRequest(message.P2PSvc, &message.MapQueryMsg{Count: MaxAddrListSizePolaris})
	}

	// if server is not discover new peer, such as of BP or backup node, it does not send addresses reqeust to other peer.
	// These types are only connect to designated peers.
	if pm.conf.NPDiscoverPeers {
		// not strictly need to check peers. so use cache instead
		for _, remotePeer := range pm.peerCache {
			pm.actorService.SendRequest(message.P2PSvc, &message.GetAddressesMsg{ToWhom: remotePeer.ID(), Size: MaxAddrListSizePeer, Offset: 0})
		}
	}
}

func (pm *peerManager) checkAndCollectPeerList(ID peer.ID) {
	if pm.hasEnoughPeers() {
		return
	}
	rPeer, ok := pm.GetPeer(ID)
	if !ok {
		pm.logger.Warn().Str(p2putil.LogFullID, ID.Pretty()).Msg("invalid peer id")
		return
	}
	pm.actorService.SendRequest(message.P2PSvc, &message.GetAddressesMsg{ToWhom: rPeer.ID(), Size: 20, Offset: 0})
}

func (pm *peerManager) hasEnoughPeers() bool {
	return len(pm.peerPool) >= pm.conf.NPPeerPool
}

// tryConnectPeers should be called in runManagePeers() only
func (pm *peerManager) tryFillPool(metas *[]p2pcommon.PeerMeta) {
	added := make([]p2pcommon.PeerMeta, 0, len(*metas))
	invalid := make([]string, 0)
	for _, meta := range *metas {
		if string(meta.ID) == "" {
			invalid = append(invalid, p2putil.ShortMetaForm(meta))
			continue
		}
		_, found := pm.peerPool[meta.ID]
		if !found {
			// change some properties
			meta.Outbound = true
			meta.Designated = false
			pm.peerPool[meta.ID] = meta
			added = append(added, meta)
		}
	}
	if len(invalid) > 0 {
		pm.logger.Warn().Strs("metas", invalid).Msg("invalid meta list was come")
	}
	pm.logger.Debug().Int("added_cnt", len(added)).Msg("Filled unknown peer addresses to peerpool")
	pm.tryConnectPeers()
}

// tryConnectPeers should be called in runManagePeers() only
func (pm *peerManager) tryConnectPeers() {
	remained := pm.conf.NPMaxPeers - len(pm.remotePeers)
	for ID, meta := range pm.peerPool {
		if _, found := pm.GetPeer(ID); found {
			delete(pm.peerPool, ID)
			continue
		}
		if meta.IPAddress == "" || meta.Port == 0 {
			pm.logger.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(meta.ID)).Str("addr", meta.IPAddress).
				Uint32("port", meta.Port).Msg("Invalid peer meta informations")
			continue
		}
		// in same go rountine.
		pm.addOutboundPeer(meta)
		remained--
		if remained <= 0 {
			break
		}
	}
}

func (pm *peerManager) GetPeer(ID peer.ID) (p2pcommon.RemotePeer, bool) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// vs code's lint does not allow direct return of map operation
	ptr, ok := pm.remotePeers[ID]
	if !ok {
		return nil, false
	}
	return ptr, ok
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
			&addr, meta.Hidden, time.Now(), bestBlk.BlockHash(), bestBlk.Header.BlockNo, types.RUNNING, true}
		peers = append(peers, selfpi)
	}
	for _, aPeer := range pm.peerCache {
		meta := aPeer.Meta()
		if noHidden && meta.Hidden {
			continue
		}
		addr := meta.ToPeerAddress()
		lastNoti := aPeer.LastNotice()
		pi := &message.PeerInfo{
			&addr, meta.Hidden, lastNoti.CheckTime, lastNoti.BlockHash, lastNoti.BlockNumber, aPeer.State(), false}
		peers = append(peers, pi)
	}
	return peers
}

// this method should be called inside pm.mutex
func (pm *peerManager) insertPeer(ID peer.ID, peer *remotePeerImpl) {
	if _, exist := pm.hiddenPeerSet[ID]; exist {
		peer.meta.Hidden = true
	}
	pm.remotePeers[ID] = peer
	pm.updatePeerCache()
}

// this method should be called inside pm.mutex
func (pm *peerManager) deletePeer(ID peer.ID) {
	pm.mm.Remove(ID)
	delete(pm.remotePeers, ID)
	pm.updatePeerCache()
}

func (pm *peerManager) updatePeerCache() {
	newSlice := make([]p2pcommon.RemotePeer, 0, len(pm.remotePeers))
	for _, rPeer := range pm.remotePeers {
		newSlice = append(newSlice, rPeer)
	}
	pm.peerCache = newSlice
}

func (pm *peerManager) addAwait(meta p2pcommon.PeerMeta) {
	pm.awaitMutex.Lock()
	defer pm.awaitMutex.Unlock()
	if _, exist := pm.awaitPeers[meta.ID]; exist {
		return
	}
	if atomic.LoadInt32(&pm.status) != running {
		return
	}
	job := newReconnectRunner(meta, pm, pm.logger)
	pm.awaitPeers[meta.ID] = job
	go job.runJob()
}

func (pm *peerManager) cancelAwait(id peer.ID) {
	pm.awaitMutex.Lock()
	defer pm.awaitMutex.Unlock()
	defer func() {
		if atomic.LoadInt32(&pm.status) == stopping &&
			len(pm.awaitPeers) == 0 {
			pm.awaitDone <- struct{}{}
		}
	}()
	job, exist := pm.awaitPeers[id]
	if !exist {
		return
	}
	delete(pm.awaitPeers, id)
	job.cancel <- struct{}{}
}

func (pm *peerManager) cancelAllAwait() {
	for id, _ := range pm.awaitPeers {
		go pm.cancelAwait(id)
	}
	<-pm.awaitDone
}
