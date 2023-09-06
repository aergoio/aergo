package p2p

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/message/messagemock"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_syncTxManager_HandleNewTxNotice(t *testing.T) {
	// check if
	//    sm sends query txs that is not in all caches.
	//    sm store to queryqueue that that are in front cache.
	//    sm drops txs that are in tx cache
	logger := log.NewLogger("tt.p2p")
	rawHashes := sampleTxs
	txHashes := sampleTxIDs

	// tt if new block notice comes
	tests := []struct {
		name      string
		front     []types.TxID
		inCache   []types.TxID
		wantFront []types.TxID
		wantCache []types.TxID
		wantToGet []types.TxID
		wantQueue []types.TxID
	}{
		// 0. Succ : valid tx hashes and not exist in local cache
		{"TSuccAllNew", nil, nil, txHashes, nil, txHashes, nil},
		// 1. : some txs are in front cache. and only txs not in front cache are sent getTx
		{"TInFront", txHashes[:2], nil, txHashes, nil, txHashes[2:], txHashes[:2]},
		// 2. Succ : valid tx hashes and partially exist in local cache
		{"TBoth", txHashes[:2], txHashes[2:4], ct(txHashes[:2], txHashes[4:]), txHashes[2:4], txHashes[4:], txHashes[:2]},
		// 3. Succ : valid tx hashes and all exist in local cache
		{"TAllFront", txHashes, nil,
			txHashes, nil, nil, txHashes},
		// 4. Succ : valid tx hashes and partially exist in local cache
		{"TCachePart", nil, txHashes[2:], txHashes[:2], txHashes[2:], txHashes[:2], nil},
		// 5. Succ : valid tx hashes and all exist in local cache
		{"TSuccExistAll", nil, txHashes, nil, txHashes, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)

			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMO := p2pmock.NewMockMsgOrder(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()
			mockPeer.EXPECT().Name().Return(sampleMeta.ID.String()).AnyTimes()

			if len(tt.wantToGet) > 0 {
				mockPeer.EXPECT().MF().Return(mockMF).MinTimes(1)
				mockPeer.EXPECT().SendMessage(mockMO).MinTimes(1)
				mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), p2pcommon.GetTXsRequest, gomock.Any()).Return(mockMO).MinTimes(1)
				mockMO.EXPECT().GetMsgID().Return(p2pcommon.NewMsgID()).MinTimes(1)
			}

			data := &types.NewTransactionsNotice{TxHashes: rawHashes}
			tm := newTxSyncManager(nil, mockActor, mockPM, logger)

			if tt.front != nil {
				for _, hash := range tt.front {
					tm.frontCache[hash] = &incomingTxNotice{hash: hash}
				}
			}
			if tt.inCache != nil {
				for _, hash := range tt.inCache {
					tm.txCache.Add(hash, true)
				}
			}
			tm.Start()

			in := make([]types.TxID, len(txHashes))
			copy(in, txHashes)
			tm.HandleNewTxNotice(mockPeer, in, data)

			// make terminate
			tm.taskChannel <- func() {
				tm.Stop()
			}
			<-tm.finishChannel

			if len(tm.frontCache) != len(tt.wantFront) {
				t.Fatalf("front len %v, want %v", len(tm.frontCache), len(tt.wantFront))
			} else {
				for _, id := range tt.wantFront {
					if _, ok := tm.frontCache[id]; !ok {
						t.Errorf("txId %v is not in frontCache, want exist", id)
					}
				}
			}
			if len(tt.wantCache) != tm.txCache.Len() {
				t.Fatalf("txCache len %v, want %v", tm.txCache.Len(), len(tt.wantCache))
			} else {
				for _, id := range tt.wantCache {
					if !tm.txCache.Contains(id) {
						t.Errorf("txId %v is not in txCache, want exist", id)
					}
				}
			}
			if (tm.toNoticeIdQueue.Len() > 0) != (len(tt.wantQueue) > 0) {
				t.Fatalf("expected queued txs is %v, want exist %v", tm.toNoticeIdQueue, len(tt.wantQueue) > 0)
			}
			if len(tt.wantQueue) > 0 {
				q := tm.toNoticeIdQueue.Front()
				if !equalTXIDs(q.Value.(*queryQueue).txIDs, tt.wantQueue) {
					t.Fatalf("expected queued txs is %v, want %v", q.Value.(*queryQueue).txIDs, tt.wantQueue)
				}

			}

		})
	}
}

func equalTXIDs(a []types.TxID, b []types.TxID) bool {
	if len(a) != len(b) {
		return false
	}
	for i, e1 := range a {
		e2 := b[i]
		if !(types.HashID(e1).Equal(types.HashID(e2))) {
			return false
		}
	}
	return true
}

var txSample []*types.Tx
var txIDSample []types.TxID
var pidSample []types.PeerID

const idSampleSize = 20
const pidSampleSize = 5

