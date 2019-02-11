/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package pmap

import (
	"bufio"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
)

// internal
const (
	PolarisConnectionTTL = time.Second * 30
	PolarisPingTTL       = PolarisConnectionTTL >> 1

	// polaris will return peers list at most this number
	ResponseMaxPeerLimit = 500
	// libp2p internal library is not always send message instantly, so closing socket soon after sent a message will cause packet loss and read error, us walkaround here till finding the real reason and fix it.
	MsgSendDelay = time.Second * 1

	PeerHealthcheckInterval = time.Minute
	//PeerHealthcheckInterval = time.Minute * 5
	ConcurrentHealthCheckCount = 20
)

var (
	EmptyMsgID = p2pcommon.MsgID(uuid.Nil)
)

var (
	// 89.16 is ceiling of declination of Polaris
	MainnetMapServer = []string{
		"/dns/polaris.aergo.io/tcp/8916/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF",
	}

	// 89.16 is ceiling of declination of Polaris
	TestnetMapServer = []string{
		"/dns/polaris.aergo.io/tcp/8916/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF",
	}

	// Hardcoded chainID of ONE MAINNET and ONE TESTNET
	ONEMainNet types.ChainID
	ONETestNet types.ChainID
)

func init() {
	// mainnet is not opened yet and have some unconfirmed values now, this values will be changed after the spec of mainnet is determined.
	//FIXME
	ONEMainNet = types.ChainID{PublicNet: true, MainNet: true, CoinbaseFee: "1000000000", Consensus: "dpos", Magic: "mainnet.aergo.io"}

	tnGen := types.GetTestNetGenesis()
	if tnGen == nil {
		panic("Failed to get TestNet GenesisInfo")
	}
	ONETestNet = tnGen.ID
}

type mapService interface {
	getPeerCheckers() []peerChecker
	registerPeer(receivedMeta p2pcommon.PeerMeta) error
	unregisterPeer(peerID peer.ID)
}

type peerChecker interface {
	lastCheck() time.Time
	// check checks peer. it will stop check at best effort when timeout is exceeded. and wg done.
	check(wg *sync.WaitGroup, timeout time.Duration)
}

// PeerMapService is
type PeerMapService struct {
	*component.BaseComponent

	PrivateNet   bool
	allowPrivate bool

	ntc p2p.NTContainer
	nt  p2p.NetworkTransport
	hc  HealthCheckManager

	rwmutex      *sync.RWMutex
	peerRegistry map[peer.ID]*peerState
}

func NewPolarisService(cfg *config.Config, ntc p2p.NTContainer) *PeerMapService {
	pms := &PeerMapService{
		rwmutex:      &sync.RWMutex{},
		peerRegistry: make(map[peer.ID]*peerState),
		allowPrivate: cfg.Polaris.AllowPrivate,
	}

	pms.BaseComponent = component.NewBaseComponent(message.PolarisSvc, pms, log.NewLogger("polaris"))

	pms.ntc = ntc
	pms.hc = NewHCM(pms, pms.nt)

	pms.PrivateNet = !ntc.ChainID().MainNet

	// initialize map Servers
	return pms
}

func (pms *PeerMapService) SetHub(hub *component.ComponentHub) {
	pms.BaseComponent.SetHub(hub)
}

func (pms *PeerMapService) BeforeStart() {}

func (pms *PeerMapService) AfterStart() {
	pms.nt = pms.ntc.GetNetworkTransport()
	pms.Logger.Info().Str("version", string(PolarisMapSub)).Msg("Starting polaris listening")
	pms.nt.AddStreamHandler(PolarisMapSub, pms.onConnect)
	pms.hc.Start()
}

func (pms *PeerMapService) BeforeStop() {
	if pms.nt != nil {
		pms.hc.Stop()
		pms.nt.RemoveStreamHandler(PolarisMapSub)
	}
}

func (pms *PeerMapService) Statistics() *map[string]interface{} {
	return nil
	//dummy := make(map[string]interface{})
	//return &dummy
}

