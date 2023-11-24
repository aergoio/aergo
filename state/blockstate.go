package state

import (
	"math/big"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/types"
	"github.com/bluele/gcache"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/willf/bloom"
)

// BlockState contains BlockInfo and statedb for block
type BlockState struct {
	LuaStateDB *StateDB
	EvmStateDB *ethstate.StateDB

	BpReward      big.Int // final bp reward, increment when tx executes
	receipts      types.Receipts
	CCProposal    *consensus.ConfChangePropose
	prevBlockHash []byte
	consensus     []byte // Consensus Header
	GasPrice      *big.Int

	timeoutTx types.Transaction
	codeCache gcache.Cache
	abiCache  gcache.Cache
}

type BlockStateOptFn func(s *BlockState)

func SetPrevBlockHash(h []byte) BlockStateOptFn {
	return func(s *BlockState) {
		s.SetPrevBlockHash(h)
	}
}

func SetGasPrice(gasPrice *big.Int) BlockStateOptFn {
	return func(s *BlockState) {
		s.SetGasPrice(gasPrice)
	}
}

// NewBlockState create new blockState contains blockInfo, account states and undo states
func NewBlockState(luaStates *StateDB, evmStates *ethstate.StateDB, options ...BlockStateOptFn) *BlockState {
	b := &BlockState{
		codeCache: gcache.New(100).LRU().Build(),
		abiCache:  gcache.New(100).LRU().Build(),
	}
	if luaStates != nil {
		b.LuaStateDB = luaStates.Clone()
	}
	if evmStates != nil {
		b.EvmStateDB = evmStates.Copy()
	}
	for _, opt := range options {
		opt(b)
	}
	return b
}

type BlockSnapshot struct {
	LuaVersion Snapshot
	luaStorage map[types.AccountID]int
	EvmVersion int
}

func (bs *BlockState) Snapshot() BlockSnapshot {
	result := BlockSnapshot{
		LuaVersion: bs.LuaStateDB.Snapshot(),
		luaStorage: bs.LuaStateDB.cache.snapshot(),
	}
	if bs.EvmStateDB != nil {
		result.EvmVersion = bs.EvmStateDB.Snapshot()
	}
	return result
}

func (bs *BlockState) Rollback(bSnap BlockSnapshot) error {
	if err := bs.LuaStateDB.cache.rollback(bSnap.luaStorage); err != nil {
		return err
	}
	if err := bs.LuaStateDB.Rollback(bSnap.LuaVersion); err != nil {
		return err
	}
	if bs.EvmStateDB != nil {
		bs.EvmStateDB.RevertToSnapshot(bSnap.EvmVersion)
	}

	return nil
}

func (bs *BlockState) GetLuaRoot() []byte {
	return bs.LuaStateDB.GetRoot()
}

func (bs *BlockState) SetLuaRoot(root []byte) {
	bs.LuaStateDB.SetRoot(root)
}

func (bs *BlockState) GetEvmRoot() []byte {
	if bs.EvmStateDB == nil {
		return nil
	}
	return bs.EvmStateDB.IntermediateRoot(false).Bytes()
}

func (bs *BlockState) Update() error {
	if bs.LuaStateDB != nil {
		err := bs.LuaStateDB.Update()
		if err != nil {
			return err
		}
	}
	if bs.EvmStateDB != nil {
		bs.EvmStateDB.Finalise(true)
	}
	return nil
}

func (bs *BlockState) Commit() error {
	if bs.LuaStateDB != nil {
		err := bs.LuaStateDB.Commit()
		if err != nil {
			return err
		}
	}
	if bs.EvmStateDB != nil {
		_, err := bs.EvmStateDB.Commit(true)
		if err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------------------------------------//
//

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

func (bs *BlockState) SetPrevBlockHash(prevHash []byte) *BlockState {
	if bs != nil {
		bs.prevBlockHash = prevHash
	}
	return bs
}

func (bs *BlockState) SetGasPrice(gasPrice *big.Int) *BlockState {
	if bs != nil {
		bs.GasPrice = gasPrice
	}
	return bs
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

func (bs *BlockState) PrevBlockHash() []byte {
	return bs.prevBlockHash
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
