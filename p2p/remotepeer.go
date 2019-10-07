/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/pkg/errors"
	"runtime/debug"
	"sync"
	"time"

	"github.com/aergoio/aergo/p2p/metric"
	"github.com/aergoio/aergo/p2p/p2pcommon"

	lru "github.com/hashicorp/golang-lru"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
)

var TimeoutError = errors.New("timeout")
var CancelError = errors.New("canceled")

type requestInfo struct {
	cTime    time.Time
	reqMO    p2pcommon.MsgOrder
	receiver p2pcommon.ResponseReceiver
}

type queryMsg struct {
	handler p2pcommon.MessageHandler
	msg     p2pcommon.Message
	msgBody p2pcommon.MessageBody
}

// remotePeerImpl represent remote peer to which is connected
type remotePeerImpl struct {
	logger       *log.Logger
	pingDuration time.Duration

	manageNum uint32
	meta      p2pcommon.PeerMeta
	name      string
	state     types.PeerState
	role      p2pcommon.PeerRole
	actor     p2pcommon.ActorService
	pm        p2pcommon.PeerManager
	mf        p2pcommon.MoFactory
	signer    p2pcommon.MsgSigner
	metric    *metric.PeerMetric
	tnt       p2pcommon.TxNoticeTracer

	stopChan chan struct{}

	// direct write channel
	dWrite     chan p2pcommon.MsgOrder
	closeWrite chan struct{}

	// used to access request data from response handlers
	requests map[p2pcommon.MsgID]*requestInfo
	reqMutex *sync.Mutex

	handlers map[p2pcommon.SubProtocol]p2pcommon.MessageHandler

	// TODO make automatic disconnect if remote peer cause too many wrong message
	blkHashCache *lru.Cache
	txHashCache  *lru.Cache
	lastStatus   *types.LastBlockStatus
	// lastBlkNoticeTime is time that local peer sent NewBlockNotice to this remote peer
	lastBlkNoticeTime time.Time
	skipCnt           int32

	txQueueLock         *sync.Mutex
	txNoticeQueue       *p2putil.PressableQueue
	maxTxNoticeHashSize int

	rw p2pcommon.MsgReadWriter
}

var _ p2pcommon.RemotePeer = (*remotePeerImpl)(nil)

// newRemotePeer create an object which represent a remote peer.
func newRemotePeer(meta p2pcommon.PeerMeta, manageNum uint32, pm p2pcommon.PeerManager, actor p2pcommon.ActorService, log *log.Logger, mf p2pcommon.MoFactory, signer p2pcommon.MsgSigner, rw p2pcommon.MsgReadWriter) *remotePeerImpl {
	rPeer := &remotePeerImpl{
		meta: meta, manageNum: manageNum, pm: pm,
		name:  fmt.Sprintf("%s#%d", p2putil.ShortForm(meta.ID), manageNum),
		actor: actor, logger: log, mf: mf, signer: signer, rw: rw,
		pingDuration: defaultPingInterval,
		state:        types.STARTING,

		lastStatus: &types.LastBlockStatus{},
		stopChan:   make(chan struct{}, 1),
		closeWrite: make(chan struct{}),

		requests: make(map[p2pcommon.MsgID]*requestInfo),
		reqMutex: &sync.Mutex{},

		handlers: make(map[p2pcommon.SubProtocol]p2pcommon.MessageHandler),

		txQueueLock:         &sync.Mutex{},
		txNoticeQueue:       p2putil.NewPressableQueue(DefaultPeerTxQueueSize),
		maxTxNoticeHashSize: DefaultPeerTxQueueSize,
	}
	rPeer.dWrite = make(chan p2pcommon.MsgOrder, writeMsgBufferSize)

	var err error
	rPeer.blkHashCache, err = lru.New(DefaultPeerBlockCacheSize)
	if err != nil {
		panic("Failed to create remote peer " + err.Error())
	}
	rPeer.txHashCache, err = lru.New(DefaultPeerTxCacheSize)
	if err != nil {
		panic("Failed to create remote peer " + err.Error())
	}

	return rPeer
}

