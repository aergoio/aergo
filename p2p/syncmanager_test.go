/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
)

func TestSyncManager_HandleBlockProducedNotice(t *testing.T) {
	// only interested in max block size
	chain.Init(1024*1024, "", false, 0, 0)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.p2p")
	sampleBlock := &types.Block{Hash: dummyBlockHash}
	txs := make([]*types.Tx, 1)
	txs[0] = &types.Tx{Hash: make([]byte, 1024*1024*2)}
	sampleBigBlock := &types.Block{Hash: dummyBlockHash, Body: &types.BlockBody{Txs: txs}}
	var blkHash = types.ToBlockID(dummyBlockHash)
	// test if new block notice comes
	tests := []struct {
		name       string
		put        *types.BlockID
		addedBlock *types.Block

		wantActorCall bool
	}{
		// 1. Succ : valid block hash and not exist in local
		{"TSucc", nil, sampleBlock, true},
		// 2. Rare case - valid block hash but already exist in local cache
		{"TExist", &blkHash, sampleBlock, false},
		{"TTooBigBlock", nil, sampleBigBlock, false},
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
	chain.Init(1024*1024, "", false, 0, 0)

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

func TestSyncManager_HandleGetBlockResponse(t *testing.T) {
	// only interested in max block size
	chain.Init(1024*1024, "", false, 0, 0)

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

func Test_syncManager_Constructor(t *testing.T) {
	tests := []struct {
		name  string
		delay time.Duration
	}{
		{"TNormal", time.Millisecond * 10},
		{"TInstant", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)

			sm := newSyncManager(mockActor, mockPM, logger).(*syncManager)

			sm.Start()
			if sm.tm == nil {
				t.Fatalf("newSyncManager member variable tm was not set")
			}
			if tt.delay > 0 {
				time.Sleep(tt.delay)
			}
			sm.Stop()
		})
	}
}
