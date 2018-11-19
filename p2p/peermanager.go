/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"

	cfg "github.com/aergoio/aergo/config"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// TODO this value better related to max peer and block produce interval, not constant
const (
	DefaultGlobalBlockCacheSize = 300
	DefaultPeerBlockCacheSize   = 100

	DefaultGlobalTxCacheSize = 50000
	DefaultPeerTxCacheSize   = 2000
	// DefaultPeerTxQueueSize is maximum size of hashes in a single tx notice message
	DefaultPeerTxQueueSize = 40000

	defaultTTL          = time.Second * 4
	defaultHandshakeTTL = time.Second * 20

	cachePlaceHolder = true
)

// PeerManager is internal service that provide peer management
type PeerManager interface {
	host.Host
	Start() error
	Stop() error

	PrivateKey() crypto.PrivKey
	PublicKey() crypto.PubKey
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
	host.Host
	privateKey  crypto.PrivKey
	publicKey   crypto.PubKey
	bindAddress net.IP
	bindPort    int
	selfMeta    PeerMeta

	handlerFactory HandlerFactory
	actorServ  ActorService
	signer     msgSigner
	mf 		   moFactory
	rm         ReconnectManager

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

func init() {
}

// NewPeerManager creates a peer manager object.
func NewPeerManager(handlerFactory HandlerFactory, iServ ActorService, cfg *cfg.Config, signer msgSigner, rm ReconnectManager, logger *log.Logger, mf moFactory) PeerManager {
	p2pConf := cfg.P2P
	//logger.SetLevel("debug")
	pm := &peerManager{
		handlerFactory: handlerFactory,
		actorServ: iServ,
		conf:      p2pConf,
		signer:    signer,
		mf:        mf,
		rm:        rm,
		logger:    logger,
		mutex:     &sync.Mutex{},

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

func (pm *peerManager) PrivateKey() crypto.PrivKey {
	return pm.privateKey
}
func (pm *peerManager) PublicKey() crypto.PubKey {
	return pm.publicKey
}
func (pm *peerManager) SelfMeta() PeerMeta {
	return pm.selfMeta
}
func (pm *peerManager) SelfNodeID() peer.ID {
	return pm.selfMeta.ID
}

func (pm *peerManager) RegisterEventListener(listener PeerEventListener) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.eventListeners = append(pm.eventListeners, listener)
}

func (pm *peerManager) init() {
	// check Key and address
	priv := NodePrivKey()
	pub := NodePubKey()
	pid := NodeID()

	pm.privateKey = priv
	pm.publicKey = pub
	// init address and port
	// if not set, it look up ip addresses of machine and choose suitable one (but not so smart) and default port 7845
	peerAddr, peerPort := pm.getProtocolAddrs()
	pm.selfMeta.IPAddress = peerAddr.String()
	pm.selfMeta.Port = uint32(peerPort)
	pm.selfMeta.ID = pid

	// if bindAddress or bindPort is not set, it will be same as NetProtocolAddr or NetProtocolPort
	if len(pm.conf.NPBindAddr) > 0 {
		bindAddr := net.ParseIP(pm.conf.NPBindAddr)
		if bindAddr == nil {
			panic("invalid NPBindAddr " + pm.conf.NPBindAddr)
		}
		pm.bindAddress = bindAddr
	} else {
		pm.bindAddress = peerAddr
	}
	if pm.conf.NPBindPort > 0 {
		pm.bindPort = pm.conf.NPBindPort
	} else {
		pm.bindPort = peerPort
	}

	// set meta info
	// TODO more survey libp2p NAT configuration

	// set designated peers
	pm.addDesignatedPeers()
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
			panic("error while finding IP address: "+err.Error())
		}
		protocolAddr = extIP
	}
	protocolPort = pm.conf.NetProtocolPort
	if protocolPort <= 0 {
		panic("invalid NetProtocolPort " + strconv.Itoa(pm.conf.NetProtocolPort))
	}
	return
}
func (pm *peerManager) run() {

	go pm.runManagePeers()
	// need to start listen after chainservice is read to init
	// FIXME: adhoc code
	go func() {
		time.Sleep(time.Second * 3)
		pm.startListener()

		// addition should start after all modules are started
		go func() {
			time.Sleep(time.Second * 2)
			for _, meta := range pm.designatedPeers {
				pm.addPeerChannel <- meta
			}
		}()
	}()
}

