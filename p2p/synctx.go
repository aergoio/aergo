package p2p

import (
	"container/list"
	"fmt"
	"runtime/debug"
	"sort"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/p2p/subproto"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
)

const minimumTxQueryInterval = time.Second >> 1
const txQueryTimeout = time.Second << 2

var unsent = time.Time{}

// syncTxManager handle operations about tx sync
type syncTxManager struct {
	logger *log.Logger

	sm        p2pcommon.SyncManager
	actor     p2pcommon.ActorService
	pm        p2pcommon.PeerManager
	msgHelper message.Helper

	txCache *lru.Cache
	// received notice but not in my mempool
	frontCache      map[types.TxID]*incomingTxNotice
	toNoticeIdQueue *list.List

	taskChannel      chan smTask
	taskQueryChannel chan smTask
	finishChannel    chan struct{}

	fcTicker *time.Ticker
}

type queryQueue struct {
	peerID types.PeerID
	txIDs  []types.TxID
}

type smTask func()

func newTxSyncManager(sm p2pcommon.SyncManager, actor p2pcommon.ActorService, pm p2pcommon.PeerManager, logger *log.Logger) *syncTxManager {
	tm := &syncTxManager{sm: sm, actor: actor, pm: pm, logger: logger,
		frontCache:       make(map[types.TxID]*incomingTxNotice),
		toNoticeIdQueue:  list.New(),
		taskChannel:      make(chan smTask, 20),
		finishChannel:    make(chan struct{}, 1),
		taskQueryChannel: make(chan smTask, 10),

		msgHelper: message.GetHelper(),
		fcTicker:  time.NewTicker(minimumTxQueryInterval),
	}
	var err error
	tm.txCache, err = lru.New(DefaultGlobalTxCacheSize)
	if err != nil {
		panic("Failed to create p2p tx cache " + err.Error())
	}
	return tm
}

func (tm *syncTxManager) Start() {
	go tm.runManager()
	go tm.runQueryLog()
}

func (tm *syncTxManager) Stop() {
	close(tm.finishChannel)
}

func (tm *syncTxManager) runManager() {
	defer func() {
		if panicMsg := recover(); panicMsg != nil {
			tm.logger.Warn().Str("callStack", string(debug.Stack())).Str("errMsg", fmt.Sprintf("%v", panicMsg)).Msg("panic ocurred tx sync task")
		}
	}()
	tm.logger.Debug().Msg("syncTXManager started")

	// set interval of trying to resend getTransaction
MANLOOP:
	for {
		select {
		case <-tm.fcTicker.C:
			tm.refineFrontCache()
		case task := <-tm.taskChannel:
			task()
		case <-tm.finishChannel:
			tm.fcTicker.Stop()
			break MANLOOP
		}
	}
	tm.logger.Debug().Msg("syncTXManager finished")
}
func (tm *syncTxManager) runQueryLog() {
	defer func() {
		if panicMsg := recover(); panicMsg != nil {
			tm.logger.Warn().Str("callStack", string(debug.Stack())).Str("errMsg", fmt.Sprintf("%v", panicMsg)).Msg("panic occurred handle get tx queries")
		}
	}()
	// set interval of trying to resend getTransaction
	tm.logger.Debug().Msg("syncTXManager starting query routine")

MANLOOP:
	for {
		select {
		case task := <-tm.taskQueryChannel:
			task()
		case <-tm.finishChannel:
			break MANLOOP
		}
	}
	tm.logger.Debug().Msg("syncTXManager finished query routine")
}

func (tm *syncTxManager) registerTxNotice(txs []*types.Tx) {
	tm.taskChannel <- func() {
		for _, tx := range txs {
			tm.moveToMPCache(tx)
		}
	}
}

// pre-allocated slices to reduce memory allocation. this buffers must used inside syncTXManager goroutine.
var (
	// for general usage
	addBuf         = make([]types.TxID, 0, DefaultPeerTxQueueSize)
	dupBuf         = make([]types.TxID, 0, DefaultPeerTxQueueSize)
	queuedBuf      = make([]types.TxID, 0, DefaultPeerTxQueueSize)
	cleanupCounter = 0
	// idsBuf is used for indivisual peer
	idsBuf    = make([][]types.TxID, 0, 10)
	bufOffset = 0
)

