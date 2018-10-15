/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)


func TestSyncManager_HandleNewBlockNotice(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	sampleBlock := &types.Block{Hash:dummyBlockHash}
	var blkHash BlockHash
	// test if new block notice comes
	tests := []struct {
		name string
		put *BlockHash
		setup func(tt *testing.T, actor *MockActorService) (BlockHash,*types.NewBlockNotice)
		verify func(tt *testing.T, actor *MockActorService)
	}{
		// 1. Succ : valid block hash and not exist in local
		{"TSucc", nil,
		func(tt *testing.T, actor *MockActorService) (BlockHash,*types.NewBlockNotice) {
			actor.On("CallRequest",message.ChainSvc, mock.AnythingOfType("*message.GetBlock")).Return(message.GetBlockRsp{Err:fmt.Errorf("not found")}, nil)
			copy(blkHash[:], dummyBlockHash)
			return blkHash, &types.NewBlockNotice{BlockHash:dummyBlockHash}
		},
		func(tt *testing.T, actor *MockActorService) {
			actor.AssertCalled(tt,"CallRequest",message.ChainSvc, mock.AnythingOfType("*message.GetBlock"))
			actor.AssertCalled(tt,"SendRequest",message.P2PSvc, mock.AnythingOfType("*message.GetBlockInfos"))
		}},
		// 1-1. Succ : valid block hash and exist in chainsvc, but not in cache
		{"TSuccExistChain", nil,
			func(tt *testing.T, actor *MockActorService) (BlockHash,*types.NewBlockNotice) {
				actor.On("CallRequest",message.ChainSvc, mock.AnythingOfType("*message.GetBlock")).Return(message.GetBlockRsp{Block:sampleBlock}, nil)
				copy(blkHash[:], dummyBlockHash)
				return blkHash, &types.NewBlockNotice{BlockHash:dummyBlockHash}
			},
			func(tt *testing.T, actor *MockActorService) {
				actor.AssertCalled(tt,"CallRequest",message.ChainSvc, mock.AnythingOfType("*message.GetBlock"))
				actor.AssertNotCalled(tt,"SendRequest",message.P2PSvc, mock.AnythingOfType("*message.GetBlockInfos"))
			}},
		// 2. SuccCachehit : valid block hash but already exist in local cache
		{"TSuccExistCache", &blkHash,
			func(tt *testing.T, actor *MockActorService) (BlockHash,*types.NewBlockNotice) {
				actor.On("CallRequest",message.ChainSvc, mock.AnythingOfType("*message.GetBlock")).Return(message.GetBlockRsp{Block:sampleBlock}, nil)
				copy(blkHash[:], dummyBlockHash)
				return blkHash, &types.NewBlockNotice{BlockHash:dummyBlockHash}
			},
			func(tt *testing.T, actor *MockActorService) {
				actor.AssertNotCalled(tt,"CallRequest",message.ChainSvc, mock.Anything)
				actor.AssertNotCalled(tt,"SendRequest",message.P2PSvc, mock.Anything)
			}},
		// 3. chainService failed
		{"TActorFail", nil,
			func(tt *testing.T, actor *MockActorService) (BlockHash,*types.NewBlockNotice) {
				actor.On("CallRequest",message.ChainSvc, mock.AnythingOfType("*message.GetBlock")).After(time.Millisecond*100).Return(nil, fmt.Errorf("actor timeout"))
				copy(blkHash[:], dummyBlockHash)
				return blkHash, &types.NewBlockNotice{BlockHash:dummyBlockHash}
			},
			func(tt *testing.T, actor *MockActorService) {
				actor.AssertCalled(tt,"CallRequest",message.ChainSvc, mock.AnythingOfType("*message.GetBlock"))
				actor.AssertNotCalled(tt,"SendRequest",message.P2PSvc, mock.Anything)
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockActor.On("SendRequest", mock.Anything, mock.AnythingOfType("*message.GetBlockInfos"))
			mockPeer := new(MockRemotePeer)
			mockPeer.On("Meta").Return(sampleMeta)
			mockPeer.On("ID").Return(sampleMeta.ID)

			hash, data := test.setup(t, mockActor)
			target := newSyncManager(mockActor, mockPM, logger)
			if test.put != nil  {
				target.(*syncManager).blkCache.Add(*test.put, true)
			}
			target.HandleNewBlockNotice(mockPeer, hash, data )
			test.verify(t,mockActor)
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
		put []TxHash
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
				actor.AssertCalled(tt,"SendRequest",message.P2PSvc, mock.MatchedBy(func(arg *message.GetTransactions) bool {
					for i,hash := range arg.Hashes {
						assert.True(tt, bytes.Equal(hash, txHashes[i][:]))
					}
					return len(arg.Hashes) == len(txHashes) - 1
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
			if test.put != nil  {
				for _, hash := range test.put {
					target.(*syncManager).txCache.Add(hash, true)
				}
			}
			target.HandleNewTxNotice(mockPeer, txHashes, data )
			test.verify(t,mockActor)
		})
	}
}
