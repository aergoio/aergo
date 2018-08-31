/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/libp2p/go-libp2p-host"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"

	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/pkg/component"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

var myPeerInfo peerInfo

type peerInfo struct {
	sync.RWMutex
	id      *peer.ID
	privKey *crypto.PrivKey
}

// TODO this value better related to max peer and block produce interval, not constant
const (
	DefaultGlobalInvCacheSize = 100
	DefaultPeerInvCacheSize   = 30
)

// PeerManager is internal service that provide peer management
type PeerManager interface {
	host.Host
	Start() error
	Stop() error
	GetStatus() component.Status

	PrivateKey() crypto.PrivKey
	PublicKey() crypto.PubKey
	SelfMeta() PeerMeta
	SelfNodeID() peer.ID

	AddNewPeer(peer PeerMeta)
	RemovePeer(peerID peer.ID)
	// NotifyPeerHandshake is called after remote peer is completed handshake and ready to receive or send
	NotifyPeerHandshake(peerID peer.ID)
	NotifyPeerAddressReceived([]PeerMeta)

	HandleNewBlockNotice(peerID peer.ID, b64hash string, data *types.NewBlockNotice)

	// GetPeer return registered(handshaked) remote peer object
	GetPeer(ID peer.ID) (*RemotePeer, bool)
	GetPeers() []*RemotePeer
	GetPeerAddresses() ([]*types.PeerAddress, []types.PeerState)

	// deprecated methods... use sendmessage helper functions instead
	SignProtoMessage(message proto.Message) ([]byte, error)
	AuthenticateMessage(message proto.Message, data *types.MessageData) bool
}

/**
 * peerManager connect to and listen from other nodes.
 * It implements  Component interface
 */
