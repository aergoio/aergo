/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"crypto/sha256"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestSyncManager_HandleBlockProducedNotice(t *testing.T) {
	// only interested in max block size
	chain.Init(1024*1024,"",false,0,0)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.p2p")
	sampleBlock := &types.Block{Hash: dummyBlockHash}
	txs := make([]*types.Tx,1)
	txs[0] = &types.Tx{Hash:make([]byte,1024*1024*2)}
	sampleBigBlock := &types.Block{Hash:dummyBlockHash,Body:&types.BlockBody{Txs:txs}}
	var blkHash = types.ToBlockID(dummyBlockHash)
	// test if new block notice comes
	tests := []struct {
		name string
		put  *types.BlockID
		addedBlock *types.Block

		wantActorCall bool
	}{
		// 1. Succ : valid block hash and not exist in local
		{"TSucc", nil, sampleBlock,true},
		// 2. Rare case - valid block hash but already exist in local cache
		{"TExist", &blkHash, sampleBlock, false},
		{"TTooBigBlock", nil, sampleBigBlock,false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()
			actorCallCnt := 0
			if test.wantActorCall {
				actorCallCnt = 1
			}
			mockActor.EXPECT().SendRequest(message.ChainSvc, gomock.Any()).Times(actorCallCnt)

			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)
			if test.put != nil {
				target.blkCache.Add(*test.put, true)
			}
			target.HandleBlockProducedNotice(mockPeer, test.addedBlock)
		})
	}
}

func TestSyncManager_HandleNewBlockNotice(t *testing.T) {
	// only interested in max block size
	chain.Init(1024*1024,"",false,0,0)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.p2p")
	sampleBlock := &types.Block{Hash: dummyBlockHash}
	var blkHash types.BlockID
	// test if new block notice comes
	tests := []struct {
		name    string
		put     *types.BlockID
		syncing bool
		setup   func(tt *testing.T, actor *p2pmock.MockActorService, ca *p2pmock.MockChainAccessor, peer *p2pmock.MockRemotePeer) (types.BlockID, *types.NewBlockNotice)
		//verify  func(tt *testing.T, actor *p2pmock.MockActorService, ca *p2pmock.MockChainAccessor)
	}{
		// 1. Succ : valid block hash and not exist in local
		{"TSucc", nil, false,
			func(tt *testing.T, actor *p2pmock.MockActorService, ca *p2pmock.MockChainAccessor, peer *p2pmock.MockRemotePeer) (types.BlockID, *types.NewBlockNotice) {
				ca.EXPECT().GetBlock(gomock.Any()).Return(nil, nil)
				actor.EXPECT().GetChainAccessor().Return(ca)
				copy(blkHash[:], dummyBlockHash)
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any())
				peer.EXPECT().Name().Return("16..aadecf@1")
				return blkHash, &types.NewBlockNotice{BlockHash: dummyBlockHash}
			}},
		// 1-1. Succ : valid block hash and exist in chainsvc, but not in cache
		{"TSuccExistChain", nil, false,
			func(tt *testing.T, actor *p2pmock.MockActorService, ca *p2pmock.MockChainAccessor, peer *p2pmock.MockRemotePeer) (types.BlockID, *types.NewBlockNotice) {
				ca.EXPECT().GetBlock(gomock.Any()).Return(sampleBlock, nil)
				copy(blkHash[:], dummyBlockHash)
				actor.EXPECT().GetChainAccessor().Return(ca)
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).MaxTimes(0)
				return blkHash, &types.NewBlockNotice{BlockHash: dummyBlockHash}
			}},
		// 2. SuccCachehit : valid block hash but already exist in local cache
		{"TSuccExistCache", &blkHash, false,
			func(tt *testing.T, actor *p2pmock.MockActorService, ca *p2pmock.MockChainAccessor, peer *p2pmock.MockRemotePeer) (types.BlockID, *types.NewBlockNotice) {
				ca.EXPECT().GetBlock(gomock.Any()).Return(sampleBlock, nil).MaxTimes(0)
				copy(blkHash[:], dummyBlockHash)

				//ca.AssertNotCalled(tt, "GetBlock", mock.AnythingOfType("[]uint8"))
				//actor.EXPECT().AssertNotCalled(tt, "SendRequest", message.P2PSvc, mock.Anything)
				actor.EXPECT().SendRequest(message.P2PSvc, mock.Anything).MaxTimes(0)
				return blkHash, &types.NewBlockNotice{BlockHash: dummyBlockHash}
			}},
		// 2. Busy : other sync worker is working
		{"TBusy", &blkHash, true,
			func(tt *testing.T, actor *p2pmock.MockActorService, ca *p2pmock.MockChainAccessor, peer *p2pmock.MockRemotePeer) (types.BlockID, *types.NewBlockNotice) {
				copy(blkHash[:], dummyBlockHash)
				actor.EXPECT().GetChainAccessor().MaxTimes(0)
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).MaxTimes(0)
				return blkHash, &types.NewBlockNotice{BlockHash: dummyBlockHash}
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID)

			_, data := test.setup(t, mockActor, mockCA, mockPeer)
			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)
			if test.put != nil {
				target.blkCache.Add(*test.put, true)
			}
			target.HandleNewBlockNotice(mockPeer, data)
			//test.verify(t, mockActor, mockCA)
		})
	}
}

