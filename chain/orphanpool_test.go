package chain

import (
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func checkExist(t *testing.T, orp *OrphanPool, blk *types.Block) {
	var orphan *types.Block
	orphan = orp.getOrphan(blk.Header.GetPrevBlockHash())

	assert.NotNil(t, orphan)
	assert.Equal(t, orphan.BlockHash(), blk.BlockHash())
}

func TestOrphanPool(t *testing.T) {
	// measure time to add block
	var stubChain *StubBlockChain
	var orp *OrphanPool
	var orphan *types.Block

	orp = NewOrphanPool(5)

	_, stubChain = testAddBlockNoTest(10)

	start := 1
	for i := start; i <= 5; i++ {
		blk := stubChain.GetBlockByNo(uint64(i))

		err := orp.addOrphan(blk)
		assert.NoError(t, err)

		checkExist(t, orp, blk)
	}

	// check pool is full
	assert.True(t, orp.isFull())

	// remove oldest and put new one
	blk := stubChain.GetBlockByNo(6)
	err := orp.addOrphan(blk)
	assert.NoError(t, err)

	checkExist(t, orp, blk)

	// first block is removed
	startBlock := stubChain.GetBlockByNo(uint64(start))
	orphan = orp.getOrphan(startBlock.Header.GetPrevBlockHash())
	assert.Nil(t, orphan)

	assert.True(t, orp.isFull())
}

func TestOrphanSamePrev(t *testing.T) {
	var orp *OrphanPool

	mainChainBest := 3
	_, mainChain := testAddBlock(t, mainChainBest)

	// make branch
	sideChain := InitStubBlockChain(mainChain.Blocks[0:mainChainBest+1], 1)

	// make fork
	mainChain.GenAddBlock()

	mBest := mainChain.BestBlock
	sBest := sideChain.BestBlock

	assert.Equal(t, mBest.PrevBlockID(), sBest.PrevBlockID())

	// No.4 blocks of mainchain and sidechain have same previous hash
	orp = NewOrphanPool(5)

	err := orp.addOrphan(mBest)
	assert.NoError(t, err)

	err = orp.addOrphan(sBest)
	assert.NoError(t, err)

	checkExist(t, orp, mBest)

	orphan := orp.getOrphan(sBest.Header.GetPrevBlockHash())
	assert.Equal(t, orphan.BlockHash(), mBest.BlockHash())
}

func BenchmarkOrphanPoolWhenPool(b *testing.B) {
	b.ResetTimer()

	// measure time to add block
	start := 1001

	var stubChain *StubBlockChain
	var orp *OrphanPool

	b.StopTimer()

	orp = NewOrphanPool(300)
	_, stubChain = testAddBlockNoTest(11000)
	// make pool to be full
	for i := 1; i <= 1000; i++ {
		blk := stubChain.GetBlockByNo(uint64(i))

		orp.addOrphan(blk)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {

		idx := start + (i % 10000)

		blk := stubChain.GetBlockByNo(uint64(idx))

		orp.addOrphan(blk)
	}
}
