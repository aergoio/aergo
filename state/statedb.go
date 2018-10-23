/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package state

import (
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

type V struct {
	sdb    *StateDB
	id     []byte
	aid    types.AccountID
	oldV   *types.State
	newV   *types.State
	newOne bool
}

func (v *V) ID() []byte {
	if len(v.id) != 33 {
		v.id = append(v.id, make([]byte, 33-len(v.id))...)
	}
	return v.id
}

func (v *V) AccountID() types.AccountID {
	return v.aid
}

func (v *V) State() *types.State {
	return v.newV
}

func (v *V) SetNonce(nonce uint64) {
	v.newV.Nonce = nonce
}

func (v *V) Balance() uint64 {
	return v.newV.Balance
}

func (v *V) AddBalance(amount uint64) {
	v.newV.Balance += amount
}

func (v *V) SubBalance(amount uint64) {
	v.newV.Balance -= amount
}

func (v *V) RP() uint64 {
	return v.newV.SqlRecoveryPoint
}

func (v *V) IsNew() bool {
	return v.newOne
}

func (v *V) Reset() {
	*v.newV = types.State(*v.oldV)
}

func (v *V) PutState() error {
	return v.sdb.PutState(v.aid, v.newV)
}

func (states *StateDB) CreateAccountStateV(id []byte) (*V, error) {
	v, err := states.GetAccountStateV(id)
	if err != nil {
		return nil, err
	}
	if !v.newOne {
		return nil, fmt.Errorf("account(%s) aleardy exists", types.EncodeAddress(v.ID()))
	}
	v.newV.SqlRecoveryPoint = 1
	return v, nil
}

func (states *StateDB) GetAccountStateV(id []byte) (*V, error) {
	aid := types.ToAccountID(id)
	st, err := states.GetState(aid)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return &V{
			sdb:    states,
			id:     id,
			aid:    aid,
			oldV:   &types.State{},
			newV:   &types.State{},
			newOne: true,
		}, nil
	}
	newV := new(types.State)
	*newV = types.State(*st)
	return &V{
		sdb:  states,
		id:   id,
		aid:  aid,
		oldV: st,
		newV: newV,
	}, nil
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
	key, err := states.trie.Get(id[:])
	if err != nil {
		return nil, err
	}
	if key == nil || len(key) == 0 {
		return nil, nil
	}
	return states.loadStateData(key)
}

// GetStateAndProof gets the state and associated proof of an account
// in the given trie root. If the account doesnt exist, a proof of
// non existence is returned.
func (states *StateDB) GetStateAndProof(id types.AccountID, root []byte) (*types.StateProof, error) {
	var state *types.State
	var ap [][]byte
	var proofKey, proofVal []byte
	var isIncluded bool
	var err error
	states.lock.RLock()
	defer states.lock.RUnlock()

	if len(root) != 0 {
		// Get the state and proof of the account for a past state
		ap, isIncluded, proofKey, proofVal, err = states.trie.MerkleProofPast(id[:], root)
		if err != nil {
			return nil, err
		}
	} else {
		// Get the state and proof of the account
		// The wallet should check that state hashes to proofVal and verify the audit path,
		// The returned proofVal shouldn't be trusted by the wallet, it is used to proove non inclusion
		ap, isIncluded, proofKey, proofVal, err = states.trie.MerkleProof(id[:])
		if err != nil {
			return nil, err
		}
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
	logger.Debug().Str("state root : ", enc.ToString(states.trie.Root)).Msg("Get State and Proof")
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
