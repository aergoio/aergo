/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/p2p/metric"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	inet "github.com/libp2p/go-libp2p-net"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"

	cfg "github.com/aergoio/aergo/config"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

// PeerManager is internal service that provide peer management
type PeerManager interface {
	Start() error
	Stop() error

	//NetworkTransport
	SelfMeta() PeerMeta
	SelfNodeID() peer.ID

	AddNewPeer(peer PeerMeta)
	RemovePeer(peerID peer.ID)
	// NotifyPeerHandshake is called after remote peer is completed handshake and ready to receive or send
	NotifyPeerHandshake(peerID peer.ID)
	NotifyPeerAddressReceived([]PeerMeta)

	// GetPeer return registered(handshaked) remote peer object
	GetPeer(ID peer.ID) (RemotePeer, bool)
	GetPeers() []RemotePeer
	GetPeerAddresses() ([]*types.PeerAddress, []*types.NewBlockNotice, []types.PeerState)
}

/**
 * peerManager connect to and listen from other nodes.
 * It implements  Component interface
 */
type peerManager struct {
	nt       NetworkTransport
	hsFactory HSHandlerFactory
	handlerFactory HandlerFactory
	actorService   ActorService
	signer         msgSigner
	mf             moFactory
	rm             ReconnectManager
	mm             metric.MetricsManager

	designatedPeers map[peer.ID]PeerMeta

	remotePeers map[peer.ID]*remotePeerImpl
	peerPool    map[peer.ID]PeerMeta
	conf        *cfg.P2PConfig
	logger      *log.Logger
	mutex       *sync.Mutex
	peerCache   []RemotePeer

	addPeerChannel    chan PeerMeta
	removePeerChannel chan peer.ID
	fillPoolChannel   chan []PeerMeta
	finishChannel     chan struct{}
	eventListeners    []PeerEventListener
}

var _ PeerManager = (*peerManager)(nil)

// PeerEventListener listen peer manage event
type PeerEventListener interface {
	// OnAddPeer is called just after the peer is added.
	OnAddPeer(peerID peer.ID)

	// OnRemovePeer is called just before the peer is removed
	OnRemovePeer(peerID peer.ID)
}

// NewPeerManager creates a peer manager object.
func NewPeerManager(handlerFactory HandlerFactory, hsFactory HSHandlerFactory, iServ ActorService, cfg *cfg.Config, signer msgSigner, nt NetworkTransport, rm ReconnectManager, mm metric.MetricsManager, logger *log.Logger, mf moFactory) PeerManager {
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
		rm:             rm,
		mm:             mm,
		logger:         logger,
		mutex:          &sync.Mutex{},

		designatedPeers: make(map[peer.ID]PeerMeta, len(cfg.P2P.NPAddPeers)),

		remotePeers: make(map[peer.ID]*remotePeerImpl, p2pConf.NPMaxPeers),
		peerPool:    make(map[peer.ID]PeerMeta, p2pConf.NPPeerPool),
		peerCache:   make([]RemotePeer, 0, p2pConf.NPMaxPeers),

		addPeerChannel:    make(chan PeerMeta, 2),
		removePeerChannel: make(chan peer.ID),
		fillPoolChannel:   make(chan []PeerMeta),
		eventListeners:    make([]PeerEventListener, 0, 4),
		finishChannel:     make(chan struct{}),
	}

	// additional initializations
	pm.init()

	return pm
}

func (pm *peerManager) SelfMeta() PeerMeta {
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
}

func (pm *peerManager) getProtocolAddrs() (protocolAddr net.IP, protocolPort int) {
	if len(pm.conf.NetProtocolAddr) != 0 {
		protocolAddr = net.ParseIP(pm.conf.NetProtocolAddr)
		if protocolAddr == nil {
			panic("invalid NetProtocolAddr " + pm.conf.NetProtocolAddr)
		}
		if protocolAddr.IsUnspecified() {
			panic("NetProtocolAddr should be a specified IP address, not 0.0.0.0")
		}
	} else {
		extIP, err := externalIP()
		if err != nil {
			panic("error while finding IP address: " + err.Error())
		}
		protocolAddr = extIP
	}
	protocolPort = pm.conf.NetProtocolPort
	if protocolPort <= 0 {
		panic("invalid NetProtocolPort " + strconv.Itoa(pm.conf.NetProtocolPort))
	}
	return
}

func (pm *peerManager) Start() error {

	go pm.runManagePeers()
	// need to start listen after chainservice is read to init
	// FIXME: adhoc code
	go func() {
		time.Sleep(time.Second * 3)
		// TODO sl을 독립시켜야 aergomap에서도 사용할 수 있음.
		pm.nt.SetStreamHandler(aergoP2PSub, pm.onConnect)

		// addition should start after all modules are started
		go func() {
			time.Sleep(time.Second * 2)
			for _, meta := range pm.designatedPeers {
				pm.addPeerChannel <- meta
			}
		}()
	}()

	return nil
}