func init() {
	txSample = make([]*types.Tx, idSampleSize)
	txIDSample = make([]types.TxID, idSampleSize)
	for i := 0; i < idSampleSize; i++ {
		txSample[i] = types.NewTx()
		txSample[i].Body.Nonce = uint64(i)
		txSample[i].Hash = txSample[i].CalculateTxHash()
		txIDSample[i] = types.ToTxID(txSample[i].Hash)
	}
	pidSample = make([]types.PeerID, pidSampleSize)
	for i := 0; i < pidSampleSize; i++ {
		pidSample[i] = types.RandomPeerID()
	}
}

// mq make list of queyrQueue. the params must be pair. (i.e first is peerid, second is list of txID, and so on.
func mq(arg ...interface{}) []*queryQueue {
	size := len(arg) / 2
	var qs = make([]*queryQueue, 0, size)
	for i := 0; i < size; i++ {
		pid := pidSample[arg[i*2].(int)]
		org := arg[i*2+1].([]types.TxID)
		ids := make([]types.TxID, len(org))
		copy(ids, org)
		q := &queryQueue{peerID: pid, txIDs: ids}
		qs = append(qs, q)
	}
	return qs
}

// ct means concatenateTxIDs
func ct(slis ...[]types.TxID) []types.TxID {
	ret := make([]types.TxID, 0, len(slis))
	for _, s := range slis {
		ret = append(ret, s...)
	}
	return ret
}

func Test_txSyncManager_refineFrontCacheConsumption(t *testing.T) {
	// this test should check if...
	//   syncManager sends getTXrequests as expected
	//     if each que has small txs, the tm combine txs with same peers and send to single query
	//     if que has cut, remain txs is in front of que
	//   queryQueue is lowered as expected

	sampleSize := 6000
	logger := log.NewLogger("tt.p2p")
	txs := make([]*types.Tx, sampleSize)
	tids := make([]types.TxID, sampleSize)
	for i := 0; i < sampleSize; i++ {
		txs[i] = types.NewTx()
		txs[i].Body.Nonce = uint64(i)
		txs[i].Hash = txs[i].CalculateTxHash()
		tids[i] = types.ToTxID(txs[i].Hash)
		if i%100 < 2 {
			t.Logf("tid %04d : %v", i, tids[i])
		}
	}
	ps := pidSample

	tests := []struct {
		name string
		ques []*queryQueue

		wantSend   []*queryQueue
		wantRemain []*queryQueue
	}{
		// 0. assumption: every txids is in front cache also
		// 1. small different notces from multiple peers. all notices should be consumed
		{"TMultiSmall", mq(0, tids[:2], 1, tids[2:3], 2, tids[3:6], 3, tids[10:20]), mq(0, tids[:2], 1, tids[2:3], 2, tids[3:6], 3, tids[10:20]), nil},
		// 2. small duplcated notices (=sent to other peer already) from multiple peers. duplicated notices will remain
		{"TDupWhole", mq(0, tids[:3], 1, tids[:3], 2, tids[:3], 3, tids[3:6]), mq(0, tids[:3], 3, tids[3:6]), mq(1, tids[:3], 2, tids[:3])},
		// 3. part of tx ids are dup
		{"TDupPart", mq(0, tids[:3], 1, tids[2:5], 2, tids[3:6], 3, tids[4:8]), mq(0, tids[:3], 1, tids[3:5], 2, tids[5:6], 3, tids[6:8]), mq(1, tids[2:3], 2, tids[3:5], 3, tids[4:6])},
		// 4. same peer sends lots of notices
		{"TFreq", mq(0, tids[:3], 0, tids[3:10], 0, tids[10:20], 0, tids[20:100], 0, tids[200:400]), mq(0, ct(tids[:100], tids[200:400])), nil},
		{"TBig", mq(0, tids[:1500], 0, tids[1500:3000], 0, tids[3000:4500], 0, tids[4500:6000]), mq(0, tids[:2000]), mq(0, tids[2000:3000], 0, tids[3000:4500], 0, tids[4500:6000])},
		{"TMultiBig", mq(0, tids[:1500], 1, tids[3000:4500], 0, tids[1500:3000], 1, tids[4500:6000]), mq(0, tids[:2000], 1, tids[3000:5000]), mq(0, tids[2000:3000], 1, tids[5000:6000])},
		{"TMultiNew", mq(0, tids[:1500], 1, tids[1500:3000], 2, tids[3000:4500], 3, tids[4500:6000]), mq(0, tids[:1500], 1, tids[1500:3000], 2, tids[3000:4500], 3, tids[4500:6000]), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()
			mockPeer.EXPECT().Name().Return(sampleMeta.ID.Pretty()).AnyTimes()
			mockPM.EXPECT().GetPeer(gomock.Any()).Return(mockPeer, true).AnyTimes()

			tm := newTxSyncManager(nil, mockActor, mockPM, logger)
			for _, q := range tt.ques {
				tm.toNoticeIdQueue.PushBack(q)
			}
			now := time.Now()
			// put front cash all peers
			for _, tid := range tids {
				info := &incomingTxNotice{hash: tid, created: now, lastSent: unsent}
				pids := make([]types.PeerID, len(ps))
				copy(pids, ps)
				info.peers = pids
				tm.frontCache[tid] = info
			}

			sent := [][][]byte{}
			if len(tt.wantSend) > 0 {
				mockMF := p2pmock.NewMockMoFactory(ctrl)
				mockMO := p2pmock.NewMockMsgOrder(ctrl)
				mockPeer.EXPECT().MF().Return(mockMF).Times(len(tt.wantSend))
				mockPeer.EXPECT().SendMessage(mockMO).Times(len(tt.wantSend))
				mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), p2pcommon.GetTXsRequest, gomock.Any()).DoAndReturn(func(respReceiver p2pcommon.ResponseReceiver, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
					req := message.(*types.GetTransactionsRequest)
					sent = append(sent, req.Hashes)
					return mockMO
				}).Times(len(tt.wantSend))
				mockMO.EXPECT().GetMsgID().Return(p2pcommon.NewMsgID()).AnyTimes()
			}

			tm.refineFrontCache()

			// verifying tx was sent as expected.
			for i, hashes := range sent {
				// send order is not preserved
				var w *queryQueue
				// infer matching txid set
				for _, cand := range tt.wantSend {
					if bytes.Equal(hashes[0], types.HashID(cand.txIDs[0]).Bytes()) {
						w = cand
						break
					}
				}
				if w == nil {
					t.Fatalf("unexpected sent request %v", enc.ToString(hashes[i]))
				}
				wTids := w.txIDs
				if len(hashes) != len(wTids) {
					t.Fatalf("sent tx len %v, want %v", len(hashes), len(wTids))
				}
				for j, hash := range hashes {
					tid := types.ToTxID(hash)
					if !types.HashID(tid).Equal(types.HashID(wTids[j])) {
						t.Fatalf("idx %d:%d, sent txID %v, want %v", i, j, tid, wTids[j])
					}
				}
			}
			if tm.toNoticeIdQueue.Len() != len(tt.wantRemain) {
				t.Fatalf("remained queue len %v, want %v", tm.toNoticeIdQueue.Len(), len(tt.wantRemain))
			} else {
				idx := 0
				for e := tm.toNoticeIdQueue.Front(); e != nil; e = e.Next() {
					v := e.Value.(*queryQueue)
					w := tt.wantRemain[idx]
					if !types.IsSamePeerID(v.peerID, w.peerID) {
						t.Fatalf("idx %d, remained peerID %v, want %v", idx, v.peerID, tt.wantRemain[idx])
					}
					if len(v.txIDs) != len(w.txIDs) {
						t.Fatalf("remained queue len %v, want %v", tm.toNoticeIdQueue.Len(), len(tt.wantRemain))
					}
					for i, tid := range v.txIDs {
						if !types.HashID(tid).Equal(types.HashID(w.txIDs[i])) {
							t.Fatalf("idx %d, tid in queue %v, want %v", idx, tid, w.txIDs[i])
						}
					}
					idx++
				}
			}

		})
	}
}

