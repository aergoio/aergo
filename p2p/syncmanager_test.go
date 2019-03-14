/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmocks"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSyncManager_HandleBlockProducedNotice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.p2p")
	sampleBlock := &types.Block{Hash: dummyBlockHash}
	var blkHash = types.ToBlockID(dummyBlockHash)
	// test if new block notice comes
	tests := []struct {
		name string
		put  *types.BlockID

		wantActorCall bool
	}{
		// 1. Succ : valid block hash and not exist in local
		{"TSucc", nil, true},
		// 2. Rare case - valid block hash but already exist in local cache
		{"TExist", &blkHash, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmocks.NewMockPeerManager(ctrl)
			mockActor := p2pmocks.NewMockActorService(ctrl)
			mockPeer := p2pmocks.NewMockRemotePeer(ctrl)
			if test.wantActorCall {
				mockPeer.EXPECT().ID().Return(sampleMeta.ID)
				mockActor.EXPECT().SendRequest(message.ChainSvc, gomock.Any())
			} else {
				mockPeer.EXPECT().Name().Return("16..aadecf@1")
				mockActor.EXPECT().SendRequest(message.ChainSvc, gomock.Any()).Times(0)
			}

			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)
			if test.put != nil {
				target.blkCache.Add(*test.put, true)
			}
			target.HandleBlockProducedNotice(mockPeer, sampleBlock)
		})
	}
}

func TestSyncManager_HandleNewBlockNotice(t *testing.T) {
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
		setup   func(tt *testing.T, actor *p2pmocks.MockActorService, ca *p2pmocks.MockChainAccessor, peer *p2pmocks.MockRemotePeer) (types.BlockID, *types.NewBlockNotice)
		//verify  func(tt *testing.T, actor *p2pmocks.MockActorService, ca *p2pmocks.MockChainAccessor)
	}{
		// 1. Succ : valid block hash and not exist in local
		{"TSucc", nil, false,
			func(tt *testing.T, actor *p2pmocks.MockActorService, ca *p2pmocks.MockChainAccessor, peer *p2pmocks.MockRemotePeer) (types.BlockID, *types.NewBlockNotice) {
				ca.EXPECT().GetBlock(gomock.Any()).Return(nil, nil)
				actor.EXPECT().GetChainAccessor().Return(ca)
				copy(blkHash[:], dummyBlockHash)
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any())
				peer.EXPECT().Name().Return("16..aadecf@1")
				return blkHash, &types.NewBlockNotice{BlockHash: dummyBlockHash}
			}},
		// 1-1. Succ : valid block hash and exist in chainsvc, but not in cache
		{"TSuccExistChain", nil, false,
			func(tt *testing.T, actor *p2pmocks.MockActorService, ca *p2pmocks.MockChainAccessor, peer *p2pmocks.MockRemotePeer) (types.BlockID, *types.NewBlockNotice) {
				ca.EXPECT().GetBlock(gomock.Any()).Return(sampleBlock, nil)
				copy(blkHash[:], dummyBlockHash)
				actor.EXPECT().GetChainAccessor().Return(ca)
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).MaxTimes(0)
				return blkHash, &types.NewBlockNotice{BlockHash: dummyBlockHash}
			}},
		// 2. SuccCachehit : valid block hash but already exist in local cache
		{"TSuccExistCache", &blkHash, false,
			func(tt *testing.T, actor *p2pmocks.MockActorService, ca *p2pmocks.MockChainAccessor, peer *p2pmocks.MockRemotePeer) (types.BlockID, *types.NewBlockNotice) {
				ca.EXPECT().GetBlock(gomock.Any()).Return(sampleBlock, nil).MaxTimes(0)
				copy(blkHash[:], dummyBlockHash)

				//ca.AssertNotCalled(tt, "GetBlock", mock.AnythingOfType("[]uint8"))
				//actor.EXPECT().AssertNotCalled(tt, "SendRequest", message.P2PSvc, mock.Anything)
				actor.EXPECT().SendRequest(message.P2PSvc, mock.Anything).MaxTimes(0)
				return blkHash, &types.NewBlockNotice{BlockHash: dummyBlockHash}
			}},
		// 2. Busy : other sync worker is working
		{"TBusy", &blkHash, true,
			func(tt *testing.T, actor *p2pmocks.MockActorService, ca *p2pmocks.MockChainAccessor, peer *p2pmocks.MockRemotePeer) (types.BlockID, *types.NewBlockNotice) {
				copy(blkHash[:], dummyBlockHash)
				actor.EXPECT().GetChainAccessor().MaxTimes(0)
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).MaxTimes(0)
				return blkHash, &types.NewBlockNotice{BlockHash: dummyBlockHash}
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmocks.NewMockPeerManager(ctrl)
			mockActor := p2pmocks.NewMockActorService(ctrl)
			mockCA := p2pmocks.NewMockChainAccessor(ctrl)
			mockPeer := p2pmocks.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID)

			_, data := test.setup(t, mockActor, mockCA, mockPeer)
			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)
			target.syncing = test.syncing
			if test.put != nil {
				target.blkCache.Add(*test.put, true)
			}
			target.HandleNewBlockNotice(mockPeer, data)
			//test.verify(t, mockActor, mockCA)
		})
	}
}