func TestSyncManager_HandleNewTxNotice(t *testing.T) {
	logger := log.NewLogger("tt.p2p")
	rawHashes := sampleTxs
	txHashes := sampleTxIDs

	// tt if new block notice comes
	tests := []struct {
		name     string
		front    []types.TxID
		inCache  []types.TxID
		setup    func(tt *testing.T, actor *p2pmock.MockActorService)
		expectedFront []types.TxID
		expected []types.TxID

	}{
		// 0. Succ : valid tx hashes and not exist in local cache
		{"TSuccAllNew", nil, nil,
			func(tt *testing.T, actor *p2pmock.MockActorService) {
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).DoAndReturn(func(name string, arg *message.GetTransactions) {
					expected := txHashes
					if len(arg.Hashes) != len(expected) {
						tt.Errorf("to send get %v, want %v", len(arg.Hashes), len(expected))
					}
					for i, hash := range arg.Hashes {
						if !bytes.Equal(hash, expected[i][:]) {
							tt.Errorf("to send get %v, want %v", enc.ToString(hash), expected[i])
						}
					}
				})
			}, txHashes, nil},
		// 1. : some txs are in front cache. and only txs not in front cache are sent getTx
		{"TInFront", txHashes[:2], nil,
			func(tt *testing.T, actor *p2pmock.MockActorService) {
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).DoAndReturn(func(name string, arg *message.GetTransactions) {
					expected := txHashes[2:]
					if len(arg.Hashes) != len(expected) {
						tt.Errorf("to send get %v, want %v", len(arg.Hashes), len(expected))
					}
					for i, hash := range arg.Hashes {
						if !bytes.Equal(hash, expected[i][:]) {
							tt.Errorf("to send get %v, want %v", enc.ToString(hash), expected[i])
						}
					}
				})
			}, txHashes, nil},
		// 2. Succ : valid tx hashes and partially exist in local cache
		{"TBoth", txHashes[:2], txHashes[2:4],
			func(tt *testing.T, actor *p2pmock.MockActorService) {
				// only hashes not in cache are sent to method, which is first 2 hashes
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).DoAndReturn(func(name string, arg *message.GetTransactions) {
					expected := txHashes[4:]
					if len(arg.Hashes) != len(expected) {
						tt.Errorf("to send get %v, want %v", len(arg.Hashes), len(expected))
					}
					for i, hash := range arg.Hashes {
						if !bytes.Equal(hash, expected[i][:]) {
							tt.Errorf("to send get %v, want %v", enc.ToString(hash), expected[i])
						}
					}
				})
			}, concatSlice(txHashes[:2],txHashes[4:]), txHashes[2:4]},
		// 3. Succ : valid tx hashes and all exist in local cache
		{"TAllFront", txHashes, nil,
			func(tt *testing.T, actor *p2pmock.MockActorService) {
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).MaxTimes(0)
			}, txHashes, nil},
		// 4. Succ : valid tx hashes and partially exist in local cache
		{"TCachePart", nil, txHashes[2:],
			func(tt *testing.T, actor *p2pmock.MockActorService) {
				// only hashes not in cache are sent to method, which is first 2 hashes
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).DoAndReturn(func(name string, arg *message.GetTransactions) {
					expected := txHashes[:2]
					if len(arg.Hashes) != len(expected) {
						tt.Errorf("to send get %v, want %v", len(arg.Hashes), len(expected))
					}
					for i, hash := range arg.Hashes {
						if !bytes.Equal(hash, expected[i][:]) {
							tt.Errorf("to send get %v, want %v", enc.ToString(hash), expected[i])
						}
					}
				})

			}, txHashes[:2], txHashes[2:]},
		// 5. Succ : valid tx hashes and all exist in local cache
		{"TSuccExistAll", nil, txHashes,
			func(tt *testing.T, actor *p2pmock.MockActorService) {
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).MaxTimes(0)
			}, nil, txHashes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID)

			data := &types.NewTransactionsNotice{TxHashes: rawHashes}

			tt.setup(t, mockActor)
			sm := newSyncManager(mockActor, mockPM, logger).(*syncManager)

			if tt.front != nil {
				for _, hash := range tt.front {
					sm.frontCache[hash] = &incomingTxNotice{hash: hash}
				}
			}
			if tt.inCache != nil {
				for _, hash := range tt.inCache {
					sm.txCache.Add(hash, true)
				}
			}
			sm.Start()

			sm.HandleNewTxNotice(mockPeer, txHashes, data)

			// make terminate
			sm.taskChannel <- func() {
				sm.Stop()
			}
			<- sm.finishChannel


			if len(sm.frontCache) != len(tt.expectedFront) {
				t.Fatalf("front len %v, want %v", len(sm.frontCache), len(tt.expectedFront))
			}

			if len(tt.expected) != sm.txCache.Len() {
				t.Fatalf("txCache len %v, want %v", sm.txCache.Len(), len(tt.expected))
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

func TestSyncManager_HandleGetBlockResponse(t *testing.T) {
	// only interested in max block size
	chain.Init(1024*1024,"",false,0,0)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	totalBlkCnt := len(sampleTxs)
	sampleBlocks := make([]*types.Block, totalBlkCnt)
	for i, hash := range sampleTxs {
		sampleBlocks[i] = &types.Block{Hash: hash}
	}
	tests := []struct {
		name       string
		respBlocks []*types.Block

		// call count directly to chainservice
		chainCallCnt int
	}{
		// 1. message triggered by NewBlockNotice (maybe)
		{"TSingleBlock", sampleBlocks[:1], 1},
		// 2. message triggered by newsyncer but not handled by it (caused by sync fail or timeout)
		{"TZeroBlock", sampleBlocks[:0], 0},
		{"TMultiBlocks", sampleBlocks, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID)

			mockActor.EXPECT().SendRequest(gomock.Any(), gomock.Any()).Times(test.chainCallCnt)
			dummyMsgID := p2pcommon.NewMsgID()
			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)

			msg := p2pcommon.NewSimpleRespMsgVal(p2pcommon.PingResponse, p2pcommon.NewMsgID(), dummyMsgID)
			resp := &types.GetBlockResponse{Blocks: test.respBlocks}
			target.HandleGetBlockResponse(mockPeer, msg, resp)

			//mockActor.AssertNumberOfCalls(t, "SendRequest", test.chainCallCnt)
		})
	}
}

