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


func TestSyncManager_HandleBlockProducedNotice(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	sampleBlock := &types.Block{Hash:dummyBlockHash}
	var blkHash BlkHash
	copy(blkHash[:], dummyBlockHash)
	// test if new block notice comes
	tests := []struct {
		name string
		put *BlkHash

		wantActorCall bool
	}{
		// 1. Succ : valid block hash and not exist in local
		{"TSucc", nil, true},
		// 2. Rare case - valid block hash but already exist in local cache
		{"TExist", &blkHash, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockActor.On("SendRequest", message.ChainSvc, mock.AnythingOfType("*message.AddBlock"))
			mockPeer := new(MockRemotePeer)
			mockPeer.On("ID").Return(sampleMeta.ID)

			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)
			if test.put != nil  {
				target.blkCache.Add(*test.put, true)
			}
			target.HandleBlockProducedNotice(mockPeer, sampleBlock)
			if test.wantActorCall {
				mockActor.AssertCalled(t,"SendRequest", message.ChainSvc, mock.AnythingOfType("*message.AddBlock"))
			} else {
				mockActor.AssertNotCalled(t, "SendRequest",mock.Anything, mock.Anything)
			}
		})
	}
}


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

			_, data := test.setup(t, mockActor, mockCA)
			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)
			target.syncing = test.syncing
			if test.put != nil  {
				target.blkCache.Add(*test.put, true)
			}
			target.HandleNewBlockNotice(mockPeer, data)
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

func TestSyncManager_HandleGetBlockResponse(t *testing.T) {
	totalBlkCnt := len(sampleTxs)
	sampleBlocks := make([]*types.Block, totalBlkCnt)
	for i, hash := range sampleTxs {
		sampleBlocks[i] = &types.Block{Hash:hash}
	}
	tests := []struct {
		name         string
		respBlocks   []*types.Block

		// call count directly to chainservice
		chainCallCnt int
	}{
		// 1. message triggered by NewBlockNotice (maybe)
		{"TSingleBlock", sampleBlocks[:1], 1, },
		// 2. message triggered by newsyncer but not handled by it (caused by sync fail or timeout)
		{"TZeroBlock", sampleBlocks[:0], 0, },
		{"TMultiBlocks", sampleBlocks, 0, },
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockActor.On("SendRequest", message.ChainSvc, mock.AnythingOfType("*message.AddBlock"))
			mockMF := new(MockMoFactory)
			mockPeer := new(MockRemotePeer)
			mockPeer.On("Meta").Return(sampleMeta)
			mockPeer.On("ID").Return(sampleMeta.ID)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("sendMessage", mock.Anything)

			dummyMsgID := NewMsgID()
			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)

			msg := &V030Message{originalID:dummyMsgID}
			resp := &types.GetBlockResponse{Blocks:test.respBlocks}
			target.HandleGetBlockResponse(mockPeer, msg, resp)

			mockActor.AssertNumberOfCalls(t, "SendRequest", test.chainCallCnt)
		})
	}
}