func (pm *peerManager) Stop() error {
	// TODO stop service
	pm.finishChannel <- struct{}{}
	return nil
}

func (pm *peerManager) initDesignatedPeerList() {
	// add remote node from config
	for _, target := range pm.conf.NPAddPeers {
		// go-multiaddr implementation does not support recent p2p protocol yet, but deprecated name ipfs.
		// This adhoc will be removed when go-multiaddr is patched.
		target = strings.Replace(target, "/p2p/", "/ipfs/", 1)
		targetAddr, err := ma.NewMultiaddr(target)
		if err != nil {
			pm.logger.Warn().Err(err).Str("target", target).Msg("invalid NPAddPeer address")
			continue
		}
		splitted := strings.Split(targetAddr.String(), "/")
		if len(splitted) != 7 {
			pm.logger.Warn().Str("target", target).Msg("invalid NPAddPeer address")
			continue
		}
		peerAddrString := splitted[2]
		peerPortString := splitted[4]
		peerPort, err := strconv.Atoi(peerPortString)
		if err != nil {
			pm.logger.Warn().Str("port", peerPortString).Msg("invalid Peer port")
			continue
		}
		peerIDString := splitted[6]
		peerID, err := peer.IDB58Decode(peerIDString)
		if err != nil {
			pm.logger.Warn().Str(LogPeerID, peerIDString).Msg("invalid PeerID")
			continue
		}
		peerMeta := PeerMeta{
			ID:         peerID,
			Port:       uint32(peerPort),
			IPAddress:  peerAddrString,
			Designated: true,
			Outbound:   true,
		}
		pm.logger.Info().Str(LogPeerID, peerID.Pretty()).Str("addr", peerAddrString).Int("port", peerPort).Msg("Adding Designated peer")
		pm.designatedPeers[peerID] = peerMeta
	}
}

func (pm *peerManager) runManagePeers() {
	addrDuration := time.Minute * 3
	addrTicker := time.NewTicker(addrDuration)
	// reconnectRunners := make(map[peer.ID]*reconnectRunner)
MANLOOP:
	for {
		select {
		case meta := <-pm.addPeerChannel:
			if pm.addOutboundPeer(meta) {
				if _, found := pm.designatedPeers[meta.ID]; found {
					pm.rm.CancelJob(meta.ID)
				}
			}
		case id := <-pm.removePeerChannel:
			if pm.removePeer(id) {
				if meta, found := pm.designatedPeers[id]; found {
					pm.rm.AddJob(meta)
				}
			}
		case <-addrTicker.C:
			pm.checkAndCollectPeerListFromAll()
		    pm.logPeerMetrics()
		case peerMetas := <-pm.fillPoolChannel:
			pm.tryFillPool(&peerMetas)
		case <-pm.finishChannel:
			addrTicker.Stop()
			pm.nt.RemoveStreamHandler(aergoP2PSub)
			pm.rm.Stop()
			// TODO need to keep loop till all remote peer objects are removed, otherwise panic or channel deadlock can come.
			break MANLOOP
		}
	}

	// cleanup peers
	for peerID := range pm.remotePeers {
		pm.removePeer(peerID)
	}
}

func (pm *peerManager) logPeerMetrics() {
	if pm.logger.IsDebugEnabled() {
		pm.logger.Debug().Msg(pm.mm.Summary())
	}
}

