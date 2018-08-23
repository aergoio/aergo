/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-protocol"
)

const defaultPingInterval = time.Second * 60

// RemotePeer represent remote peer to which is connected
type RemotePeer struct {
	log          *log.Logger
	pingDuration time.Duration

	meta      PeerMeta
	state     types.PeerState
	actorServ ActorService
	ps        PeerManager
	stopChan  chan struct{}

	write      chan msgOrder
	closeWrite chan struct{}
	read       chan readMsg
	op         chan OpOrder

	hsLock   *sync.Mutex
	readLock dummyMutex

	// used to access request data from response handlers
	requests    map[string]msgOrder
	consumeChan chan string

	sentStatus, gotStatus bool
	failCounter           uint32

	blkHashCache *lru.Cache
}

type dummyMutex struct{}

func (d *dummyMutex) Lock()   {}
func (d *dummyMutex) Unlock() {}

// msgOrder is abstraction information about the message that will be sent to peer
type msgOrder interface {
	GetRequestID() string
	Timestamp() int64
	IsRequest() bool
	IsGossip() bool
	IsNeedSign() bool
	// ResponseExpected means that remote peer is expected to send response to this request.
	ResponseExpected() bool
	GetProtocolID() protocol.ID
	SignWith(ps PeerManager) error
	SendOver(s inet.Stream) error
}

type readMsg struct {
	in inet.Stream
}

type OpType int

// Op series are for asking to process operation about remote peer.
const (
	// OpInitHS initiate handshaking, sending status message to remote peer
	OpInitHS OpType = iota
	// OpHandleHS handle status message from remote peer.
	OpHandleHS
	// OpStop stops peer
	OpStop
)

type OpOrder struct {
	op     OpType
	param1 interface{}
	param2 interface{}
}

const (
	cleanRequestDuration = time.Hour
)

// newRemotePeer create an object which represent a remote peer.
func newRemotePeer(meta PeerMeta, p2ps PeerManager, iServ ActorService, log *log.Logger) *RemotePeer {
	peer := &RemotePeer{
		meta: meta, ps: p2ps, actorServ: iServ, log: log,
		pingDuration: defaultPingInterval,
		state:        types.STARTING,

		stopChan:   make(chan struct{}),
		write:      make(chan msgOrder),
		closeWrite: make(chan struct{}),
		hsLock:     &sync.Mutex{},
		op:         make(chan OpOrder, 20),
		read:       make(chan readMsg),

		requests:    make(map[string]msgOrder),
		consumeChan: make(chan string, 10),
	}

	var err error
	peer.blkHashCache, err = lru.New(DefaultPeerInvCacheSize)
	if err != nil {
		panic("Failed to create remotepeer " + err.Error())
	}
	return peer
}

// State returns current state of peer
func (p *RemotePeer) State() types.PeerState {
	return p.state.Get()
}

func (p *RemotePeer) setState(newState types.PeerState) {
	p.state.SetAndGet(newState)
}

func (p *RemotePeer) checkState() error {
	switch p.State() {
	case types.HANDSHAKING:
		return fmt.Errorf("not handshaked")
	case types.STOPPED:
		return fmt.Errorf("peer stopped")
	default:
		return nil
	}
}

// runPeer should be called by go routine
func (p *RemotePeer) runPeer() {
	p.log.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Starting peer")
	pingTicker := time.NewTicker(p.pingDuration)
	go p.runWrite()
READNOPLOOP:
	for {
		select {
		case <-pingTicker.C:
			p.sendPing()
		case op := <-p.op:
			p.processOp(op)
		// case hsMsg := <-p.hsChan:
		// 	p.startHandshake(hsMsg)
		case <-p.stopChan:
			p.setState(types.STOPPED)
			break READNOPLOOP
		}
	}
	p.log.Info().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Finishing peer")
	pingTicker.Stop()
	p.closeWrite <- struct{}{}
	close(p.stopChan)
}

