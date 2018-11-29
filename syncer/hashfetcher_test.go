package syncer

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"testing"
	"time"
)

func TestHashFetcher_normal(t *testing.T) {
	remoteChainLen := 100
	localChainLen := 99
	targetNo := uint64(99)

	//ancestor = 0
	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:1], localChainLen)

	remoteChains := []*StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	//set debug property
	testCfg := *SyncerCfg
	testCfg.maxHashReqSize = TestMaxHashReqSize
	testCfg.maxBlockReqSize = TestMaxBlockFetchSize
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}
	testCfg.debugContext.debugHashFetcher = true
	testCfg.debugContext.targetNo = targetNo

	//set ctx because finder is skipped
	ctx := types.NewSyncCtx("peer-0", targetNo, uint64(localChain.best))
	ancestorInfo := remoteChain.GetBlockInfo(0)

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.realSyncer.ctx = ctx

	syncer.start()

	//ancestor of ctx will be set by FinderResult
	syncer.stubRequester.TellTo(message.SyncerSvc, &message.FinderResult{ancestorInfo, nil})

	syncer.waitStop()
}

//test if hashfetcher stops successfully while waiting to send HashSet to resultCh
func TestHashFetcher_quit(t *testing.T) {
	remoteChainLen := 100
	localChainLen := 99
	targetNo := uint64(99)

	//ancestor = 0
	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:1], localChainLen)

	remoteChains := []*StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	//set debug property
	testCfg := *SyncerCfg
	testCfg.maxHashReqSize = TestMaxHashReqSize
	testCfg.maxBlockReqSize = TestMaxBlockFetchSize
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}
	testCfg.debugContext.debugHashFetcher = true
	testCfg.debugContext.BfWaitTime = time.Second * 1000

	//set ctx because finder is skipped
	ctx := types.NewSyncCtx("peer-0", targetNo, uint64(localChain.best))
	ancestorInfo := remoteChain.GetBlockInfo(0)

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.realSyncer.ctx = ctx

	syncer.start()

	//ancestor of ctx will be set by FinderResult
	syncer.stubRequester.TellTo(message.SyncerSvc, &message.FinderResult{ancestorInfo, nil})

	//test if hashfetcher stop
	go func() {
		time.Sleep(time.Second * 1)
		stopSyncer(syncer.stubRequester, NameBlockFetcher, ErrQuitBlockFetcher)
	}()
	syncer.waitStop()
}

func TestHashFetcher_ResponseError(t *testing.T) {
	//TODO test hashfetcher error
	/*
		//make remoteBlockChain
		remoteChain := initStubBlockChain(nil, 10)
		ancestor := remoteChain.GetBlockByNo(0)

		ctx := types.NewSyncCtx("p1", 5, 1)
		ctx.SetAncestor(ancestor)

		syncer := NewStubSyncer(ctx, false, true, false, nil, remoteChain, TestMaxHashReqSize, TestMaxBlockFetchSize)
		syncer.hf.Start()

		//hashset 2~4, 5~7, 8~9
		//receive GetHash message
		msg := syncer.stubRequester.recvMessage()
		assert.IsTypef(t, &message.GetHashes{}, msg, "invalid message from hf")
		syncer.handleMessageManual(t, msg, ErrGetHashesRspError)

		//stop
		msg = syncer.stubRequester.recvMessage()
		assert.IsTypef(t, &message.SyncStop{}, msg, "need syncer stop msg")
		syncer.handleMessageManual(t, msg, nil)

		assert.True(t, syncer.isStop, "hashfetcher finished")
	*/
}