func Test_txSyncManager_refineFrontCache(t *testing.T) {
	t.Skip("This test is for old implementation. This will be deleted if the code is sure no need to useful")
	// this test should check if...
	//   syncManager sends getTXrequests as expected
	//     if each que has small txs, the tm combine txs with same peers and send to single query
	//     if que has cut, remain txs is in front of que
	//   queryQueue is lowered as expected

	logger := log.NewLogger("tt.p2p")
	digest := sha256.New()
	oldTXIds := make([]types.TxID, 5)
	for i := 0; i < 5; i++ {
		digest.Write([]byte{byte(i)})
		b := digest.Sum(nil)
		oldTXIds[i] = types.ToTxID(b)
	}
	newTXIds := make([]types.TxID, 5)
	for i := 0; i < 5; i++ {
		digest.Write([]byte{byte(i + 10)})
		b := digest.Sum(nil)
		newTXIds[i] = types.ToTxID(b)
	}
	ps := make([]types.PeerID, 4)
	for i := 0; i < 4; i++ {
		ps[i] = types.RandomPeerID()
	}

	tests := []struct {
		name     string
		oldTx    []types.TxID
		peersOld [][]types.PeerID
		newTx    []types.TxID
		peersNew [][]types.PeerID

		wantSend  []types.TxID
		wantFront []types.TxID
	}{
		// 1. Nothing is old. no resend no delete
		{"TAllNew", nil, nil, newTXIds, nil, nil, newTXIds},
		// 2. Some old notices are in, but from single peer, so deleting it.
		{"TOldFromSingle", oldTXIds, nil, nil, nil, nil, nil},
		{"TOldFromSingle2", oldTXIds, nil, newTXIds, nil, nil, newTXIds},
		// 3. Some old notices are in, but from multiple peers, resend gettx to next peer. check if peer id was deleted.
		{"TOldFromMulti", oldTXIds, [][]types.PeerID{ps[:2], ps[2:3], nil, ps[2:], ps}, nil, nil, ct(oldTXIds[:2], oldTXIds[3:5]), ct(oldTXIds[:2], oldTXIds[3:5])},
		{"TOldFromMulti2", oldTXIds, [][]types.PeerID{ps[:2], ps[2:3], nil, ps[2:], ps}, newTXIds, nil, ct(oldTXIds[:2], oldTXIds[3:5]), ct(oldTXIds[:2], oldTXIds[3:5], newTXIds)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()
			mockPM.EXPECT().GetPeer(gomock.Any()).Return(mockPeer, true).AnyTimes()

			tm := newTxSyncManager(nil, mockActor, mockPM, logger)

			if tt.oldTx != nil {
				lt := time.Now().Add(-time.Second * 61)
				for i, hash := range tt.oldTx {
					tm.frontCache[hash] = &incomingTxNotice{hash: hash, created: lt, lastSent: lt}
					if i < len(tt.peersOld) {
						tm.frontCache[hash].peers = tt.peersOld[i]
					}
				}
			}
			if tt.newTx != nil {
				lt := time.Now()
				for i, hash := range tt.newTx {
					tm.frontCache[hash] = &incomingTxNotice{hash: hash, created: lt, lastSent: lt}
					if i < len(tt.peersNew) {
						tm.frontCache[hash].peers = tt.peersNew[i]
					}
				}
			}

			var sentMap = make(map[types.TxID]int32)
			if len(tt.wantSend) > 0 {
				mockMF := p2pmock.NewMockMoFactory(ctrl)
				mockMO := p2pmock.NewMockMsgOrder(ctrl)
				mockPeer.EXPECT().MF().Return(mockMF).MinTimes(1)
				mockPeer.EXPECT().SendMessage(mockMO).MinTimes(1)
				mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), p2pcommon.GetTXsRequest, gomock.Any()).DoAndReturn(func(respReceiver p2pcommon.ResponseReceiver, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
					tr, ok := message.(*types.GetTransactionsRequest)
					if !ok {
						t.Fatalf("unexpected message data type, want *types.GetTransactionsRequest")
					}
					for _, hash := range tr.Hashes {
						found := false
						for _, id := range tt.wantSend {
							if bytes.Equal(hash, id[:]) {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("req hash %v, is not in wanted hash %v", enc.ToString(hash), tt.wantSend)
						}
						sentMap[types.ToTxID(hash)] = 1
					}
					return mockMO
				}).MinTimes(1)
				mockMO.EXPECT().GetMsgID().Return(p2pcommon.NewMsgID()).MinTimes(1)
			}

			tm.refineFrontCache()

			if len(sentMap) != len(tt.wantSend) {
				t.Fatalf("sent tx %v, want %v", len(sentMap), len(tt.wantSend))
			}
			for _, id := range tt.wantSend {
				if _, exist := sentMap[id]; !exist {
					t.Fatalf("to send get %v, but is not expected hash", id)
				}
			}
			if len(tm.frontCache) != len(tt.wantFront) {
				t.Fatalf("front len %v, want %v", len(tm.frontCache), len(tt.wantFront))
			}
		})
	}
}

