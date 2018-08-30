/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
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
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/multiformats/go-multicodec/protobuf"
)

const defaultPingInterval = time.Second * 60

// RemotePeer represent remote peer to which is connected
type RemotePeer struct {
	logger       *log.Logger
	pingDuration time.Duration

	meta      PeerMeta
	state     types.PeerState
	actorServ ActorService
	pm        PeerManager
	stopChan  chan struct{}

	write      chan msgOrder
	closeWrite chan struct{}

	hsLock *sync.Mutex

	// used to access request data from response handlers
	requests    map[string]msgOrder
	consumeChan chan string

	handlers map[SubProtocol]MessageHandler

	// TODO make automatic disconnect if remote peer cause too many wrong message

	blkHashCache *lru.Cache

	rw *bufio.ReadWriter
}

// msgOrder is abstraction information about the message that will be sent to peer
type msgOrder interface {
	GetRequestID() string
	Timestamp() int64
	IsRequest() bool
	IsGossip() bool
	IsNeedSign() bool
	// ResponseExpected means that remote peer is expected to send response to this request.
	ResponseExpected() bool
	GetProtocolID() SubProtocol
	SignWith(pm PeerManager) error
	SendOver(rw *bufio.ReadWriter) error
}

const (
	cleanRequestDuration = time.Hour
)

// newRemotePeer create an object which represent a remote peer.
func newRemotePeer(meta PeerMeta, p2ps PeerManager, iServ ActorService, log *log.Logger) *RemotePeer {
	peer := &RemotePeer{
		meta: meta, pm: p2ps, actorServ: iServ, logger: log,
		pingDuration: defaultPingInterval,
		state:        types.STARTING,

		stopChan:   make(chan struct{}),
		write:      make(chan msgOrder),
		closeWrite: make(chan struct{}),
		hsLock:     &sync.Mutex{},

		requests:    make(map[string]msgOrder),
		consumeChan: make(chan string, 10),

		handlers: make(map[SubProtocol]MessageHandler),
	}

	var err error
	peer.blkHashCache, err = lru.New(DefaultPeerInvCacheSize)
	if err != nil {
		panic("Failed to create remotepeer " + err.Error())
	}
	return peer
}

// ID return id of peer, same as peer.meta.ID
func (p *RemotePeer) ID() peer.ID {
	return p.meta.ID
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
	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Starting peer")
	pingTicker := time.NewTicker(p.pingDuration)
	go p.runWrite()
	go p.runRead()
READNOPLOOP:
	for {
		select {
		case <-pingTicker.C:
			p.sendPing()
		case <-p.stopChan:
			p.setState(types.STOPPED)
			break READNOPLOOP
		}
	}
	p.logger.Info().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Finishing peer")
	pingTicker.Stop()

	// send channel twice. one for read and another for write
	p.closeWrite <- struct{}{}
	close(p.stopChan)
}

func (p *RemotePeer) runWrite() {
	cleanupTicker := time.NewTicker(cleanRequestDuration)
	defer func() {
		if r := recover(); r != nil {
			p.logger.Panic().Str("recover", fmt.Sprint(r)).Msg("There were panic in runWrite ")
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
			p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Quitting runWrite")
			break WRITELOOP
		}
	}
	cleanupTicker.Stop()

	// closing channel is to golang runtime
	// close(p.write)
	// close(p.consumeChan)
}

func (p *RemotePeer) runRead() {
	for {
		msg, err := p.readMsg()
		if err != nil {
			p.logger.Error().Err(err).Msg("Failed to read message")
			p.pm.RemovePeer(p.ID())
			return
		}

		if err = p.handleMsg(msg); err != nil {
			p.logger.Error().Err(err).Msg("Failed to handle message")
			p.pm.RemovePeer(p.ID())
			return
		}
	}

	// closing channel is to golang runtime
	// close(p.write)
	// close(p.consumeChan)
}

func (p *RemotePeer) readMsg() (*types.P2PMessage, error) {
	data := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(p.rw)
	err := decoder.Decode(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *RemotePeer) handleMsg(msg *types.P2PMessage) error {
	var err error
	proto := SubProtocol(msg.Header.Subprotocol)
	defer func() {
		if r := recover(); r != nil {
			p.logger.Warn().Str("panic", fmt.Sprint(r)).Msg("There were panic in handler")
			err = fmt.Errorf("internal error")
		}
	}()
	p.logger.Debug().Str(LogPeerID, p.ID().Pretty()).Str("protocol", proto.String()).Msg("Handling messge")

	handler, found := p.handlers[proto]
	if !found {
		return fmt.Errorf("invalid protocol %s", proto)
	}
	handler(msg)
	return err
}

// Stop stops aPeer works
func (p *RemotePeer) stop() {
	p.stopChan <- struct{}{}
}

const writeChannelTimeout = time.Second * 2

func (p *RemotePeer) sendMessage(msg msgOrder) {
	select {
	case p.write <- msg:
		return
	case <-time.After(writeChannelTimeout):
		p.logger.Warn().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogMsgID, msg.GetRequestID()).Str(LogProtoID, msg.GetProtocolID().String()).Msg("Peer too busy or deadlock, stalled message is dropped")
	}
}

// consumeRequest remove request from request history.
func (p *RemotePeer) consumeRequest(requestID string) {
	p.consumeChan <- requestID
}