// getIDsBuf return empty slice with capacity DefaultPeerTxQueueSize
func getIDsBuf(idx int) []types.TxID {
	for idx >= len(idsBuf) {
		idsBuf = append(idsBuf, make([]types.TxID, 0, DefaultPeerTxQueueSize))
	}
	return idsBuf[idx][:0]
}

func (tm *syncTxManager) HandleNewTxNotice(peer p2pcommon.RemotePeer, txIDs []types.TxID, data *types.NewTransactionsNotice) {
	tm.taskChannel <- func() {
		peerID := peer.ID()
		now := time.Now()
		newComer := addBuf[:0]
		duplicated := dupBuf[:0]
		queued := queuedBuf[:0]

		for _, txID := range txIDs {
			// If you want to strict check, query tx to cahinservice. It is skipped since it's so time consuming
			// mempool has tx already
			if ok := tm.txCache.Contains(txID); ok {
				duplicated = append(duplicated, txID)
				continue
			}
			// check if tx is in front cache
			if info, ok := tm.frontCache[txID]; ok {
				// other peer sent notice already. so add peerid to next waiting list
				appendPeerID(info, peerID)
				queued = append(queued, txID)
				continue
			}

			info := &incomingTxNotice{hash: txID, created: now, lastSent: now}
			tm.frontCache[txID] = info
			newComer = append(newComer, txID)
		}

		if len(newComer) > 0 {
			if len(newComer) <= len(txIDs) {
				copy(txIDs, newComer)
				txIDs = txIDs[:len(newComer)]
			}
			tm.sendGetTx(peer, txIDs)
		}
		if len(queued) > 0 {
			toQueue := make([]types.TxID, len(queued))
			copy(toQueue, queued)
			tm.toNoticeIdQueue.PushBack(&queryQueue{peerID: peerID, txIDs: toQueue})
		}

		tm.logger.Trace().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Int("newCnt", len(newComer)).Int("queCnt", len(queued)).Int("dupCnt", len(duplicated)).Array("newComer", types.NewLogTxIDsMarshaller(newComer, 10)).Array("duplicated", types.NewLogTxIDsMarshaller(duplicated, 10)).Array("queued", types.NewLogTxIDsMarshaller(queued, 10)).Int("frontCacheSize", len(tm.frontCache)).Msg("push txs, to query next time")
	}
}

func (tm *syncTxManager) sendGetTxs(peer p2pcommon.RemotePeer, ids []types.TxID) {
	tm.logger.Debug().Int("tx_cnt", len(ids)).Array("hashes", types.NewLogTxIDsMarshaller(ids, 10)).Msg("syncManager request back unknown tx hashes")
	receiver := NewGetTxsReceiver(tm.actor, peer, tm.sm, tm.logger, ids, p2pcommon.DefaultActorMsgTTL)
	receiver.StartGet()
}

func (tm *syncTxManager) HandleGetTxReq(peer p2pcommon.RemotePeer, msgID p2pcommon.MsgID, data *types.GetTransactionsRequest) error {
	select {
	case tm.taskQueryChannel <- func() {
		reqHashes := data.Hashes
		tm.handleTxReq(peer, msgID, reqHashes)
	}:
		return nil
	default:
		return p2pcommon.SyncManagerBusyError

	}
}

func (tm *syncTxManager) retryGetTx(peerID types.PeerID, hashes [][]byte) {
	tm.taskChannel <- func() {
		txIDs := make([]types.TxID, len(hashes))
		for i, hash := range hashes {
			txIDs[i] = types.ToTxID(hash)
		}
		tm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Array("txIDs", types.NewLogTxIDsMarshaller(txIDs, 10)).Msg("push txs that are failed to get by server busy")
		tm.pushBackToFrontCache(peerID, txIDs)
	}
}

func (tm *syncTxManager) pushBackToFrontCache(peerID types.PeerID, txIDs []types.TxID) {
	// this method is called when the sending is failed by remote peer is busy.
	// resetting last sent time will trigger immediate query of that tx.
	// push back
	pushedCount := 0
	for _, txID := range txIDs {
		// only search front cache.
		if info, ok := tm.frontCache[txID]; ok {
			// other peer sent notice already and ready to
			appendPeerID(info, peerID)
			info.lastSent = unsent
			pushedCount++
		}
	}
	if pushedCount > 0 {
		tm.toNoticeIdQueue.PushFront(&queryQueue{peerID: peerID, txIDs: txIDs})
	}
}