func Test_syncTxManager_pushBackToFrontCache(t *testing.T) {
	logger := log.NewLogger("tt.p2p")
	txIds := txIDSample[:5]

	tests := []struct {
		name     string
		arg      []types.TxID
		front    []types.TxID
		wantBack bool
	}{
		// 1. All are in front cache
		{"TAll", txIds, txIds, true},
		// 2. removed from front cache (maybe containd in block)
		{"TRemoved", txIds[3:], txIds[:3], false},
		{"TPartial", txIds[3:], txIds[:5], true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)

			//mockMF := p2pmock.NewMockMoFactory(ctrl)
			//mockMO := p2pmock.NewMockMsgOrder(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()
			mockPeer.EXPECT().Name().Return(sampleMeta.ID.String()).AnyTimes()

			tm := newTxSyncManager(nil, mockActor, mockPM, logger)

			if tt.front != nil {
				for _, hash := range tt.front {
					tm.frontCache[hash] = &incomingTxNotice{hash: hash}
				}
			}
			tm.pushBackToFrontCache(dummyPeerID, tt.arg)

			if (tm.toNoticeIdQueue.Len() > 0) != tt.wantBack {
				t.Fatalf("txs added %v, want %v", (tm.toNoticeIdQueue.Len() > 0), tt.wantBack)
			}
			if tt.wantBack {
				if !equalTXIDs(tm.toNoticeIdQueue.Front().Value.(*queryQueue).txIDs, tt.arg) {
					t.Fatalf("expected queued txs is %v, want %v", tm.toNoticeIdQueue.Front().Value.(*queryQueue).txIDs, tt.arg)
				}
			}

		})
	}
}

