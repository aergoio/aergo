package syncer

import (
	"github.com/aergoio/aergo/message"
	"testing"
)

func testFullscanSucceed(t *testing.T, expAncestor uint64) {
	logger.Debug().Uint64("expAncestor", expAncestor).Msg("testfullscan")

	remoteChainLen := 11
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:expAncestor+1], localChainLen-int(expAncestor+1))

	remoteChains := []*StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	//set debug property
	testCfg := *SyncerCfg
	testCfg.useFullScanOnly = true
	testCfg.debugContext = &SyncerDebug{t: t, stopByFinder: true, expAncestor: int(expAncestor)}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)

	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: targetNo}
	syncer.testhub.Tell(message.SyncerSvc, syncReq)

	syncer.waitStop()
}

func TestFinder_fullscan_found(t *testing.T) {
	for i := 0; i < 10; i++ {
		testFullscanSucceed(t, uint64(i))
	}
}

func TestFinder_fullscan_notfound(t *testing.T) {
	remoteChainLen := 1002
	localChainLen := 1000
	targetNo := uint64(1000)

	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(nil, localChainLen)

	remoteChains := []*StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	//set debug property
	testCfg := *SyncerCfg
	testCfg.useFullScanOnly = true
	testCfg.debugContext = &SyncerDebug{t: t, stopByFinder: true, expAncestor: -1}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)

	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: targetNo}
	syncer.testhub.Tell(message.SyncerSvc, syncReq)

	syncer.waitStop()
}