func (p *RemotePeer) updateMetaInfo(statusMsg *types.Status) {
	// check address. and apply current
	receivedMeta := FromPeerAddress(statusMsg.Sender)
	p.meta.IPAddress = receivedMeta.IPAddress
	p.meta.Port = receivedMeta.Port
}

func (p *RemotePeer) writeToPeer(m msgOrder) {
	// check peer's status
	// TODO code smell. hardcoded check and need memory barrier for peer state
	if p.State() != types.RUNNING {
		p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, m.GetProtocolID().String()).
			Str(LogMsgID, m.GetRequestID()).Str("peer_state", p.State().String()).Msg("Cancel sending messge, since peer is not running state")
		return
	}

	// sign the data
	// TODO signing can be done earlier. Consider change signing point to reduce cpu load
	if m.IsNeedSign() {
		err := m.SignWith(p.pm)
		if err != nil {
			p.logger.Warn().Err(err).Msg("fail to sign")
			return
		}
	}

	err := m.SendOver(p.rw)
	if err != nil {
		p.logger.Warn().Err(err).Msg("fail to SendOver")
		return
	}
	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, m.GetProtocolID().String()).
		Str(LogMsgID, m.GetRequestID()).Msg("Send message")
	//p.logger.Debugf("Sent message %v:%v to peer %s", m.GetProtocolID(), m.GetRequestID(), p.meta.ID.Pretty())
	if m.ResponseExpected() {
		p.requests[m.GetRequestID()] = m
	}
}

func (p *RemotePeer) tryGetStream(msgID string, protocol protocol.ID, timeout time.Duration) inet.Stream {
	streamChannel := make(chan inet.Stream)
	var s inet.Stream = nil
	go p.getStreamForWriting(msgID, protocol, streamChannel)
	select {
	case s = <-streamChannel:
		return s
	case <-time.After(timeout):
		p.logger.Warn().Str(LogMsgID, msgID).Msg("stream get timeout")
	}
	return s
}

func (p *RemotePeer) getStreamForWriting(msgID string, protocol protocol.ID, schannel chan inet.Stream) {
	ctx := context.Background()
	s, err := p.pm.NewStream(ctx, p.meta.ID, protocol)
	if err != nil {
		p.logger.Warn().Err(err).Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, string(protocol)).Msg("Error while get stream")
		schannel <- nil
	}
	schannel <- s
}

// this method MUST be called in same go routine as AergoPeer.RunPeer()
func (p *RemotePeer) sendPing() {
	// find my best block
	bestBlock, err := extractBlockFromRequest(p.actorServ.CallRequest(message.ChainSvc, &message.GetBestBlock{}))
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to get best block")
		return
	}
	// create message data
	pingMsg := &types.Ping{
		MessageData:   &types.MessageData{},
		BestBlockHash: bestBlock.BlockHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	p.sendMessage(newPbMsgRequestOrder(true, false, pingRequest, pingMsg))
}

// sendStatus is called once when a peer is added.()
func (p *RemotePeer) sendStatus() {
	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Sending status message for handshaking")

	// create message data
	statusMsg, err := createStatusMsg(p.pm, p.actorServ)
	if err != nil {
		p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Err(err).Msg("Cancel sending status")
		return
	}

	p.sendMessage(newPbMsgRequestOrder(false, true, statusRequest, statusMsg))
}

// send notice message and then disconnect. this routine should only run in RunPeer go routine
func (p *RemotePeer) goAwayMsg(msg string) {
	p.logger.Info().Str(LogPeerID, p.meta.ID.Pretty()).Str("msg", msg).Msg("Peer is closing")
	p.sendMessage(newPbMsgRequestOrder(false, true, goAway, &types.GoAwayNotice{MessageData: &types.MessageData{}, Message: msg}))
	p.pm.RemovePeer(p.meta.ID)
}

func (p *RemotePeer) pruneRequests() {
	debugLog := p.logger.IsDebugEnabled()
	deletedCnt := 0
	var deletedReqs []string
	expireTime := time.Now().Add(-1 * time.Hour).Unix()
	for key, m := range p.requests {
		if m.Timestamp() < expireTime {
			delete(p.requests, key)
			if debugLog {
				deletedReqs = append(deletedReqs, m.GetProtocolID().String()+"/"+key+time.Unix(m.Timestamp(), 0).String())
			}
			deletedCnt++
		}
	}
	//p.logger.Infof("Pruned %d requests but no response to peer %s until %v", deletedCnt, p.meta.ID.Pretty(), time.Unix(expireTime, 0))
	p.logger.Info().Int("count", deletedCnt).Str(LogPeerID, p.meta.ID.Pretty()).
		Time("until", time.Unix(expireTime, 0)).Msg("Pruned requests, but no response to peer")
	//.Msg("Pruned %d requests but no response to peer %s until %v", deletedCnt, p.meta.ID.Pretty(), time.Unix(expireTime, 0))
	if debugLog {
		p.logger.Debug().Msg(strings.Join(deletedReqs[:], ","))
	}

}

func (p *RemotePeer) handleNewBlockNotice(data *types.NewBlockNotice) {
	// lru cache can accept hashable key
	b64hash := enc.ToString(data.BlockHash)

	p.blkHashCache.Add(b64hash, data.BlockHash)
	p.pm.HandleNewBlockNotice(p.meta.ID, b64hash, data)
}

func (p *RemotePeer) sendGoAway(msg string) {
	// TODO: send goaway message and close connection
}