func Test_syncTxManager_RegisterTxNotice(t *testing.T) {
	logger := log.NewLogger("tt.p2p")

	sampleTXs := txSample[:10]
	sampleIDs := txIDSample[:10]
	peerIDs := make([]types.PeerID, 4)
	for i := 0; i < 4; i++ {
		peerIDs[i] = types.RandomPeerID()
	}

	tests := []struct {
		name    string
		frontTx []types.TxID

		argIn       []*types.Tx
		expectFront []types.TxID
		expectCache []types.TxID
	}{
		// 1. remove tx in front and add to txCache
		{"TRmFront", sampleIDs, sampleTXs[:5], sampleIDs[5:], sampleIDs[:5]},
		// 2. nothing to remove, but only add to txCache
		{"TRmFront", sampleIDs[5:], sampleTXs[:5], sampleIDs[5:], sampleIDs[:5]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()

			tm := newTxSyncManager(nil, mockActor, mockPM, logger)

			if tt.frontTx != nil {
				lt := time.Now()
				for _, hash := range tt.frontTx {
					tm.frontCache[hash] = &incomingTxNotice{hash: hash, created: lt, lastSent: lt}
				}
			}
			tm.Start()

			tm.registerTxNotice(tt.argIn)

			// make terminate
			tm.taskChannel <- func() {
				tm.Stop()
			}
			<-tm.finishChannel

			if len(tm.frontCache) != len(tt.expectFront) {
				t.Fatalf("to frontCache %v, want %v", len(tm.frontCache), len(tt.expectFront))
			}
			if len(tt.expectCache) != tm.txCache.Len() {
				t.Fatalf("txCache len %v, want %v", tm.txCache.Len(), len(tt.expectCache))
			}

		})
	}
}

