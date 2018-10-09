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
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
)

const defaultPingInterval = time.Second * 60

type RemotePeer interface {
	ID() peer.ID
	Meta() PeerMeta
	State() types.PeerState

	runPeer()
	stop()

	sendMessage(msg msgOrder)
	pushTxsNotice(txHashes []TxHash)
	consumeRequest(msgID string)

	handleNewBlockNotice(data *types.NewBlockNotice)
	handleNewTxNotice(data *types.NewTransactionsNotice)

	// TODO
	MF() moFactory
}

// remotePeerImpl represent remote peer to which is connected
type remotePeerImpl struct {
	logger       *log.Logger
	pingDuration time.Duration

	meta      PeerMeta
	state     types.PeerState
	actorServ ActorService
	pm        PeerManager
	mf        moFactory
	signer    msgSigner

	stopChan chan struct{}

	write      p2putil.ChannelPipe
	closeWrite chan struct{}

	// used to access request data from response handlers
	requests    map[string]msgOrder
	consumeChan chan string

	handlers map[SubProtocol]MessageHandler

	// TODO make automatic disconnect if remote peer cause too many wrong message

	blkHashCache *lru.Cache
	txHashCache  *lru.Cache

	txQueueLock *sync.Mutex
	txNoticeQueue *p2putil.PressableQueue
	maxTxNoticeHashSize int

	rw MsgReadWriter
}

var _ RemotePeer = (*remotePeerImpl)(nil)

const (
	cleanRequestInterval = time.Hour
	txNoticeInterval     = time.Second * 5
)

// newRemotePeer create an object which represent a remote peer.
func newRemotePeer(meta PeerMeta, pm PeerManager, actor ActorService, log *log.Logger, mf moFactory, signer msgSigner, rw MsgReadWriter) *remotePeerImpl {
	peer := &remotePeerImpl{
		meta: meta, pm: pm, actorServ: actor, logger: log, mf: mf, signer: signer, rw: rw,
		pingDuration: defaultPingInterval,
		state:        types.STARTING,

		stopChan:   make(chan struct{}),
		closeWrite: make(chan struct{}),

		requests:    make(map[string]msgOrder),
		consumeChan: make(chan string, 10),

		handlers: make(map[SubProtocol]MessageHandler),

		txQueueLock: &sync.Mutex{},
		txNoticeQueue: p2putil.NewPressableQueue(DefaultPeerTxQueueSize),
		maxTxNoticeHashSize: DefaultPeerTxQueueSize,
	}
	peer.write = p2putil.NewDefaultChannelPipe(20, newHangresolver(peer, log))

	var err error
	peer.blkHashCache, err = lru.New(DefaultPeerInvCacheSize)
	if err != nil {
		panic("Failed to create remotepeer " + err.Error())
	}
	peer.txHashCache, err = lru.New(DefaultPeerInvCacheSize)
	if err != nil {
		panic("Failed to create remotepeer " + err.Error())
	}

	return peer
}

// ID return id of peer, same as peer.meta.ID
func (p *remotePeerImpl) ID() peer.ID {
	return p.meta.ID
}

func (p *remotePeerImpl) Meta() PeerMeta {
	return p.meta
}

func (p *remotePeerImpl) MF() moFactory {
	return p.mf
}

// State returns current state of peer
func (p *remotePeerImpl) State() types.PeerState {
	return p.state.Get()
}