// ID return id of peer, same as peer.meta.ID
func (p *remotePeerImpl) ID() types.PeerID {
	return p.meta.ID
}

func (p *remotePeerImpl) Meta() p2pcommon.PeerMeta {
	return p.meta
}

func (p *remotePeerImpl) ManageNumber() uint32 {
	return p.manageNum
}

func (p *remotePeerImpl) Name() string {
	return p.name
}

func (p *remotePeerImpl) Version() string {
	return p.meta.Version
}

func (p *remotePeerImpl) Role() p2pcommon.PeerRole {
	return p.role
}
func (p *remotePeerImpl) ChangeRole(role p2pcommon.PeerRole) {
	p.role = role
}

func (p *remotePeerImpl) AddMessageHandler(subProtocol p2pcommon.SubProtocol, handler p2pcommon.MessageHandler) {
	p.handlers[subProtocol] = handler
}

func (p *remotePeerImpl) MF() p2pcommon.MoFactory {
	return p.mf
}

// State returns current state of peer
func (p *remotePeerImpl) State() types.PeerState {
	return p.state.Get()
}

func (p *remotePeerImpl) LastStatus() *types.LastBlockStatus {
	return p.lastStatus
}

// runPeer should be called by go routine
func (p *remotePeerImpl) RunPeer() {
	p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Msg("Starting peer")
	pingTicker := time.NewTicker(p.pingDuration)

	go p.runWrite()
	go p.runRead()

	txNoticeTicker := time.NewTicker(txNoticeInterval)

	// peer state is changed to RUNNING after all sub goroutine is ready, and to STOPPED before fll sub goroutine is stopped.
	p.state.SetAndGet(types.RUNNING)
READNOPLOOP:
	for {
		select {
		case <-pingTicker.C:
			p.sendPing()
			// no operation for now
		case <-txNoticeTicker.C:
			p.trySendTxNotices()
		case <-p.stopChan:
			break READNOPLOOP
		}
	}

	p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Msg("Finishing peer")
	txNoticeTicker.Stop()
	pingTicker.Stop()
	// finish goroutine write. read goroutine will be closed automatically when disconnect
	close(p.closeWrite)
	close(p.stopChan)
	p.state.SetAndGet(types.STOPPED)

	p.pm.RemovePeer(p)
}

func (p *remotePeerImpl) runWrite() {
	cleanupTicker := time.NewTicker(cleanRequestInterval)
	defer func() {
		if r := recover(); r != nil {
			p.logger.Panic().Str("callStack", string(debug.Stack())).Str(p2putil.LogPeerName, p.Name()).Str("recover", fmt.Sprint(r)).Msg("There were panic in runWrite ")
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
			p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Msg("Quitting runWrite")
			break WRITELOOP
		}
	}
	cleanupTicker.Stop()
	p.cleanupWrite()
	p.rw.Close()

	// closing channel is up to golang runtime
	// close(p.write)
	// close(p.consumeChan)
}

func (p *remotePeerImpl) cleanupWrite() {
	// 1. cleaning receive handlers. TODO add code

	// 2. canceling not sent orders
	for {
		select {
		case m := <-p.dWrite:
			m.CancelSend(p)
		default:
			return
		}
	}
}

func (p *remotePeerImpl) runRead() {
	for {
		msg, err := p.rw.ReadMsg()
		if err != nil {
			// TODO set different log level by case (i.e. it can be expected if peer is disconnecting )
			p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Err(err).Msg("Failed to read message")
			p.Stop()
			return
		}
		if err = p.handleMsg(msg); err != nil {
			// TODO set different log level by case (i.e. it can be expected if peer is disconnecting )
			p.logger.Warn().Str(p2putil.LogPeerName, p.Name()).Err(err).Msg("Failed to handle message")
			p.Stop()
			return
		}
	}
}

