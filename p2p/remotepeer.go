/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"sync"
	"time"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/metric"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-net"

	"github.com/hashicorp/golang-lru"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

var TimeoutError error

func init() {
	TimeoutError = fmt.Errorf("timeout")
}

type RemotePeer interface {
	ID() peer.ID
	Meta() PeerMeta
	ManageNumber() uint32

	State() types.PeerState
	LastNotice() *types.NewBlockNotice

	runPeer()
	stop()

	sendMessage(msg msgOrder)
	sendAndWaitMessage(msg msgOrder, ttl time.Duration) error

	pushTxsNotice(txHashes []TxHash)
	// utility method

	consumeRequest(msgID MsgID)
	GetReceiver(id MsgID) ResponseReceiver

	// updateBlkCache add hash to block cache and return true if this hash already exists.
	updateBlkCache(hash BlkHash, blkNotice *types.NewBlockNotice) bool
	// updateTxCache add hashes to transaction cache and return newly added hashes.
	updateTxCache(hashes []TxHash) []TxHash

	// TODO
	MF() moFactory
}

type requestInfo struct {
	cTime    time.Time
	reqMO    msgOrder
	receiver ResponseReceiver
}

// ResponseReceiver returns true when receiver handled it, or false if this receiver is not the expected handler.
// NOTE: the return value is temporal works for old implementation and will be remove later.
type ResponseReceiver func(msg Message, msgBody proto.Message) bool

func dummyResponseReceiver(msg Message, msgBody proto.Message) bool {
	return false
}

// remotePeerImpl represent remote peer to which is connected
type remotePeerImpl struct {
	logger       *log.Logger
	pingDuration time.Duration

	manageNum uint32
	meta      PeerMeta
	state     types.PeerState
	actorServ ActorService
	pm        PeerManager
	mf        moFactory
	signer    msgSigner
	metric    *metric.PeerMetric

	stopChan chan struct{}

	// direct write channel
	dWrite     chan msgOrder
	closeWrite chan struct{}

	// used to access request data from response handlers
	requests map[MsgID]*requestInfo
	reqMutex *sync.Mutex

	handlers map[SubProtocol]MessageHandler

	// TODO make automatic disconnect if remote peer cause too many wrong message

	blkHashCache *lru.Cache
	txHashCache  *lru.Cache
	lastNotice   *types.NewBlockNotice

	txQueueLock         *sync.Mutex
	txNoticeQueue       *p2putil.PressableQueue
	maxTxNoticeHashSize int

	s  net.Stream
	rw MsgReadWriter
}

var _ RemotePeer = (*remotePeerImpl)(nil)

// newRemotePeer create an object which represent a remote peer.
func newRemotePeer(meta PeerMeta, pm PeerManager, actor ActorService, log *log.Logger, mf moFactory, signer msgSigner, s net.Stream, rw MsgReadWriter) *remotePeerImpl {
	rPeer := &remotePeerImpl{
		meta: meta, pm: pm, actorServ: actor, logger: log, mf: mf, signer: signer, s: s, rw: rw,
		pingDuration: defaultPingInterval,
		state:        types.STARTING,

		stopChan:   make(chan struct{}, 1),
		closeWrite: make(chan struct{}),

		requests: make(map[MsgID]*requestInfo),
		reqMutex: &sync.Mutex{},

		handlers: make(map[SubProtocol]MessageHandler),

		txQueueLock:         &sync.Mutex{},
		txNoticeQueue:       p2putil.NewPressableQueue(DefaultPeerTxQueueSize),
		maxTxNoticeHashSize: DefaultPeerTxQueueSize,
	}
	//rPeer.write =make(chan msgp2putil.NewDefaultChannelPipe(20, newHangresolver(rPeer, log))
	rPeer.dWrite = make(chan msgOrder, writeMsgBufferSize)

	var err error
	rPeer.blkHashCache, err = lru.New(DefaultPeerBlockCacheSize)
	if err != nil {
		panic("Failed to create remotepeer " + err.Error())
	}
	rPeer.txHashCache, err = lru.New(DefaultPeerTxCacheSize)
	if err != nil {
		panic("Failed to create remotepeer " + err.Error())
	}

	return rPeer
}

// ID return id of peer, same as peer.meta.ID
func (p *remotePeerImpl) ID() peer.ID {
	return p.meta.ID
}

func (p *remotePeerImpl) Meta() PeerMeta {
	return p.meta
}

func (p *remotePeerImpl) ManageNumber() uint32 {
	return p.manageNum
}

func (p *remotePeerImpl) MF() moFactory {
	return p.mf
}

