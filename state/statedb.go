/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package state

import (
	"errors"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/pkg/trie"
	"github.com/aergoio/aergo/types"
)

const (
	stateName   = "state"
	stateLatest = stateName + ".latest"
)

var (
	logger = log.NewLogger(stateName)
)

var (
	emptyHashID    = types.HashID{}
	emptyBlockID   = types.BlockID{}
	emptyAccountID = types.AccountID{}
)

var (
	errSaveData      = errors.New("Failed to save data: invalid key")
	errLoadData      = errors.New("Failed to load data: invalid key")
	errSaveBlockInfo = errors.New("Failed to save blockInfo: invalid BlockID")
	errLoadBlockInfo = errors.New("Failed to load blockInfo: invalid BlockID")
	errSaveStateData = errors.New("Failed to save StateData: invalid HashID")
	errLoadStateData = errors.New("Failed to load StateData: invalid HashID")
)

var (
	errInvalidArgs = errors.New("invalid arguments")
	errInvalidRoot = errors.New("invalid root")
	errSetRoot     = errors.New("Failed to set root: invalid root")
	errLoadRoot    = errors.New("Failed to load root: invalid root")
	errGetState    = errors.New("Failed to get state: invalid account id")
	errPutState    = errors.New("Failed to put state: invalid account id")
)

// StateDB manages trie of states
type StateDB struct {
	lock   sync.RWMutex
	buffer *stateBuffer
	trie   *trie.Trie
	store  *db.DB
}

// NewStateDB craete StateDB instance
func NewStateDB(dbstore *db.DB, root []byte) *StateDB {
	sdb := StateDB{
		buffer: newStateBuffer(),
		trie:   trie.NewTrie(root, common.Hasher, *dbstore),
		store:  dbstore,
	}
	return &sdb
}

// Clone returns a new StateDB which has same store and Root
func (states *StateDB) Clone() *StateDB {
	states.lock.RLock()
	defer states.lock.RUnlock()

	return NewStateDB(states.store, states.GetRoot())
}

// GetRoot returns root hash of trie
func (states *StateDB) GetRoot() []byte {
	states.lock.RLock()
	defer states.lock.RUnlock()
	return states.trie.Root
}

// SetRoot updates root node of trie as a given root hash
func (states *StateDB) SetRoot(root []byte) error {
	states.lock.Lock()
	defer states.lock.Unlock()
	// update root node
	states.trie.Root = root
	// reset buffer
	return states.buffer.reset()
}

// LoadCache reads first layer of trie given root hash
// and also updates root node of trie as a given root hash
func (states *StateDB) LoadCache(root []byte) error {
	states.lock.Lock()
	defer states.lock.Unlock()
	// update root node and load cache
	err := states.trie.LoadCache(root)
	if err != nil {
		return err
	}
	// reset buffer
	return states.buffer.reset()
}

// Revert rollbacks trie to previous root hash
func (states *StateDB) Revert(root types.HashID) error {
	states.lock.Lock()
	defer states.lock.Unlock()
	// // handle nil bytes
	// targetRoot := root.Bytes()

	// // revert trie
	// err := states.trie.Revert(targetRoot)
	// if err != nil {
	// 	// when targetRoot is not contained in the cached tries.
	// 	states.trie.Root = targetRoot
	// }

	// just update root node as targetRoot.
	// revert trie consumes unnecessarily long time.
	states.trie.Root = root.Bytes()

	// reset buffer
	return states.buffer.reset()
}

// PutState puts account id and its state into state buffer.
func (states *StateDB) PutState(id types.AccountID, state *types.State) error {
	states.lock.Lock()
	defer states.lock.Unlock()
	if id == emptyAccountID {
		return errPutState
	}
	return states.buffer.put(types.HashID(id), state)
}

// GetAccountState gets state of account id from statedb.
// empty state is returned when there is no state corresponding to account id.
func (states *StateDB) GetAccountState(id types.AccountID) (*types.State, error) {
	st, err := states.GetState(id)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return &types.State{}, nil
	}
	return st, nil
}