func (p *remotePeerImpl) handleMsg(msg p2pcommon.Message) (err error) {
	subProto := msg.Subprotocol()
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error().Str(p2putil.LogProtoID, subProto.String()).Str("callStack", string(debug.Stack())).Interface("panic", r).Msg("There were panic in handler.")
			err = fmt.Errorf("internal error")
		}
	}()

	if p.State() > types.RUNNING {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogMsgID, msg.ID().String()).Str(p2putil.LogProtoID, subProto.String()).Str("current_state", p.State().String()).Msg("peer is not running. silently drop input message")
		return nil
	}

	handler, found := p.handlers[subProto]
	if !found {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogMsgID, msg.ID().String()).Str(p2putil.LogProtoID, subProto.String()).Msg("invalid protocol")
		return fmt.Errorf("invalid protocol %s", subProto)
	}

	handler.PreHandle()

	payload, err := handler.ParsePayload(msg.Payload())
	if err != nil {
		p.logger.Warn().Err(err).Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogMsgID, msg.ID().String()).Str(p2putil.LogProtoID, subProto.String()).Msg("invalid message data")
		return fmt.Errorf("invalid message data")
	}
	//err = p.signer.verifyMsg(msg, p.meta.ID)
	//if err != nil {
	//	p.logger.Warn().Err(err).Str(LogPeerName, p.Name()).Str(LogMsgID, msg.ID().String()).Str(LogProtoID, subProto.String()).Msg("Failed to check signature")
	//	return fmt.Errorf("Failed to check signature")
	//}
	err = handler.CheckAuth(msg, payload)
	if err != nil {
		p.logger.Warn().Err(err).Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogMsgID, msg.ID().String()).Str(p2putil.LogProtoID, subProto.String()).Msg("Failed to authenticate message")
		return fmt.Errorf("Failed to authenticate message.")
	}

	handler.Handle(msg, payload)

	handler.PostHandle(msg, payload)
	return nil
}

// Stop stops aPeer works
func (p *remotePeerImpl) Stop() {
	prevState := p.state.SetAndGet(types.STOPPING)
	if prevState <= types.RUNNING {
		p.stopChan <- struct{}{}
	}

}

func (p *remotePeerImpl) SendMessage(msg p2pcommon.MsgOrder) {
	if p.State() > types.RUNNING {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, msg.GetProtocolID().String()).
			Str(p2putil.LogMsgID, msg.GetMsgID().String()).Str("current_state", p.State().String()).Msg("Cancel sending message, since peer is not running state")
		return
	}
	select {
	case p.dWrite <- msg:
		// it's OK
	default:
		p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, msg.GetProtocolID().String()).
			Str(p2putil.LogMsgID, msg.GetMsgID().String()).Msg("Remote peer is busy or down")
		// TODO find more elegant way to handled flooding queue. in lots of cases, pending for dropped tx notice or newBlock notice (not blockProduced notice) are not critical in lots of cases.
		p.Stop()
	}
}

func (p *remotePeerImpl) SendAndWaitMessage(msg p2pcommon.MsgOrder, timeout time.Duration) error {
	if p.State() > types.RUNNING {
		p.logger.Debug().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, msg.GetProtocolID().String()).
			Str(p2putil.LogMsgID, msg.GetMsgID().String()).Str("current_state", p.State().String()).Msg("Cancel sending message, since peer is not running state")
		return fmt.Errorf("not running")
	}
	select {
	case p.dWrite <- msg:
		return nil
	case <-time.NewTimer(timeout).C:
		p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Str(p2putil.LogProtoID, msg.GetProtocolID().String()).
			Str(p2putil.LogMsgID, msg.GetMsgID().String()).Msg("Remote peer is busy or down")
		// TODO find more elegant way to handled flooding queue. in lots of cases, pending for dropped tx notice or newBlock notice (not blockProduced notice) are not critical in lots of cases.
		p.Stop()
		return TimeoutError
	}
}