func Test_syncTxManager_handleTxReq(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//var dummyPeerID, _ = types.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")
	var sampleMsgID = p2pcommon.NewMsgID()
	var sampleHeader = p2pmock.NewMockMessage(ctrl)
	sampleHeader.EXPECT().ID().Return(sampleMsgID).AnyTimes()
	sampleHeader.EXPECT().Subprotocol().Return(p2pcommon.GetTXsResponse).AnyTimes()

	var sampleTxsB58 = []string{
		"4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45",
		"6xfk39kuyDST7NwCu8tx3wqwFZ5dwKPDjxUS14tU7NZb8",
		"E8dbBGe9Hnuhk35cJoekPjL3VoL4xAxtnRuP47UoxzHd",
		"HB7Hg5GUbHuxwe8Lp5PcYUoAaQ7EZjRNG6RuvS6DnDRf",
		"BxKmDg9VbWHxrWnStEeTzJ2Ze7RF7YK4rpyjcsWSsnxs",
		"DwmGqFU4WgADpYN36FXKsYxMjeppvh9Najg4KxJ8gtX3",
	}
	var sampleTxs = make([][]byte, len(sampleTxsB58))
	var sampleTxHashes = make([]types.TxID, len(sampleTxsB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleTxs[i] = hash
		copy(sampleTxHashes[i][:], hash)
	}

	//dummyPeerID2, _ = types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	//dummyPeerID3, _ = types.IDB58Decode("16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD")

	logger := log.NewLogger("tt.subproto")
	//dummyMeta := p2pcommon.PeerMeta{ID: dummyPeerID, IPAddress: "192.168.1.2", Port: 4321}
	mockMo := p2pmock.NewMockMsgOrder(ctrl)
	mockMo.EXPECT().GetProtocolID().Return(p2pcommon.GetTXsResponse).AnyTimes()
	mockMo.EXPECT().GetMsgID().Return(sampleMsgID).AnyTimes()
	//mockSigner := p2pmock.NewmockMsgSigner(ctrl)
	//mockSigner.EXPECT().signMsg",gomock.Any()).Return(nil)
	tests := []struct {
		name   string
		setup  func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest)
		verify func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter)
	}{
		// 1. success case (single tx)
		{"TSucc1", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			// receive request for one tx , query to mempool get send response to remote peer
			dummyTxs := make([]*types.Tx, 1)
			dummyTxs[0] = &types.Tx{Hash: sampleTxs[0]}
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(&message.MemPoolExistExRsp{Txs: dummyTxs}, nil).Times(1)
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(gomock.AssignableToTypeOf(&message.MemPoolExistExRsp{}), nil).Return(dummyTxs, nil).Times(1)
			hashes := sampleTxs[:1]
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, p2pcommon.GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_OK, resp.Status)
				assert.Equal(tt, 1, len(resp.Hashes))
				assert.Equal(tt, sampleTxs[0], resp.Hashes[0])
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
			// verification is defined in setup
		}},
		// 1-1 success case2 (multiple tx)
		{"TSucc2", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			dummyTxs := make([]*types.Tx, len(sampleTxs))
			for i, txHash := range sampleTxs {
				dummyTxs[i] = &types.Tx{Hash: txHash}
			}
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(&message.MemPoolExistExRsp{Txs: dummyTxs}, nil).Times(1)
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(gomock.AssignableToTypeOf(&message.MemPoolExistExRsp{}), nil).Return(dummyTxs, nil).Times(1)
			hashes := sampleTxs
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, p2pcommon.GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_OK, resp.Status)
				assert.Equal(tt, len(sampleTxs), len(resp.Hashes))
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
		}},
		// 2. hash not found (partial)
		{"TPartialExist", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			dummyTxs := make([]*types.Tx, 0, len(sampleTxs))
			hashes := make([][]byte, 0, len(sampleTxs))
			for i, txHash := range sampleTxs {
				if i%2 == 0 {
					dummyTxs = append(dummyTxs, &types.Tx{Hash: txHash})
					hashes = append(hashes, txHash)
				}
			}
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(&message.MemPoolExistExRsp{Txs: dummyTxs}, nil).Times(1)
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(gomock.AssignableToTypeOf(&message.MemPoolExistExRsp{}), nil).Return(dummyTxs, nil).Times(1)
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, p2pcommon.GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_OK, resp.Status)
				assert.Equal(tt, len(dummyTxs), len(resp.Hashes))
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
		}},
		// 3. hash not found (all)
		{"TNoExist", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			//dummyTx := &types.Tx{Hash:nil}
			// emulate second tx is not exists.
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(&message.MemPoolExistExRsp{}, nil).Times(1)
			//msgHelper.EXPECT().ExtractTxsFromResponseAndError", mock.MatchedBy(func(m *message.MemPoolExistExRsp) bool {
			//	if len(m.Txs) == 0 {
			//		return false
			//	}
			//	return true
			//}), nil).Return(dummyTx, nil)
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(&MempoolRspTxCountMatcher{0}, nil).Return(nil, nil).Times(1)
			hashes := sampleTxs
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, p2pcommon.GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				assert.Equal(tt, types.ResultStatus_NOT_FOUND, resp.Status)
				assert.Equal(tt, 0, len(resp.Hashes))
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
		}},
		// 4. actor failure
		{"TActorError", func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) (p2pcommon.Message, *types.GetTransactionsRequest) {
			//dummyTx := &types.Tx{Hash:nil}
			actor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(nil, fmt.Errorf("timeout")).Times(1)
			//msgHelper.EXPECT().ExtractTxsFromResponseAndError", nil, gomock.AssignableToTypeOf("error")).Return(nil, fmt.Errorf("error"))
			msgHelper.EXPECT().ExtractTxsFromResponseAndError(nil, &WantErrMatcher{true}).Return(nil, fmt.Errorf("error")).Times(0)
			hashes := sampleTxs
			mockMF.EXPECT().NewMsgResponseOrder(sampleMsgID, p2pcommon.GetTXsResponse, gomock.AssignableToTypeOf(&types.GetTransactionsResponse{})).Do(func(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) {
				resp := message.(*types.GetTransactionsResponse)
				// TODO check if the changed behavior is fair or not.
				assert.Equal(tt, types.ResultStatus_NOT_FOUND, resp.Status)
			}).Return(mockMo).Times(1)
			return sampleHeader, &types.GetTransactionsRequest{Hashes: hashes}
		}, func(tt *testing.T, pm *p2pmock.MockPeerManager, actor *p2pmock.MockActorService, msgHelper *messagemock.Helper, mockMF *p2pmock.MockMoFactory, mockRW *p2pmock.MockMsgReadWriter) {
			// break at first eval
			// TODO need check that error response was sent
		}},

		// 5. invalid parameter (no input hash, or etc.)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMsgHelper := messagemock.NewHelper(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().Name().Return("mockPeer").AnyTimes()
			mockPeer.EXPECT().MF().Return(mockMF).AnyTimes()
			mockPeer.EXPECT().SendMessage(mockMo)

			_, body := tt.setup(t, mockPM, mockActor, mockMsgHelper, mockMF, mockRW)
			h := newTxSyncManager(nil, mockActor, mockPM, logger)
			h.msgHelper = mockMsgHelper

			//h.Handle(header, body)
			h.handleTxReq(mockPeer, sampleMsgID, body.Hashes)
			// wait for handle finished

			tt.verify(t, mockPM, mockActor, mockMsgHelper, mockMF, mockRW)
		})
	}
}