// GetState gets state of account id from state buffer and trie.
// nil value is returned when there is no state corresponding to account id.
func (states *StateDB) GetState(id types.AccountID) (*types.State, error) {
	states.lock.RLock()
	defer states.lock.RUnlock()
	if id == emptyAccountID {
		return nil, errGetState
	}
	// get state from buffer
	entry := states.buffer.get(types.HashID(id))
	if entry != nil {
		return entry.getData().(*types.State), nil
	}
	// get state from trie
	return states.getState(id)
}

// getState gets state of account id from trie.
// nil value is returned when there is no state corresponding to account id.
func (states *StateDB) getState(id types.AccountID) (*types.State, error) {
	states.lock.RLock()
	defer states.lock.RUnlock()
	key, err := states.trie.Get(id[:])
	if err != nil {
		return nil, err
	}
	if key == nil || len(key) == 0 {
		return nil, nil
	}
	return states.loadStateData(key)
	// st := types.State{}
	// err = loadData(states.store, key, st)
	// if err != nil {
	// 	return nil, err
	// }
	// return &st, nil
}

// GetStateAndProof gets the state and associated proof of an account
// in the last produced block. If the account doesnt exist, a proof of
// non existence is returned.
func (states *StateDB) GetStateAndProof(id types.AccountID) (*types.StateProof, error) {
	var state *types.State
	states.lock.RLock()
	defer states.lock.RUnlock()
	// Get the state and proof of the account
	// The wallet should check that state hashes to proofVal and verify the audit path,
	// The returned proofVal shouldn't be trusted by the wallet, it is used to proove non inclusion
	ap, isIncluded, proofKey, proofVal, err := states.trie.MerkleProof(id[:])
	if err != nil {
		return nil, err
	}
	if isIncluded {
		state, err = states.loadStateData(proofVal)
		if err != nil {
			return nil, err
		}
	}
	stateProof := &types.StateProof{
		State:     state,
		Inclusion: isIncluded,
		ProofKey:  proofKey,
		ProofVal:  proofVal,
		AuditPath: ap,
	}
	return stateProof, nil
}

// Snapshot represents revision number of statedb
type Snapshot int

// Snapshot returns revision number of state buffer
func (states *StateDB) Snapshot() Snapshot {
	states.lock.RLock()
	defer states.lock.RUnlock()
	return Snapshot(states.buffer.snapshot())
}

// Rollback discards changes of state buffer to revision number
func (states *StateDB) Rollback(revision Snapshot) error {
	states.lock.Lock()
	defer states.lock.Unlock()
	return states.buffer.rollback(int(revision))
}

// Update applies changes of state buffer to trie
func (states *StateDB) Update() error {
	states.lock.Lock()
	defer states.lock.Unlock()
	keys, vals := states.buffer.export()
	if len(keys) == 0 || len(vals) == 0 {
		// nothing to update
		return nil
	}
	_, err := states.trie.Update(keys, vals)
	if err != nil {
		return err
	}
	return nil
}

// Commit writes state buffer and trie to db
func (states *StateDB) Commit() error {
	states.lock.Lock()
	defer states.lock.Unlock()
	err := states.trie.Commit()
	if err != nil {
		return err
	}
	err = states.buffer.commit(states.store)
	if err != nil {
		return err
	}
	return states.buffer.reset()
}

