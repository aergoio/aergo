package syncer

import (
	"fmt"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/message"
	"github.com/stretchr/testify/assert"
)

func TestSyncer_sync1000(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:1], localChainLen-1)

	remoteChains := []*chain.StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: 1000}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	assert.Equal(t, int(targetNo), syncer.localChain.Best, "sync failed")
}

func TestSyncer_sync10000(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	remoteChainLen := 10002
	localChainLen := 10000
	targetNo := uint64(10000)
	ancestorNo := 100

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:ancestorNo+1], localChainLen-(ancestorNo+1))

	remoteChains := []*chain.StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: ancestorNo}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: targetNo}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	assert.Equal(t, int(targetNo), syncer.localChain.Best, "sync failed")
}

func TestSyncer_sync_multiPeer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:1], localChainLen-1)

	remoteChains := []*chain.StubBlockChain{remoteChain, remoteChain, remoteChain, remoteChain}
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

	assert.Equal(t, int(targetNo), syncer.localChain.Best, "sync failed")
}

// case : peer1 is slow (timeout)
func TestSyncer_sync_slowPeer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:1], localChainLen-1)

	remoteChains := []*chain.StubBlockChain{remoteChain, remoteChain, remoteChain, remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}
	testCfg.debugContext.logBadPeers = make(map[int]bool)
	testCfg.fetchTimeOut = time.Millisecond * 500
	expBadPeer := 1
	peers[expBadPeer].timeDelaySec = time.Second * 1

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

	assert.Equal(t, int(targetNo), syncer.localChain.Best, "sync failed")
}

func TestSyncer_sync_allPeerBad(t *testing.T) {
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:1], localChainLen-1)

	remoteChains := []*chain.StubBlockChain{remoteChain, remoteChain, remoteChain, remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}
	testCfg.debugContext.logBadPeers = make(map[int]bool)

	testCfg.fetchTimeOut = time.Millisecond * 500
	peers[0].timeDelaySec = time.Second * 1
	peers[1].timeDelaySec = time.Second * 1
	peers[2].timeDelaySec = time.Second * 1
	peers[3].timeDelaySec = time.Second * 1

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

	assert.NotEqual(t, int(targetNo), syncer.localChain.Best, "sync must fail")
}

func TestSyncerAllPeerBadByResponseError(t *testing.T) {
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:1], localChainLen-1)

	remoteChains := []*chain.StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}
	testCfg.debugContext.logBadPeers = make(map[int]bool)

	testCfg.fetchTimeOut = time.Millisecond * 500

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)

	peers[0].HookGetBlockChunkRsp = func(msgReq *message.GetBlockChunks) {
		rsp := &message.GetBlockChunksRsp{Seq: msgReq.Seq, ToWhom: msgReq.ToWhom, Blocks: nil, Err: message.RemotePeerFailError}
		syncer.stubRequester.TellTo(message.SyncerSvc, rsp)
	}

	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: 1000}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	//check if all peers is used
	for i, peer := range peers {
		assert.True(t, peer.blockFetched, fmt.Sprintf("%d is not used", i))
	}

	for _, peerNo := range []int{0} {
		assert.True(t, testCfg.debugContext.logBadPeers[peerNo], "check bad peer")
	}

	assert.NotEqual(t, int(targetNo), syncer.localChain.Best, "sync must fail")
}

func TestSyncerAlreadySynched(t *testing.T) {
	//When sync is already done before finder runs
	remoteChainLen := 1010
	//localChainLen := 999
	targetNo := uint64(1000)

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:1000], 0)

	remoteChains := []*chain.StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0, expErrResult: ErrAlreadySyncDone}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	syncer.getAnchorsHookFn = func(stubSyncer *StubSyncer) {
		stubSyncer.localChain.AddBlock(remoteChain.Blocks[1000])
		stubSyncer.localChain.AddBlock(remoteChain.Blocks[1001])
		stubSyncer.localChain.AddBlock(remoteChain.Blocks[1002])
	}
	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: targetNo}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	assert.Equal(t, 1002, syncer.localChain.Best, "sync failed")
}

func TestSyncer_invalid_seq_getancestor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	remoteChainLen := 1002
	localChainLen := 10
	targetNo := uint64(1000)

	remoteChain := chain.InitStubBlockChain(nil, remoteChainLen)
	localChain := chain.InitStubBlockChain(remoteChain.Blocks[0:1], localChainLen-1)

	remoteChains := []*chain.StubBlockChain{remoteChain}
	peers := makeStubPeerSet(remoteChains)

	testCfg := *SyncerCfg
	testCfg.debugContext = &SyncerDebug{t: t, expAncestor: 0}

	syncer := NewTestSyncer(t, localChain, remoteChain, peers, &testCfg)
	// set prev sequence
	syncer.realSyncer.Seq = 99
	syncer.getSyncAncestorHookFn = func(stubSyncer *StubSyncer, msg *message.GetSyncAncestor) {
		// send prev sequence
		newMsg := &message.GetSyncAncestor{Seq: 98, ToWhom: msg.ToWhom, Hashes: msg.Hashes}
		syncer.SendGetSyncAncestorRsp(newMsg)

		newMsg = &message.GetSyncAncestor{Seq: 99, ToWhom: msg.ToWhom, Hashes: msg.Hashes}
		syncer.SendGetSyncAncestorRsp(newMsg)
	}

	syncer.start()

	syncReq := &message.SyncStart{PeerID: targetPeerID, TargetNo: 1000}
	syncer.stubRequester.TellTo(message.SyncerSvc, syncReq)

	syncer.waitStop()

	assert.Equal(t, int(targetNo), syncer.localChain.Best, "sync failed")
}
