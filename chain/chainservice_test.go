package chain

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	testCfg *config.Config
)

const (
	testPeer = "testpeer1"
)

type StubConsensus struct{}

func (stubC *StubConsensus) SetStateDB(sdb *state.ChainStateDB) {

}
func (stubC *StubConsensus) IsTransactionValid(tx *types.Tx) bool {
	return true
}
func (stubC *StubConsensus) VerifyTimestamp(block *types.Block) bool {
	return true
}
func (stubC *StubConsensus) VerifySign(block *types.Block) error {
	return nil
}
func (stubC *StubConsensus) IsBlockValid(block *types.Block, bestBlock *types.Block) error {
	return nil
}
func (stubC *StubConsensus) Update(block *types.Block) {

}
func (stubC *StubConsensus) Save(tx consensus.TxWriter) error {
	return nil
}
func (stubC *StubConsensus) NeedReorganization(rootNo types.BlockNo) bool {
	return true
}
func (stubC *StubConsensus) Info() string {
	return ""
}
func (stubC *StubConsensus) GetType() consensus.ConsensusType {
	return consensus.ConsensusSBP
}
func (stubC *StubConsensus) NeedNotify() bool {
	return true
}
func (stubC *StubConsensus) HasWAL() bool {
	return false
}
func (stubC *StubConsensus) IsConnectedBlock(block *types.Block) bool {
	return false
}
func (stubC *StubConsensus) IsForkEnable() bool {
	return true
}

func (stubC *StubConsensus) MakeConfChangeProposal(req *types.MembershipChange) (*consensus.ConfChangePropose, error) {
	return nil, consensus.ErrNotSupportedMethod
}

func makeBlockChain() *ChainService {
	serverCtx := config.NewServerContext("", "")
	testCfg = serverCtx.GetDefaultConfig().(*config.Config)
	testCfg.DbType = "memorydb"
	//TODO use testnet genesis for test for now
	testCfg.UseTestnet = true

	//TODO drop data when close memorydb when test mode
	cs := NewChainService(testCfg)

	stubConsensus := &StubConsensus{}

	cs.SetChainConsensus(stubConsensus)

	logger.Debug().Str("chainsvc name", cs.BaseComponent.GetName()).Msg("test")

	return cs
}

func testAddBlock(t *testing.T, best int) (*ChainService, *StubBlockChain) {
	cs := makeBlockChain()

	genesisBlk, _ := cs.getBlockByNo(0)

	assert.NotNil(t, genesisBlk)

	stubChain := InitStubBlockChain([]*types.Block{genesisBlk}, best)

	for i := 1; i <= best; i++ {
		newBlock := stubChain.GetBlockByNo(uint64(i))
		err := cs.addBlock(newBlock, nil, testPeer)
		assert.NoError(t, err)

		testBlockIsOnMasterChain(t, cs, newBlock)

		//best block height
		blk, err := cs.GetBestBlock()
		assert.NoError(t, err)
		assert.Equal(t, blk.BlockNo(), uint64(i))

		//block hash/no mapping
		var noblk *types.Block
		noblk, err = cs.getBlockByNo(uint64(i))
		assert.NoError(t, err)
		assert.Equal(t, blk.BlockHash(), noblk.BlockHash())
	}

	return cs, stubChain
}

// Test add block to height 0 chain
func testAddBlockNoTest(best int) (*ChainService, *StubBlockChain) {
	cs := makeBlockChain()

	genesisBlk, _ := cs.getBlockByNo(0)

	stubChain := InitStubBlockChain([]*types.Block{genesisBlk}, best)

	for i := 1; i <= best; i++ {
		newBlock := stubChain.GetBlockByNo(uint64(i))
		_ = cs.addBlock(newBlock, nil, testPeer)
	}

	return cs, stubChain
}

func TestAddBlock(t *testing.T) {
	testAddBlock(t, 1)
	testAddBlock(t, 10)
	testAddBlock(t, 100)
}

// test if block exist on sideChain
func testBlockIsOnMasterChain(t *testing.T, cs *ChainService, block *types.Block) {
	//check if block added in DB
	chainBlock, err := cs.GetBlock(block.BlockHash())
	assert.NoError(t, err)
	assert.Equal(t, chainBlock.GetHeader().BlockNo, block.GetHeader().BlockNo)

	//check if block added on master chain
	chainBlock, err = cs.getBlockByNo(block.GetHeader().BlockNo)
	assert.NoError(t, err)
	assert.Equal(t, chainBlock.BlockHash(), block.BlockHash())
}

// test if block exist on sideChain
func testBlockIsOnSideChain(t *testing.T, cs *ChainService, block *types.Block) {
	//check if block added in DB
	chainBlock, err := cs.GetBlock(block.BlockHash())
	assert.NoError(t, err)
	assert.Equal(t, chainBlock.GetHeader().BlockNo, block.GetHeader().BlockNo)

	//check if block added on side chain
	chainBlock, err = cs.getBlockByNo(block.GetHeader().BlockNo)
	assert.NoError(t, err)
	assert.NotEqual(t, chainBlock.BlockHash(), block.BlockHash(), fmt.Sprintf("no=%d", block.GetHeader().BlockNo))
}

