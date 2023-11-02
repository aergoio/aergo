package syncer

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/message"
)

func testFullscanSucceed(t *testing.T, expAncestor uint64) {
	logger.Debug().Uint64("expAncestor", expAncestor).Msg("testfullscan")

	remoteChainLen := 11
	localChainLen := 10
	targetNo := uint64(11)

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:expAncestor+1], localChainLen-int(expAncestor+1))

	remoteChains := []*chain.StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	//set debug property
	testCfg := *SyncerCfg
	testCfg.useFullScanOnly = true
	testCfg.debugContext = &SyncerDebug{t: t, debugFinder: true, expAncestor: int(expAncestor)}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)

	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: targetNo}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

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

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(nil, localChainLen)

	remoteChains := []*chain.StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	//set debug property
	testCfg := *SyncerCfg
	testCfg.useFullScanOnly = true
	testCfg.debugContext = &SyncerDebug{t: t, debugFinder: true, expAncestor: -1}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)

	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: targetNo}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()
}

// test finder stop when close finder.quitCh
func TestFinder_timeout(t *testing.T) {
	logger.Debug().Int("expAncestor", -1).Msg("testfullscan")

	remoteChainLen := 1001
	localChainLen := 1000
	targetNo := uint64(1000)

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:1], localChainLen-1)

	remoteChains := []*chain.StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	//set debug property
	testCfg := *SyncerCfg
	testCfg.fetchTimeOut = time.Millisecond * 500
	testCfg.debugContext = &SyncerDebug{t: t, debugFinder: true, expAncestor: -1, expErrResult: ErrHubFutureTimeOut}
	peers[0].timeDelaySec = time.Second * 1

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)

	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: targetNo}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()
}