type peerManager struct {
	host.Host
	privateKey crypto.PrivKey
	publicKey  crypto.PubKey
	selfMeta   PeerMeta
	actorServ  ActorService
	rm         ReconnectManager

	designatedPeers map[peer.ID]PeerMeta

	remotePeers map[peer.ID]*RemotePeer
	peerPool    map[peer.ID]PeerMeta
	conf        *cfg.P2PConfig
	logger      *log.Logger
	mutex       *sync.Mutex
	peerCache   []*RemotePeer

	status component.Status

	addPeerChannel    chan PeerMeta
	removePeerChannel chan peer.ID
	fillPoolChannel   chan []PeerMeta
	finishChannel     chan struct{}
	eventListeners    []PeerEventListener

	invCache *lru.Cache
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
func NewPeerManager(iServ ActorService, cfg *cfg.Config, rm ReconnectManager, logger *log.Logger) PeerManager {
	p2pConf := cfg.P2P
	//logger.SetLevel("debug")
	pm := &peerManager{
		actorServ: iServ,
		conf:      p2pConf,
		rm:        rm,
		logger:    logger,
		mutex:     &sync.Mutex{},

		designatedPeers: make(map[peer.ID]PeerMeta, len(cfg.P2P.NPAddPeers)),

		remotePeers: make(map[peer.ID]*RemotePeer, p2pConf.NPMaxPeers),
		peerPool:    make(map[peer.ID]PeerMeta, p2pConf.NPPeerPool),
		peerCache:   make([]*RemotePeer, 0, p2pConf.NPMaxPeers),

		status:            component.StoppedStatus,
		addPeerChannel:    make(chan PeerMeta, 2),
		removePeerChannel: make(chan peer.ID),
		fillPoolChannel:   make(chan []PeerMeta),
		eventListeners:    make([]PeerEventListener, 0, 4),
		finishChannel:     make(chan struct{}),
	}

	var err error
	pm.invCache, err = lru.New(DefaultGlobalInvCacheSize)
	if err != nil {
		panic("Failed to create peermanager " + err.Error())
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
	var priv crypto.PrivKey
	var pub crypto.PubKey
	if pm.conf.NPKey != "" {
		dat, err := ioutil.ReadFile(pm.conf.NPKey)
		if err == nil {
			priv, err = crypto.UnmarshalPrivateKey(dat)
			if err != nil {
				pm.logger.Warn().Str("npkey", pm.conf.NPKey).Msg("invalid keyfile. It's not private key file")
			}
			pub = priv.GetPublic()
		} else {
			pm.logger.Warn().Str("npkey", pm.conf.NPKey).Msg("invalid keyfile path")
		}
	}
	if nil == priv {
		pm.logger.Info().Msg("No valid private key file is found. use temporary pk instead")
		priv, pub, _ = crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	}
	pid, _ := peer.IDFromPublicKey(pub)
	myPeerInfo.set(&pid, &priv)

	listenAddr := net.ParseIP(pm.conf.NetProtocolAddr)
	listenPort := pm.conf.NetProtocolPort
	var err error
	if nil == listenAddr {
		panic("invalid NetProtocolAddr " + pm.conf.NetProtocolAddr)
	} else if !listenAddr.IsUnspecified() {
		pm.logger.Info().Str("pm.conf.NetProtocolAddr", pm.conf.NetProtocolAddr).Int("listenPort", listenPort).Msg("Using NetProtocolAddr in configfile")
	} else {
		listenAddr, err = externalIP()
		pm.logger.Info().Str("addr", listenAddr.To4().String()).Int("port", listenPort).Msg("No NetProtocolAddr is specified")
		if err != nil {
			panic("Couldn't find listening ip address: " + err.Error())
		}
	}
	pm.privateKey = priv
	pm.publicKey = pub
	pm.selfMeta.IPAddress = listenAddr.String()
	pm.selfMeta.Port = uint32(listenPort)
	pm.selfMeta.ID = pid

	// set designated peers
	pm.addDesignatedPeers()
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
			pm.rm.Stop()
			break MANLOOP
		}
	}
	addrTicker.Stop()

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
	newPeer, ok := pm.remotePeers[peerID]
	if ok {
		// peer is already exist
		pm.logger.Info().Str(LogPeerID, newPeer.meta.ID.Pretty()).Msg("Peer is already managed by peerService")
		if meta.Designated {
			// If remote peer was connected first. designated flag is not set yet.
			newPeer.meta.Designated = true
		}
		pm.mutex.Unlock()
		return true
	}
	pm.mutex.Unlock()

	// if peer exists in peerstore already, reuse that peer again.
	if !pm.checkInPeerstore(peerID) {
		pm.Peerstore().AddAddr(peerID, peerAddr, meta.TTL())
	}

	ctx := context.Background()
	s, err := pm.NewStream(ctx, meta.ID, aergoP2PSub)
	if err != nil {
		pm.logger.Info().Err(err).Str(LogPeerID, meta.ID.Pretty()).Str(LogProtoID, string(aergoP2PSub)).Msg("Error while get stream")
		return false
	}
	rw := &bufio.ReadWriter{Reader: bufio.NewReader(s), Writer: bufio.NewWriter(s)}

	remoteStatus, success := initiateHandshake(pm, peerID, rw)
	if !success {
		pm.sendGoAway(rw, "Failed to handshake")
		s.Close()
		return false
	}

	pm.mutex.Lock()
	newPeer, ok = pm.remotePeers[peerID]
	if ok {
		if ComparePeerID(pm.selfMeta.ID, meta.ID) <= 0 {
			pm.logger.Info().Str(LogPeerID, newPeer.meta.ID.Pretty()).Msg("Peer is added while handshaking")
			s.Close()
			pm.mutex.Unlock()
			return true
		}
	}

	newPeer = newRemotePeer(meta, pm, pm.actorServ, pm.logger)
	newPeer.rw = &bufio.ReadWriter{Reader: bufio.NewReader(s), Writer: bufio.NewWriter(s)}
	// insert Handlers
	pm.insertHandlers(newPeer)
	go newPeer.runPeer()
	newPeer.setState(types.RUNNING)

	pm.insertPeer(peerID, newPeer)
	pm.logger.Info().Str(LogPeerID, peerID.Pretty()).Str("addr", net.ParseIP(meta.IPAddress).String()+":"+strconv.Itoa(int(meta.Port))).Msg("Outbound peer is  added to peerService")
	pm.mutex.Unlock()

	// peer is ready
	pm.actorServ.SendRequest(message.ChainSvc, &message.SyncBlockState{PeerID: peerID, BlockNo: remoteStatus.BestHeight, BlockHash: remoteStatus.BestBlockHash})

	// notice to p2pmanager that handshaking is finished
	pm.NotifyPeerHandshake(peerID)

	return true
}