func testBlockIsOrphan(t *testing.T, cs *ChainService, block *types.Block) {
	//check if block added in DB
	_, err := cs.GetBlock(block.BlockHash())
	assert.Equal(t, &ErrNoBlock{id: block.BlockHash()}, err)

	//check if block exist on orphan pool
	orphan := cs.op.getOrphan(block.Header.GetPrevBlockHash())
	assert.Equal(t, orphan.BlockHash(), block.BlockHash())
}

// test to add blocks to sidechain until best of sideChain is equal to the mainChain
func testSideBranch(t *testing.T, mainChainBest int) (cs *ChainService, mainChain *StubBlockChain, sideChain *StubBlockChain) {
	cs, mainChain = testAddBlock(t, mainChainBest)

	//common ancestor of master chain and side chain is 0
	sideChain = InitStubBlockChain(mainChain.Blocks[0:1], mainChainBest)

	//add sideChainBlock
	for _, block := range sideChain.Blocks[1 : sideChain.Best+1] {
		err := cs.addBlock(block, nil, testPeer)
		assert.NoError(t, err)

		//block added on sidechain
		testBlockIsOnSideChain(t, cs, block)
	}

	assert.Equal(t, mainChain.Best, sideChain.Best)

	return cs, mainChain, sideChain
}

func TestSideBranch(t *testing.T) {
	testSideBranch(t, 5)
}

func TestOrphan(t *testing.T) {
	mainChainBest := 5
	cs, mainChain := testAddBlock(t, mainChainBest)

	//make branch
	sideChain := InitStubBlockChain(mainChain.Blocks[0:1], mainChainBest)

	//add orphan
	for _, block := range sideChain.Blocks[2 : sideChain.Best+1] {
		err := cs.addBlock(block, nil, testPeer)
		assert.NoError(t, err)

		//block added on sidechain
		testBlockIsOrphan(t, cs, block)
	}
}

func TestSideChainReorg(t *testing.T) {
	cs, mainChain, sideChain := testSideBranch(t, 5)

	// add heigher block to sideChain
	sideChain.GenAddBlock()
	assert.Equal(t, mainChain.Best+1, sideChain.Best)

	sideBestBlock, err := sideChain.GetBestBlock()
	assert.NoError(t, err)

	//check top block before reorg
	mainBestBlock, _ := cs.GetBestBlock()
	assert.Equal(t, mainChain.Best, int(mainBestBlock.GetHeader().BlockNo))
	assert.Equal(t, mainChain.BestBlock.BlockHash(), mainBestBlock.BlockHash())
	assert.Equal(t, mainBestBlock.GetHeader().BlockNo+1, sideBestBlock.GetHeader().BlockNo)

	err = cs.addBlock(sideBestBlock, nil, testPeer)
	assert.NoError(t, err)

	//check if reorg is succeed
	mainBestBlock, _ = cs.GetBestBlock()
	assert.Equal(t, sideBestBlock.GetHeader().BlockNo, mainBestBlock.GetHeader().BlockNo)

	assert.Equal(t, sideBestBlock.BlockHash(), mainBestBlock.BlockHash())
}

func TestAddErroredBlock(t *testing.T) {
	// make chain
	cs, stubChain := testAddBlock(t, 10)

	// add block which occur validation error
	stubChain.GenAddBlock()

	newBlock, _ := stubChain.GetBestBlock()
	newBlock.SetBlocksRootHash([]byte("xxx"))

	err := cs.addBlock(newBlock, nil, testPeer)
	assert.Equal(t, ErrorBlockVerifyStateRoot, err)

	err = cs.addBlock(newBlock, nil, testPeer)
	assert.Equal(t, ErrBlockCachedErrLRU, err)

	cs.errBlocks.Purge()
	// check error when server is rebooted
	err = cs.addBlock(newBlock, nil, testPeer)
	assert.Equal(t, ErrorBlockVerifyStateRoot, err)
}

func TestResetChain(t *testing.T) {
	mainChainBest := 5
	cs, mainChain := testAddBlock(t, mainChainBest)

	resetHeight := uint64(3)
	err := cs.cdb.ResetBest(resetHeight)
	assert.NoError(t, err)

	// check best
	assert.Equal(t, resetHeight, cs.cdb.getBestBlockNo())

	for i := uint64(mainChainBest); i > resetHeight; i-- {
		err := cs.cdb.checkBlockDropped(mainChain.GetBlockByNo(i))
		assert.NoError(t, err)
	}
}

