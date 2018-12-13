/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)


func TestSyncManager_HandleNewBlockNotice(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	sampleBlock := &types.Block{Hash:dummyBlockHash}
	var blkHash BlkHash
	// test if new block notice comes
	tests := []struct {
		name string
		put *BlkHash
		syncing bool
		setup func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor) (BlkHash,*types.NewBlockNotice)
		verify func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor)
	}{
		// 1. Succ : valid block hash and not exist in local
		{"TSucc", nil, false,
		func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor) (BlkHash,*types.NewBlockNotice) {
			ca.On("GetBlock", mock.AnythingOfType("[]uint8")).Return(nil, nil)
			copy(blkHash[:], dummyBlockHash)
			return blkHash, &types.NewBlockNotice{BlockHash:dummyBlockHash}
		},
		func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor) {
			ca.AssertCalled(tt,"GetBlock", mock.AnythingOfType("[]uint8"))
			actor.AssertCalled(tt,"SendRequest",message.P2PSvc, mock.AnythingOfType("*message.GetBlockInfos"))
		}},
		// 1-1. Succ : valid block hash and exist in chainsvc, but not in cache
		{"TSuccExistChain", nil,false,
			func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor) (BlkHash,*types.NewBlockNotice) {
				ca.On("GetBlock", mock.AnythingOfType("[]uint8")).Return(sampleBlock, nil)
				copy(blkHash[:], dummyBlockHash)
				return blkHash, &types.NewBlockNotice{BlockHash:dummyBlockHash}
			},
			func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor) {
				ca.AssertCalled(tt,"GetBlock", mock.AnythingOfType("[]uint8"))
				actor.AssertNotCalled(tt,"SendRequest",message.P2PSvc, mock.AnythingOfType("*message.GetBlockInfos"))
			}},
		// 2. SuccCachehit : valid block hash but already exist in local cache
		{"TSuccExistCache", &blkHash,false,
			func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor) (BlkHash,*types.NewBlockNotice) {
				ca.On("GetBlock", mock.AnythingOfType("[]uint8")).Return(sampleBlock, nil)
				copy(blkHash[:], dummyBlockHash)
				return blkHash, &types.NewBlockNotice{BlockHash:dummyBlockHash}
			},
			func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor) {
				ca.AssertNotCalled(tt,"GetBlock", mock.AnythingOfType("[]uint8"))
				actor.AssertNotCalled(tt,"SendRequest",message.P2PSvc, mock.Anything)
			}},
		// 2. Busy : other sync worker is working
		{"TBusy", &blkHash,true,
			func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor) (BlkHash,*types.NewBlockNotice) {
				ca.On("GetBlock", mock.AnythingOfType("[]uint8")).Return(sampleBlock, nil)
				copy(blkHash[:], dummyBlockHash)
				return blkHash, &types.NewBlockNotice{BlockHash:dummyBlockHash}
			},
			func(tt *testing.T, actor *MockActorService, ca *MockChainAccessor) {
				actor.AssertNotCalled(tt,"GetChainAccessor")
				actor.AssertNotCalled(tt,"SendRequest",message.P2PSvc, mock.Anything)
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockCA := new(MockChainAccessor)
			mockActor.On("GetChainAccessor").Return(mockCA)
			mockActor.On("SendRequest", mock.Anything, mock.AnythingOfType("*message.GetBlockInfos"))
			mockPeer := new(MockRemotePeer)
			mockPeer.On("Meta").Return(sampleMeta)
			mockPeer.On("ID").Return(sampleMeta.ID)

			hash, data := test.setup(t, mockActor, mockCA)
			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)
			target.syncing = test.syncing
			if test.put != nil  {
				target.blkCache.Add(*test.put, true)
			}
			target.HandleNewBlockNotice(mockPeer, hash, data )
			test.verify(t,mockActor,mockCA)
		})
	}
}


func TestSyncManager_HandleNewTxNotice(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	rawHashes := sampleTxs
	txHashes := sampleTxHashes

	// test if new block notice comes
	tests := []struct {
		name string
		inCache []TxHash
		verify func(tt *testing.T, actor *MockActorService)
		expected []TxHash
	}{
		// 1. Succ : valid tx hashes and not exist in local cache
		{"TSuccAllNew", nil,
			func(tt *testing.T, actor *MockActorService) {
				actor.AssertCalled(tt,"SendRequest",message.P2PSvc, mock.MatchedBy(func(arg *message.GetTransactions) bool {
					for i,hash := range arg.Hashes {
						assert.True(tt, bytes.Equal(hash, txHashes[i][:]))
					}
					return len(arg.Hashes) == len(txHashes)
				}))
			} , sampleTxHashes},
		// 2. Succ : valid tx hashes and partially exist in local cache
		{"TSuccExistPart", txHashes[2:],
			func(tt *testing.T, actor *MockActorService) {
				// only hashes not in cache are sent to method, which is first 2 hashes
				actor.AssertCalled(tt,"SendRequest",message.P2PSvc, mock.MatchedBy(func(arg *message.GetTransactions) bool {
					for i,hash := range arg.Hashes {
						assert.True(tt, bytes.Equal(hash, txHashes[i][:]))
					}
					return len(arg.Hashes) == 2
				}))
			}, sampleTxHashes[:len(sampleTxHashes)-1]},
		// 3. Succ : valid tx hashes and all exist in local cache
		{"TSuccExistAll", txHashes,
			func(tt *testing.T, actor *MockActorService) {
				actor.AssertNotCalled(tt,"SendRequest",message.P2PSvc, mock.AnythingOfType("*message.GetTransactions"))
			}, sampleTxHashes[:0]},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockActor.On("SendRequest", message.P2PSvc, mock.AnythingOfType("*message.GetTransactions"))
			mockPeer := new(MockRemotePeer)
			mockPeer.On("Meta").Return(sampleMeta)
			mockPeer.On("ID").Return(sampleMeta.ID)
			data := &types.NewTransactionsNotice{TxHashes:rawHashes}

			target := newSyncManager(mockActor, mockPM, logger)
			if test.inCache != nil  {
				for _, hash := range test.inCache {
					target.(*syncManager).txCache.Add(hash, true)
				}
			}
			target.HandleNewTxNotice(mockPeer, txHashes, data )
			test.verify(t,mockActor)
		})
	}
}

