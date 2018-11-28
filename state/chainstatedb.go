package state

import (
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
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

		sdb.states = NewStateDB(&sdb.store, sroot, sdb.testmode)
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

// OpenNewStateDB returns new instance of statedb given state root hash
func (sdb *ChainStateDB) OpenNewStateDB(root []byte) *StateDB {
	return NewStateDB(&sdb.store, root, sdb.testmode)
}

func (sdb *ChainStateDB) SetGenesis(genesisBlock *types.Genesis) error {
	block := genesisBlock.Block()

	// create state of genesis block
	gbState := NewBlockState(sdb.OpenNewStateDB(sdb.GetRoot()))
	for address, balance := range genesisBlock.Balance {
		bytes := types.ToAddress(address)
		id := types.ToAccountID(bytes)
		if err := gbState.PutState(id, balance); err != nil {
			return err
		}
	}

	if genesisBlock.VoteState() != nil {
		aid := types.ToAccountID([]byte(types.AergoSystem))
		if err := gbState.PutState(aid, genesisBlock.VoteState()); err != nil {
			return err
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

func (sdb *ChainStateDB) Rollback(targetBlockRoot []byte) error {
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

func (sdb *ChainStateDB) NewBlockState(root []byte) *BlockState {
	bState := NewBlockState(sdb.OpenNewStateDB(root))

	return bState
}