// State returns current state of peer
func (p *remotePeerImpl) State() types.PeerState {
	return p.state.Get()
}

func (p *remotePeerImpl) GetBlocks(hashes []message.BlockHash, ttl time.Duration) ([]*types.Block, error) {
	//    remotePeer 객체가 상대 peer에 보내기 위한 메세지 생성.
	hashesToGet := make([][]byte, len(hashes))
	for i, hash := range hashes {
		hashesToGet[i] = ([]byte)(hash)
	}
	// create message data
	req := &types.GetBlockRequest{Hashes: hashesToGet}

	p.sendMessage(p.mf.newMsgRequestOrder(true, GetBlocksRequest, req))
	//p.mf.newMsgRequestOrder()
	//    remotePeer 가 peer에 메세지 전송하고 메세지 아이디 저장.
	//    상대 peer에서 보낸 메세지에 대한 응답대기
	//    응답이 도착하면 응답을 리턴
	panic("implement me")
}

func (p *remotePeerImpl) LastNotice() *types.NewBlockNotice {
	return p.lastNotice
}

// runPeer should be called by go routine
func (p *remotePeerImpl) runPeer() {
	p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Starting peer")
	pingTicker := time.NewTicker(p.pingDuration)

	go p.runWrite()
	go p.runRead()

	txNoticeTicker := time.NewTicker(txNoticeInterval)

	// peer state is changed to RUNNIG after all sub goroutine is ready, and to STOPPED before fll sub goroutine is stopped.
	p.state.SetAndGet(types.RUNNING)
READNOPLOOP:
	for {
		select {
		case <-pingTicker.C:
			// no operation for now
		case <-txNoticeTicker.C:
			p.trySendTxNotices()
		case <-p.stopChan:
			break READNOPLOOP
		}
	}

	p.logger.Info().Uint32("manage_num", p.manageNum).Str(LogPeerID, p.meta.ID.Pretty()).Msg("Finishing peer")
	txNoticeTicker.Stop()
	pingTicker.Stop()
	// finish goroutine write. read goroutine will be closed automatically when disconnect
	p.closeWrite <- struct{}{}
	close(p.stopChan)
	p.state.SetAndGet(types.STOPPED)

	p.pm.RemovePeer(p)
}

func (p *remotePeerImpl) runWrite() {
	cleanupTicker := time.NewTicker(cleanRequestInterval)
	defer func() {
		if r := recover(); r != nil {
			p.logger.Panic().Str(LogPeerID, p.ID().Pretty()).Str("recover", fmt.Sprint(r)).Msg("There were panic in runWrite ")
		}
	}()

WRITELOOP:
	for {
		select {
		case m := <-p.dWrite:
			p.writeToPeer(m)
		case <-cleanupTicker.C:
			p.pruneRequests()
		case <-p.closeWrite:
			p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Msg("Quitting runWrite")
			break WRITELOOP
		}
	}
	cleanupTicker.Stop()
	p.cleanupWrite()
	p.s.Close()

	// closing channel is up to golang runtime
	// close(p.write)
	// close(p.consumeChan)
}

func (p *remotePeerImpl) cleanupWrite() {
	// 1. cleaning receive handlers. TODO add code

	// 2. canceling not sent orders TODO add code

	for {
		select {
		case m := <-p.dWrite:
			m.IsRequest()
		default:
			return
		}
	}
}

func (p *remotePeerImpl) runRead() {
	for {
		msg, err := p.rw.ReadMsg()
		if err != nil {
			p.logger.Error().Str(LogPeerID, p.ID().Pretty()).Err(err).Msg("Failed to read message")
			p.stop()
			return
		}
		if err = p.handleMsg(msg); err != nil {
			p.logger.Error().Str(LogPeerID, p.ID().Pretty()).Err(err).Msg("Failed to handle message")
			p.stop()
			return
		}
	}
}

