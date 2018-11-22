package syncer

import (
	"fmt"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	targetPeerID = peer.ID([]byte(fmt.Sprintf("peer-%d", 0)))

	remoteChainLen = 20
	localChainLen  = 10

	targetNo = uint64(20)
)

func testFullscanSucceed(t *testing.T, expAncestor uint64) {
	logger.Debug().Uint64("expAncestor", expAncestor).Msg("testfullscan")

	//case 1: 0 is ancestor
	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(remoteChain.blocks[0:expAncestor+1], localChainLen-int(expAncestor+1))

	ctx := types.NewSyncCtx(targetPeerID, targetNo, uint64(localChain.best))

	syncer := NewStubSyncer(ctx, true, false, false, localChain, remoteChain)

	assert.Equal(t, targetPeerID, peer.ID(syncer.stubPeers[0].addr.GetPeerID()), "target peerid=peer-0")

	syncer.finder.setFullScanOnly(10)

	syncer.finder.start()

	//syncer stop after finding ancestor
	syncer.stopFoundAncestor = true
	syncer.start(t)

	assert.Equal(t, expAncestor, syncer.ctx.CommonAncestor.BlockNo(), "wrong ancestor")
}

func TestFinder_fullscan_found(t *testing.T) {
	//case 1: 0 is ancestor
	for i := 0; i < 10; i++ {
		testFullscanSucceed(t, uint64(i))
	}
}

func TestFinder_fullscan_notfound(t *testing.T) {
	//case 1: 0 is ancestor
	remoteChain := initStubBlockChain(nil, remoteChainLen)
	localChain := initStubBlockChain(nil, localChainLen)

	ctx := types.NewSyncCtx(targetPeerID, targetNo, uint64(localChain.best))

	syncer := NewStubSyncer(ctx, true, false, false, localChain, remoteChain)

	assert.Equal(t, targetPeerID, peer.ID(syncer.stubPeers[0].addr.GetPeerID()), "target peerid=peer-0")

	syncer.finder.setFullScanOnly(10)

	syncer.finder.start()

	//syncer stop after finding ancestor
	syncer.stopFoundAncestor = true
	syncer.start(t)

	assert.Equal(t, (*types.Block)(nil), syncer.ctx.CommonAncestor, "not nil ancestor")
}
