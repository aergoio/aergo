package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog"
	"runtime/debug"
	"sort"
	"time"
)

const minimumTxQueryInterval = time.Second >> 2
const txQueryTimeout = time.Second << 2
var unsent = time.Time{}

// syncTxManager handle operations about tx sync
type syncTxManager struct {
	logger *log.Logger

	sm        p2pcommon.SyncManager
	actor     p2pcommon.ActorService
	pm        p2pcommon.PeerManager
	msgHelper message.Helper

	txCache       *lru.Cache
	// received notice but not in my mempool
	frontCache       map[types.TxID]*incomingTxNotice
	taskChannel      chan smTask
	taskQueryChannel chan smTask
	finishChannel    chan struct{}

	getTicker   *time.Ticker
}

type smTask func()

func newTxSyncManager(sm p2pcommon.SyncManager, actor p2pcommon.ActorService, pm p2pcommon.PeerManager, logger *log.Logger) *syncTxManager {
	tm := &syncTxManager{sm:sm, actor: actor, pm: pm, logger: logger,
		frontCache:       make(map[types.TxID]*incomingTxNotice),
		taskChannel:      make(chan smTask, 20),
		finishChannel:    make(chan struct{}, 1),
		taskQueryChannel: make(chan smTask, 10),

		msgHelper:   message.GetHelper(),
		getTicker:   time.NewTicker(minimumTxQueryInterval),
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

func (tm *syncTxManager) runManager() {
	defer func() {
		if panicMsg := recover(); panicMsg != nil {
			tm.logger.Warn().Str("callStack", string(debug.Stack())).Str("errMsg",fmt.Sprintf("%v",panicMsg)).Msg("panic ocurred tx sync task")
		}
	}()
	tm.logger.Debug().Msg("syncTXManager started")

	// set interval of trying to resend getTransaction
MANLOOP:
	for {
		select {
		case <-tm.getTicker.C:
			tm.refineFrontCache()
		case task := <-tm.taskChannel:
			task()
		case <-tm.finishChannel:
			tm.getTicker.Stop()
			break MANLOOP
		}
	}
	tm.logger.Debug().Msg("syncTXManager finished")
}
func (tm *syncTxManager) runQueryLog() {
	defer func() {
		if panicMsg := recover(); panicMsg != nil {
			tm.logger.Warn().Str("callStack", string(debug.Stack())).Str("errMsg",fmt.Sprintf("%v",panicMsg)).Msg("panic occurred handle get tx queries")
		}
	}()
	// set interval of trying to resend getTransaction
MANLOOP:
	for {
		select {
		case task := <-tm.taskQueryChannel:
			task()
		case <-tm.finishChannel:
			break MANLOOP
		}
	}
}

func (tm *syncTxManager) Stop() {
	close(tm.finishChannel)
}

func (tm *syncTxManager) registerTxNotice(txs []*types.Tx) {
	tm.taskChannel <- func() {
		for _, tx := range txs {
			tm.moveToMPCache(tx)
		}
		// tm.logger.Debug().Array("txIDs", types.NewLogTxIDsMarshaller(txIDs, 10)).Msg("syncManager caches txs")
	}
}

// pre-allocated slices to reduce memory allocation. this buffers must used inside syncTXManager goroutine.
var (
	// for general usage
	addBuf = make([]types.TxID,0,DefaultPeerTxQueueSize)
	dupBuf = make([]types.TxID,0,DefaultPeerTxQueueSize)
	queuedBuf = make([]types.TxID,0,DefaultPeerTxQueueSize)

	// idsBuf is used for indivisual peer
	idsBuf = make([][]types.TxID,0,10)
 	bufOffset = 0
)

// getIDsBuf return empty slice with capacity DefaultPeerTxQueueSize
func getIDsBuf(idx int) []types.TxID {
	for idx >= len(idsBuf) {
		idsBuf = append(idsBuf,make([]types.TxID,0,DefaultPeerTxQueueSize))
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
				// other peer sent notice already and ready to
				appendPeerID(info, peerID)
				queued = append(queued,txID)
				continue
			}

			info := &incomingTxNotice{hash: txID, created: now, lastSent: unsent}
			tm.frontCache[txID] = info
			appendPeerID(info, peerID)
			newComer = append(newComer,txID)
		}
		tm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Array("newComer", types.NewLogTxIDsMarshaller(newComer, 10)).Array("duplicated", types.NewLogTxIDsMarshaller(duplicated, 10)).Array("queued", types.NewLogTxIDsMarshaller(queued, 10)).Int("frontCacheSize",len(tm.frontCache)).Msg("push txs, to query next time")
	}
}

func (tm *syncTxManager) sendGetTxs(peer p2pcommon.RemotePeer, ids []types.TxID) {
	tm.logger.Debug().Int("tx_cnt", len(ids)).Array("hashes", types.NewLogTxIDsMarshaller(ids, 10)).Msg("syncManager request back unknown tx hashes")
	receiver := NewGetTxsReceiver(tm.actor, peer, tm.sm, ids ,p2pcommon.DefaultActorMsgTTL)
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
		txIDs := make([]types.TxID,len(hashes))
		for i, hash := range hashes {
			txIDs[i] = types.ToTxID(hash)
		}
		tm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Array("txIDs", types.NewLogTxIDsMarshaller(txIDs, 10)).Msg("push txs that are failed to get by server busy")
		tm.pushBackToFrontCache(peerID, txIDs)
	}
}

