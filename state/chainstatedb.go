package state

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
)

// ChainStateDB manages statedb and additional informations about blocks like a state root hash
type ChainStateDB struct {
	sync.RWMutex
	luaStore  db.DB
	luaStates *StateDB
	evmStore  ethstate.Database
	evmStates *ethstate.StateDB
	testmode  bool
}

// NewChainStateDB creates instance of ChainStateDB
func NewChainStateDB() *ChainStateDB {
	return &ChainStateDB{}
}

// Init initialize database and load statedb of latest block
func (sdb *ChainStateDB) Clone() *ChainStateDB {
	sdb.Lock()
	defer sdb.Unlock()

	newSdb := &ChainStateDB{
		luaStore:  sdb.luaStore,
		luaStates: sdb.GetLuaStateDB().Clone(),
	}
	return newSdb
}

// Init initialize database and load statedb of latest block
func (sdb *ChainStateDB) Init(dbType string, dataDir string, bestBlock *types.Block, test bool) error {
	sdb.Lock()
	defer sdb.Unlock()

	sdb.testmode = test
	// init db
	if sdb.luaStore == nil {
		dbPath := common.PathMkdirAll(dataDir, stateName)
		sdb.luaStore = db.NewDB(db.ImplType(dbType), dbPath)
	}

	// init trie
	if sdb.luaStates == nil {
		var sroot []byte
		if bestBlock != nil {
			sroot = bestBlock.GetHeader().GetBlocksRootHash()
		}

		sdb.luaStates = NewStateDB(sdb.luaStore, sroot, sdb.testmode)
	}
	return nil
}

// Close saves latest block information of the chain
func (sdb *ChainStateDB) Close() error {
	sdb.Lock()
	defer sdb.Unlock()

	// close db
	if sdb.luaStore != nil {
		sdb.luaStore.Close()
	}
	return nil
}

// GetLuaStateDB returns statedb stores account states
func (sdb *ChainStateDB) GetLuaStateDB() *StateDB {
	return sdb.luaStates
}

// GetStateDB returns statedb stores account states
func (sdb *ChainStateDB) GetEvmStateDB() *ethstate.StateDB {
	return sdb.evmStates
}

// GetSystemAccountState returns the state of the aergo system account.
func (sdb *ChainStateDB) GetSystemAccountState() (*ContractState, error) {
	return sdb.GetLuaStateDB().GetSystemAccountState()
}

// OpenLuaStateDB returns new instance of statedb given state root hash
func (sdb *ChainStateDB) OpenLuaStateDB(root []byte) *StateDB {
	return NewStateDB(sdb.luaStore, root, sdb.testmode)
}

func (sdb *ChainStateDB) OpenEvmStateDB(root []byte) *ethstate.StateDB {
	esdb, _ := ethstate.New(ethcommon.BytesToHash(root), sdb.evmStore, nil)
	return esdb
}

func (sdb *ChainStateDB) SetGenesis(genesis *types.Genesis, bpInit func(*StateDB, *types.Genesis) error) error {
	block := genesis.Block()
	luaStateDB := sdb.OpenLuaStateDB(sdb.GetLuaRoot())

	// create state of genesis block
	gbState := sdb.NewBlockState(luaStateDB.GetRoot(), nil)

	if len(genesis.BPs) > 0 && bpInit != nil {
		// To avoid cyclic dedendency, BP initilization is called via function
		// pointer.
		if err := bpInit(luaStateDB, genesis); err != nil {
			return err
		}

		aid := types.ToAccountID([]byte(types.AergoSystem))
		scs, err := luaStateDB.GetSystemAccountState()
		if err != nil {
			return err
		}

		if err := gbState.LuaStateDB.PutState(aid, scs.State); err != nil {
			return err
		}
	}

	for address, balance := range genesis.Balance {
		bytes := types.ToAddress(address)
		id := types.ToAccountID(bytes)
		if v, ok := new(big.Int).SetString(balance, 10); ok {
			if err := gbState.LuaStateDB.PutState(id, &types.State{Balance: v.Bytes()}); err != nil {
				return err
			}
			genesis.AddBalance(v)
		} else {
			return fmt.Errorf("balance conversion failed for %s (address: %s)", balance, address)
		}
	}

	// save state of genesis block
	// FIXME don't use chainstate API
	if err := sdb.Apply(gbState); err != nil {
		return err
	}

	block.SetBlocksRootHash(sdb.GetLuaRoot())

	return nil
}

// Apply specific blockstate to statedb of main chain
func (sdb *ChainStateDB) Apply(bstate *BlockState) error {
	sdb.Lock()
	defer sdb.Unlock()

	// // rollback and revert trie requires state root before apply
	// if bstate.Undo.StateRoot == emptyHashID {
	// 	bstate.Undo.StateRoot = types.ToHashID(sdb.states.trie.Root)
	// }

	// apply blockState to trie
	if err := bstate.Update(); err != nil {
		return err
	}
	if err := bstate.Commit(); err != nil {
		return err
	}

	if err := sdb.UpdateRoot(bstate); err != nil {
		return err
	}

	return nil
}

func (sdb *ChainStateDB) UpdateRoot(bstate *BlockState) error {
	// // check state root
	// if bstate.BlockInfo.StateRoot != types.ToHashID(bstate.GetRoot()) {
	// 	// TODO: if validation failed, than revert statedb.
	// 	bstate.BlockInfo.StateRoot = types.ToHashID(sdb.GetRoot())
	// }

	logger.Debug().Str("before", base58.Encode(sdb.luaStates.GetRoot())).
		Str("luaStateRoot", base58.Encode(bstate.GetLuaRoot())).Msg("apply block state")

	if err := sdb.luaStates.SetRoot(bstate.GetLuaRoot()); err != nil {
		return err
	}

	// no need to update evm root

	return nil
}

func (sdb *ChainStateDB) SetLuaRoot(targetBlockRoot []byte) error {
	sdb.Lock()
	defer sdb.Unlock()

	logger.Debug().Str("before", base58.Encode(sdb.luaStates.GetRoot())).
		Str("target", base58.Encode(targetBlockRoot)).Msg("rollback state")

	sdb.luaStates.SetRoot(targetBlockRoot)
	return nil
}

// GetLuaRoot returns state root hash
func (sdb *ChainStateDB) GetLuaRoot() []byte {
	return sdb.luaStates.GetRoot()
}

// GetRoot returns state root hash
func (sdb *ChainStateDB) GetEvmRoot() []byte {
	return sdb.evmStates.IntermediateRoot(false).Bytes()
}

func (sdb *ChainStateDB) IsExistState(hash []byte) bool {
	//TODO : StateRootValidation
	return false
}

func (sdb *ChainStateDB) NewBlockState(luaRoot []byte, evmRoot []byte, options ...BlockStateOptFn) *BlockState {
	var ls *StateDB
	if luaRoot != nil {
		ls = sdb.OpenLuaStateDB(luaRoot)
	}
	var es *ethstate.StateDB
	if evmRoot != nil {
		es = sdb.OpenEvmStateDB(evmRoot)
	}

	return NewBlockState(ls, es, options...)
}
