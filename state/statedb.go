/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package state

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/pkg/trie"
	"github.com/aergoio/aergo/v2/types"
)

const (
	stateName   = "state"
	stateLatest = stateName + ".latest"
)

var (
	stateMarker = []byte{0x54, 0x45} // marker: tail end
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
	errSaveData = errors.New("Failed to save data: invalid key")
	errLoadData = errors.New("Failed to load data: invalid key")

	errLoadStateData = errors.New("Failed to load StateData: invalid HashID")
	// errSaveStateData = errors.New("Failed to save StateData: invalid HashID")

	// errInvalidArgs = errors.New("invalid arguments")
	// errInvalidRoot = errors.New("invalid root")
	// errSetRoot     = errors.New("Failed to set root: invalid root")
	// errLoadRoot    = errors.New("Failed to load root: invalid root")

	errGetState = errors.New("Failed to get state: invalid account id")
	errPutState = errors.New("Failed to put state: invalid account id")
)

// StateDB manages trie of states
type StateDB struct {
	lock     sync.RWMutex
	buffer   *stateBuffer
	cache    *storageCache
	trie     *trie.Trie
	store    db.DB
	batchtx  db.Transaction
	testmode bool
}

// NewStateDB craete StateDB instance
func NewStateDB(dbstore db.DB, root []byte, test bool) *StateDB {
	sdb := StateDB{
		buffer:   newStateBuffer(),
		cache:    newStorageCache(),
		trie:     trie.NewTrie(root, common.Hasher, dbstore),
		store:    dbstore,
		testmode: test,
	}
	return &sdb
}

// Clone returns a new StateDB which has same store and Root
func (states *StateDB) Clone() *StateDB {
	states.lock.RLock()
	defer states.lock.RUnlock()

	return NewStateDB(states.store, states.GetRoot(), states.testmode)
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
	states.buffer.put(newValueEntry(types.HashID(id), state))
	return nil
}