func (pm *peerManager) insertHandlers(peer *RemotePeer) {
	// PingHandler
	ph := NewPingHandler(pm, peer, pm.logger)
	peer.handlers[pingRequest] = ph.handlePing
	peer.handlers[pingResponse] = ph.handlePingResponse
	peer.handlers[goAway] = ph.handleGoAway
	peer.handlers[addressesRequest] = ph.handleAddressesRequest
	peer.handlers[addressesResponse] = ph.handleAddressesResponse

	// BlockHandler
	bh := NewBlockHandler(pm, peer, pm.logger)
	peer.handlers[getBlocksRequest] = bh.handleBlockRequest
	peer.handlers[getBlocksResponse] = bh.handleGetBlockResponse
	peer.handlers[getBlockHeadersRequest] = bh.handleGetBlockHeadersRequest
	peer.handlers[getBlockHeadersResponse] = bh.handleGetBlockHeadersResponse
	peer.handlers[getMissingRequest] = bh.handleGetMissingRequest
	// peer.handlers[getMissingResponse] = // no function yet
	peer.handlers[newBlockNotice] = bh.handleNewBlockNotice

	th := NewTxHandler(pm, peer, pm.logger)
	peer.handlers[getTXsRequest] = th.handleGetTXsRequest
	peer.handlers[getTxsResponse] = th.handleGetTXsResponse
	peer.handlers[newTxNotice] = th.handleNewTXsNotice
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
	listen, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d",
		pm.selfMeta.IPAddress, pm.selfMeta.Port))
	if err != nil {
		panic("Can't estabilish listening address: " + err.Error())
	}
	listens = append(listens, listen)
	listen, _ = ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", pm.selfMeta.Port))
	listens = append(listens, listen)

	peerStore := pstore.NewPeerstore()

	newHost, err := libp2p.New(context.Background(), libp2p.Identity(pm.privateKey), libp2p.Peerstore(peerStore), libp2p.ListenAddrs(listens...))
	if err != nil {
		pm.logger.Fatal().Err(err).Str("addr", listen.String()).Msg("Couldn't listen from")
		panic(err.Error())
	}

	pm.logger.Info().Str("pid", pm.SelfNodeID().Pretty()).Str("addr[0]", listens[0].String()).Str("addr[1]", listens[1].String()).
		Msg("Set self node's pid, and listening for connections")
	pm.Host = newHost

	pm.SetStreamHandler(aergoP2PSub, pm.onHandshake)
}

func (pi *peerInfo) set(id *peer.ID, privKey *crypto.PrivKey) {
	pi.Lock()
	pi.id = id
	pi.privKey = privKey
	pi.Unlock()
}

// GetMyID returns the peer id of the current node
func GetMyID() (peer.ID, crypto.PrivKey) {
	var id *peer.ID
	var pk *crypto.PrivKey

	for pk == nil || id == nil {
		myPeerInfo.RLock()
		id = myPeerInfo.id
		pk = myPeerInfo.privKey
		myPeerInfo.RUnlock()

		// To prevent high cpu usage
		time.Sleep(100 * time.Millisecond)
	}

	return *id, *pk
}