func (tm *syncTxManager) burnFailedTxFrontCache(peerID types.PeerID, txIDs []types.TxID) {
	for _, txID := range txIDs {
		// only search front cache.
		if info, ok := tm.frontCache[txID]; ok {
			if len(info.peers) > 0 {
				// make send gettx to other peer
				info.lastSent = unsent
			} else {
				delete(tm.frontCache, txID)
			}
		}
	}
}

// this function must called only if ticket can be retrieved.
func (tm *syncTxManager) handleTxReq(remotePeer p2pcommon.RemotePeer, mID p2pcommon.MsgID, reqHashes [][]byte) {
	// NOTE size estimation is tied to protobuf3 it should be changed when protobuf is changed.
	// find transactions from chainservice
	idx := 0
	status := types.ResultStatus_OK
	var hashes, mpReqs []types.TxHash
	var txInfos []*types.Tx
	var reqIDs = make([]types.TxID, len(reqHashes))
	var txs = make(map[types.TxID]*types.Tx)
	payloadSize := subproto.EmptyGetBlockResponseSize
	var txSize, fieldSize int

	bucket := message.MaxReqestHashes
	var futures []interface{}

	var inCache, inMempool = 0, 0
	// 1. first check in cache
	for i, h := range reqHashes {
		reqIDs[i] = types.ToTxID(h)
		tx, ok := tm.txCache.Get(reqIDs[i])
		if ok {
			txs[reqIDs[i]] = tx.(*types.Tx)
			inCache++
		} else {
			mpReqs = append(mpReqs, h)
		}
	}

	for _, h := range mpReqs {
		hashes = append(hashes, h)
		if len(hashes) == bucket {
			if f, err := tm.actor.CallRequestDefaultTimeout(message.MemPoolSvc,
				&message.MemPoolExistEx{Hashes: hashes}); err == nil {
				futures = append(futures, f)
			}
			hashes = nil
		}
	}
	if hashes != nil {
		if f, err := tm.actor.CallRequestDefaultTimeout(message.MemPoolSvc,
			&message.MemPoolExistEx{Hashes: hashes}); err == nil {
			futures = append(futures, f)
		}
	}
	hashes = nil
	idx = 0
	for _, f := range futures {
		if tmp, err := tm.msgHelper.ExtractTxsFromResponseAndError(f, nil); err == nil {
			for _, tx := range tmp {
				if tx == nil {
					continue
				}
				txs[types.ToTxID(tx.Hash)] = tx
				inMempool++
			}
		} else {
			tm.logger.Debug().Err(err).Msg("ErrExtract tx in future")
		}
	}
	msgCnt := 0
	for _, tid := range reqIDs {
		tx, ok := txs[tid]
		if !ok {
			continue
		}
		hash := tx.GetHash()
		txSize = proto.Size(tx)

		fieldSize = txSize + p2putil.CalculateFieldDescSize(txSize)
		fieldSize += len(hash) + p2putil.CalculateFieldDescSize(len(hash))

		if uint32(payloadSize+fieldSize) > p2pcommon.MaxPayloadLength {
			// send partial list
			resp := &types.GetTransactionsResponse{
				Status: status,
				Hashes: hashes,
				Txs:    txInfos, HasNext: true}
			tm.logger.Trace().Int(p2putil.LogTxCount, len(hashes)).
				Str(p2putil.LogOrgReqID, mID.String()).Msg("Sending partial response")

			remotePeer.SendMessage(remotePeer.MF().
				NewMsgResponseOrder(mID, p2pcommon.GetTXsResponse, resp))
			msgCnt++
			hashes, txInfos, payloadSize = nil, nil, subproto.EmptyGetBlockResponseSize
		}

		hashes = append(hashes, hash)
		txInfos = append(txInfos, tx)
		payloadSize += fieldSize
		idx++
	}
	// generate response message
	if 0 == idx {
		// if no tx is found, set status tu not found
		status = types.ResultStatus_NOT_FOUND
	}
	resp := &types.GetTransactionsResponse{
		Status: status,
		Hashes: hashes,
		Txs:    txInfos, HasNext: false}
	tm.logger.Trace().Int(p2putil.LogTxCount, len(hashes)).
		Str(p2putil.LogOrgReqID, mID.String()).Str(p2putil.LogRespStatus, status.String()).Msg("Sending last part response")
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(mID, p2pcommon.GetTXsResponse, resp))
	msgCnt++
	tm.logger.Debug().Int("respMsgCnt", msgCnt).
		Int("inCache", inCache).Int("inMempool", inMempool).
		Str(p2putil.LogOrgReqID, mID.String()).Str(p2putil.LogRespStatus, status.String()).
		Msg("handled getTx query")
}

