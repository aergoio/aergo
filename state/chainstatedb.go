package state

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/types"
)

// ChainStateDB manages statedb and additional informations about blocks like a state root hash
type ChainStateDB struct {
	sync.RWMutex
	states   *StateDB
	store    db.DB
	testmode bool
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
		store:  sdb.store,
		states: sdb.GetStateDB().Clone(),
	}
	return newSdb
}

// Init initialize database and load statedb of latest block
func (sdb *ChainStateDB) Init(dbType string, dataDir string, bestBlock *types.Block, test bool) error {
	sdb.Lock()
	defer sdb.Unlock()

	sdb.testmode = test
	// init db
	if sdb.store == nil {
		dbPath := common.PathMkdirAll(dataDir, stateName)
		sdb.store = db.NewDB(db.ImplType(dbType), dbPath)
	}

	// init trie
	if sdb.states == nil {
		var sroot []byte
		if bestBlock != nil {
			sroot = bestBlock.GetHeader().GetBlocksRootHash()
		}

		sdb.states = NewStateDB(sdb.store, sroot, sdb.testmode)
	}
	return nil
}

// Close saves latest block information of the chain
func (sdb *ChainStateDB) Close() error {
	sdb.Lock()
	defer sdb.Unlock()

	// close db
	if sdb.store != nil {
		sdb.store.Close()
	}
	return nil
}

// GetStateDB returns statedb stores account states
func (sdb *ChainStateDB) GetStateDB() *StateDB {
	return sdb.states
}

// GetSystemAccountState returns the state of the aergo system account.
func (sdb *ChainStateDB) GetSystemAccountState() (*ContractState, error) {
	return sdb.GetStateDB().GetSystemAccountState()
}

// OpenNewStateDB returns new instance of statedb given state root hash
func (sdb *ChainStateDB) OpenNewStateDB(root []byte) *StateDB {
	return NewStateDB(sdb.store, root, sdb.testmode)
}

func (sdb *ChainStateDB) SetGenesis(genesis *types.Genesis, bpInit func(*StateDB, *types.Genesis) error) error {
	block := genesis.Block()
	stateDB := sdb.OpenNewStateDB(sdb.GetRoot())

	// create state of genesis block
	gbState := sdb.NewBlockState(stateDB.GetRoot())

	if len(genesis.BPs) > 0 && bpInit != nil {
		// To avoid cyclic dedendency, BP initilization is called via function
		// pointer.
		if err := bpInit(stateDB, genesis); err != nil {
			return err
		}

		aid := types.ToAccountID([]byte(types.AergoSystem))
		scs, err := stateDB.OpenContractStateAccount(aid)
		if err != nil {
			return err
		}

		if err := gbState.PutState(aid, scs.State); err != nil {
			return err
		}
	}

	for address, balance := range genesis.Balance {
		bytes := types.ToAddress(address)
		id := types.ToAccountID(bytes)
		if v, ok := new(big.Int).SetString(balance, 10); ok {
			if err := gbState.PutState(id, &types.State{Balance: v.Bytes()}); err != nil {
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

	block.SetBlocksRootHash(sdb.GetRoot())

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

	logger.Debug().Str("before", enc.ToString(sdb.states.GetRoot())).
		Str("stateRoot", enc.ToString(bstate.GetRoot())).Msg("apply block state")

	if err := sdb.states.SetRoot(bstate.GetRoot()); err != nil {
		return err
	}

	return nil
}

func (sdb *ChainStateDB) SetRoot(targetBlockRoot []byte) error {
	sdb.Lock()
	defer sdb.Unlock()

	logger.Debug().Str("before", enc.ToString(sdb.states.GetRoot())).
		Str("target", enc.ToString(targetBlockRoot)).Msg("rollback state")

	sdb.states.SetRoot(targetBlockRoot)
	return nil
}

// GetRoot returns state root hash
func (sdb *ChainStateDB) GetRoot() []byte {
	return sdb.states.GetRoot()
}

func (sdb *ChainStateDB) IsExistState(hash []byte) bool {
	//TODO : StateRootValidation
	return false
}

func (sdb *ChainStateDB) NewBlockState(root []byte, options ...BlockStateOptFn) *BlockState {
	return NewBlockState(sdb.OpenNewStateDB(root), options...)
}