func (p *RemotePeer) runWrite() {
	cleanupTicker := time.NewTicker(cleanRequestDuration)
	defer func() {
		if r := recover(); r != nil {
			p.log.Panic().Str("recover", fmt.Sprint(r)).Msg("There were panic in runWrite ")
		}
	}()

WRITELOOP:
	for {
		select {
		case m := <-p.write:
			p.writeToPeer(m)
		case rID := <-p.consumeChan:
			delete(p.requests, rID)
		case <-cleanupTicker.C:
			p.pruneRequests()
		case <-p.closeWrite:
			p.log.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Quitting runWrite")
			break WRITELOOP
		}
	}
	cleanupTicker.Stop()
	close(p.write)
	close(p.consumeChan)
}

func (p *RemotePeer) processOp(op OpOrder) {
	switch op.op {
	case OpInitHS:
		p.initiateHandshake()
	case OpHandleHS:
		p.handleHandshake(op.param1.(*types.Status))
	case OpStop:
		// do stop
	}
}

// Stop stops aPeer works
func (p *RemotePeer) stop() {
	p.stopChan <- struct{}{}
}

func (p *RemotePeer) sendMessage(msg msgOrder) {
	p.write <- msg
}

// consumeRequest remove request from request history.
func (p *RemotePeer) consumeRequest(requestID string) {
	p.consumeChan <- requestID
}

func (p *RemotePeer) initiateHandshake() {
	// FIXME change read operations and then remove it
	p.hsLock.Lock()
	defer p.hsLock.Unlock()

	if p.State() != types.STARTING {
		p.goAwayMsg("Invalid status msg")
		return
	}
	p.sendStatus()
	p.setState(types.HANDSHAKING)
}

func (p *RemotePeer) updateMetaInfo(statusMsg *types.Status) {
	// check address. and apply current
	receivedMeta := FromPeerAddress(statusMsg.Sender)
	p.meta.IPAddress = receivedMeta.IPAddress
	p.meta.Port = receivedMeta.Port
}

// startHandshake is run only in AergoPeer.RunPeer go routine
func (p *RemotePeer) handleHandshake(statusMsg *types.Status) {
	// FIXME change read operations and then remove it
	p.hsLock.Lock()
	defer p.hsLock.Unlock()

	peerState := p.State()
	if peerState != types.STARTING && peerState != types.HANDSHAKING {
		p.goAwayMsg("Invalid status msg")
		return
	}

	p.updateMetaInfo(statusMsg)
	// TODO: check protocol version, blacklist, key authentication or etc.
	err := p.checkProtocolVersion()
	if err != nil {
		p.log.Info().Str(LogPeerID, p.meta.ID.Pretty()).Msg("invalid protocol version of peer")
		p.goAwayMsg("Handshake failed")
		return
	}

	// if state is han
	if peerState == types.STARTING {
		p.sendStatus()
	}

	// If all checked and validated. it's now handshaked. and then run sync.
	p.log.Info().Str(LogPeerID, p.meta.ID.Pretty()).Msg("peer is handshaked and now running status")
	p.setState(types.RUNNING)

	// notice to p2pmanager that handshaking is finished
	p.ps.NotifyPeerHandshake(p.meta.ID)

	p.actorServ.SendRequest(message.ChainSvc, &message.SyncBlockState{PeerID: p.meta.ID, BlockNo: statusMsg.BestHeight, BlockHash: statusMsg.BestBlockHash})
}

func (p *RemotePeer) writeToPeer(m msgOrder) {
	// check peer's status
	// TODO code smell. hardcoded check and need memory barrier for peer state
	if m.GetProtocolID() != statusRequest && p.State() != types.RUNNING {
		p.log.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, string(m.GetProtocolID())).
			Str(LogMsgID, m.GetRequestID()).Str("peer_state", p.State().String()).Msg("Cancel sending messge, since peer is not running state")
		return
	}

	// sign the data
	// TODO signing can be done earlier. Consider change signing point to reduce cpu load
	if m.IsNeedSign() {
		err := m.SignWith(p.ps)
		if err != nil {
			p.log.Warn().Err(err).Msg("fail to sign")
			return
		}
	}

	s, err := p.ps.NewStream(context.Background(), p.meta.ID, m.GetProtocolID())
	if err != nil {
		p.log.Warn().Err(err).Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, string(m.GetProtocolID())).Str(LogMsgID, m.GetRequestID()).Msg("Error while sending")
		// problem in connection starting disconnect
		p.ps.RemovePeer(p.meta.ID)
		return
	}
	//defer s.Close()

	err = m.SendOver(s)
	if err != nil {
		p.log.Warn().Err(err).Msg("fail to SendOver")
		return
	}
	p.log.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, string(m.GetProtocolID())).
		Str(LogMsgID, m.GetRequestID()).Msg("Send message")
	//p.log.Debugf("Sent message %v:%v to peer %s", m.GetProtocolID(), m.GetRequestID(), p.meta.ID.Pretty())
	if m.ResponseExpected() {
		p.requests[m.GetRequestID()] = m
	}
}

