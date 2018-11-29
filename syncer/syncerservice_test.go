package syncer

import (
	"fmt"
	"github.com/aergoio/aergo/message"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSyncer_sync1000(t *testing.T) {
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:1], localChainLen-1)

	remoteChains := []*StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: 1000}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	assert.Equal(t, int(targetNo), syncer.localChain.best, "sync failed")
}

func TestSyncer_sync10000(t *testing.T) {
	remoteChainLen := 10002
	localChainLen := 10000
	targetNo := uint64(10000)
	ancestorNo := 100

	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:ancestorNo+1], localChainLen-(ancestorNo+1))

	remoteChains := []*StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: ancestorNo}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: targetNo}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	assert.Equal(t, int(targetNo), syncer.localChain.best, "sync failed")
}

func TestSyncer_sync_multiPeer(t *testing.T) {
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:1], localChainLen-1)

	remoteChains := []*StubBlockChain{remoteChain, remoteChain, remoteChain, remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: 1000}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	//check if all peers is used
	for i, peer := range peers {
		assert.True(t, peer.blockFetched, fmt.Sprintf("%d is not used", i))
	}

	assert.Equal(t, int(targetNo), syncer.localChain.best, "sync failed")
}

//case : peer1 is slow (timeout)
func TestSyncer_sync_slowPeer(t *testing.T) {
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:1], localChainLen-1)

	remoteChains := []*StubBlockChain{remoteChain, remoteChain, remoteChain, remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}
	testCfg.debugContext.logBadPeers = make(map[int]bool)
	testCfg.fetchTimeOut = time.Second * 2
	expBadPeer := 1
	peers[expBadPeer].timeDelaySec = time.Second * 4

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: 1000}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	//check if all peers is used
	for i, peer := range peers {
		assert.True(t, peer.blockFetched, fmt.Sprintf("%d is not used", i))
	}

	//check bad peer
	assert.True(t, testCfg.debugContext.logBadPeers[expBadPeer], "check bad peer")

	assert.Equal(t, int(targetNo), syncer.localChain.best, "sync failed")
}

func TestSyncer_sync_allPeerBad(t *testing.T) {
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:1], localChainLen-1)

	remoteChains := []*StubBlockChain{remoteChain, remoteChain, remoteChain, remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}
	testCfg.debugContext.logBadPeers = make(map[int]bool)

	testCfg.fetchTimeOut = time.Second * 1
	peers[0].timeDelaySec = time.Second * 2
	peers[1].timeDelaySec = time.Second * 2
	peers[2].timeDelaySec = time.Second * 2
	peers[3].timeDelaySec = time.Second * 2

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: 1000}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	//check if all peers is used
	for i, peer := range peers {
		assert.True(t, peer.blockFetched, fmt.Sprintf("%d is not used", i))
	}

	for _, peerNo := range []int{0, 1, 2, 3} {
		assert.True(t, testCfg.debugContext.logBadPeers[peerNo], "check bad peer")
	}

	assert.NotEqual(t, int(targetNo), syncer.localChain.best, "sync must fail")
}