func (p *remotePeerImpl) PushTxsNotice(txHashes []types.TxID) {
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

// ConsumeRequest remove request from request history.
func (p *remotePeerImpl) ConsumeRequest(originalID p2pcommon.MsgID) {
	p.reqMutex.Lock()
	delete(p.requests, originalID)
	p.reqMutex.Unlock()
}

// requestIDNotFoundReceiver is to handle response msg which the original message is not identified
func (p *remotePeerImpl) requestIDNotFoundReceiver(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) bool {
	return true
}

// passThroughReceiver is bypass message to legacy handler.
func (p *remotePeerImpl) passThroughReceiver(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) bool {
	return false
}

func (p *remotePeerImpl) GetReceiver(originalID p2pcommon.MsgID) p2pcommon.ResponseReceiver {
	p.reqMutex.Lock()
	defer p.reqMutex.Unlock()
	req, found := p.requests[originalID]
	if !found {
		return p.requestIDNotFoundReceiver
	} else if req.receiver == nil {
		return p.passThroughReceiver
	} else {
		return req.receiver
	}
}

func (p *remotePeerImpl) writeToPeer(m p2pcommon.MsgOrder) {
	if err := m.SendTo(p); err != nil {
		// write fail
		p.Stop()
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
		skippedTxIDs := make([]types.TxID, 0)
		for element := p.txNoticeQueue.Poll(); element != nil; element = p.txNoticeQueue.Poll() {
			hash := element.(types.TxID)
			if p.txHashCache.Contains(hash) {
				skippedTxIDs = append(skippedTxIDs, hash)
				continue
			}
			hashes = append(hashes, hash[:])
			p.txHashCache.Add(hash, cachePlaceHolder)
		}
		if len(hashes) > 0 {
			mo := p.mf.NewMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes: hashes})
			p.SendMessage(mo)
		}
		if len(skippedTxIDs) > 0 {
			// if tx is in cache, the remote peer will have that tx.
			p.tnt.ReportSend(skippedTxIDs, p.ID())
		}
	}
}

// this method MUST be called in same go routine as AergoPeer.RunPeer()
func (p *remotePeerImpl) sendPing() {
	// find my best block
	bestBlock, err := p.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		p.logger.Warn().Err(err).Msg("cancel ping. failed to get best block")
		return
	}
	// create message data
	pingMsg := &types.Ping{
		BestBlockHash: bestBlock.BlockHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	p.SendMessage(p.mf.NewMsgRequestOrder(true, p2pcommon.PingRequest, pingMsg))
}

// send notice message and then disconnect. this routine should only run in RunPeer go routine
func (p *remotePeerImpl) goAwayMsg(msg string) {
	p.logger.Info().Str(p2putil.LogPeerName, p.Name()).Str("msg", msg).Msg("Peer is closing")
	p.SendAndWaitMessage(p.mf.NewMsgRequestOrder(false, p2pcommon.GoAway, &types.GoAwayNotice{Message: msg}), time.Second)
	p.Stop()
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
	p.logger.Info().Int("count", deletedCnt).Str(p2putil.LogPeerName, p.Name()).
		Time("until", expireTime).Msg("Pruned requests which response was not came")
	//.Msg("Pruned %d requests but no response to peer %s until %v", deletedCnt, p.meta.ID.Pretty(), time.Unix(expireTime, 0))
	if debugLog {
		p.logger.Debug().Strs("reqs", deletedReqs).Msg("Pruned")
	}
}

func (p *remotePeerImpl) UpdateBlkCache(blkHash []byte, blkNumber uint64) bool {
	p.UpdateLastNotice(blkHash, blkNumber)
	hash := types.ToBlockID(blkHash)
	// lru cache can't accept byte slice key
	found, _ := p.blkHashCache.ContainsOrAdd(hash, true)
	return found
}

func (p *remotePeerImpl) UpdateTxCache(hashes []types.TxID) []types.TxID {
	// lru cache can't accept byte slice key
	added := make([]types.TxID, 0, len(hashes))
	for _, hash := range hashes {
		if found, _ := p.txHashCache.ContainsOrAdd(hash, true); !found {
			added = append(added, hash)
		}
	}
	return added
}

func (p *remotePeerImpl) UpdateLastNotice(blkHash []byte, blkNumber uint64) {
	p.lastStatus = &types.LastBlockStatus{time.Now(), blkHash, blkNumber}
}

func (p *remotePeerImpl) sendGoAway(msg string) {
	// TODO: send goaway message and close connection
}