func (p *remotePeerImpl) checkState() error {
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
func (p *remotePeerImpl) runPeer() {
	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Starting peer")
	pingTicker := time.NewTicker(p.pingDuration)

	go p.runWrite()
	p.write.Open()
	go p.runRead()

	txNoticeTicker := time.NewTicker(txNoticeInterval)
	defer txNoticeTicker.Stop()

	// peer state is changed to RUNNIG after all sub goroutine is ready, and to STOPPED before fll sub goroutine is stopped.
	p.state.SetAndGet(types.RUNNING)
READNOPLOOP:
	for {
		select {
		case <-pingTicker.C:
			// no operation for now
		case <- txNoticeTicker.C:
			p.trySendTxNotices()
		case <-p.stopChan:
			break READNOPLOOP
		}
	}

	p.logger.Info().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Finishing peer")
	p.state.SetAndGet(types.STOPPED)
	pingTicker.Stop()
	// finish goroutine write. read goroutine will be closed automatically when disconnect
	p.write.Close()
	p.closeWrite <- struct{}{}
	close(p.stopChan)
}

func (p *remotePeerImpl) runWrite() {
	cleanupTicker := time.NewTicker(cleanRequestInterval)
	defer func() {
		if r := recover(); r != nil {
			p.logger.Panic().Str("recover", fmt.Sprint(r)).Msg("There were panic in runWrite ")
		}
	}()

WRITELOOP:
	for {
		select {
		case m := <-p.write.Out():
			p.writeToPeer(m.(msgOrder))
			p.write.Done() <- m
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

func (p *remotePeerImpl) runRead() {
	for {
		msg, err := p.rw.ReadMsg()
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

}

func (p *remotePeerImpl) handleMsg(msg *types.P2PMessage) error {
	var err error
	proto := SubProtocol(msg.Header.Subprotocol)
	defer func() {
		if r := recover(); r != nil {
			p.logger.Warn().Interface("panic", r).Msg("There were panic in handler")
			err = fmt.Errorf("internal error")
		}
	}()

	handler, found := p.handlers[proto]
	if !found {
		p.logger.Debug().Str(LogPeerID, p.ID().Pretty()).Str(LogMsgID, msg.Header.GetId()).Str(LogProtoID, proto.String()).Msg("Invalid protocol")
		return fmt.Errorf("invalid protocol %s", proto)
	}
	payload, err := handler.parsePayload(msg.Data)
	if err != nil {
		p.logger.Warn().Err(err).Str(LogPeerID, p.ID().Pretty()).Str(LogMsgID, msg.Header.GetId()).Str(LogProtoID, proto.String()).Msg("Invalid message data")
		return fmt.Errorf("Invalid message data")
	}
	err = p.signer.verifyMsg(msg, p.meta.ID)
	if err != nil {
		p.logger.Warn().Err(err).Str(LogPeerID, p.ID().Pretty()).Str(LogMsgID, msg.Header.GetId()).Str(LogProtoID, proto.String()).Msg("Failed to check signature")
		return fmt.Errorf("Failed to check signature")
	}
	err = handler.checkAuth(msg, payload)
	if err != nil {
		p.logger.Warn().Err(err).Str(LogPeerID, p.ID().Pretty()).Str(LogMsgID, msg.Header.GetId()).Str(LogProtoID, proto.String()).Msg("Failed to authenticate message")
		return fmt.Errorf("Failed to authenticate message")
	}

	handler.handle(msg.Header, payload)
	return nil
}

// Stop stops aPeer works
func (p *remotePeerImpl) stop() {
	p.stopChan <- struct{}{}
}

const writeChannelTimeout = time.Second * 2

func (p *remotePeerImpl) sendMessage(msg msgOrder) {
	if p.state.Get() != types.RUNNING {
		p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, msg.GetProtocolID().String()).
			Str(LogMsgID, msg.GetMsgID()).Interface("peer_state", p.State()).Msg("Cancel sending messge, since peer is not running state")
		return
	}
	p.write.In() <- msg
}


func (p *remotePeerImpl) pushTxsNotice(txHashes []TxHash) {
	p.txQueueLock.Lock()
	defer p.txQueueLock.Unlock()
	for _,hash := range txHashes {
		p.txNoticeQueue.Press(hash)
	}
}



// consumeRequest remove request from request history.
func (p *remotePeerImpl) consumeRequest(requestID string) {
	p.consumeChan <- requestID
}

func (p *remotePeerImpl) updateMetaInfo(statusMsg *types.Status) {
	// check address. and apply current
	receivedMeta := FromPeerAddress(statusMsg.Sender)
	p.meta.IPAddress = receivedMeta.IPAddress
	p.meta.Port = receivedMeta.Port
}

func (p *remotePeerImpl) writeToPeer(m msgOrder) {
	m.SendTo(p)
}

func (p *remotePeerImpl) trySendTxNotices() {
	p.txQueueLock.Lock()
	defer p.txQueueLock.Unlock()
	// make create multiple message if queue is too many hashes.
	for {
		// no need to send if queue is empty
		if p.txNoticeQueue.Size() == 0 {
			return
		}
		hashes := make([][]byte,0,p.txNoticeQueue.Size())
		idx := 0
		for  element := p.txNoticeQueue.Poll(); element != nil; element = p.txNoticeQueue.Poll() {
			hash := element.(TxHash)
			if p.txHashCache.Contains(hash) {
				continue
			}
			hashes = append(hashes, hash[:])
			idx++
			p.txHashCache.Add(hash, cachePlaceHolder)
			if idx >= p.maxTxNoticeHashSize {
				break
			}
		}
		mo := p.mf.newMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes:hashes})
		p.sendMessage(mo)
	}
}

func (p *remotePeerImpl) tryGetStream(msgID string, protocol protocol.ID, timeout time.Duration) inet.Stream {
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

func (p *remotePeerImpl) getStreamForWriting(msgID string, protocol protocol.ID, schannel chan inet.Stream) {
	ctx := context.Background()
	s, err := p.pm.NewStream(ctx, p.meta.ID, protocol)
	if err != nil {
		p.logger.Warn().Err(err).Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, string(protocol)).Msg("Error while get stream")
		schannel <- nil
	}
	schannel <- s
}

// this method MUST be called in same go routine as AergoPeer.RunPeer()
func (p *remotePeerImpl) sendPing() {
	// find my best block
	//bestBlock, err := extractBlockFromRequest(p.actorServ.CallRequest(message.ChainSvc, &message.GetBestBlock{}))
	//if err != nil {
	//	p.logger.Error().Err(err).Msg("Failed to get best block")
	//	return
	//}
	// create message data
	pingMsg := &types.Ping{
		//BestBlockHash: bestBlock.BlockHash(),
		//BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	p.sendMessage(p.mf.newMsgRequestOrder(true, PingRequest, pingMsg))
}

// sendStatus is called once when a peer is added.()
func (p *remotePeerImpl) sendStatus() {
	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Sending status message for handshaking")

	// create message data
	statusMsg, err := createStatusMsg(p.pm, p.actorServ)
	if err != nil {
		p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Err(err).Msg("Cancel sending status")
		return
	}

	p.sendMessage(p.mf.newMsgRequestOrder(false, StatusRequest, statusMsg))
}

// send notice message and then disconnect. this routine should only run in RunPeer go routine
func (p *remotePeerImpl) goAwayMsg(msg string) {
	p.logger.Info().Str(LogPeerID, p.meta.ID.Pretty()).Str("msg", msg).Msg("Peer is closing")
	p.sendMessage(p.mf.newMsgRequestOrder(false, GoAway, &types.GoAwayNotice{Message: msg}))
	p.pm.RemovePeer(p.meta.ID)
}

func (p *remotePeerImpl) pruneRequests() {
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

func (p *remotePeerImpl) handleNewBlockNotice(data *types.NewBlockNotice) {
	// lru cache can accept hashable key
	var hashArr BlockHash
	copy(hashArr[:], data.BlockHash)

	p.blkHashCache.Add(hashArr, true)
	p.pm.HandleNewBlockNotice(p.meta.ID, hashArr, data)
}

func (p *remotePeerImpl) handleNewTxNotice(data *types.NewTransactionsNotice) {
	if len(data.TxHashes) == 0 {
		return
	}
	// lru cache can accept hashable key
	hashArrs := make([]TxHash, len(data.TxHashes))
	for i, hash := range data.TxHashes {
		copy(hashArrs[i][:], hash)
		p.blkHashCache.Add(hashArrs[i], true)
	}
	p.pm.HandleNewTxNotice(p.meta.ID, hashArrs, data)
}

func (p *remotePeerImpl) sendGoAway(msg string) {
	// TODO: send goaway message and close connection
}

func newHangresolver(peer *remotePeerImpl, logger *log.Logger) *hangResolver {
	return &hangResolver{p: peer, logger: logger}
}

type hangResolver struct {
	p      *remotePeerImpl
	logger *log.Logger

	consecutiveDrops uint64
}

func (l *hangResolver) OnIn(element interface{}) {
}

func (l *hangResolver) OnDrop(element interface{}) {
	mo := element.(msgOrder)
	now := time.Now().Unix()
	// if last send hang too long. drop this peer
	if l.consecutiveDrops > 20 || (now-mo.Timestamp()) > 60 {
		l.logger.Info().Str(LogPeerID, l.p.ID().Pretty()).Msg("Peer seems to hang, drop this peer")
		l.p.pm.RemovePeer(l.p.ID())
	} else {
		l.logger.Debug().Str(LogPeerID, l.p.ID().Pretty()).Str(LogMsgID, mo.GetMsgID()).Str(LogProtoID, mo.GetProtocolID().String()).Msg("Peer too busy or deadlock, stalled message is dropped")
	}
}

func (l *hangResolver) OnOut(element interface{}) {
	l.consecutiveDrops = 0
}