func Test_syncTxManager_handleBySize(t *testing.T) {
	logger := log.NewLogger("test.subproto")
	var dummyPeerID, _ = types.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")

	bigTxBytes := make([]byte, 2*1024*1024)
	//dummyMO := p2pmock.NewMockMsgOrder(ctrl)
	tests := []struct {
		name              string
		hashCnt           int
		validCallCount    int
		expectedSendCount int
	}{
		{"TSingle", 1, 1, 1},
		{"TNotFounds", 100, 0, 1},
		{"TFound10", 10, 10, 3},
		{"TFoundAll", 20, 100, 5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockMF := &testDoubleMOFactory{}
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer.EXPECT().MF().Return(mockMF).AnyTimes()
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(test.expectedSendCount)

			validBigMempoolRsp := &message.MemPoolExistExRsp{}
			inHashes := make([][]byte, 0, test.hashCnt)
			txs := make([]*types.Tx, 0, test.hashCnt)
			for i := 0; i < test.hashCnt; i++ {
				tx := &types.Tx{Body: &types.TxBody{Nonce: uint64(i + 1), Payload: bigTxBytes}}
				tx.Hash = tx.CalculateTxHash()
				inHashes = append(inHashes, tx.Hash)
				if i < test.validCallCount {
					txs = append(txs, tx)
				} else {
				}
				txs = append(txs, (*types.Tx)(nil))
			}
			validBigMempoolRsp.Txs = txs

			mockActor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(validBigMempoolRsp, nil)

			tm := newTxSyncManager(nil, mockActor, mockPM, logger)
			dummyMsg := &testMessage{subProtocol: p2pcommon.GetTXsRequest, id: p2pcommon.NewMsgID()}
			msgBody := &types.GetTransactionsRequest{Hashes: inHashes}
			//h.Handle(dummyMsg, msgBody)
			tm.handleTxReq(mockPeer, dummyMsg.id, msgBody.Hashes)
			// wait for handle finished
		})
	}
}

func Test_syncTxManager_assignTxToPeer(t *testing.T) {
	logger := log.NewLogger("test.subproto")
	dummy := types.TxID{}
	p0 := types.RandomPeerID()
	p1 := types.RandomPeerID()
	p2 := types.RandomPeerID()
	p3 := types.RandomPeerID()
	p4 := types.RandomPeerID()
	pids := []types.PeerID{p0, p1, p2, p3, p4}

	// check if assign is expected successful or failed, and if success, check peerID
	tests := []struct {
		name string

		arg       []types.PeerID
		want      bool
		wantPeers []types.PeerID
	}{
		// 0. add case.
		{"TSingle", ToP(p0), true, nil},
		// 1. first peer
		{"TFirst", ToP(p0, p1, p2, p3, p4), true, ToP(p1, p2, p3, p4)},
		// 2. mid peer
		{"TMid", ToP(p4, p3, p0, p1, p2), true, ToP(p4, p3, p1, p2)},
		// 3. lastPeer
		{"TLast", ToP(p4, p3, p0), true, ToP(p4, p3)},
		// 4. all full
		{"TFull", ToP(p3, p4), false, ToP(p3, p4)},
		// nowhere peer
		// TODO change method impl and add case.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)

			argSendMap := make(map[types.PeerID][]types.TxID)
			// pid 0,1,2, is empty, 3,4 is full
			for i, pid := range pids {
				var arr []types.TxID
				if i < 3 {
					arr = make([]types.TxID, 0)
				} else {
					arr = make([]types.TxID, DefaultPeerTxQueueSize)
					for j := 0; j < DefaultPeerTxQueueSize; j++ {
						arr[j] = dummy
					}
				}
				argSendMap[pid] = arr
			}

			in := incomingTxNotice{peers: tt.arg}
			tm := newTxSyncManager(nil, mockActor, mockPM, logger)

			if got := tm.assignTxToPeer(&in, argSendMap); got != tt.want {
				t.Errorf("assignTxToPeer() = %v, want %v", got, tt.want)
			}
			if len(in.peers) != len(tt.wantPeers) {
				t.Errorf("assignTxToPeer() peers = %v, want %v", in.peers, tt.wantPeers)
			} else {
				for i, pid := range in.peers {
					o := tt.wantPeers[i]
					if !types.IsSamePeerID(pid, o) {
						t.Errorf("assignTxToPeer() peers = %v, want %v", pid, o)
						break
					}
				}
			}
		})
	}
}

func ToP(in ...types.PeerID) []types.PeerID {
	sli := make([]types.PeerID, len(in))
	copy(sli, in)
	return sli
}

type MempoolRspTxCountMatcher struct {
	matchCnt int
}

func (tcm MempoolRspTxCountMatcher) Matches(x interface{}) bool {
	m, ok := x.(*message.MemPoolExistExRsp)
	if !ok {
		return false
	}
	return tcm.matchCnt == len(m.Txs)
}