func (pm *peerManager) addDesignatedPeers() {
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
		case peerMetas := <-pm.fillPoolChannel:
			pm.tryFillPool(&peerMetas)
		case <-pm.finishChannel:
			addrTicker.Stop()
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

// addOutboundPeer try to connect and handshake to remote peer. it can be called after peermanager is inited.
// It return true if peer is added or already exist, or return false if failed to add peer.
func (pm *peerManager) addOutboundPeer(meta PeerMeta) bool {
	addrString := fmt.Sprintf("/ip4/%s/tcp/%d", meta.IPAddress, meta.Port)
	var peerAddr, err = ma.NewMultiaddr(addrString)
	if err != nil {
		pm.logger.Warn().Err(err).Str("addr", addrString).Msg("invalid NPAddPeer address")
		return false
	}
	var peerID = meta.ID
	pm.mutex.Lock()
	inboundPeer, ok := pm.remotePeers[peerID]
	if ok {
		// peer is already exist (and maybe inbound peer)
		pm.logger.Info().Str(LogPeerID, inboundPeer.meta.ID.Pretty()).Msg("Peer is already managed by peermanager")
		if meta.Designated {
			// If remote peer was connected first. designated flag is not set yet.
			inboundPeer.meta.Designated = true
		}
		pm.mutex.Unlock()
		return true
	}
	pm.mutex.Unlock()

	// if peer exists in peerstore already, clear previous (and possibly wrong) addresses.
	// commented out clearing code, since there seems to be race condition, this code willed be deleted or reborn after
	// more investigation.
	//if pm.checkInPeerstore(peerID) {
	//	addrs := pm.Peerstore().Addrs(peerID)
	//	addrStrs := make([]string, len(addrs))
	//	for i, addr := range addrs {
	//		addrStrs[i] = addr.String()
	//	}
	//	pm.logger.Debug().Strs("addrs", addrStrs).Str(LogPeerID, peerID.Pretty()).Msg("clearing all prev address of peer before connect")
	//	pm.Peerstore().ClearAddrs(peerID)
	//}
	pm.Peerstore().AddAddr(peerID, peerAddr, meta.TTL())

	ctx := context.Background()
	s, err := pm.NewStream(ctx, meta.ID, aergoP2PSub)
	if err != nil {
		pm.logger.Info().Err(err).Str(LogPeerID, meta.ID.Pretty()).Str(LogProtoID, string(aergoP2PSub)).Msg("Error while get stream")
		return false
	}
	h := newHandshaker(pm, pm.actorServ, pm.logger, peerID)
	rw, remoteStatus, err := h.handshakeOutboundPeerTimeout(s, s, defaultHandshakeTTL)
	if err != nil {
		pm.logger.Debug().Err(err).Str(LogPeerID, meta.ID.Pretty()).Msg("Failed to handshake")
		//pm.sendGoAway(rw, "Failed to handshake")
		s.Close()
		return false
	}

	pm.mutex.Lock()
	inboundPeer, ok = pm.remotePeers[peerID]
	if ok {
		if ComparePeerID(pm.selfMeta.ID, meta.ID) <= 0 {
			pm.logger.Info().Str(LogPeerID, inboundPeer.meta.ID.Pretty()).Msg("Inbound connection was already handshaked while handshaking outbound connection, and remote peer is higher priority so closing this outbound connection.")
			pm.mutex.Unlock()
			pm.sendGoAway(rw, "Already handshaked")
			s.Close()
			return true
		} else {
			pm.logger.Info().Str(LogPeerID, inboundPeer.meta.ID.Pretty()).Msg("Inbound connection was already handshaked while handshaking outbound connection, but local peer is higher priority so closing that inbound connection")
			// disconnect lower valued connection
			pm.deletePeer(meta.ID)
			inboundPeer.stop()
		}
	}

	// update peer info to remote sent infor
	meta = FromPeerAddress(remoteStatus.Sender)

	inboundPeer = newRemotePeer(meta, pm, pm.actorServ, pm.logger, pm.mf, pm.signer, rw)
	// insert Handlers
	pm.handlerFactory.insertHandlers(inboundPeer)
	go inboundPeer.runPeer()
	pm.insertPeer(peerID, inboundPeer)
	pm.logger.Info().Str(LogPeerID, peerID.Pretty()).Str("addr", net.ParseIP(meta.IPAddress).String()+":"+strconv.Itoa(int(meta.Port))).Msg("Outbound peer is  added to peerService")
	pm.mutex.Unlock()

	addrs := pm.Peerstore().Addrs(peerID)
	addrStrs := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrs[i] = addr.String()
	}
	pm.logger.Debug().Strs("addrs" , addrStrs).Str(LogPeerID, inboundPeer.meta.ID.Pretty()).Msg("addresses of peer")

	// peer is ready
	h.doInitialSync()

	// notice to p2pmanager that handshaking is finished
	pm.NotifyPeerHandshake(peerID)

	return true
}