func (tm *syncTxManager) refineFrontCache() {
	now := time.Now()
	expireTime := now.Add(-txQueryTimeout)
	if tm.toNoticeIdQueue.Len() == 0 { // nothing to resend
		cleanupCounter++
		if cleanupCounter%20 == 0 {
			cleanupCounter = 0
			if len(tm.frontCache) > 0 {
				tm.cleanupFrontCache(expireTime)
			}
		}
		return
	}
	tm.logger.Trace().Int("noticeQueues", tm.toNoticeIdQueue.Len()).Int("frontCache", len(tm.frontCache)).Msg("refining front cache")

	// init
	expired := dupBuf[:0]
	done := addBuf[:0]
	bufOffset = 0

	// assume peer is all available for now
	sendMap := make(map[types.PeerID]*[]types.TxID)
	// find txs that should query to peers
	// tx in front cache has tri-state: unsent, waitingResp, expiredWaiting

	var next *list.Element
	for e := tm.toNoticeIdQueue.Front(); e != nil; e = next {
		next = e.Next()
		queAgain := queuedBuf[:0]
		queuedIDs := e.Value.(*queryQueue)
		toSend := tm.allocIDSlice(queuedIDs.peerID, sendMap)
		if len(*toSend) >= DefaultPeerTxQueueSize {
			// list is full. skip this peer
			continue
		}

		idSize := len(queuedIDs.txIDs)
		toSendCnt := 0
		for j := 0; j < idSize; j++ {
			txID := queuedIDs.txIDs[j]
			info := tm.frontCache[txID]
			if info == nil { // tx is done to mempool or block. this txid is safe to delete
				done = append(done, txID)
				continue
			}
			if info.lastSent.After(expireTime) {
				// txs that wait for getTXResp and not expired will wait more time.
				queAgain = append(queAgain, txID)
				continue
			}
			if len(info.peers) == 0 {
				// remove old or unsent tx that has no peer to query.
				expired = append(expired, txID)
				delete(tm.frontCache, txID)
			}

			if tm.addToList(info, queuedIDs.peerID, toSend) {
				info.lastSent = now
				toSendCnt++
				if len(*toSend) >= DefaultPeerTxQueueSize {
					queAgain = append(queAgain, queuedIDs.txIDs[j+1:]...)
					break
				}
			} else {
				queAgain = append(queAgain, txID)
			}
		}

		// if not all txs is filled, the unsent will be pushed front to try send in next turn.
		if len(queAgain) > 0 {
			// reuse allocated slice
			toQueue := queuedIDs.txIDs[:len(queAgain)]
			copy(toQueue, queAgain)
			tm.logger.Trace().Array("queAgain", types.NewLogTxIDsMarshaller(toQueue, 10)).Msg("syncManager enqueue txIDs again that waiting for response")

			e.Value = &queryQueue{peerID: queuedIDs.peerID, txIDs: toQueue}
		} else {
			tm.toNoticeIdQueue.Remove(e)
		}
	}

	if len(expired) > 0 {
		tm.logger.Debug().Array("done", types.NewLogTxIDsMarshaller(done, 10)).Array("expired", types.NewLogTxIDsMarshaller(expired, 10)).Msg("syncManager deletes txIDs that are not needed anymore")
	}

	for peerID, idsP := range sendMap {
		ids := *idsP
		if len(ids) == 0 {
			// no tx to send
			continue
		}
		if peer, ok := tm.pm.GetPeer(peerID); ok {
			tm.sendGetTx(peer, ids)
		} else {
			// peer probably disconnected.
			tm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Array("hashes", types.NewLogTxIDsMarshaller(ids, 10)).Msg("syncManager failed to send get tx, since peer is disconnected just before")
			toRetry := make([]types.TxID, len(ids))
			copy(toRetry, ids)
			tm.burnFailedTxFrontCache(peerID, toRetry)
		}
	}
}