func (tm *syncTxManager) pushBackToFrontCache(peerID types.PeerID, txIDs []types.TxID) {
	// this method is called when the sending is failed by remote peer is busy or disconnected.
	// resetting last sent time will trigger immediate query of that tx.
	// push back
	for _, txID := range txIDs {
		// only search front cache.
		if info, ok := tm.frontCache[txID]; ok {
			// other peer sent notice already and ready to
			appendPeerID(info, peerID)
			info.lastSent = unsent
		}
	}
}
func (tm *syncTxManager) burnFailedTxFrontCache(peerID types.PeerID, txIDs []types.TxID) {
	// this method is called when the sending is failed by remote peer is busy or disconnected.
	// resetting last sent time will trigger immediate query of that tx.
	// push back
	for _, txID := range txIDs {
		// only search front cache.
		if info, ok := tm.frontCache[txID]; ok {
			// other peer sent notice already and ready to
			info.lastSent = unsent
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

	// 1. first check in cache
	for i, h := range reqHashes {
		reqIDs[i] = types.ToTxID(h)
		tx, ok := tm.txCache.Get(reqIDs[i])
		if ok {
			txs[reqIDs[i]] = tx.(*types.Tx)
		} else {
			mpReqs = append(mpReqs,h)
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
			}
		} else {
			tm.logger.Debug().Err(err).Msg("ErrExtract tx in future")
		}
	}
	for _, tid := range reqIDs {
		tx, ok := txs[tid]
		if !ok {
			continue
		}
		hash := tx.GetHash()
		txSize = proto.Size(tx)

		fieldSize = txSize + p2putil.CalculateFieldDescSize(txSize)
		fieldSize += len(hash) + p2putil.CalculateFieldDescSize(len(hash))

		if (payloadSize + fieldSize) > p2pcommon.MaxPayloadLength {
			// send partial list
			resp := &types.GetTransactionsResponse{
				Status: status,
				Hashes: hashes,
				Txs:    txInfos, HasNext: true}
			tm.logger.Debug().Int(p2putil.LogTxCount, len(hashes)).
				Str(p2putil.LogOrgReqID, mID.String()).Msg("Sending partial response")

			remotePeer.SendMessage(remotePeer.MF().
				NewMsgResponseOrder(mID, p2pcommon.GetTXsResponse, resp))
			hashes, txInfos, payloadSize = nil, nil, subproto.EmptyGetBlockResponseSize
		}

		hashes = append(hashes, hash)
		txInfos = append(txInfos, tx)
		payloadSize += fieldSize
		idx++
	}
	if 0 == idx {
		status = types.ResultStatus_NOT_FOUND
	}
	tm.logger.Debug().Int(p2putil.LogTxCount, len(hashes)).
		Str(p2putil.LogOrgReqID, mID.String()).Str(p2putil.LogRespStatus, status.String()).Msg("Sending last part response")
	// generate response message

	resp := &types.GetTransactionsResponse{
		Status: status,
		Hashes: hashes,
		Txs:    txInfos, HasNext: false}
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(mID, p2pcommon.GetTXsResponse, resp))
}

//
func (tm *syncTxManager) refineFrontCache() {
	if len(tm.frontCache) == 0 {
		return
	}
	//tm.logger.Debug().Int("frontCache",len(tm.frontCache)).Msg("refining front cache")

	// init
	deleted := queuedBuf[:0]
	bufOffset = 0

	now := time.Now()
	// assume peer is all available for now
	sendMap := make(map[types.PeerID][]types.TxID)
	// find txs that should query to peers
	expireTime := now.Add(-txQueryTimeout)
	// tx in front cache has tri-state: unsent, waitingResp, expiredWaiting
	for txID, info := range tm.frontCache {
		if info.lastSent.After(expireTime) {
			// txs that wait for getTXResp and not expired will wait more time.
			continue
		}
		if len(info.peers) == 0 {
			// remove old or unsent tx that has no peer to query.
			deleted = append(deleted, txID)
			delete(tm.frontCache, txID)
		}
		if tm.assignTxToPeer(info, sendMap) {
			info.lastSent = now

		}
	}

	if len(deleted) > 0 {
		tm.logger.Debug().Array("hashes", types.NewLogTxIDsMarshaller(deleted, 10)).Msg("syncManager deletes txs that was expired and has no additional peers to query")
	}

	for peerID, ids := range sendMap {
		if len(ids) == 0 {
			// no tx to send
			continue
		}
		if peer, ok := tm.pm.GetPeer(peerID); ok {
			tm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Array("hashes", types.NewLogTxIDsMarshaller(ids, 10)).Msg("syncManager try to get tx to other peers")
			// create message data
			receiver := NewGetTxsReceiver(tm.actor, peer, tm.sm, ids, p2pcommon.DefaultActorMsgTTL)
			receiver.StartGet()
		} else {
			// peer probably disconnected.
			tm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Array("hashes", types.NewLogTxIDsMarshaller(ids, 10)).Msg("syncManager failed to send get tx, since peer is disconnected just before")
			tm.burnFailedTxFrontCache(peerID, ids)
		}
	}
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
}

func appendPeerID(info *incomingTxNotice, peerID types.PeerID) {
	info.peers = append(info.peers, peerID)
	if len(info.peers) >= inTxPeerBufSize {
		info.peers = info.peers[1:]
	}
}

type logTXHashesMarshaler struct {
	arr   []message.TXHash
	limit int
}

func newLogTXHashesMarshaler(bbarray []message.TXHash, limit int) *logTXHashesMarshaler {
	return &logTXHashesMarshaler{arr: bbarray, limit: limit}
}

func (m logTXHashesMarshaler) MarshalZerologArray(a *zerolog.Array) {
	size := len(m.arr)
	if size > m.limit {
		for i := 0; i < m.limit-1; i++ {
			a.Str(enc.ToString(m.arr[i]))
		}
		a.Str(fmt.Sprintf("(and %d more)", size-m.limit+1))
	} else {
		for _, element := range m.arr {
			a.Str(enc.ToString(element))
		}
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