func (p *remotePeerImpl) handleMsg(msg Message) error {
	var err error
	subProto := msg.Subprotocol()
	defer func() {
		if r := recover(); r != nil {
			p.logger.Warn().Interface("panic", r).Msg("There were panic in handler.")
			err = fmt.Errorf("internal error")
		}
	}()

	if p.State() > types.RUNNING {
		p.logger.Debug().Str(LogPeerID, p.ID().Pretty()).Str(LogMsgID, msg.ID().String()).Str(LogProtoID, subProto.String()).Str("current_state", p.State().String()).Msg("peer is not running. silently drop input message")
		return nil
	}

	handler, found := p.handlers[subProto]
	if !found {
		p.logger.Debug().Str(LogPeerID, p.ID().Pretty()).Str(LogMsgID, msg.ID().String()).Str(LogProtoID, subProto.String()).Msg("invalid protocol")
		return fmt.Errorf("invalid protocol %s", subProto)
	}

	handler.preHandle()

	payload, err := handler.parsePayload(msg.Payload())
	if err != nil {
		p.logger.Warn().Err(err).Str(LogPeerID, p.ID().Pretty()).Str(LogMsgID, msg.ID().String()).Str(LogProtoID, subProto.String()).Msg("invalid message data")
		return fmt.Errorf("invalid message data")
	}
	//err = p.signer.verifyMsg(msg, p.meta.ID)
	//if err != nil {
	//	p.logger.Warn().Err(err).Str(LogPeerID, p.ID().Pretty()).Str(LogMsgID, msg.ID().String()).Str(LogProtoID, subProto.String()).Msg("Failed to check signature")
	//	return fmt.Errorf("Failed to check signature")
	//}
	err = handler.checkAuth(msg, payload)
	if err != nil {
		p.logger.Warn().Err(err).Str(LogPeerID, p.ID().Pretty()).Str(LogMsgID, msg.ID().String()).Str(LogProtoID, subProto.String()).Msg("Failed to authenticate message")
		return fmt.Errorf("Failed to authenticate message.")
	}

	handler.handle(msg, payload)

	handler.postHandle(msg, payload)
	return nil
}

// Stop stops aPeer works
func (p *remotePeerImpl) stop() {
	prevState := p.state.SetAndGet(types.STOPPING)
	if prevState <= types.RUNNING {
		p.stopChan <- struct{}{}
	}
}

func (p *remotePeerImpl) sendMessage(msg msgOrder) {
	if p.State() > types.RUNNING {
		p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, msg.GetProtocolID().String()).
			Str(LogMsgID, msg.GetMsgID().String()).Str("current_state", p.State().String()).Msg("Cancel sending messge, since peer is not running state")
		return
	}
	select {
	case p.dWrite <- msg:
		// it's OK
	default:
		p.logger.Info().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, msg.GetProtocolID().String()).
			Str(LogMsgID, msg.GetMsgID().String()).Msg("Remote peer is busy or down")
		// TODO find more elegant way to handled flooding queue. in lots of cases, pending for dropped tx notice or newblocknotice (not blockproducednotice) are not critical in lots of cases.
		p.stop()
	}
}

func (p *remotePeerImpl) sendAndWaitMessage(msg msgOrder, timeout time.Duration) error {
	if p.State() > types.RUNNING {
		p.logger.Debug().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, msg.GetProtocolID().String()).
			Str(LogMsgID, msg.GetMsgID().String()).Str("current_state", p.State().String()).Msg("Cancel sending messge, since peer is not running state")
		return fmt.Errorf("not running")
	}
	select {
	case p.dWrite <- msg:
		return nil
	case <-time.NewTimer(timeout).C:
		p.logger.Info().Str(LogPeerID, p.meta.ID.Pretty()).Str(LogProtoID, msg.GetProtocolID().String()).
			Str(LogMsgID, msg.GetMsgID().String()).Msg("Remote peer is busy or down")
		// TODO find more elegant way to handled flooding queue. in lots of cases, pending for dropped tx notice or newblocknotice (not blockproducednotice) are not critical in lots of cases.
		p.stop()
		return TimeoutError
	}
}

func (p *remotePeerImpl) pushTxsNotice(txHashes []TxHash) {
	p.txQueueLock.Lock()
	defer p.txQueueLock.Unlock()
	for _, hash := range txHashes {
		if !p.txNoticeQueue.Offer(hash) {
			p.sendTxNotices()
			// this Offer is always succeeded by invariant
			p.txNoticeQueue.Offer(hash)
		}
	}
}

// consumeRequest remove request from request history.
func (p *remotePeerImpl) consumeRequest(originalID MsgID) {
	p.reqMutex.Lock()
	delete(p.requests, originalID)
	p.reqMutex.Unlock()
}

func (p *remotePeerImpl) notFoundReceiver(msg Message, msgBody proto.Message) bool {
	//	p.logger.Debug().Str(LogPeerID, p.ID().Pretty()).Str("req_id", msg.OriginalID().String()).Str(LogMsgID, msg.ID().String()).Msg("not found suitable reciever. toss message to legacy handler")
	return false
}

func (p *remotePeerImpl) GetReceiver(originalID MsgID) ResponseReceiver {
	p.reqMutex.Lock()
	defer p.reqMutex.Unlock()
	req, found := p.requests[originalID]
	if !found || req.receiver == nil {
		return p.notFoundReceiver
	}
	return req.receiver
}