// ChainStateDB manages statedb and additional informations about blocks like a state root hash
type ChainStateDB struct {
	sync.RWMutex
	states *StateDB
	store  db.DB
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
func (sdb *ChainStateDB) Init(dataDir string, bestBlock *types.Block) error {
	sdb.Lock()
	defer sdb.Unlock()

	// init db
	if sdb.store == nil {
		dbPath := common.PathMkdirAll(dataDir, stateName)
		sdb.store = db.NewDB(db.BadgerImpl, dbPath)
	}

	// init trie
	if sdb.states == nil {
		var sroot []byte
		if bestBlock != nil {
			sroot = bestBlock.GetHeader().GetBlocksRootHash()
		}

		sdb.states = NewStateDB(&sdb.store, sroot)
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
	return NewStateDB(&sdb.store, root)
}

func (sdb *ChainStateDB) SetGenesis(genesisBlock *types.Genesis) error {
	block := genesisBlock.Block

	// create state of genesis block
	gbState := NewBlockState(sdb.OpenNewStateDB(sdb.GetRoot()), nil)
	for address, balance := range genesisBlock.Balance {
		bytes := types.ToAddress(address)
		id := types.ToAccountID(bytes)
		if err := gbState.PutState(id, balance); err != nil {
			return err
		}
	}

	if genesisBlock.VoteState != nil {
		aid := types.ToAccountID([]byte(types.AergoSystem))
		if err := gbState.PutState(aid, genesisBlock.VoteState); err != nil {
			return err
		}
	}
	// save state of genesis block
	// FIXME don't use chainstate API
	if err := sdb.apply(gbState); err != nil {
		return err
	}

	block.SetBlocksRootHash(sdb.GetRoot())

	return nil
}

// func (sdb *ChainStateDB) getAccountState(aid types.AccountID) (*types.State, error) {
// 	state, err := sdb.states.GetAccountState(aid)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return state, nil
// }
// func (sdb *ChainStateDB) GetAccountStateClone(aid types.AccountID) (*types.State, error) {
// 	state, err := sdb.states.GetAccountState(aid)
// 	if err != nil {
// 		return nil, err
// 	}
// 	res := types.State(*state)
// 	return &res, nil
// }

// func (sdb *ChainStateDB) getBlockAccount(bs *BlockState, aid types.AccountID) (*types.State, error) {
// 	if aid == emptyAccountID {
// 		return nil, fmt.Errorf("Failed to get block account: invalid account id")
// 	}

// 	if prev, ok := bs.GetAccountState(aid); ok {
// 		return prev, nil
// 	}
// 	return sdb.states.GetAccountState(aid)
// }
// func (sdb *ChainStateDB) GetBlockAccountClone(bs *BlockState, aid types.AccountID) (*types.State, error) {
// 	state, err := sdb.getBlockAccount(bs, aid)
// 	if err != nil {
// 		return nil, err
// 	}
// 	res := types.State(*state)
// 	return &res, nil
// }

// func (sdb *ChainStateDB) updateStateDB(bstate *BlockState) error {
// 	accounts := bstate.GetAccountStates()
// 	if len(accounts) <= 0 {
// 		// do nothing
// 		return nil
// 	}

// 	var err error
// 	// put states to buffer
// 	for k, v := range accounts {
// 		err = sdb.states.PutState(k, v)
// 		if err != nil {
// 			err2 := sdb.states.Rollback(0)
// 			if err2 != nil {
// 				return fmt.Errorf("%v + %v", err.Error(), err2.Error())
// 			}
// 			return err
// 		}
// 	}
// 	// update state db
// 	err = bstate.Update()
// 	if err != nil {
// 		// rollback to latest
// 		err2 := sdb.states.Revert(sdb.latest.StateRoot)
// 		if err2 != nil {
// 			return fmt.Errorf("%v + %v", err.Error(), err2.Error())
// 		}
// 		return err
// 	}
// 	// commit state db
// 	return sdb.states.Commit()
// }

func (sdb *ChainStateDB) Apply(bstate *BlockState) error {
	return sdb.apply(bstate)
}

func (sdb *ChainStateDB) apply(bstate *BlockState) error {
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

func (sdb *ChainStateDB) NewBlockState(root []byte, recpTx db.Transaction) *BlockState {
	bState := NewBlockState(sdb.OpenNewStateDB(root), recpTx)

	return bState
}