// addOutboundPeer try to connect and handshake to remote peer. it can be called after peermanager is inited.
// It return true if peer is added or return false if failed to add peer or more suitable connection already exists.
func (pm *peerManager) addOutboundPeer(meta PeerMeta) bool {
	s, err := pm.nt.GetOrCreateStream(meta, aergoP2PSub)
	if err != nil {
		pm.logger.Info().Err(err).Str(LogPeerID, meta.ID.Pretty()).Msg("Failed to get stream.")
		return false
	}

	completeMeta, added := pm.tryAddPeer(true, meta, s)
	if !added {
		s.Close()
		return false
	} else {
		if meta.IPAddress != completeMeta.IPAddress {
			pm.logger.Debug().Str("before",meta.IPAddress).Str("after", completeMeta.IPAddress).Msg("IP address of remote peer is changed to ")
		}
	}
	return true
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer setup some
func (pm *peerManager) tryAddPeer(outbound bool, meta PeerMeta, s inet.Stream) (PeerMeta, bool) {
	var peerID = meta.ID
	rd := metric.NewReader(s)
	wt := metric.NewWriter(s)
	h := pm.hsFactory.CreateHSHandler(outbound, pm, pm.actorService, pm.logger, peerID)
	rw, remoteStatus, err := h.Handle(rd, wt, defaultHandshakeTTL)
	if err != nil {
		pm.logger.Debug().Err(err).Str(LogPeerID, meta.ID.Pretty()).Msg("Failed to handshake")
		pm.sendGoAway(rw, "Failed to handshake")
		return meta, false
	}
	// update peer meta info using sent information from remote peer
	receivedMeta := FromPeerAddress(remoteStatus.Sender)
	if receivedMeta.ID != peerID {
		pm.logger.Debug().Str("received_peer_id", receivedMeta.ID.Pretty()).Str(LogPeerID, meta.ID.Pretty()).Msg("Inconsistent peerID")
		pm.sendGoAway(rw, "Inconsistent peerID")
		return meta, false
	}
	receivedMeta.Outbound = outbound
	_, receivedMeta.Designated = pm.designatedPeers[peerID]

	// adding peer to peer list
	newPeer, err := pm.registerPeer(peerID, receivedMeta, rw)
	if err != nil {
			pm.sendGoAway(rw, err.Error() )
			return meta, false
	}
	newPeer.metric = pm.mm.Add(peerID, rd, wt)

	if pm.logger.IsDebugEnabled() {
		addrStrs := pm.nt.GetAddressesOfPeer(peerID)
		pm.logger.Debug().Strs("addrs", addrStrs).Str(LogPeerID, peerID.Pretty()).Msg("addresses of peer")
	}

	pm.doPostHandshake(peerID, remoteStatus)
	// notice to p2pmanager that handshaking is finished
	pm.NotifyPeerHandshake(peerID)

	return receivedMeta, true
}

func (pm *peerManager) registerPeer(peerID peer.ID, receivedMeta PeerMeta, rw MsgReadWriter) (*remotePeerImpl, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	preExistPeer, ok := pm.remotePeers[peerID]
	if ok {
		pm.logger.Info().Str(LogPeerID, peerID.Pretty()).Msg("Peer add collision. Outbound connection of higher hash will survive.")
		iAmLower := ComparePeerID(pm.SelfNodeID(), receivedMeta.ID) <= 0
		if iAmLower == receivedMeta.Outbound {
			pm.logger.Info().Str("local_peer_id",pm.SelfNodeID().Pretty()).Str(LogPeerID, peerID.Pretty()).Bool("outbound",receivedMeta.Outbound).Msg("Close connection and keep earlier handshake connection.")
			return nil, fmt.Errorf("Already handshake peer %s ",peerID.Pretty())
		} else {
			pm.logger.Info().Str("local_peer_id",pm.SelfNodeID().Pretty()).Str(LogPeerID, peerID.Pretty()).Bool("outbound",receivedMeta.Outbound).Msg("Keep connection and close earlier handshake connection.")
			// TODO send goaway messge to pre-exist peer
			// disconnect lower valued connection
			pm.deletePeer(receivedMeta.ID)
			preExistPeer.stop()
		}
	}

	outboundPeer := newRemotePeer(receivedMeta, pm, pm.actorService, pm.logger, pm.mf, pm.signer, rw)
	// insert Handlers
	pm.handlerFactory.insertHandlers(outboundPeer)
	go outboundPeer.runPeer()
	pm.insertPeer(peerID, outboundPeer)
	pm.logger.Info().Bool("outbound",receivedMeta.Outbound).Str(LogPeerID, peerID.Pretty()).Str("addr", net.ParseIP(receivedMeta.IPAddress).String()+":"+strconv.Itoa(int(receivedMeta.Port))).Msg("peer is  added to peerService")

	return outboundPeer, nil
}

// doPostHandshake is additional work after peer is added.
func (pm *peerManager) doPostHandshake(peerID peer.ID, remoteStatus *types.Status) {

	if chain.UseFastSyncer {
		pm.logger.Debug().Uint64("target", remoteStatus.BestHeight).Msg("request new syncer")
		pm.actorService.SendRequest(message.SyncerSvc, &message.SyncStart{PeerID: peerID, TargetNo: remoteStatus.BestHeight})
	} else {
		// sync block infos
		pm.actorService.SendRequest(message.ChainSvc, &message.SyncBlockState{PeerID: peerID, BlockNo: remoteStatus.BestHeight, BlockHash: remoteStatus.BestBlockHash})
	}

	// sync mempool tx infos
	// TODO add tx handling
}

func (pm *peerManager) sendGoAway(rw MsgReadWriter, msg string) {
	goMsg := &types.GoAwayNotice{Message: msg}
	// TODO code smell. non safe casting.
	mo := pm.mf.newMsgRequestOrder(false, GoAway, goMsg).(*pbRequestOrder)
	container := mo.message

	rw.WriteMsg(container)
}

func (pm *peerManager) AddNewPeer(peer PeerMeta) {
	pm.addPeerChannel <- peer
}

func (pm *peerManager) RemovePeer(peerID peer.ID) {
	pm.removePeerChannel <- peerID
}

func (pm *peerManager) NotifyPeerHandshake(peerID peer.ID) {
	pm.checkAndCollectPeerList(peerID)
}

func (pm *peerManager) NotifyPeerAddressReceived(metas []PeerMeta) {
	pm.fillPoolChannel <- metas
}

// removePeer remove and disconnect managed remote peer connection
// It return true if peer is exist and managed by peermanager
func (pm *peerManager) removePeer(peerID peer.ID) bool {
	pm.mutex.Lock()
	target, ok := pm.remotePeers[peerID]
	if !ok {
		pm.mutex.Unlock()
		return false
	}
	pm.deletePeer(peerID)
	// No internal module access this peer anymore, but remote message can be received.
	target.stop()
	pm.mutex.Unlock()
	for _, listener := range pm.eventListeners {
		listener.OnRemovePeer(peerID)
	}

	// also disconnect connection
	pm.nt.ClosePeerConnection(peerID)
	return true
}

func (pm *peerManager) onConnect(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	tempMeta := PeerMeta{ID:peerID}
	completeMeta, added := pm.tryAddPeer(false, tempMeta, s)
	if !added {
		s.Close()
	} else {
		if tempMeta.IPAddress != completeMeta.IPAddress {
			pm.logger.Debug().Str("after", completeMeta.IPAddress).Msg("IP address of remote peer")
		}
	}
}

func (pm *peerManager) checkAndCollectPeerListFromAll() {
	if pm.hasEnoughPeers() {
		return
	}
	for _, remotePeer := range pm.remotePeers {
		pm.actorService.SendRequest(message.P2PSvc, &message.GetAddressesMsg{ToWhom: remotePeer.meta.ID, Size: 20, Offset: 0})
	}
}

func (pm *peerManager) checkAndCollectPeerList(ID peer.ID) {
	if pm.hasEnoughPeers() {
		return
	}
	rPeer, ok := pm.GetPeer(ID)
	if !ok {
		//pm.logger.Warnf("invalid peer id %s", ID.Pretty())
		pm.logger.Warn().Str(LogPeerID, ID.Pretty()).Msg("invalid peer id")
		return
	}
	pm.actorService.SendRequest(message.P2PSvc, &message.GetAddressesMsg{ToWhom: rPeer.ID(), Size: 20, Offset: 0})
}

func (pm *peerManager) hasEnoughPeers() bool {
	return len(pm.peerPool) >= pm.conf.NPPeerPool
}

// tryConnectPeers should be called in runManagePeers() only
func (pm *peerManager) tryFillPool(metas *[]PeerMeta) {
	added := make([]PeerMeta, 0, len(*metas))
	invalid := make([]string, 0)
	for _, meta := range *metas {
		if string(meta.ID) == "" {
			invalid = append(invalid, meta.String())
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
			pm.logger.Warn().Str(LogPeerID, meta.ID.Pretty()).Str("addr", meta.IPAddress).
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

func (pm *peerManager) GetPeer(ID peer.ID) (RemotePeer, bool) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// vs code's lint does not allow direct return of map operation
	ptr, ok := pm.remotePeers[ID]
	if !ok {
		return nil, false
	}
	return ptr, ok
}

func (pm *peerManager) GetPeers() []RemotePeer {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	return pm.peerCache
}

func (pm *peerManager) GetPeerAddresses() ([]*types.PeerAddress, []*types.NewBlockNotice, []types.PeerState) {
	peers := make([]*types.PeerAddress, 0, len(pm.remotePeers))
	blks := make([]*types.NewBlockNotice, 0, len(pm.remotePeers))
	states := make([]types.PeerState, 0, len(pm.remotePeers))
	for _, aPeer := range pm.remotePeers {
		addr := aPeer.meta.ToPeerAddress()
		peers = append(peers, &addr)
		blks = append(blks, aPeer.lastNotice)
		states = append(states, aPeer.state)
	}
	return peers, blks, states
}

// this method should be called inside pm.mutex
func (pm *peerManager) insertPeer(ID peer.ID, peer *remotePeerImpl) {
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
	newSlice := make([]RemotePeer, 0, len(pm.remotePeers))
	for _, rPeer := range pm.remotePeers {
		newSlice = append(newSlice, rPeer)
	}
	pm.peerCache = newSlice
}

