package chain

import (
	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	testCfg *config.Config
)

type StubConsensus struct{}

func (stubC *StubConsensus) SetStateDB(sdb *state.ChainStateDB) {

}
func (stubC *StubConsensus) IsTransactionValid(tx *types.Tx) bool {
	return true
}
func (stubC *StubConsensus) IsBlockValid(block *types.Block, bestBlock *types.Block) error {
	return nil
}
func (stubC *StubConsensus) Update(block *types.Block) {

}
func (stubC *StubConsensus) Save(tx db.Transaction) error {
	return nil
}
func (stubC *StubConsensus) NeedReorganization(rootNo types.BlockNo) bool {
	return true
}

func makeBlockChain(best int) *ChainService {
	serverCtx := config.NewServerContext("", "")
	testCfg = serverCtx.GetDefaultConfig().(*config.Config)
	testCfg.DbType = "memorydb"

	//TODO drop data when close memorydb when test mode
	cs := NewChainService(testCfg)

	stubConsensus := &StubConsensus{}

	cs.SetChainConsensus(stubConsensus)

	logger.Debug().Str("chainsvc name", cs.BaseComponent.GetName()).Msg("test")

	return cs
}

// Test add block to height 0 chain
func TestAddBlock(t *testing.T) {
	best := 10
	cs := makeBlockChain(0)

	genesisBlk, _ := cs.getBlockByNo(0)
	stubChain := InitStubBlockChain([]*types.Block{genesisBlk}, best)

	for i := 1; i <= best; i++ {
		err := cs.addBlock(stubChain.GetBlockByNo(uint64(i)), nil, "testpeer1")
		if err != nil {
			assert.Error(t, err, "add block failed")
			return
		}

		//best block height
		blk, err := cs.GetBestBlock()
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.Equal(t, blk.BlockNo(), uint64(i))

		//block hash/no mapping
		var noblk *types.Block
		noblk, err = cs.getBlockByNo(uint64(i))
		assert.Equal(t, blk.BlockHash(), noblk.BlockHash())
	}
}
