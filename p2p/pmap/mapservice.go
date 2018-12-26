/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package pmap

import (
	"bufio"
	"fmt"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
	"sync"
	"time"
)

// subprotocol for polaris
const (
	PolarisMapSub  protocol.ID = "/polaris/0.1"
	PolarisPingSub protocol.ID = "/ping/0.1"
)
const (
	MapQuery p2p.SubProtocol = 0x0100 + iota
	MapResponse
)

// internal
const (
	PolarisConnectionTTL = time.Second * 30

	// polaris will return peers list at most this number
	ResponseMaxPeerLimit = 500
	// libp2p internal library is not always send message instantly, so closing socket soon after sent a message will cause packet loss and read error, us walkaround here till finding the real reason and fix it.
	MsgSendDelay = time.Second * 1

	PeerHealthcheckInterval = time.Minute
	//PeerHealthcheckInterval = time.Minute * 5
	ConcurrentHealthCheckCount = 20
)

var (
	EmptyMsgID = p2p.MsgID(uuid.Nil)
)

var (
	// 89.15 is floor of declination of Polaris
	MainnetMapServer = []string{
		"/dns/polaris.aergo.io/tcp/8915/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF",
	}

	// 89.16 is ceiling of declination of Polaris
	TestnetMapServer = []string{
		"/dns/polaris.aergo.io/tcp/8916/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF",
	}
)

type mapService interface {
	getPeerCheckers() []peerChecker
	registerPeer(receivedMeta p2p.PeerMeta) error
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

	ChainID    []byte
	PrivateNet bool

	mapServers []p2p.PeerMeta

	ntc    p2p.NTContainer
	listen bool
	nt     p2p.NetworkTransport
	hc     HealthCheckManager

	rwmutex *sync.RWMutex
	peerRegistry map[peer.ID]*peerState
}

func NewMapServiceCli(cfg *config.P2PConfig, ntc p2p.NTContainer) *PeerMapService {
	return NewMapService(cfg, ntc, false)
}
func NewMapService(cfg *config.P2PConfig, ntc p2p.NTContainer, listen bool) *PeerMapService {
	pms := &PeerMapService{
		rwmutex:      &sync.RWMutex{},
		peerRegistry: make(map[peer.ID]*peerState),
		PrivateNet:   cfg.NPPrivateNet,
		listen:       listen,
	}
	pms.BaseComponent = component.NewBaseComponent(message.MapSvc, pms, log.NewLogger("map"))


	// init
	pms.ntc = ntc
	pms.initializeMapServers(cfg)
	pms.hc= NewHCM(pms, pms.nt)
	// initialize map Servers
	return pms
}

func (pms *PeerMapService) initializeMapServers(cfg *config.P2PConfig) {
	if cfg.NPUsePolaris {
		// private network does not use public polaris
		if !pms.PrivateNet {
			// TODO select default built-in servers
			servers := TestnetMapServer
			for _, addrStr := range servers {
				meta, err := p2p.FromMultiAddrString(addrStr)
				if err != nil {
					pms.Logger.Info().Str("addr_str", addrStr).Msg("invalid polaris server address in base setting ")
					continue
				}
				pms.mapServers = append(pms.mapServers, meta)
			}
		}
		for _, addrStr := range cfg.NPAddPolarises {
			meta, err := p2p.FromMultiAddrString(addrStr)
			if err != nil {
				pms.Logger.Info().Str("addr_str", addrStr).Msg("invalid polaris server address in config file ")
				continue
			}
			pms.mapServers = append(pms.mapServers, meta)
		}

		if len(pms.mapServers) == 0 {
			pms.Logger.Warn().Msg("no active polaris server found. node discovery by polaris is disabled")
		}
	} else {
		pms.Logger.Info().Msg("node discovery by polaris is disabled configuration.")
	}
}

func (pms *PeerMapService) BeforeStart() {}

func (pms *PeerMapService) AfterStart() {
	pms.nt = pms.ntc.GetNetworkTransport()
	if pms.listen {
		pms.nt.AddStreamHandler(PolarisMapSub, pms.onConnect)
		pms.hc.Start()
	}
	pms.nt.AddStreamHandler(PolarisPingSub, pms.onPing)

}

func (pms *PeerMapService) BeforeStop() {
	pms.nt.RemoveStreamHandler(PolarisPingSub)
	if pms.listen {
		if pms.nt != nil {
			pms.hc.Stop()
			pms.nt.RemoveStreamHandler(PolarisMapSub)
		}
	}
}

