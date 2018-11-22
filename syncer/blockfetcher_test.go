package syncer

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func testHandleAddBlock(t *testing.T, syncer *StubSyncer, blockNo uint64) {
	//AddBlock
	msg := syncer.testhub.recvMessage()
	assert.IsTypef(t, &message.AddBlock{}, msg, "add block")
	assert.Equal(t, blockNo, msg.(*message.AddBlock).Block.GetHeader().BlockNo, "check add blockno")

	//AddBlockRsp
	syncer.handleMessageManual(t, msg, nil)
}

// test blockfetcher
// sync target : 1 ~ 5
// GetBlockChunk : 1 ~ 2 (peer0), 3 (peer1), 4 ~ 5 (peer0)
func TestBlockFetcher_normal(t *testing.T) {
	//init test ----------------------------------
	testTargetNo := uint64(5)

	//make remoteBlockChain
	remoteChain := initStubBlockChain(nil, 10)

	ancestor := remoteChain.GetBlockByNo(0)

	ctx := types.NewSyncCtx("p1", testTargetNo, 1)
	ctx.SetAncestor(ancestor)

	syncer := NewStubSyncer(ctx, false, false, true, nil, remoteChain)
	syncer.bf.Start()

	bf := syncer.bf

	//start test ----------------------------------
	//register peers
	msg := syncer.testhub.recvMessage()
	assert.IsTypef(t, &message.GetPeers{}, msg, "get peers from BF")
	//reply peers
	syncer.handleMessageManual(t, msg, nil)

	testHashSet := func(prev *types.BlockInfo, count uint64) {
		//push hashSet
		hashes, _ := syncer.remoteChain.GetHashes(prev, count)

		bf.hfCh <- &HashSet{len(hashes), hashes, prev.No + 1}

		taskCount := int(math.Ceil(float64(count) / float64(TestMaxFetchSize)))
		msgs := make([]interface{}, 0)
		for i := 0; i < taskCount; i++ {
			//loop until consume all hashSet (3 block)
			msg = syncer.testhub.recvMessage()
			assert.IsTypef(t, &message.GetBlockChunks{}, msg, "get block chunks")
			assert.True(t, TestMaxFetchSize >= len(msg.(*message.GetBlockChunks).Hashes), "get block chunks")

			msgs = append(msgs, msg)
		}

		for i := 0; i < taskCount; i++ {
			syncer.handleMessageManual(t, msgs[i], nil)
		}

		//AddBlockReq - must run #len(hashes) times
		for i := 0; i < len(hashes); i++ {
			testHandleAddBlock(t, syncer, prev.No+uint64(i)+1)
		}

		//check last == prev
		assert.Equal(t, prev.No+uint64(len(hashes)), syncer.bf.stat.getMaxChunkRsp().GetHeader().BlockNo, "max block chunk response")
	}

	//1~3 : 1~2 / 3
	testHashSet(&types.BlockInfo{Hash: ancestor.GetHash(), No: ancestor.BlockNo()}, 3)

	prevInfo := &types.BlockInfo{Hash: syncer.bf.stat.getMaxChunkRsp().GetHash(),
		No: syncer.bf.stat.getMaxChunkRsp().GetHeader().BlockNo}

	//4~5 : 4~5 end
	testHashSet(prevInfo, 2)

	//stop
	msg = syncer.testhub.recvMessage()
	assert.IsTypef(t, &message.SyncStop{}, msg, "need syncer stop msg")
	syncer.handleMessageManual(t, msg, nil)
}