func (pms *PeerMapService) onConnect(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	remoteAddrStr := s.Conn().RemoteMultiaddr().String()
	remotePeerMeta := p2pcommon.PeerMeta{ID: peerID}
	pms.Logger.Debug().Str("addr", remoteAddrStr).Str(p2p.LogPeerID, peerID.String()).Msg("Received map query")

	rw := p2p.NewV030ReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	defer s.Close()

	// receive input
	container, query, err := pms.readRequest(remotePeerMeta, rw)
	if err != nil {
		pms.Logger.Debug().Err(err).Str(p2p.LogPeerID, peerID.String()).Msg("failed to read query")
		return
	}

	resp, err := pms.handleQuery(container, query)
	if err != nil {
		pms.Logger.Debug().Err(err).Str(p2p.LogPeerID, peerID.String()).Msg("failed to handle query")
		return
	}

	// response to peer
	if err = pms.writeResponse(container, remotePeerMeta, resp, rw); err != nil {
		pms.Logger.Debug().Err(err).Str(p2p.LogPeerID, peerID.String()).Msg("failed to write query")
		return
	}
	pms.Logger.Debug().Str("status", resp.Status.String()).Str(p2p.LogPeerID, peerID.String()).Int("peer_cnt", len(resp.Addresses)).Msg("Sent map response")

	// TODO send goodbye message.
	time.Sleep(time.Second * 3)

	// disconnect!
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer setup some
func (pms *PeerMapService) readRequest(meta p2pcommon.PeerMeta, rd p2p.MsgReader) (p2pcommon.Message, *types.MapQuery, error) {
	data, err := rd.ReadMsg()
	if err != nil {
		return nil, nil, err
	}
	queryReq := &types.MapQuery{}
	err = p2p.UnmarshalMessage(data.Payload(), queryReq)
	if err != nil {
		return data, nil, err
	}

	return data, queryReq, nil
}

// handleQuery check query parameters, return registered peers, and register this peer if requested.
func (pms *PeerMapService) handleQuery(container p2pcommon.Message, query *types.MapQuery) (*types.MapResponse, error) {
	if query.Status == nil {
		return nil, fmt.Errorf("malformed query %v", query)
	}
	receivedMeta := p2pcommon.FromPeerAddress(query.Status.Sender)
	maxPeers := int(query.Size)
	if maxPeers <= 0 {
		return nil, fmt.Errorf("invalid argument count %d", maxPeers)
	} else if maxPeers > ResponseMaxPeerLimit {
		pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Int("req_size", maxPeers).Int("clipped", ResponseMaxPeerLimit).Msg("Clipping too high count of query ")
		maxPeers = ResponseMaxPeerLimit
	}

	pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Msg("Handling query.")

	// make response
	resp := &types.MapResponse{}

	// compare chainID
	sameChain, err := pms.checkChain(query.Status.ChainID)
	if err != nil {
		pms.Logger.Debug().Err(err).Str(p2p.LogPeerID, receivedMeta.ID.String()).Bytes("chainid", query.Status.ChainID).Msg("err parsing chainid")
		resp.Status = types.ResultStatus_INVALID_ARGUMENT
		resp.Message = "invalid chainid"
		return resp, nil
	} else if !sameChain {
		pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Msg("err different chain")
		resp.Status = types.ResultStatus_UNAUTHENTICATED
		resp.Message = "different chain"
		return resp, nil
	}

	resp.Addresses = pms.retrieveList(maxPeers, receivedMeta.ID)

	if query.AddMe {
		// check Sender
		// check peer is really capable to aergosvr
		if !pms.checkConnectness(receivedMeta) {
			pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Msg("AddMe is set, but cant connect back to peer")
			resp.Status = types.ResultStatus_OK
			resp.Message = "can't connect back, so not registered"
			return resp, nil
		}
		pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Msg("AddMe is set, and register peer to peer registry")
		pms.registerPeer(receivedMeta)
	}

	resp.Status = types.ResultStatus_OK
	return resp, nil
}

func (pms *PeerMapService) retrieveList(maxPeers int, exclude peer.ID) []*types.PeerAddress {
	list := make([]*types.PeerAddress, 0, maxPeers)
	pms.rwmutex.Lock()
	defer pms.rwmutex.Unlock()
	for id, ps := range pms.peerRegistry {
		if id == exclude {
			continue
		}
		list = append(list, &ps.addr)
		if len(list) >= maxPeers {
			return list
		}
	}
	return list
}

func (pms *PeerMapService) registerPeer(receivedMeta p2pcommon.PeerMeta) error {
	peerID := receivedMeta.ID
	pms.rwmutex.Lock()
	defer pms.rwmutex.Unlock()
	now := time.Now()
	prev, ok := pms.peerRegistry[peerID]
	if !ok {
		newState := &peerState{connected: now, PeerMapService: pms, meta: receivedMeta, addr: receivedMeta.ToPeerAddress(), lCheckTime: now}
		pms.Logger.Info().Str("meta", receivedMeta.String()).Msg("Registering new peer info")
		pms.peerRegistry[peerID] = newState
	} else {
		if prev.meta != receivedMeta {
			pms.Logger.Info().Str("meta", prev.meta.String()).Msg("Replacing previous peer info")
			prev.meta = receivedMeta
			prev.addr = receivedMeta.ToPeerAddress()
		}
		prev.lCheckTime = now
	}
	return nil
}

func (pms *PeerMapService) unregisterPeer(peerID peer.ID) {
	pms.rwmutex.Lock()
	defer pms.rwmutex.Unlock()
	pms.Logger.Info().Str(p2p.LogPeerID, p2putil.ShortForm(peerID)).Msg("Unregistering bad peer")
	delete(pms.peerRegistry, peerID)

}

func (pms *PeerMapService) writeResponse(reqContainer p2pcommon.Message, meta p2pcommon.PeerMeta, resp *types.MapResponse, wt p2p.MsgWriter) error {
	msgID := p2pcommon.NewMsgID()
	respMsg, err := createV030Message(msgID, reqContainer.ID(), MapResponse, resp)
	if err != nil {
		return err
	}

	return wt.WriteMsg(respMsg)
}

// TODO code duplication. it can result in a bug.
func createV030Message(msgID, orgID p2pcommon.MsgID, subProtocol p2pcommon.SubProtocol, innerMsg proto.Message) (*p2p.V030Message, error) {
	bytes, err := p2p.MarshalMessage(innerMsg)
	if err != nil {
		return nil, err
	}

	msg := p2p.NewV030Message(msgID, orgID, time.Now().UnixNano(), subProtocol, bytes)
	return msg, nil
}

func (pms *PeerMapService) Receive(context actor.Context) {
	rawMsg := context.Message()
	switch msg := rawMsg.(type) {
	case *message.CurrentListMsg:
		pms.Logger.Debug().Msg("Got current message")
		context.Respond(pms.getCurrentPeers(msg))
	case *message.WhiteListMsg:
		pms.Logger.Debug().Msg("Got whitelist message")
		context.Respond(pms.getWhiteList(msg))
	case *message.BlackListMsg:
		pms.Logger.Debug().Msg("Got blacklist message")
		context.Respond(pms.getBlackList(msg))
	default:
		pms.Logger.Debug().Interface("msg", msg) // TODO: temporal code for resolve compile error
	}
}

func (pms *PeerMapService) onPing(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	pms.Logger.Debug().Str(p2p.LogPeerID, peerID.String()).Msg("Received ping from polaris (maybe)")

	rw := p2p.NewV030ReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	defer s.Close()

	req, err := rw.ReadMsg()
	if err != nil {
		return
	}
	pingReq := &types.Ping{}
	err = p2p.UnmarshalMessage(req.Payload(), pingReq)
	if err != nil {
		return
	}
	// TODO: check if sender is known polaris or peer and it not, ban or write to blacklist .
	pingResp := &types.Ping{}
	msgID := p2pcommon.NewMsgID()
	respMsg, err := createV030Message(msgID, req.ID(), p2p.PingResponse, pingResp)
	if err != nil {
		return
	}

	err = rw.WriteMsg(respMsg)
	if err != nil {
		return
	}

}

func (pms *PeerMapService) getCurrentPeers(param *message.CurrentListMsg) *types.PolarisPeerList {
	retSize := int(param.Size)
	totalSize := len(pms.peerRegistry)
	listSize := calcMinimum(retSize, totalSize, ResponseMaxPeerLimit)
	pList := make([]*types.PolarisPeer, listSize)
	addSize := 0
	pms.rwmutex.Lock()
	pms.rwmutex.Unlock()
	for _, rPeer := range pms.peerRegistry {
		pList[addSize] = &types.PolarisPeer{Address: &rPeer.addr, Connected: rPeer.connected.UnixNano(), LastCheck: rPeer.lastCheck().UnixNano()}
		addSize++
		if addSize >= listSize {
			break
		}
	}
	if addSize < listSize {
		pList = pList[:addSize]
	}
	result := &types.PolarisPeerList{Peers: pList, HasNext: false, Total: uint32(totalSize)}
	return result
}

func (pms *PeerMapService) getWhiteList(param *message.WhiteListMsg) *types.PolarisPeerList {
	// TODO implement!
	return &types.PolarisPeerList{}
}

func (pms *PeerMapService) getBlackList(param *message.BlackListMsg) *types.PolarisPeerList {
	// TODO implement!
	return &types.PolarisPeerList{}
}

func calcMinimum(values ...int) int {
	min := math.MaxUint32
	for _, val := range values {
		if min > val {
			min = val
		}
	}
	return min
}
func (pms *PeerMapService) getPeerCheckers() []peerChecker {
	pms.rwmutex.Lock()
	pms.rwmutex.Unlock()
	newSlice := make([]peerChecker, 0, len(pms.peerRegistry))
	for _, rPeer := range pms.peerRegistry {
		newSlice = append(newSlice, rPeer)
	}
	return newSlice
}

func makeGoAwayMsg(message string) (p2pcommon.Message, error) {
	awayMsg := &types.GoAwayNotice{Message: message}
	msgID := p2pcommon.NewMsgID()
	return createV030Message(msgID, EmptyMsgID, p2p.GoAway, awayMsg)
}

// send notice message and then disconnect. this routine should only run in RunPeer go routine
func (pms *PeerMapService) SendGoAwayMsg(message string, wt p2p.MsgWriter) error {
	msg, err := makeGoAwayMsg(message)
	if err != nil {
		return err
	}
	wt.WriteMsg(msg)
	time.Sleep(MsgSendDelay)
	return nil
}

//
func (pms *PeerMapService) checkChain(chainIDBytes []byte) (bool, error) {
	chainID := types.NewChainID()
	if err := chainID.Read(chainIDBytes); err != nil {
		return false, err
	}
	sameChain := pms.ntc.ChainID().Equals(chainID)
	if !sameChain && pms.Logger.IsDebugEnabled() {
		pms.Logger.Debug().Str("chain_id", chainID.ToJSON()).Msg("chainid differ")

	}
	return sameChain, nil
}

func (pms *PeerMapService) checkConnectness(meta p2pcommon.PeerMeta) bool {
	if !pms.allowPrivate && !p2putil.IsExternalAddr(meta.IPAddress) {
		pms.Logger.Debug().Str("peer_meta", meta.String()).Msg("peer is private address")
		return false
	}
	tempState := &peerState{PeerMapService: pms, meta: meta, addr: meta.ToPeerAddress(), lCheckTime: time.Now(), temporary: true}
	_, err := tempState.checkConnect(PolarisPingTTL)
	if err != nil {
		pms.Logger.Debug().Err(err).Str(p2p.LogPeerID, p2putil.ShortForm(meta.ID)).Msg("Ping check was failed.")
		return false
	} else {
		pms.Logger.Debug().Str(p2p.LogPeerID, p2putil.ShortForm(meta.ID)).Msg("Ping check is succeeded.")
		return true
	}
}