func (pm *peerManager) sendGoAway(rw MsgReadWriter, msg string) {
	goMsg := &types.GoAwayNotice{Message: msg}
	// TODO code smell. non safe casting.
	mo := pm.mf.newMsgRequestOrder(false, GoAway, goMsg).(*pbRequestOrder)
	container := mo.message

	rw.WriteMsg(container)
}

func (pm *peerManager) checkInPeerstore(peerID peer.ID) bool {
	found := false
	for _, existingPeerID := range pm.Peerstore().Peers() {
		if existingPeerID == peerID {
			found = true
			break
		}
	}
	return found
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

	// also disconnect connection
	for _, existingPeerID := range pm.Peerstore().Peers() {
		if existingPeerID == peerID {
			for _, listener := range pm.eventListeners {
				listener.OnRemovePeer(peerID)
			}
			pm.Network().ClosePeer(peerID)
			return true
		}
	}
	return true
}

func (pm *peerManager) Peerstore() pstore.Peerstore {
	return pm.Host.Peerstore()
}

func (pm *peerManager) startListener() {
	var err error
	listens := make([]ma.Multiaddr, 0, 2)
	// FIXME: should also support ip6 later
	listen, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d",pm.bindAddress, pm.bindPort))
	if err != nil {
		panic("Can't estabilish listening address: " + err.Error())
	}
	listens = append(listens, listen)

	peerStore := pstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewPeerMetadata())

	newHost, err := libp2p.New(context.Background(), libp2p.Identity(pm.privateKey), libp2p.Peerstore(peerStore), libp2p.ListenAddrs(listens...))
	if err != nil {
		pm.logger.Fatal().Err(err).Str("addr", listen.String()).Msg("Couldn't listen from")
		panic(err.Error())
	}

	pm.logger.Info().Str("pid", pm.SelfNodeID().Pretty()).Str("addr[0]", listens[0].String()).
		Msg("Set self node's pid, and listening for connections")
	pm.Host = newHost

	pm.SetStreamHandler(aergoP2PSub, pm.onHandshake)
}

