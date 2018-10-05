/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package state

import (
	"bytes"
	"errors"
	"fmt"
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
	errSaveData       = errors.New("Failed to save data: invalid key")
	errLoadData       = errors.New("Failed to load data: invalid key")
	errSaveBlockState = errors.New("Failed to save blockState: invalid BlockID")
	errLoadBlockState = errors.New("Failed to load blockState: invalid BlockID")
	errSaveStateData  = errors.New("Failed to save StateData: invalid HashID")
	errLoadStateData  = errors.New("Failed to load StateData: invalid HashID")
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

// GetRoot returns root hash of trie
func (states *StateDB) GetRoot() []byte {
	states.lock.RLock()
	defer states.lock.RUnlock()
	return states.trie.Root
}

// SetRoot updates root node of trie as a given root hash
func (states *StateDB) SetRoot(root []byte) {
	states.lock.Lock()
	defer states.lock.Unlock()
	states.trie.Root = root
}

// LoadCache reads first layer of trie given root hash
// and also updates root node of trie as a given root hash
func (states *StateDB) LoadCache(root []byte) error {
	states.lock.Lock()
	defer states.lock.Unlock()
	return states.trie.LoadCache(root)
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

// Snapshot returns revision number of state buffer
func (states *StateDB) Snapshot() int {
	states.lock.RLock()
	defer states.lock.RUnlock()
	return states.buffer.snapshot()
}

// Rollback discards changes of state buffer to revision number
func (states *StateDB) Rollback(snapshot int) error {
	states.lock.Lock()
	defer states.lock.Unlock()
	return states.buffer.rollback(snapshot)
}

// Update applies changes of state buffer to trie
func (states *StateDB) Update() error {
	states.lock.Lock()
	defer states.lock.Unlock()
	keys, vals := states.buffer.export()
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
	latest *types.BlockInfo
	states *StateDB
	store  db.DB
}

// NewChainStateDB creates instance of ChainStateDB
func NewChainStateDB() *ChainStateDB {
	return &ChainStateDB{}
}

// Init initialize database and load statedb of latest block
func (sdb *ChainStateDB) Init(dataDir string) error {
	sdb.Lock()
	defer sdb.Unlock()

	// init db
	if sdb.store == nil {
		dbPath := common.PathMkdirAll(dataDir, stateName)
		sdb.store = db.NewDB(db.BadgerImpl, dbPath)
	}

	// load latest data from db
	err := sdb.loadStateLatest()
	if err != nil {
		return err
	}

	// init trie
	if sdb.states == nil {
		sdb.states = NewStateDB(&sdb.store, sdb.latest.GetStateRoot())
	}
	return nil
}

// Close saves latest block information of the chain
func (sdb *ChainStateDB) Close() error {
	sdb.Lock()
	defer sdb.Unlock()

	// save data to db
	err := sdb.saveStateLatest()
	if err != nil {
		return err
	}

	// close db
	if sdb.store != nil {
		sdb.store.Close()
	}
	return nil
}

// OpenNewStateDB returns new instance of statedb given state root hash
func (sdb *ChainStateDB) OpenNewStateDB(root []byte) *StateDB {
	return NewStateDB(&sdb.store, root)
}

// GetStateRoot returns state root hash
func (sdb *ChainStateDB) GetStateRoot(blockID []byte) ([]byte, error) {
	bstate, err := sdb.loadBlockState(types.ToBlockID(blockID))
	if err != nil {
		return nil, err
	}
	return bstate.StateRoot.Bytes(), nil
}

func (sdb *ChainStateDB) SetGenesis(genesisBlock *types.Genesis) error {
	block := genesisBlock.Block
	gbInfo := &types.BlockInfo{
		BlockNo:   0,
		BlockHash: types.ToBlockID(block.BlockHash()),
	}
	sdb.latest = gbInfo

	// create state of genesis block
	gbState := types.NewBlockState(gbInfo, nil)
	for address, balance := range genesisBlock.Balance {
		bytes := types.ToAddress(address)
		id := types.ToAccountID(bytes)
		gbState.PutAccount(id, &types.State{}, balance)
	}

	if genesisBlock.VoteState != nil {
		aid := types.ToAccountID([]byte(types.AergoSystem))
		gbState.PutAccount(aid, &types.State{}, genesisBlock.VoteState)
	}
	// save state of genesis block
	if err := sdb.apply(gbState); err != nil {
		return err
	}
	return nil
}

func (sdb *ChainStateDB) getAccountState(aid types.AccountID) (*types.State, error) {
	state, err := sdb.states.getState(aid)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return &types.State{}, nil
	}
	return state, nil
}
func (sdb *ChainStateDB) GetAccountStateClone(aid types.AccountID) (*types.State, error) {
	state, err := sdb.getAccountState(aid)
	if err != nil {
		return nil, err
	}
	res := types.State(*state)
	return &res, nil
}
func (sdb *ChainStateDB) getBlockAccount(bs *types.BlockState, aid types.AccountID) (*types.State, error) {
	if aid == emptyAccountID {
		return nil, fmt.Errorf("Failed to get block account: invalid account id")
	}

	if prev, ok := bs.GetAccount(aid); ok {
		return prev, nil
	}
	return sdb.getAccountState(aid)
}
func (sdb *ChainStateDB) GetBlockAccountClone(bs *types.BlockState, aid types.AccountID) (*types.State, error) {
	state, err := sdb.getBlockAccount(bs, aid)
	if err != nil {
		return nil, err
	}
	res := types.State(*state)
	return &res, nil
}

func (sdb *ChainStateDB) updateStateDB(bstate *types.BlockState) error {
	accounts := bstate.GetAccountStates()
	if len(accounts) <= 0 {
		// do nothing
		return nil
	}

	var err error
	// put states to buffer
	for k, v := range accounts {
		err = sdb.states.PutState(k, v)
		if err != nil {
			err2 := sdb.states.Rollback(0)
			if err2 != nil {
				return fmt.Errorf("%v + %v", err.Error(), err2.Error())
			}
			return err
		}
	}
	// update state db
	err = sdb.states.Update()
	if err != nil {
		// rollback to latest
		err2 := sdb.states.Revert(sdb.latest.StateRoot)
		if err2 != nil {
			return fmt.Errorf("%v + %v", err.Error(), err2.Error())
		}
		return err
	}
	// commit state db
	return sdb.states.Commit()
}

func (sdb *ChainStateDB) revertStateDB(prevBlockStateRoot types.HashID) error {
	if sdb.states.trie.Root == nil && prevBlockStateRoot.Equal(emptyHashID) {
		// nil and empty bytes, do nothing
		return nil
	}
	if bytes.Equal(sdb.states.trie.Root, prevBlockStateRoot[:]) {
		// same root, do nothing
		return nil
	}
	// revert state db
	return sdb.states.Revert(prevBlockStateRoot)
}

func (sdb *ChainStateDB) Apply(bstate *types.BlockState) error {
	if sdb.latest.BlockNo+1 != bstate.BlockNo {
		return fmt.Errorf("Failed to apply: invalid block no - latest=%v, this=%v", sdb.latest.BlockNo, bstate.BlockNo)
	}
	if sdb.latest.BlockHash != bstate.PrevHash {
		return fmt.Errorf("Failed to apply: invalid previous block latest=%v, bstate=%v",
			sdb.latest.BlockHash, bstate.PrevHash)
	}
	return sdb.apply(bstate)
}

func (sdb *ChainStateDB) apply(bstate *types.BlockState) error {
	sdb.Lock()
	defer sdb.Unlock()

	// rollback and revert trie requires state root before apply
	if bstate.Undo.StateRoot == emptyHashID {
		bstate.Undo.StateRoot = types.ToHashID(sdb.states.trie.Root)
	}

	// apply blockState to trie
	err := sdb.updateStateDB(bstate)
	if err != nil {
		return err
	}

	// check state root
	if bstate.BlockInfo.StateRoot != types.ToHashID(sdb.GetRoot()) {
		// TODO: if validation failed, than revert statedb.
		bstate.BlockInfo.StateRoot = types.ToHashID(sdb.GetRoot())
	}
	logger.Debug().Str("stateRoot", enc.ToString(sdb.GetRoot())).Msg("apply block state")

	// save blockState
	err = sdb.saveBlockState(bstate)
	if err != nil {
		return err
	}

	sdb.latest = &bstate.BlockInfo
	err = sdb.saveStateLatest()
	return err
}

func (sdb *ChainStateDB) Rollback(targetBlockID types.BlockID) error {
	sdb.Lock()
	defer sdb.Unlock()

	target, err := sdb.loadBlockState(targetBlockID)
	if err != nil {
		return err
	}
	logger.Debug().Str("before", sdb.latest.StateRoot.String()).
		Str("target", target.StateRoot.String()).Msg("rollback state")

	err = sdb.revertStateDB(target.StateRoot)
	if err != nil {
		return err
	}

	sdb.latest = &target.BlockInfo
	return sdb.saveStateLatest()
}

// GetRoot returns state root hash
func (sdb *ChainStateDB) GetRoot() []byte {
	return sdb.states.GetRoot()
}

func (sdb *ChainStateDB) IsExistState(hash []byte) bool {
	//TODO : StateRootValidation
	return false
}