func TestSyncManager_HandleNewTxNotice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.p2p")
	rawHashes := sampleTxs
	txHashes := sampleTxHashes

	// test if new block notice comes
	tests := []struct {
		name     string
		inCache  []types.TxID
		setup    func(tt *testing.T, actor *p2pmocks.MockActorService)
		expected []types.TxID
	}{
		// 1. Succ : valid tx hashes and not exist in local cache
		{"TSuccAllNew", nil,
			func(tt *testing.T, actor *p2pmocks.MockActorService) {
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).DoAndReturn(func(name string, arg *message.GetTransactions) {
					for i, hash := range arg.Hashes {
						assert.True(tt, bytes.Equal(hash, txHashes[i][:]))
					}
					assert.True(tt, len(arg.Hashes) == len(txHashes))
				})
			}, sampleTxHashes},
		// 2. Succ : valid tx hashes and partially exist in local cache
		{"TSuccExistPart", txHashes[2:],
			func(tt *testing.T, actor *p2pmocks.MockActorService) {
				// only hashes not in cache are sent to method, which is first 2 hashes
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).DoAndReturn(func(name string, arg *message.GetTransactions) {
					for i, hash := range arg.Hashes {
						assert.True(tt, bytes.Equal(hash, txHashes[i][:]))
					}
					assert.True(tt, len(arg.Hashes) == 2)
				})

			}, sampleTxHashes[:len(sampleTxHashes)-1]},
		// 3. Succ : valid tx hashes and all exist in local cache
		{"TSuccExistAll", txHashes,
			func(tt *testing.T, actor *p2pmocks.MockActorService) {
				actor.EXPECT().SendRequest(message.P2PSvc, gomock.Any()).MaxTimes(0)
			}, sampleTxHashes[:0]},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmocks.NewMockPeerManager(ctrl)
			mockActor := p2pmocks.NewMockActorService(ctrl)
			mockPeer := p2pmocks.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID)

			data := &types.NewTransactionsNotice{TxHashes: rawHashes}

			test.setup(t, mockActor)
			target := newSyncManager(mockActor, mockPM, logger)
			if test.inCache != nil {
				for _, hash := range test.inCache {
					target.(*syncManager).txCache.Add(hash, true)
				}
			}
			target.HandleNewTxNotice(mockPeer, txHashes, data)
		})
	}
}

func TestSyncManager_HandleGetBlockResponse(t *testing.T) {
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
			mockPM := p2pmocks.NewMockPeerManager(ctrl)
			mockActor := p2pmocks.NewMockActorService(ctrl)
			mockPeer := p2pmocks.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(sampleMeta.ID)

			mockActor.EXPECT().SendRequest(gomock.Any(), gomock.Any()).Times(test.chainCallCnt)
			dummyMsgID := p2pcommon.NewMsgID()
			target := newSyncManager(mockActor, mockPM, logger).(*syncManager)

			msg := &V030Message{originalID: dummyMsgID}
			resp := &types.GetBlockResponse{Blocks: test.respBlocks}
			target.HandleGetBlockResponse(mockPeer, msg, resp)

			//mockActor.AssertNumberOfCalls(t, "SendRequest", test.chainCallCnt)
		})
	}
}