func TestSyncManager_DoSync(t *testing.T) {
	hashes := make([]message.BlockHash, len(sampleTxs))
	for i, hash := range sampleTxs {
		hashes[i] = message.BlockHash(hash)
	}
	stopHash := message.BlockHash(dummyBlockHash)

	mockPM := new(MockPeerManager)
	mockActor := new(MockActorService)
	mockActor.On("TellRequest", message.P2PSvc, mock.AnythingOfType("*message.GetMissingRequest"))
	mockMF := new(MockMoFactory)
	mockMF.On("newMsgRequestOrder",mock.Anything,mock.Anything,mock.Anything).Return(new(MockMsgOrder))
	mockPeer := new(MockRemotePeer)
	mockPeer.On("Meta").Return(sampleMeta)
	mockPeer.On("ID").Return(sampleMeta.ID)
	mockPeer.On("MF").Return(mockMF)
	mockPeer.On("sendMessage", mock.Anything)

	target := newSyncManager(mockActor, mockPM, logger).(*syncManager)
	dummySW := &syncWorker{}
	// assume first sync was done
	target.sw = dummySW

	// second do sync should be failed
	target.DoSync(mockPeer, hashes, stopHash)
	assert.NotNil(t, target.sw)
	assert.ObjectsAreEqual(dummySW, target.sw)

	mockActor.AssertNumberOfCalls(t,"TellRequest", 0)
	mockMF.AssertNumberOfCalls(t,"newMsgRequestOrder", 0)

}

func TestSyncManager_HandleGetBlockResponse(t *testing.T) {
	totalBlkCnt := len(sampleTxs)
	sampleBlocks := make([]*types.Block, totalBlkCnt)
	for i, hash := range sampleTxs {
		sampleBlocks[i] = &types.Block{Hash:hash}
	}
	tests := []struct {
		name         string
		sw           *syncWorker
		// call count directly to chainservice
		chainCallCnt int
		// call SendRequest count indirectly to chainservice vial syncWorker
		swSendCnt int
		// call CallRequest count indirectly to chainservice vial syncWorker
		swCallCnt int

	}{
		// 1. worker is exist, pass to worker
		{"TExist", &syncWorker{finish: make(chan interface{},1)}, 0, totalBlkCnt-1, 1},
		// 1. worker is not and directly call chainservice
		{"TNotExist", nil,totalBlkCnt, 0, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dummyRsp := &message.AddBlockRsp{BlockNo:1, Err:nil}
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockActor.On("TellRequest", message.P2PSvc, mock.AnythingOfType("*message.GetMissingRequest"))
			mockActor.On("SendRequest", message.ChainSvc, mock.AnythingOfType("*message.AddBlock"))
			mockActor.On("CallRequestDefaultTimeout", message.ChainSvc, mock.AnythingOfType("*message.AddBlock")).Return(dummyRsp, nil)
			mockActor.On("CallRequest", message.ChainSvc, mock.AnythingOfType("*message.AddBlock"), mock.AnythingOfType("time.Duration")).Return(dummyRsp, nil)
			mockMF := new(MockMoFactory)
			mockMF.On("newMsgRequestOrder",mock.Anything,mock.Anything,mock.Anything).Return(new(MockMsgOrder))
			mockPeer := new(MockRemotePeer)
			mockPeer.On("Meta").Return(sampleMeta)
			mockPeer.On("ID").Return(sampleMeta.ID)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("sendMessage", mock.Anything)

			dummyMsgID := NewMsgID()
			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)
			target.sw = test.sw
			if test.sw != nil {
				target.sw.sm = target
				target.syncing = true
				target.sw.requestMsgID = dummyMsgID
				target.sw.cancel = make(chan interface{},10)
				target.sw.retain = make(chan interface{},10)
				target.sw.finish = make(chan interface{},10)
			}

			msg := &V030Message{originalID:dummyMsgID}
			resp := &types.GetBlockResponse{Blocks:sampleBlocks}
			target.HandleGetBlockResponse(mockPeer, msg, resp)

			mockActor.AssertNumberOfCalls(t, "CallRequestDefaultTimeout", test.chainCallCnt)
			mockActor.AssertNumberOfCalls(t, "SendRequest", test.swSendCnt)
			mockActor.AssertNumberOfCalls(t, "CallRequest", test.swCallCnt)
		})
	}
}