func Test_syncManager_refineFrontCache(t *testing.T) {
	logger := log.NewLogger("tt.p2p")
	digest := sha256.New()
	oldTXIds := make([]types.TxID,5)
	for i:=0; i<5; i++ {
		digest.Write([]byte{byte(i)})
		b := digest.Sum(nil)
		oldTXIds[i] = types.ToTxID(b)
	}
	newTXIds := make([]types.TxID,5)
	for i:=0; i<5; i++ {
		digest.Write([]byte{byte(i+10)})
		b := digest.Sum(nil)
		newTXIds[i] = types.ToTxID(b)
	}
	ps := make([]types.PeerID,4)
	for i:=0; i<4; i++ {
		ps[i] = types.RandomPeerID()
	}

	tests := []struct {
		name   string
		oldTx  []types.TxID
		peersOld  [][]types.PeerID
		newTx  []types.TxID
		peersNew  [][]types.PeerID

		expectSend []types.TxID
		expectFront []types.TxID
	}{
		// 1. Nothing is old. no resend no delete
		{"TAllNew", nil, nil, newTXIds, nil, nil, newTXIds },
		// 2. Some old notices are in, but from single peer, so deleting it.
		{"TOldFromSingle", oldTXIds, nil, nil, nil, nil, nil },
		{"TOldFromSingle2", oldTXIds, nil, newTXIds, nil, nil, newTXIds },
		// 3. Some old notices are in, but from multiple peers, resend gettx to next peer. check if peer id was deleted.
		{"TOldFromMulti", oldTXIds, [][]types.PeerID{ps[:2],ps[2:3],nil,ps[2:],ps}, nil, nil, concatSlice(oldTXIds[:2], oldTXIds[3:5]), concatSlice(oldTXIds[:2], oldTXIds[3:5]) },
		{"TOldFromMulti", oldTXIds, [][]types.PeerID{ps[:2],ps[2:3],nil,ps[2:],ps}, newTXIds, nil, concatSlice(oldTXIds[:2], oldTXIds[3:5]), concatSlice(oldTXIds[:2], oldTXIds[3:5],newTXIds) },

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()

			sm := newSyncManager(mockActor, mockPM, logger).(*syncManager)

			if tt.oldTx != nil {
				lt := time.Now().Add(-time.Second * 61)
				for i, hash := range tt.oldTx {
					sm.frontCache[hash] = &incomingTxNotice{hash: hash, created:lt, lastSent:lt}
					if i < len(tt.peersOld) {
						sm.frontCache[hash].peers = tt.peersOld[i]
					}
				}
			}
			if tt.newTx != nil {
				lt := time.Now()
				for i, hash := range tt.newTx {
					sm.frontCache[hash] = &incomingTxNotice{hash: hash, created:lt, lastSent:lt}
					if i < len(tt.peersNew) {
						sm.frontCache[hash].peers = tt.peersNew[i]
					}
				}
			}

			var sentMap = make(map[types.TxID]int32)
			if len(tt.expectSend) > 0 {
				mockActor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).DoAndReturn(func(name string, arg *message.GetTransactions) {
					for _, hash := range arg.Hashes {
						id := types.ToTxID(hash)
						sentMap[id] = sentMap[id] + 1
					}
				}).MinTimes(1)
			}

			sm.refineFrontCache()

			if len(sentMap) != len(tt.expectSend) {
				t.Fatalf("sent tx %v, want %v", len(sentMap), len(tt.expectSend))
			}
			for _, id := range tt.expectSend {
				if _, exist := sentMap[id]; !exist {
					t.Fatalf("to send get %v, but is not expected hash", id)
				}
			}
			if len(sm.frontCache) != len(tt.expectFront) {
				t.Fatalf("front len %v, want %v", len(sm.frontCache), len(tt.expectFront))
			}
		})
	}
}