func (tcm MempoolRspTxCountMatcher) String() string {
	return fmt.Sprintf("tx count = %d", tcm.matchCnt)
}

type TxIDCntMatcher struct {
	matchCnt int
}

func (scm TxIDCntMatcher) Matches(x interface{}) bool {
	m, ok := x.([]types.TxID)
	if !ok {
		return false
	}
	return scm.matchCnt == len(m)
}

func (scm TxIDCntMatcher) String() string {
	return fmt.Sprintf("len(slice) = %d", scm.matchCnt)
}

type WantErrMatcher struct {
	wantErr bool
}

func (tcm WantErrMatcher) Matches(x interface{}) bool {
	m, ok := x.(*error)
	if !ok {
		return false
	}
	return tcm.wantErr == (m != nil)
}

func (tcm WantErrMatcher) String() string {
	return fmt.Sprintf("want error = %v", tcm.wantErr)
}

func BenchmarkRefine(b *testing.B) {
	b.Skipf("internal structure and logic is heavily changed. so it need to change paramters")
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockPM := p2pmock.NewMockPeerManager(ctrl)
	mockPeer := p2pmock.NewMockRemotePeer(ctrl)
	mockActor := p2pmock.NewMockActorService(ctrl)

	mockPM.EXPECT().GetPeer(gomock.Any()).Return(mockPeer, true).AnyTimes()
	mockPeer.EXPECT().MF().Return(&testDoubleMOFactory{}).AnyTimes()
	mockPeer.EXPECT().SendMessage(gomock.Any()).AnyTimes()

	pids := make([]types.PeerID, 100)
	infos := make(map[types.PeerID][]*incomingTxNotice)
	for i := 0; i < 100; i++ {
		pids[i] = types.RandomPeerID()
		list := make([]*incomingTxNotice, 2000)
		infos[pids[i]] = list
		dummyTX := &types.Tx{Body: &types.TxBody{Account: []byte(pids[i])}}
		for j := 0; j < 2000; j++ {
			dummyTX.Body.Nonce = uint64(j + 1)
			hash := dummyTX.CalculateTxHash()
			info := &incomingTxNotice{hash: types.ToTxID(hash), lastSent: unsent, peers: []types.PeerID{pids[i]}}
			list[j] = info
		}
	}

	benchmarks := []struct {
		name    string
		old     bool
		inCache map[types.TxID]*incomingTxNotice
	}{
		// 1. 10 peers, 10 in cache for each peer
		{"BP10F10", true, combine(pids[:10], infos, 10)},
		// 2. 10 peers, 200 in cache
		{"BP10F200", true, combine(pids[:10], infos, 200)},
		// 3. 10 peers, 2000 in cache
		{"BP10F2000", true, combine(pids[:10], infos, 2000)},
		// 5. 100 peers, 2000 in cache
		{"BP100F10", true, combine(pids, infos, 10)},
		// 6. 100 peers, 20000 in cache
		{"BP100F200", true, combine(pids, infos, 200)},
		// 7. 100 peers, 200000 in cache
		{"BP100F2000", true, combine(pids, infos, 2000)},
		// 8. 100 peers, heavy single (5000 in a single)
		{"BPHeavy1", true, heavy(5000, pids, 5000)},
		// 9. 100 peers, heavy single (5000, 5000, 100 overlapped peers)
		{"BPHeavy2", true, heavy(5000, pids, 5000, 5000, 100)},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				h := newTxSyncManager(nil, mockActor, mockPM, logger)
				h.frontCache = bm.inCache
				h.refineFrontCache()
			}
		})
	}
}

func combine(pids []types.PeerID, dummies map[types.PeerID][]*incomingTxNotice, size int) map[types.TxID]*incomingTxNotice {
	m := make(map[types.TxID]*incomingTxNotice)
	for _, pid := range pids {
		infos := dummies[pid]
		for i := 0; i < size; i++ {
			info := infos[i]
			m[info.hash] = info
		}
	}
	return m
}

// heavy make tx notices for inserting front cache. the length of sizes and pids must equals
func heavy(txSize int, pids []types.PeerID, sizes ...int) map[types.TxID]*incomingTxNotice {
	m := make(map[types.TxID]*incomingTxNotice)
	dummyTX := &types.Tx{Body: &types.TxBody{Account: []byte(pids[0])}}
	tempList := make([]*incomingTxNotice, txSize)
	for j := 0; j < txSize; j++ {
		dummyTX.Body.Nonce = uint64(j + 1)
		hash := dummyTX.CalculateTxHash()
		info := &incomingTxNotice{hash: types.ToTxID(hash), lastSent: unsent, peers: []types.PeerID{}}
		tempList[j] = info
		m[info.hash] = info
	}
	for i, size := range sizes {
		for j := 0; j < size; j++ {
			tempList[j].peers = append(tempList[j].peers, pids[i])
		}
	}
	return m
}
