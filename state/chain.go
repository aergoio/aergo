package state

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
)

var (
	logger = log.NewLogger(statedb.StateName)
)

// ChainStateDB manages statedb and additional informations about blocks like a state root hash
type ChainStateDB struct {
	sync.RWMutex
	luaStore  db.DB
	luaStates *statedb.StateDB
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
		luaStore: sdb.luaStore,
		evmStore: sdb.evmStore,
	}
	if sdb.luaStates != nil {
		newSdb.luaStates = sdb.luaStates.Clone()
	}
	if sdb.evmStates != nil {
		newSdb.evmStates = sdb.evmStates.Copy()
	}
	return newSdb
}

// Init initialize database and load statedb of latest block
func (sdb *ChainStateDB) Init(dbType string, dataDir string, bestBlock *types.Block, test bool) error {
	sdb.Lock()
	defer sdb.Unlock()
	var err error

	sdb.testmode = test
	// init lua db
	if sdb.luaStore == nil {
		dbPath := common.PathMkdirAll(dataDir, statedb.StateName)
		sdb.luaStore = db.NewDB(db.ImplType(dbType), dbPath)
	}

	// init trie
	if sdb.luaStates == nil {
		var sroot []byte
		if bestBlock != nil {
			sroot = bestBlock.GetHeader().GetBlocksRootHash()
		}

		sdb.luaStates = statedb.NewStateDB(sdb.luaStore, sroot, sdb.testmode)
	}

	if sdb.evmStore == nil {
		dbPath := common.PathMkdirAll(dataDir, "state_evm")

		var testDB ethdb.Database
		if db.ImplType(dbType) == db.MemoryImpl {
			testDB = rawdb.NewMemoryDatabase()
		} else {
			testDB, err = rawdb.NewLevelDBDatabase(dbPath, 128, 1024, "", false)
			if err != nil {
				return err
			}
		}
		sdb.evmStore = ethstate.NewDatabase(testDB)
	}

	if sdb.evmStates == nil {
		var sroot ethcommon.Hash
		if bestBlock != nil {
			sroot = ethcommon.BytesToHash(bestBlock.GetHeader().GetEvmRootHash())
		}
		sdb.evmStates, err = ethstate.New(sroot, sdb.evmStore, nil)
		if err != nil {
			return err
		}
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
	if sdb.evmStore != nil {
		sdb.evmStore.DiskDB().Close()
	}
	return nil
}

// GetStateDB returns statedb stores account states
func (sdb *ChainStateDB) GetStateDB() *statedb.StateDB {
	return sdb.luaStates
}

func (sdb *ChainStateDB) GetEvmStateDB() *ethstate.StateDB {
	return sdb.evmStates
}

// GetSystemAccountState returns the state of the aergo system account.
func (sdb *ChainStateDB) GetSystemAccountState() (*ContractState, error) {
	return GetSystemAccountState(sdb.GetStateDB())
}

// OpenNewStateDB returns new instance of statedb given state root hash
func (sdb *ChainStateDB) OpenNewStateDB(root []byte) *statedb.StateDB {
	return statedb.NewStateDB(sdb.luaStore, root, sdb.testmode)
}

func (sdb *ChainStateDB) OpenEvmStateDB(root []byte) *ethstate.StateDB {
	esdb, _ := ethstate.New(ethcommon.BytesToHash(root), sdb.evmStore, nil)
	return esdb
}

func (sdb *ChainStateDB) SetGenesis(genesis *types.Genesis, bpInit func(*statedb.StateDB, *types.Genesis) error) error {
	block := genesis.Block()
	stateDB := sdb.OpenNewStateDB(sdb.GetLuaRoot())

	// create state of genesis block
	gbState := sdb.NewBlockState(stateDB.GetRoot(), nil)

	if len(genesis.BPs) > 0 && bpInit != nil {
		// To avoid cyclic dedendency, BP initilization is called via function
		// pointer.
		if err := bpInit(stateDB, genesis); err != nil {
			return err
		}

		aid := types.ToAccountID([]byte(types.AergoSystem))
		scs, err := GetSystemAccountState(stateDB)
		if err != nil {
			return err
		}

		if err := gbState.LuaStateDB.PutState(aid, scs.State); err != nil {
			return err
		}
	}

	for address, balance := range genesis.Balance {
		if v, ok := new(big.Int).SetString(balance, 10); ok {
			accountState, err := GetAccountState(types.ToAddress(address), gbState.LuaStateDB)
			if err != nil {
				return err
			}
			accountState.AddBalance(v)
			if err := accountState.PutState(); err != nil {
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
		Str("luaStateRoot", base58.Encode(bstate.LuaStateDB.GetRoot())).Msg("apply block state")

	if err := sdb.luaStates.SetRoot(bstate.LuaStateDB.GetRoot()); err != nil {
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

func (sdb *ChainStateDB) NewBlockState(blockRoot []byte, evmRoot []byte, options ...BlockStateOptFn) *BlockState {
	ls := sdb.OpenNewStateDB(blockRoot)
	var es *ethstate.StateDB
	if evmRoot != nil {
		es = sdb.OpenEvmStateDB(evmRoot)
	}

	return NewBlockState(ls, es, options...)
}