// this method MUST be called in same go routine as AergoPeer.RunPeer()
func (p *RemotePeer) sendPing() {
	// find my best block
	bestBlock, err := extractBlockFromRequest(p.actorServ.CallRequest(message.ChainSvc, &message.GetBestBlock{}))
	if err != nil {
		p.log.Error().Err(err).Msg("Failed to get best block")
		return
	}
	// create message data
	pingMsg := &types.Ping{
		MessageData:   &types.MessageData{},
		BestBlockHash: bestBlock.GetHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	p.sendMessage(newPbMsgRequestOrder(true, false, pingRequest, pingMsg))
}

// sendStatus is called once when a peer is added.()
func (p *RemotePeer) sendStatus() {
	p.log.Info().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Sending status message for handshaking")

	// find my best block
	bestBlock, err := extractBlockFromRequest(p.actorServ.CallRequest(message.ChainSvc, &message.GetBestBlock{}))
	if err != nil {
		p.log.Error().Err(err).Msg("Failed to get best block")
		return
	}
	selfAddr := p.ps.SelfMeta().ToPeerAddress()
	// create message data
	statusMsg := &types.Status{
		MessageData:   &types.MessageData{},
		Sender:        &selfAddr,
		BestBlockHash: bestBlock.GetHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	p.sendMessage(newPbMsgRequestOrder(false, true, statusRequest, statusMsg))
}

// send notice message and then disconnect. this routine should only run in RunPeer go routine
func (p *RemotePeer) goAwayMsg(msg string) {
	p.log.Info().Str(LogPeerID, p.meta.ID.Pretty()).Str("msg", msg).Msg("Peer is closing")
	p.sendMessage(newPbMsgRequestOrder(false, true, goAway, &types.GoAwayNotice{MessageData: &types.MessageData{}, Message: msg}))
	p.ps.RemovePeer(p.meta.ID)
}

func (p *RemotePeer) checkProtocolVersion() error {
	// TODO modify interface and put check code here
	return nil
}

func (p *RemotePeer) pruneRequests() {
	debugLog := p.log.IsDebugEnabled()
	deletedCnt := 0
	var deletedReqs []string
	expireTime := time.Now().Add(-1 * time.Hour).Unix()
	for key, m := range p.requests {
		if m.Timestamp() < expireTime {
			delete(p.requests, key)
			if debugLog {
				deletedReqs = append(deletedReqs, string(m.GetProtocolID())+"/"+key+time.Unix(m.Timestamp(), 0).String())
			}
			deletedCnt++
		}
	}
	//p.log.Infof("Pruned %d requests but no response to peer %s until %v", deletedCnt, p.meta.ID.Pretty(), time.Unix(expireTime, 0))
	p.log.Info().Int("count", deletedCnt).Str(LogPeerID, p.meta.ID.Pretty()).
		Time("until", time.Unix(expireTime, 0)).Msg("Pruned requests, but no response to peer")
	//.Msg("Pruned %d requests but no response to peer %s until %v", deletedCnt, p.meta.ID.Pretty(), time.Unix(expireTime, 0))
	if debugLog {
		p.log.Debug().Msg(strings.Join(deletedReqs[:], ","))
	}

}

func (p *RemotePeer) handleNewBlockNotice(data *types.NewBlockNotice) {
	// lru cache can accept hashable key
	b64hash := enc.ToString(data.BlockHash)

	p.blkHashCache.Add(b64hash, data.BlockHash)
	p.ps.HandleNewBlockNotice(p.meta.ID, b64hash, data)
}