func (p *remotePeerImpl) updateMetaInfo(statusMsg *types.Status) {
	// check address. and apply current
	receivedMeta := FromPeerAddress(statusMsg.Sender)
	p.meta.IPAddress = receivedMeta.IPAddress
	p.meta.Port = receivedMeta.Port
}

func (p *remotePeerImpl) writeToPeer(m msgOrder) {
	if err := m.SendTo(p); err != nil {
		// write fail
		p.stop()
	}
}

func (p *remotePeerImpl) trySendTxNotices() {
	p.txQueueLock.Lock()
	defer p.txQueueLock.Unlock()
	p.sendTxNotices()
}

// sendTxNotices must be called in txQueueLock
func (p *remotePeerImpl) sendTxNotices() {
	// make create multiple message if queue is too many hashes.
	for {
		// no need to send if queue is empty
		if p.txNoticeQueue.Size() == 0 {
			return
		}
		hashes := make([][]byte, 0, p.txNoticeQueue.Size())
		idx := 0
		for element := p.txNoticeQueue.Poll(); element != nil; element = p.txNoticeQueue.Poll() {
			hash := element.(TxHash)
			if p.txHashCache.Contains(hash) {
				continue
			}
			hashes = append(hashes, hash[:])
			idx++
			p.txHashCache.Add(hash, cachePlaceHolder)
			//if idx >= p.maxTxNoticeHashSize {
			//	break
			//}
		}
		if idx > 0 {
			mo := p.mf.newMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes: hashes})
			p.sendMessage(mo)
		}
	}
}

// this method MUST be called in same go routine as AergoPeer.RunPeer()
func (p *remotePeerImpl) sendPing() {
	// find my best block
	//bestBlock, err := extractBlockFromRequest(p.actorService.CallRequest(message.ChainSvc, &message.GetBestBlock{}))
	//if err != nil {
	//	p.logger.Error().Err(err).Msg("Failed to get best block")
	//	return
	//}
	// create message data
	pingMsg := &types.Ping{
		//BestBlockHash: bestBlock.BlkHash(),
		//BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	p.sendMessage(p.mf.newMsgRequestOrder(true, PingRequest, pingMsg))
}

// send notice message and then disconnect. this routine should only run in RunPeer go routine
func (p *remotePeerImpl) goAwayMsg(msg string) {
	p.logger.Info().Str(LogPeerID, p.meta.ID.Pretty()).Str("msg", msg).Msg("Peer is closing")
	p.sendAndWaitMessage(p.mf.newMsgRequestOrder(false, GoAway, &types.GoAwayNotice{Message: msg}), time.Second)
	p.stop()
}

func (p *remotePeerImpl) pruneRequests() {
	debugLog := p.logger.IsDebugEnabled()
	deletedCnt := 0
	var deletedReqs []string
	expireTime := time.Now().Add(-1 * time.Hour)
	p.reqMutex.Lock()
	defer p.reqMutex.Unlock()
	for key, m := range p.requests {
		if m.cTime.Before(expireTime) {
			delete(p.requests, key)
			if debugLog {
				deletedReqs = append(deletedReqs, m.reqMO.GetProtocolID().String()+"/"+key.String()+m.cTime.String())
			}
			deletedCnt++
		}
	}
	p.logger.Info().Int("count", deletedCnt).Str(LogPeerID, p.meta.ID.Pretty()).
		Time("until", expireTime).Msg("Pruned requests which response was not came")
	//.Msg("Pruned %d requests but no response to peer %s until %v", deletedCnt, p.meta.ID.Pretty(), time.Unix(expireTime, 0))
	if debugLog {
		p.logger.Debug().Strs("reqs", deletedReqs).Msg("Pruned")
	}
}

func (p *remotePeerImpl) updateBlkCache(hash BlkHash, blkNotice *types.NewBlockNotice) bool {
	p.lastNotice = blkNotice
	// lru cache can accept hashable key
	found, _ := p.blkHashCache.ContainsOrAdd(hash, true)
	return found
}

func (p *remotePeerImpl) updateTxCache(hashes []TxHash) []TxHash {
	// lru cache can accept hashable key
	added := make([]TxHash, 0, len(hashes))
	for _, hash := range hashes {
		if found, _ := p.txHashCache.ContainsOrAdd(hash, true); !found {
			added = append(added, hash)
		}
	}
	return added
}

func (p *remotePeerImpl) sendGoAway(msg string) {
	// TODO: send goaway message and close connection
}