func TestReorgCrashRecoverBeforeReorgMarker(t *testing.T) {
	cs, mainChain, sideChain := testSideBranch(t, 5)

	// add heigher block to sideChain
	sideChain.GenAddBlock()
	assert.Equal(t, mainChain.Best+1, sideChain.Best)

	sideBestBlock, err := sideChain.GetBestBlock()
	assert.NoError(t, err)

	//check top block before reorg
	orgBestBlock, _ := cs.GetBestBlock()
	assert.Equal(t, mainChain.Best, int(orgBestBlock.GetHeader().BlockNo))
	assert.Equal(t, mainChain.BestBlock.BlockHash(), orgBestBlock.BlockHash())
	assert.Equal(t, orgBestBlock.GetHeader().BlockNo+1, sideBestBlock.GetHeader().BlockNo)

	TestDebugger = newDebugger()
	TestDebugger.Set(DEBUG_CHAIN_STOP, 1, false)

	err = cs.addBlock(sideBestBlock, nil, testPeer)
	assert.Error(t, &ErrReorg{})
	assert.Equal(t, err.(*ErrReorg).err, &ErrDebug{cond: DEBUG_CHAIN_STOP, value: 1})

	// check if chain meta is not changed
	newBestBlock, _ := cs.GetBestBlock()
	assert.Equal(t, newBestBlock.GetHeader().BlockNo, orgBestBlock.GetHeader().BlockNo)

	TestDebugger.clear()
	cs.errBlocks.Purge()

	// chain swap is not complete, so has nothing to do
	err = cs.Recover()
	assert.Nil(t, err)
}

func TestReorgCrashRecoverAfterReorgMarker(t *testing.T) {
	testReorgCrashRecoverCond(t, DEBUG_CHAIN_STOP, 2)
	testReorgCrashRecoverCond(t, DEBUG_CHAIN_STOP, 3)
}

func testReorgCrashRecoverCond(t *testing.T, cond StopCond, value int) {
	cs, mainChain, sideChain := testSideBranch(t, 5)

	// add heigher block to sideChain
	sideChain.GenAddBlock()
	assert.Equal(t, mainChain.Best+1, sideChain.Best)

	sideBestBlock, err := sideChain.GetBestBlock()
	assert.NoError(t, err)

	//check top block before reorg
	orgBestBlock, _ := cs.GetBestBlock()
	assert.Equal(t, mainChain.Best, int(orgBestBlock.GetHeader().BlockNo))
	assert.Equal(t, mainChain.BestBlock.BlockHash(), orgBestBlock.BlockHash())
	assert.Equal(t, orgBestBlock.GetHeader().BlockNo+1, sideBestBlock.GetHeader().BlockNo)

	TestDebugger = newDebugger()
	TestDebugger.Set(cond, value, false)

	err = cs.addBlock(sideBestBlock, nil, testPeer)
	assert.Error(t, &ErrReorg{})
	assert.Equal(t, err.(*ErrReorg).err, &ErrDebug{cond: cond, value: value})

	assert.True(t, !checkRecoveryDone(t, cs.cdb, sideChain))

	TestDebugger.clear()
	cs.errBlocks.Purge()

	// must recover chainDB before chainservice.Recover()
	err = cs.cdb.recover()
	assert.Nil(t, err)

	// chain swap is not complete, so has nothing to do
	err = cs.Recover()
	assert.Nil(t, err)

	assert.True(t, checkRecoveryDone(t, cs.cdb, sideChain))

	var marker *ReorgMarker
	marker, err = cs.cdb.getReorgMarker()
	assert.Nil(t, err)
	assert.Nil(t, marker)
}

// checkRecoveryDone checks if recovery is complete.
// 1. all blocks of chain has (no/hash) mapping
// 2. old receipts is deleted and new receipt is added if blocks have tx
// 3. old tx mapping is deleted and new tx mapping is added
func checkRecoveryDone(t *testing.T, cdb *ChainDB, chain *StubBlockChain) bool {
	// check block mapping
	for i := 0; i <= chain.Best; i++ {
		block := chain.Blocks[i]
		dbBlk, err := cdb.GetBlockByNo(block.GetHeader().GetBlockNo())
		assert.Nil(t, err)
		assert.NotNil(t, dbBlk)

		if !checkBlockEqual(t, block, dbBlk) {
			return false
		}
	}

	if marker, err := cdb.getReorgMarker(); err == nil && marker != nil {
		return false
	}

	return true
}

func checkBlockEqual(t *testing.T, x *types.Block, y *types.Block) bool {
	if (x == nil) != (y == nil) {
		t.Log("x or y is nil")
		return false
	}

	if !bytes.Equal(x.BlockHash(), y.BlockHash()) || x.Header.GetBlockNo() != y.Header.GetBlockNo() {
		t.Logf("stubchain<no=%d, %s>:db<no=%d, %s>", x.GetHeader().GetBlockNo(), x.ID(), y.GetHeader().GetBlockNo(), y.ID())
		return false
	}

	return true
}