func (tm *syncTxManager) sendGetTx(peer p2pcommon.RemotePeer, ids []types.TxID) {
	tm.logger.Trace().Str(p2putil.LogPeerName, peer.Name()).Array("hashes", types.NewLogTxIDsMarshaller(ids, 10)).Msg("syncManager try to get tx to remote peer")
	// create message data
	receiver := NewGetTxsReceiver(tm.actor, peer, tm.sm, tm.logger, ids, p2pcommon.DefaultActorMsgTTL)
	receiver.StartGet()
}

// assignTxToPeer set tx how to select peer for querying
func (tm *syncTxManager) allocIDSlice(peerID types.PeerID, sendMap map[types.PeerID]*[]types.TxID) *[]types.TxID {
	idsP, ok := sendMap[peerID]
	if !ok {
		list := getIDsBuf(bufOffset)
		bufOffset++
		idsP = &list
		sendMap[peerID] = idsP
	}
	return idsP
}

// addToList check
func (tm *syncTxManager) addToList(info *incomingTxNotice, target types.PeerID, ids *[]types.TxID) bool {
	for i, peerID := range info.peers {
		if types.IsSamePeerID(peerID, target) {
			// remove peerID from wait queue
			info.peers = append(info.peers[:i], info.peers[i+1:]...)
			*ids = append(*ids, info.hash)
			return true
		}
	}
	return false
}

// assignTxToPeer set tx how to select peer for querying
func (tm *syncTxManager) assignTxToPeer(info *incomingTxNotice, sendMap map[types.PeerID][]types.TxID) bool {
	for i, peerID := range info.peers {
		list, ok := sendMap[peerID]
		if !ok {
			list = getIDsBuf(bufOffset)
			bufOffset++
		}
		if len(list) >= DefaultPeerTxQueueSize {
			// reached max count in a single query
			continue
		}
		list = append(list, info.hash)
		info.peers = append(info.peers[:i], info.peers[i+1:]...)
		sendMap[peerID] = list
		return true
	}
	return false
}

func (tm *syncTxManager) moveToMPCache(tx *types.Tx) {
	txID := types.ToTxID(tx.Hash)
	delete(tm.frontCache, txID)
	tm.txCache.Add(txID, tx)
	tm.logger.Trace().Str("txID", txID.String()).Msg("syncManager caches tx")
}

// cleanupFrontCache clean unnecessary frontCache items. These are txs that sent request
func (tm *syncTxManager) cleanupFrontCache(expireTime time.Time) {
	testCnt, expired := 0, 0
	for txID, info := range tm.frontCache {
		if (!info.lastSent.After(expireTime)) && len(info.peers) == 0 {
			// remove old or unsent tx that has no peer to query.
			expired++
			delete(tm.frontCache, txID)
		}
		testCnt++
		if testCnt >= 10000 {
			break
		}
	}
	tm.logger.Debug().Int("testCnt", testCnt).Int("expireCnt", expired).Msg("syncManager clean up some of expired items in frontCache")

}

func appendPeerID(info *incomingTxNotice, peerID types.PeerID) {
	info.peers = append(info.peers, peerID)
	if len(info.peers) >= inTxPeerBufSize {
		info.peers = info.peers[1:]
	}
}

type incomingTxNotice struct {
	hash     types.TxID
	created  time.Time
	lastSent time.Time
	trial    int
	peers    []types.PeerID
}

// By is the type of a "less" function that defines the ordering of its Planet arguments.
type By func(p1, p2 *incomingTxNotice) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(notices []incomingTxNotice) {
	ps := &txSorter{
		notices: notices,
		by:      by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// planetSorter joins a By function and a slice of Planets to be sorted.
type txSorter struct {
	notices []incomingTxNotice
	by      func(p1, p2 *incomingTxNotice) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *txSorter) Len() int {
	return len(s.notices)
}

// Swap is part of sort.Interface.
func (s *txSorter) Swap(i, j int) {
	s.notices[i], s.notices[j] = s.notices[j], s.notices[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *txSorter) Less(i, j int) bool {
	return s.by(&s.notices[i], &s.notices[j])
}
