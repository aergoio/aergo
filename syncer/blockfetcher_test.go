package syncer

import (
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

// test blockfetcher without finder/hashfetcher
// - must create SyncCtx manually, because finder will be skipped
func TestBlockFetcher_simple(t *testing.T) {
	remoteChainLen := 10
	targetNo := uint64(5)

	//ancestor = 0
	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:1], 0)

	remoteChains := []*StubBlockChain{remoteChain, remoteChain} //peer count = 2
	peers := makeStubPeerSet(remoteChains)

	//set debug property
	testCfg := *SyncerCfg
	testCfg.maxHashReqSize = TestMaxHashReqSize
	testCfg.maxBlockReqSize = TestMaxBlockFetchSize
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}
	testCfg.debugContext.targetNo = targetNo

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)

	//set ctx manually because finder will be skipped
	ctx := types.NewSyncCtx("peer-0", targetNo, uint64(localChain.best))
	ancestor := remoteChain.blocks[0]
	ctx.SetAncestor(ancestor)

	//run blockFetcher direct
	syncer.runTestBlockFetcher(ctx)

	syncer.checkResultFn = func(stubSyncer *StubSyncer) {
		//check blockFetcher status
		bf := stubSyncer.realSyncer.blockFetcher
		assert.Equal(stubSyncer.t, uint64(stubSyncer.cfg.debugContext.targetNo), bf.stat.getMaxChunkRsp().BlockNo(), "response mismatch")
		assert.Equal(stubSyncer.t, uint64(stubSyncer.cfg.debugContext.targetNo), bf.stat.getLastAddBlock().BlockNo(), "last add block mismatch")
	}

	syncer.start()

	testHashSet := func(prev *types.BlockInfo, count uint64) {
		//push hashSet next from prev
		hashes, _ := syncer.remoteChain.GetHashes(prev, count)

		syncer.sendHashSetToBlockFetcher(&HashSet{len(hashes), hashes, prev.No + 1})
	}

	testHashSet(&types.BlockInfo{Hash: ancestor.GetHash(), No: ancestor.BlockNo()}, 3)

	prevInfo := remoteChain.GetBlockInfo(ancestor.BlockNo() + 3)
	testHashSet(prevInfo, 2)

	syncer.waitStop()
}