func (pm *peerManager) Start() error {
	pm.run()
	pm.status = component.StartedStatus
	//pm.conf.NPAddPeers
	return nil
}
func (pm *peerManager) Stop() error {
	// TODO stop service
	// close(pm.addPeerChannel)
	// close(pm.removePeerChannel)
	pm.status = component.StoppedStatus
	pm.finishChannel <- struct{}{}
	return nil
}

func (pm *peerManager) GetStatus() component.Status {
	return pm.status
}

func (pm *peerManager) Started() bool {
	return pm.status == component.StartedStatus
}

func (pm *peerManager) Ended() bool {
	return pm.status == component.StoppedStatus
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
	pm.actorServ.SendRequest(message.P2PSvc, &message.GetAddressesMsg{ToWhom: peer.meta.ID, Size: 20, Offset: 0})
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

func (pm *peerManager) GetPeer(ID peer.ID) (*RemotePeer, bool) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// vs code's lint does not allow direct return of map operation
	ptr, ok := pm.remotePeers[ID]
	if !ok {
		return nil, false
	}
	return ptr, ok
}

func (pm *peerManager) GetPeers() []*RemotePeer {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	return pm.peerCache
}

func (pm *peerManager) GetPeerAddresses() ([]*types.PeerAddress, []types.PeerState) {
	peers := make([]*types.PeerAddress, 0, len(pm.remotePeers))
	states := make([]types.PeerState, 0, len(pm.remotePeers))
	for _, aPeer := range pm.remotePeers {
		addr := aPeer.meta.ToPeerAddress()
		peers = append(peers, &addr)
		states = append(states, aPeer.state)
	}
	return peers, states
}

func (pm *peerManager) HandleNewBlockNotice(peerID peer.ID, b64hash string, data *types.NewBlockNotice) {
	// TODO check if evicted return value is needed.
	ok, _ := pm.invCache.ContainsOrAdd(b64hash, data.BlockHash)
	if ok {
		pm.logger.Debug().Str(LogBlkHash, b64hash).Str(LogPeerID, peerID.Pretty()).Msg("Got NewBlock notice, but sent already from other peer")
		// this notice is already sent to chainservice
		return
	}

	// request block info if selfnode does not have block already
	rawResp, err := pm.actorServ.CallRequest(message.ChainSvc, &message.GetBlock{BlockHash: message.BlockHash(data.BlockHash)})
	if err != nil {
		pm.logger.Warn().Err(err).Msg("actor return error on getblock")
		return
	}
	resp, ok := rawResp.(message.GetBlockRsp)
	if !ok {
		pm.logger.Warn().Str("expected", "message.GetBlockRsp").Str("actual", reflect.TypeOf(rawResp).Name()).Msg("chainservice returned unexpected type")
		return
	}
	if resp.Err != nil {
		pm.logger.Debug().Str(LogBlkHash, b64hash).Str(LogPeerID, peerID.Pretty()).Msg("chainservice responded that block not found. request back to notifier")
		pm.actorServ.SendRequest(message.P2PSvc, &message.GetBlockInfos{ToWhom: peerID,
			Hashes: []message.BlockHash{message.BlockHash(data.BlockHash)}})
	}

}

// this method should be called inside pm.mutex
func (pm *peerManager) insertPeer(ID peer.ID, peer *RemotePeer) {
	pm.remotePeers[ID] = peer
	pm.updatePeerCache()
}

// this method should be called inside pm.mutex
func (pm *peerManager) deletePeer(ID peer.ID) {
	delete(pm.remotePeers, ID)
	pm.updatePeerCache()
}

func (pm *peerManager) updatePeerCache() {
	newSlice := make([]*RemotePeer, 0, len(pm.remotePeers))
	for _, peer := range pm.remotePeers {
		newSlice = append(newSlice, peer)
	}
	pm.peerCache = newSlice
}
