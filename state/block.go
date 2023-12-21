package state

import (
	"math/big"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/bluele/gcache"
	"github.com/willf/bloom"
)

// BlockState contains BlockInfo and statedb for block
type BlockState struct {
	LuaStateDB *statedb.StateDB
	EthStateDB *ethdb.StateDB

	block      *types.BlockHeader
	BpReward   big.Int // final bp reward, increment when tx executes
	receipts   types.Receipts
	CCProposal *consensus.ConfChangePropose
	consensus  []byte // Consensus Header
	gasPrice   *big.Int

	timeoutTx types.Transaction
	codeCache gcache.Cache
	abiCache  gcache.Cache
}

type BlockStateOptFn func(s *BlockState)

func SetBlock(block *types.BlockHeader) BlockStateOptFn {
	return func(s *BlockState) {
		s.block = block
	}
}

func SetGasPrice(gasPrice *big.Int) BlockStateOptFn {
	return func(s *BlockState) {
		s.SetGasPrice(gasPrice)
	}
}

// NewBlockState create new blockState contains blockInfo, account states and undo states
func NewBlockState(luaStates *statedb.StateDB, ethStates *ethdb.StateDB, options ...BlockStateOptFn) *BlockState {
	b := &BlockState{
		LuaStateDB: luaStates,
		EthStateDB: ethStates,
		codeCache:  gcache.New(100).LRU().Build(),
		abiCache:   gcache.New(100).LRU().Build(),
	}
	for _, opt := range options {
		opt(b)
	}
	return b
}

type BlockSnapshot struct {
	LuaVersion statedb.Snapshot
	luaStorage map[types.AccountID]int
	EthVersion int
}

func (bs *BlockState) Snapshot() *BlockSnapshot {
	result := &BlockSnapshot{
		LuaVersion: bs.LuaStateDB.Snapshot(),
		luaStorage: bs.LuaStateDB.Cache.Snapshot(),
	}
	if bs.EthStateDB != nil {
		result.EthVersion = bs.EthStateDB.Snapshot()
	}
	return result
}

func (bs *BlockState) Rollback(bSnap *BlockSnapshot) error {
	if err := bs.LuaStateDB.Cache.Rollback(bSnap.luaStorage); err != nil {
		return err
	}
	if err := bs.LuaStateDB.Rollback(bSnap.LuaVersion); err != nil {
		return err
	}
	if bs.EthStateDB != nil {
		bs.EthStateDB.Rollback(bSnap.EthVersion)
	}

	return nil
}

func (bs *BlockState) GetLuaRoot() []byte {
	return bs.LuaStateDB.GetRoot()
}

func (bs *BlockState) GetEthRoot() []byte {
	if bs.EthStateDB == nil {
		return nil
	}
	return bs.EthStateDB.Root()
}

func (bs *BlockState) Update() error {
	err := bs.LuaStateDB.Update()
	if err != nil {
		return err
	}
	return nil
}

func (bs *BlockState) Commit() error {
	if bs.LuaStateDB != nil {
		if err := bs.LuaStateDB.Commit(); err != nil {
			return err
		}
	}
	if bs.EthStateDB != nil {
		var blockNo uint64
		if bs.block != nil {
			blockNo = bs.block.BlockNo
		}
		_, err := bs.EthStateDB.Commit(blockNo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockState) Consensus() []byte {
	return bs.consensus
}

func (bs *BlockState) SetConsensus(ch []byte) {
	bs.consensus = ch
}

func (bs *BlockState) AddReceipt(r *types.Receipt) error {
	if len(r.Events) > 0 {
		rBloom := bloom.New(types.BloomBitBits, types.BloomHashKNum)
		for _, e := range r.Events {
			rBloom.Add(e.ContractAddress)
			rBloom.Add([]byte(e.EventName))
		}
		binary, _ := rBloom.GobEncode()
		r.Bloom = binary[24:]
		err := bs.receipts.MergeBloom(rBloom)
		if err != nil {
			return err
		}
	}
	bs.receipts.Set(append(bs.receipts.Get(), r))
	return nil
}

func (bs *BlockState) Receipts() *types.Receipts {
	if bs == nil {
		return nil
	}
	return &bs.receipts
}

func (bs *BlockState) SetBlock(block *types.BlockHeader) {
	if bs != nil {
		bs.block = block
	}
}

func (bs *BlockState) Block() *types.BlockHeader {
	return bs.block
}

func (bs *BlockState) SetGasPrice(gasPrice *big.Int) *BlockState {
	if bs != nil {
		bs.gasPrice = gasPrice
	}
	return bs
}

func (bs *BlockState) GasPrice() *big.Int {
	if bs == nil {
		return nil
	}
	return bs.gasPrice
}

func (bs *BlockState) TimeoutTx() types.Transaction {
	if bs == nil {
		return nil
	}
	return bs.timeoutTx
}

func (bs *BlockState) SetTimeoutTx(tx types.Transaction) {
	bs.timeoutTx = tx
}

func (bs *BlockState) GetCode(key types.AccountID) []byte {
	if bs == nil {
		return nil
	}
	code, err := bs.codeCache.Get(key)
	if err != nil {
		return nil
	}
	return code.([]byte)
}

func (bs *BlockState) AddCode(key types.AccountID, code []byte) {
	if bs == nil {
		return
	}
	bs.codeCache.Set(key, code)
}

func (bs *BlockState) GetABI(key types.AccountID) *types.ABI {
	if bs == nil {
		return nil
	}
	abi, err := bs.abiCache.Get(key)
	if err != nil {
		return nil
	}
	return abi.(*types.ABI)
}

func (bs *BlockState) AddABI(key types.AccountID, abi *types.ABI) {
	if bs == nil {
		return
	}
	bs.abiCache.Set(key, abi)
}

func (bs *BlockState) RemoveCache(key types.AccountID) {
	if bs == nil {
		return
	}
	bs.codeCache.Remove(key)
	bs.abiCache.Remove(key)
}