func Test_syncManager_RegisterTxNotice(t *testing.T) {
	logger := log.NewLogger("tt.p2p")
	digest := sha256.New()
	sampleIDs := make([]types.TxID,10)
	for i:=0; i<10; i++ {
		digest.Write([]byte{byte(i)})
		b := digest.Sum(nil)
		sampleIDs[i] = types.ToTxID(b)
	}
	peerIDs := make([]types.PeerID,4)
	for i:=0; i<4; i++ {
		peerIDs[i] = types.RandomPeerID()
	}

	tests := []struct {
		name   string
		frontTx  []types.TxID

		argIn []types.TxID
		expectFront []types.TxID
		expectCache []types.TxID
	}{
		// 1. remove tx in front and add to txCache
		{"TRmFront", sampleIDs, sampleIDs[:5], sampleIDs[5:], sampleIDs[:5] },
		// 2. nothing to remove, but only add to txCache
		{"TRmFront", sampleIDs[5:], sampleIDs[:5], sampleIDs[5:], sampleIDs[:5] },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID).AnyTimes()

			sm := newSyncManager(mockActor, mockPM, logger).(*syncManager)

			if tt.frontTx != nil {
				lt := time.Now()
				for _, hash := range tt.frontTx {
					sm.frontCache[hash] = &incomingTxNotice{hash: hash, created:lt, lastSent:lt}
				}
			}
			sm.Start()

			sm.RegisterTxNotice(tt.argIn)

			// make terminate
			sm.taskChannel <- func() {
				sm.Stop()
			}
			<- sm.finishChannel

			if len(sm.frontCache) != len(tt.expectFront) {
				t.Fatalf("to frontCache %v, want %v", len(sm.frontCache), len(tt.expectFront))
			}
			if len(tt.expectCache) != sm.txCache.Len() {
				t.Fatalf("txCache len %v, want %v", sm.txCache.Len(), len(tt.expectCache))
			}

		})
	}
}