func (pms *PeerMapService) Statistics() *map[string]interface{} {
	return nil
	//dummy := make(map[string]interface{})
	//return &dummy
}

func (pms *PeerMapService) onConnect(s net.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeerMeta := p2p.PeerMeta{ID: peerID}
	pms.Logger.Debug().Str(p2p.LogPeerID, peerID.String()).Msg("Received map query")

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
	pms.Logger.Debug().Str(p2p.LogPeerID, peerID.String()).Int("peer_cnt", len(resp.Addresses)).Msg("Sent map response")

	// TODO send goodbye message.
	time.Sleep(time.Second * 3)

	// disconnect!
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer setup some
func (pms *PeerMapService) readRequest(meta p2p.PeerMeta, rd p2p.MsgReader) (p2p.Message, *types.MapQuery, error) {
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

func (pms *PeerMapService) handleQuery(container p2p.Message, query *types.MapQuery) (*types.MapResponse, error) {
	if query.Status == nil {
		return nil, fmt.Errorf("malformed query %v", query)
	}
	receivedMeta := p2p.FromPeerAddress(query.Status.Sender)
	maxPeers := int(query.Size)
	if maxPeers <= 0 {
		return nil, fmt.Errorf("invalid argument count %d", maxPeers)
	} else if maxPeers > ResponseMaxPeerLimit {
		pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Int("req_size", maxPeers).Int("clipped", ResponseMaxPeerLimit).Msg("Clipping too high count of query ")
		maxPeers = ResponseMaxPeerLimit
	}

	pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Msg("Handling query.")
	// TODO check more varification or request peer
	// must check peer is really capable to aergosvr

	// make response
	resp := &types.MapResponse{}

	resp.Addresses = pms.retrieveList(maxPeers, receivedMeta.ID)

	if query.AddMe {
		pms.Logger.Debug().Str(p2p.LogPeerID, receivedMeta.ID.String()).Msg("AddMe is set, and register peer to peer registry")
		pms.registerPeer(receivedMeta)
	}

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

func (pms *PeerMapService) registerPeer(receivedMeta p2p.PeerMeta) error {
	peerID := receivedMeta.ID
	pms.rwmutex.Lock()
	defer pms.rwmutex.Unlock()
	prev, ok := pms.peerRegistry[peerID]
	if !ok {
		newState := &peerState{PeerMapService:pms, meta: receivedMeta, addr: receivedMeta.ToPeerAddress(), lCheckTime: time.Now()}
		pms.peerRegistry[peerID] = newState
	} else {
		pms.Logger.Info().Str("meta", prev.meta.String()).Msg("Replacing previous peer info")
		newState := &peerState{PeerMapService:pms, meta: receivedMeta, addr: receivedMeta.ToPeerAddress(), lCheckTime: time.Now()}
		pms.peerRegistry[peerID] = newState
	}
	return nil
}

func (pms *PeerMapService) unregisterPeer(peerID peer.ID) {
	pms.rwmutex.Lock()
	defer pms.rwmutex.Unlock()
	pms.Logger.Info().Str(p2p.LogPeerID, peerID.Pretty()).Msg("Unregister bad peer")
	delete(pms.peerRegistry, peerID)

}

func (pms *PeerMapService) writeResponse(reqContainer p2p.Message, meta p2p.PeerMeta, resp *types.MapResponse, wt p2p.MsgWriter) error {
	msgID := p2p.NewMsgID()
	respMsg, err := createV030Message(msgID, reqContainer.ID(), MapResponse, resp)
	if err != nil {
		return err
	}

	return wt.WriteMsg(respMsg)
}

// TODO code duplication. it can result in a bug.
func createV030Message(msgID, orgID p2p.MsgID, subProtocol p2p.SubProtocol, innerMsg proto.Message) (*p2p.V030Message, error) {
	bytes, err := p2p.MarshalMessage(innerMsg)
	if err != nil {
		return nil, err
	}

	msg := p2p.NewV030Message(msgID, orgID, time.Now().Unix(), subProtocol, bytes)
	return msg, nil
}

func (pms *PeerMapService) Receive(context actor.Context) {
	rawMsg := context.Message()
	switch msg := rawMsg.(type) {
	case *message.MapQueryMsg:
		pms.Hub().Tell(message.P2PSvc, pms.queryPeers(msg))
	default:
		//		pms.Logger.Debug().Interface("msg", msg) // TODO: temporal code for resolve compile error
	}
}

func (pms *PeerMapService) queryPeers(msg *message.MapQueryMsg) *message.MapQueryRsp {
	succ := 0
	resultPeers := make([]*types.PeerAddress, 0, msg.Count)
	for _, meta := range pms.mapServers {
		addrs, err := pms.connectAndQuery(meta, msg.BestBlock.Hash, msg.BestBlock.Header.BlockNo)
		if err != nil {
			pms.Logger.Warn().Err(err).Str("map_id", meta.ID.Pretty()).Msg("faild to get peer addresses")
			continue
		}
		// FIXME delete duplicated peers
		resultPeers = append(resultPeers, addrs...)
		succ++
	}
	err := error(nil)
	if succ == 0 {
		err = fmt.Errorf("all servers of polaris are down")
	}
	pms.Logger.Debug().Int("peer_cnt", len(resultPeers)).Msg("Got map response and send back")
	resp := &message.MapQueryRsp{Peers: resultPeers, Err: err}
	return resp
}

func (pms *PeerMapService) connectAndQuery(mapServerMeta p2p.PeerMeta, bestHash []byte, bestHeight uint64) ([]*types.PeerAddress, error) {
	s, err := pms.nt.GetOrCreateStreamWithTTL(mapServerMeta, PolarisMapSub, PolarisConnectionTTL)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	if peerID != mapServerMeta.ID {
		return nil, fmt.Errorf("internal error peerid mismatch, exp %s, actual %s", mapServerMeta.ID.Pretty(), peerID.Pretty())
	}
	pms.Logger.Debug().Str(p2p.LogPeerID, peerID.String()).Msg("Sending map query")

	rw := p2p.NewV030ReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	peerAddress := pms.nt.SelfMeta().ToPeerAddress()
	peerStatus := &types.Status{Sender: &peerAddress, BestBlockHash: bestHash, BestHeight: bestHeight}
	// receive input
	err = pms.sendRequest(peerStatus, mapServerMeta, true, 100, rw)
	if err != nil {
		return nil, err
	}
	_, resp, err := pms.readResponse(mapServerMeta, rw)
	if err != nil {
		pms.SendGoAwayMsg(err.Error(), rw)
		return nil, err
	}
	if resp.Status == types.ResultStatus_OK {
		pms.SendGoAwayMsg("response is jot ok", rw)
		return resp.Addresses, nil
	}
	return nil, fmt.Errorf("remote error %s", resp.Status.String())
}

func (pms *PeerMapService) sendRequest(status *types.Status, mapServerMeta p2p.PeerMeta, register bool, size int, wt p2p.MsgWriter) error {
	msgID := p2p.NewMsgID()
	queryReq := &types.MapQuery{Status: status, Size: int32(size), AddMe: register, Excludes: [][]byte{[]byte(mapServerMeta.ID)}}
	respMsg, err := createV030Message(msgID, EmptyMsgID, MapQuery, queryReq)
	if err != nil {
		return err
	}

	return wt.WriteMsg(respMsg)
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer setup some
func (pms *PeerMapService) readResponse(mapServerMeta p2p.PeerMeta, rd p2p.MsgReader) (p2p.Message, *types.MapResponse, error) {
	data, err := rd.ReadMsg()
	if err != nil {
		return nil, nil, err
	}
	queryResp := &types.MapResponse{}
	err = p2p.UnmarshalMessage(data.Payload(), queryResp)
	if err != nil {
		return data, nil, err
	}
	pms.Logger.Debug().Str(p2p.LogPeerID, mapServerMeta.ID.String()).Int("peer_cnt", len(queryResp.Addresses)).Msg("Received map query response")

	return data, queryResp, nil
}

func (pms *PeerMapService) onPing(s net.Stream) {
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
	msgID := p2p.NewMsgID()
	respMsg, err := createV030Message(msgID, req.ID(), p2p.PingResponse, pingResp)
	if err != nil {
		return
	}

	err = rw.WriteMsg(respMsg)
	if err != nil {
		return
	}

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

func makeGoAwayMsg(message string) (p2p.Message, error) {
	awayMsg := &types.GoAwayNotice{Message: message}
	msgID := p2p.NewMsgID()
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