// GetAccountState gets state of account id from statedb.
// empty state is returned when there is no state corresponding to account id.
func (states *StateDB) GetAccountState(id types.AccountID) (*types.State, error) {
	st, err := states.GetState(id)
	if err != nil {
		return nil, err
	}
	if st == nil {
		if states.testmode {
			amount := new(big.Int).Add(types.StakingMinimum, types.StakingMinimum)
			return &types.State{Balance: amount.Bytes()}, nil
		}
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
	deploy int8
	buffer *stateBuffer
}

const (
	deployFlag = 0x01 << iota
	redeployFlag
)

func (v *V) ID() []byte {
	if len(v.id) < types.AddressLength {
		v.id = types.AddressPadding(v.id)
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

func (v *V) Balance() *big.Int {
	return new(big.Int).SetBytes(v.newV.Balance)
}

func (v *V) AddBalance(amount *big.Int) {
	balance := new(big.Int).SetBytes(v.newV.Balance)
	v.newV.Balance = new(big.Int).Add(balance, amount).Bytes()
}

func (v *V) SubBalance(amount *big.Int) {
	balance := new(big.Int).SetBytes(v.newV.Balance)
	v.newV.Balance = new(big.Int).Sub(balance, amount).Bytes()
}

func (v *V) RP() uint64 {
	return v.newV.SqlRecoveryPoint
}

func (v *V) IsNew() bool {
	return v.newOne
}

func (v *V) IsDeploy() bool {
	return v.deploy&deployFlag != 0
}

func (v *V) SetRedeploy() {
	v.deploy = deployFlag | redeployFlag
}

func (v *V) IsRedeploy() bool {
	return v.deploy&redeployFlag != 0
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
	v.deploy = deployFlag
	return v, nil
}

func (states *StateDB) GetAccountStateV(id []byte) (*V, error) {
	aid := types.ToAccountID(id)
	st, err := states.GetState(aid)
	if err != nil {
		return nil, err
	}
	if st == nil {
		if states.testmode {
			amount := new(big.Int).Add(types.StakingMinimum, types.StakingMinimum)
			return &V{
				sdb:    states,
				id:     id,
				aid:    aid,
				oldV:   &types.State{Balance: amount.Bytes()},
				newV:   &types.State{Balance: amount.Bytes()},
				newOne: true,
			}, nil
		}
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

func (states *StateDB) InitAccountStateV(id []byte, old *types.State, new *types.State) *V {
	return &V{
		sdb:  states,
		id:   id,
		aid:  types.ToAccountID(id),
		oldV: old,
		newV: new,
	}
}

func (v *V) ClearAid() {
	v.aid = emptyAccountID
}

// GetState gets state of account id from state buffer and trie.
// nil value is returned when there is no state corresponding to account id.
func (states *StateDB) GetState(id types.AccountID) (*types.State, error) {
	states.lock.RLock()
	defer states.lock.RUnlock()
	if id == emptyAccountID {
		return nil, errGetState
	}
	return states.getState(id)
}

// getState returns state of account id from buffer and trie.
// nil value is returned when there is no state corresponding to account id.
func (states *StateDB) getState(id types.AccountID) (*types.State, error) {
	// get state from buffer
	if entry := states.buffer.get(types.HashID(id)); entry != nil {
		return entry.Value().(*types.State), nil
	}
	// get state from trie
	return states.getTrieState(id)
}

// getTrieState gets state of account id from trie.
// nil value is returned when there is no state corresponding to account id.
func (states *StateDB) getTrieState(id types.AccountID) (*types.State, error) {
	key, err := states.trie.Get(id[:])
	if err != nil {
		return nil, err
	}
	if key == nil || len(key) == 0 {
		return nil, nil
	}
	return states.loadStateData(key)
}

func (states *StateDB) TrieQuery(id []byte, root []byte, compressed bool) ([]byte, [][]byte, int, bool, []byte, []byte, error) {
	var ap [][]byte
	var proofKey, proofVal, bitmap []byte
	var isIncluded bool
	var err error
	var height int
	states.lock.RLock()
	defer states.lock.RUnlock()

	if len(root) != 0 {
		if compressed {
			bitmap, ap, height, isIncluded, proofKey, proofVal, err = states.trie.MerkleProofCompressedR(id, root)
		} else {
			// Get the state and proof of the account for a past state
			ap, isIncluded, proofKey, proofVal, err = states.trie.MerkleProofR(id, root)
		}
	} else {
		// Get the state and proof of the account at the latest trie
		// The wallet should check that state hashes to proofVal and verify the audit path,
		// The returned proofVal shouldn't be trusted by the wallet, it is used to proove non inclusion
		if compressed {
			bitmap, ap, height, isIncluded, proofKey, proofVal, err = states.trie.MerkleProofCompressed(id)
		} else {
			ap, isIncluded, proofKey, proofVal, err = states.trie.MerkleProof(id)
		}
	}
	return bitmap, ap, height, isIncluded, proofKey, proofVal, err
}

// GetVarAndProof gets the value of a variable in the given contract trie root.
func (states *StateDB) GetVarAndProof(id []byte, root []byte, compressed bool) (*types.ContractVarProof, error) {
	var value []byte
	bitmap, ap, height, isIncluded, proofKey, dbKey, err := states.TrieQuery(id, root, compressed)
	if err != nil {
		return nil, err
	}
	if isIncluded {
		value = []byte{}
		if err := loadData(states.store, dbKey, &value); err != nil {
			return nil, err
		}
		// proofKey and proofVal are only not nil for prooving exclusion with another leaf on the path
		dbKey = nil
	}
	contractVarProof := &types.ContractVarProof{
		Value:     value,
		Inclusion: isIncluded,
		ProofKey:  proofKey,
		ProofVal:  dbKey,
		Bitmap:    bitmap,
		Height:    uint32(height),
		AuditPath: ap,
	}
	logger.Debug().Str("contract root : ", enc.ToString(root)).Msg("Get contract variable and Proof")
	return contractVarProof, nil

}

// GetAccountAndProof gets the state and associated proof of an account
// in the given trie root. If the account doesnt exist, a proof of
// non existence is returned.
func (states *StateDB) GetAccountAndProof(id []byte, root []byte, compressed bool) (*types.AccountProof, error) {
	var state *types.State
	bitmap, ap, height, isIncluded, proofKey, dbKey, err := states.TrieQuery(id, root, compressed)
	if err != nil {
		return nil, err
	}
	if isIncluded {
		state, err = states.loadStateData(dbKey)
		if err != nil {
			return nil, err
		}
		dbKey = nil
	}
	accountProof := &types.AccountProof{
		State:     state,
		Inclusion: isIncluded,
		ProofKey:  proofKey,
		ProofVal:  dbKey,
		Bitmap:    bitmap,
		Height:    uint32(height),
		AuditPath: ap,
	}
	logger.Debug().Str("state root : ", enc.ToString(root)).Msg("Get Account and Proof")
	return accountProof, nil
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

	if err := states.update(); err != nil {
		return err
	}
	return nil
}

func (states *StateDB) update() error {
	// update storage and put state with changed storage root
	if err := states.updateStorage(); err != nil {
		return err
	}
	// export buffer and update to trie
	if err := states.buffer.updateTrie(states.trie); err != nil {
		return err
	}
	return nil
}

func (states *StateDB) updateStorage() error {
	before := states.buffer.snapshot()
	for id, storage := range states.cache.storages {
		// update storage
		if err := storage.update(); err != nil {
			states.buffer.rollback(before)
			return err
		}
		// update state if storage root changed
		if storage.isDirty() {
			st, err := states.getState(id)
			if err != nil {
				states.buffer.rollback(before)
				return err
			}
			if st == nil {
				st = &types.State{}
			}
			// put state with storage root
			st.StorageRoot = storage.trie.Root
			states.buffer.put(newValueEntry(types.HashID(id), st))
		}
	}
	return nil
}

// Commit writes state buffer and trie to db
func (states *StateDB) Commit() error {
	states.lock.Lock()
	defer states.lock.Unlock()

	bulk := states.store.NewBulk()
	for _, storage := range states.cache.storages {
		// stage changes
		if err := storage.stage(bulk); err != nil {
			bulk.DiscardLast()
			return err
		}
	}
	if err := states.stage(bulk); err != nil {
		bulk.DiscardLast()
		return err
	}
	bulk.Flush()
	return nil
}

func (states *StateDB) stage(txn trie.DbTx) error {
	// stage trie and buffer
	states.trie.StageUpdates(txn)
	if err := states.buffer.stage(txn); err != nil {
		return err
	}
	// set marker
	states.setMarker(txn)
	// reset buffer
	if err := states.buffer.reset(); err != nil {
		return err
	}
	return nil
}

// setMarker store the marker that represents finalization of the state root.
func (states *StateDB) setMarker(txn trie.DbTx) {
	if states.trie.Root == nil {
		return
	}
	// logger.Debug().Str("stateRoot", enc.ToString(states.trie.Root)).Msg("setMarker")
	txn.Set(common.Hasher(states.trie.Root), stateMarker)
}

// HasMarker represents that the state root is finalized or not.
func (states *StateDB) HasMarker(root []byte) bool {
	if root == nil {
		return false
	}
	marker := states.store.Get(common.Hasher(root))
	if marker != nil && bytes.Equal(marker, stateMarker) {
		// logger.Debug().Str("stateRoot", enc.ToString(root)).Str("marker", hex.EncodeToString(marker)).Msg("IsMarked")
		return true
	}
	return false
}
