package p2p

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/message/messagemock"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_txSyncManager_HandleNewTxNotice(t *testing.T) {
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
	}{
		// 0. Succ : valid tx hashes and not exist in local cache
		{"TSuccAllNew", nil, nil,  txHashes, nil},
		// 1. : some txs are in front cache. and only txs not in front cache are sent getTx
		{"TInFront", txHashes[:2], nil,  txHashes, nil},
		// 2. Succ : valid tx hashes and partially exist in local cache
		{"TBoth", txHashes[:2], txHashes[2:4], concatSlice(txHashes[:2], txHashes[4:]), txHashes[2:4]},
		// 3. Succ : valid tx hashes and all exist in local cache
		{"TAllFront", txHashes, nil,
			txHashes, nil},
		// 4. Succ : valid tx hashes and partially exist in local cache
		{"TCachePart", nil, txHashes[2:],  txHashes[:2], txHashes[2:]},
		// 5. Succ : valid tx hashes and all exist in local cache
		{"TSuccExistAll", nil, txHashes, nil, txHashes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()


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

			tm.HandleNewTxNotice(mockPeer, txHashes, data)

			// make terminate
			tm.taskChannel <- func() {
				tm.Stop()
			}
			<-tm.finishChannel

			if len(tm.frontCache) != len(tt.wantFront) {
				t.Fatalf("front len %v, want %v", len(tm.frontCache), len(tt.wantFront))
			}

			if len(tt.wantCache) != tm.txCache.Len() {
				t.Fatalf("txCache len %v, want %v", tm.txCache.Len(), len(tt.wantCache))
			}

		})
	}
}

func concatSlice(slis ...[]types.TxID) []types.TxID {
	ret := make([]types.TxID, 0, len(slis))
	for _, s := range slis {
		ret = append(ret, s...)
	}
	return ret
}

func Test_txSyncManager_refineFrontCache(t *testing.T) {
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
		{"TOldFromMulti", oldTXIds, [][]types.PeerID{ps[:2], ps[2:3], nil, ps[2:], ps}, nil, nil, concatSlice(oldTXIds[:2], oldTXIds[3:5]), concatSlice(oldTXIds[:2], oldTXIds[3:5])},
		{"TOldFromMulti2", oldTXIds, [][]types.PeerID{ps[:2], ps[2:3], nil, ps[2:], ps}, newTXIds, nil, concatSlice(oldTXIds[:2], oldTXIds[3:5]), concatSlice(oldTXIds[:2], oldTXIds[3:5], newTXIds)},
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

func Test_txSyncManager_RegisterTxNotice(t *testing.T) {
	logger := log.NewLogger("tt.p2p")
	digest := sha256.New()
	sampleIDs := make([]types.TxID, 10)
	for i := 0; i < 10; i++ {
		digest.Write([]byte{byte(i)})
		b := digest.Sum(nil)
		sampleIDs[i] = types.ToTxID(b)
	}
	peerIDs := make([]types.PeerID, 4)
	for i := 0; i < 4; i++ {
		peerIDs[i] = types.RandomPeerID()
	}

	tests := []struct {
		name    string
		frontTx []types.TxID

		argIn       []types.TxID
		expectFront []types.TxID
		expectCache []types.TxID
	}{
		// 1. remove tx in front and add to txCache
		{"TRmFront", sampleIDs, sampleIDs[:5], sampleIDs[5:], sampleIDs[:5]},
		// 2. nothing to remove, but only add to txCache
		{"TRmFront", sampleIDs[5:], sampleIDs[:5], sampleIDs[5:], sampleIDs[:5]},
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

func TestTxSyncManager_handleTxReq(t *testing.T) {
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

func TestTxRequestHandler_handleBySize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.subproto")
	var dummyPeerID, _ = types.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")
	var dummyTxHash, _ = enc.ToBytes("4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45")

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
		{"TFound10", 10, 10, 4},
		{"TFoundAll", 20, 100, 7},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockMF := &testDoubleMOFactory{}
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer.EXPECT().MF().Return(mockMF).AnyTimes()
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(test.expectedSendCount)

			validBigMempoolRsp := &message.MemPoolExistExRsp{}
			txs := make([]*types.Tx, 0, test.hashCnt)
			for i := 0; i < test.hashCnt; i++ {
				if i >= test.validCallCount {
					break
				}
				txs = append(txs, &types.Tx{Hash: dummyTxHash, Body: &types.TxBody{Payload: bigTxBytes}})
			}
			validBigMempoolRsp.Txs = txs

			mockActor.EXPECT().CallRequestDefaultTimeout(message.MemPoolSvc, gomock.AssignableToTypeOf(&message.MemPoolExistEx{})).Return(validBigMempoolRsp, nil)

			h := newTxSyncManager(nil, mockActor, mockPM, logger)
			dummyMsg := &testMessage{subProtocol: p2pcommon.GetTXsRequest, id: p2pcommon.NewMsgID()}
			msgBody := &types.GetTransactionsRequest{Hashes: make([][]byte, test.hashCnt)}
			//h.Handle(dummyMsg, msgBody)
			h.handleTxReq(mockPeer, dummyMsg.id, msgBody.Hashes)
			// wait for handle finished
		})
	}
}

func Test_syncTxManager_assignTxToPeer(t *testing.T) {
	logger := log.NewLogger("test.subproto")
	dummy := types.TxID{}
	p0 :=  types.RandomPeerID()
	p1 :=  types.RandomPeerID()
	p2 :=  types.RandomPeerID()
	p3 :=  types.RandomPeerID()
	p4 :=  types.RandomPeerID()
	pids := []types.PeerID{p0,p1,p2,p3,p4}

	// check if assign is expected successful or failed, and if success, check peerID
	tests := []struct {
		name string

		arg []types.PeerID
		want bool
		wantPeers []types.PeerID
	}{
		// 0. add case.
		{"TSingle", ToP(p0), true, nil },
		// 1. first peer
		{"TFirst", ToP(p0,p1,p2,p3,p4), true, ToP(p1,p2,p3,p4) },
		// 2. mid peer
		{"TMid", ToP(p4,p3,p0,p1,p2), true, ToP(p4,p3,p1,p2) },
		// 3. lastPeer
		{"TLast", ToP(p4,p3,p0), true, ToP(p4,p3) },
		// 4. all full
		{"TFull", ToP(p3,p4), false, ToP(p3,p4) },
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

			in := incomingTxNotice{peers:tt.arg}
			tm := newTxSyncManager(nil, mockActor, mockPM, logger)

			if got := tm.assignTxToPeer(&in, argSendMap); got != tt.want {
				t.Errorf("assignTxToPeer() = %v, want %v",got, tt.want)
			}
			if len(in.peers) != len(tt.wantPeers) {
				t.Errorf("assignTxToPeer() peers = %v, want %v",in.peers, tt.wantPeers)
			} else {
				for i,pid := range in.peers {
					o := tt.wantPeers[i]
					if !types.IsSamePeerID(pid, o) {
						t.Errorf("assignTxToPeer() peers = %v, want %v",pid, o)
						break
					}
				}
			}
		})
	}
}

func ToP(in ...types.PeerID) []types.PeerID {
	sli := make([]types.PeerID,len(in))
	copy(sli,in)
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