func (pm *peerManager) onHandshake(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	h := newHandshaker(pm, pm.actorServ, pm.logger, peerID)
	rw, statusMsg, err := h.handshakeInboundPeer(s, s)
	if err != nil {
		pm.logger.Info().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("fail to handshake")
		s.Close()
		return
	}
	// TODO: check status
	meta := FromPeerAddress(statusMsg.Sender)
	// try Add peer
	if !pm.tryAddInboundPeer(meta, rw) {
		// failed to add
		pm.sendGoAway(rw, "Concurrent handshake")
		s.Close()
		return
	}

	h.doInitialSync()
	// notice to p2pmanager that handshaking is finished
	pm.NotifyPeerHandshake(peerID)
}

func (pm *peerManager) tryAddInboundPeer(meta PeerMeta, rw MsgReadWriter) bool {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	peerID := meta.ID
	outboundPeer, found := pm.remotePeers[peerID]

	if found {
		if ComparePeerID(pm.selfMeta.ID, meta.ID) <= 0 {
			pm.logger.Info().Str(LogPeerID, peerID.Pretty()).Msg("Outbound connection was already handshaked while handshaking inbound connection, and remote peer is higher priority so closing that outbound connection.")
			pm.sendGoAway(rw, "Already handshaked")
			pm.deletePeer(meta.ID)
			outboundPeer.stop()
		} else {
			pm.logger.Info().Str(LogPeerID, peerID.Pretty()).Msg("Outbound connection  was already handshaked while handshaking inbound connection, but local peer is higher priority and closing this inbound connection.")
			// disconnect lower valued connection
			return false
		}
	}
	inboundPeer := newRemotePeer(meta, pm, pm.actorServ, pm.logger, pm.mf, pm.signer, rw)
	pm.handlerFactory.insertHandlers(inboundPeer)
	go inboundPeer.runPeer()
	pm.insertPeer(peerID, inboundPeer)
	peerAddr := meta.ToPeerAddress()

	addrs := pm.Peerstore().Addrs(peerID)
	addrStrs := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrs[i] = addr.String()
	}
	pm.logger.Debug().Strs("addrs" , addrStrs).Str(LogPeerID, inboundPeer.meta.ID.Pretty()).Msg("addresses of peer")
	pm.logger.Info().Str(LogPeerID, peerID.Pretty()).Str("addr", getIP(&peerAddr).String()+":"+strconv.Itoa(int(peerAddr.Port))).Msg("Inbound peer is  added to peerService")
	return true
}

func (pm *peerManager) Start() error {
	pm.run()
	//pm.conf.NPAddPeers
	return nil
}
func (pm *peerManager) Stop() error {
	// TODO stop service
	// close(pm.addPeerChannel)
	// close(pm.removePeerChannel)
	pm.finishChannel <- struct{}{}
	return nil
}

func (pm *peerManager) GetName() string {
	return "p2p service"
}

func (pm *peerManager) checkAndCollectPeerListFromAll() {
	if pm.hasEnoughPeers() {
		return
	}
	for _, remotePeer := range pm.remotePeers {
		pm.actorServ.SendRequest(message.P2PSvc, &message.GetAddressesMsg{ToWhom: remotePeer.meta.ID, Size: 20, Offset: 0})
	}
}

func (pm *peerManager) checkAndCollectPeerList(ID peer.ID) {
	if pm.hasEnoughPeers() {
		return
	}
	peer, ok := pm.GetPeer(ID)
	if !ok {
		//pm.logger.Warnf("invalid peer id %s", ID.Pretty())
		pm.logger.Warn().Str(LogPeerID, ID.Pretty()).Msg("invalid peer id")
		return
	}
	pm.actorServ.SendRequest(message.P2PSvc, &message.GetAddressesMsg{ToWhom: peer.ID(), Size: 20, Offset: 0})
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
	delete(pm.remotePeers, ID)
	pm.updatePeerCache()
}

func (pm *peerManager) updatePeerCache() {
	newSlice := make([]RemotePeer, 0, len(pm.remotePeers))
	for _, peer := range pm.remotePeers {
		newSlice = append(newSlice, peer)
	}
	pm.peerCache = newSlice
}
