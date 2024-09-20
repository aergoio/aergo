package statedb

import (
	"bytes"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/aergoio/aergo/v2/types"
)

type ContractState struct {
	*types.State
	id      []byte
	account types.AccountID
	code    []byte
	storage *bufferedStorage
	store   db.DB
}

func (cs *ContractState) SetAccountID(newID types.AccountID) {
	cs.account = newID
}

func (cs *ContractState) SetCode(code []byte) error {
	codeHash := common.Hasher(code)
	storedCode, err := cs.GetRawKV(codeHash[:])
	if err == nil && !bytes.Equal(code, storedCode) {
		err = cs.SetRawKV(codeHash[:], code)
	}
	if err != nil {
		return err
	}
	cs.State.CodeHash = codeHash[:]
	cs.code = code
	return nil
}

func (cs *ContractState) GetCode() ([]byte, error) {
	if cs.code != nil {
		// already loaded.
		return cs.code, nil
	}
	codeHash := cs.GetCodeHash()
	if codeHash == nil {
		// not defined. do nothing.
		return nil, nil
	}
	err := loadData(cs.store, cs.State.GetCodeHash(), &cs.code)
	if err != nil {
		return nil, err
	}
	return cs.code, nil
}

func (cs *ContractState) GetAccountID() types.AccountID {
	return cs.account
}

func (cs *ContractState) GetID() []byte {
	return cs.id
}

// SetRawKV saves (key, value) to st.store without any kind of encoding.
func (cs *ContractState) SetRawKV(key []byte, value []byte) error {
	return saveData(cs.store, key, value)
}

// GetRawKV loads (key, value) from st.store.
func (cs *ContractState) GetRawKV(key []byte) ([]byte, error) {
	var b []byte
	if err := loadData(cs.store, key, &b); err != nil {
		return nil, err
	}
	return b, nil
}

// HasKey returns existence of the key
func (cs *ContractState) HasKey(key []byte) bool {
	return cs.storage.has(types.GetHashID(key), true)
}

// SetData store key and value pair to the storage.
func (cs *ContractState) SetData(key, value []byte) error {
	cs.storage.put(newValueEntry(types.GetHashID(key), value))
	return nil
}

// GetData returns the value corresponding to the key from the buffered storage.
func (cs *ContractState) GetData(key []byte) ([]byte, error) {
	id := types.GetHashID(key)
	if entry := cs.storage.get(id); entry != nil {
		if value := entry.Value(); value != nil {
			return value.([]byte), nil
		}
		return nil, nil
	}
	return cs.getInitialData(id[:])
}

func (cs *ContractState) getInitialData(id []byte) ([]byte, error) {
	dkey, err := cs.storage.Trie.Get(id)
	if err != nil {
		return nil, err
	}
	if len(dkey) == 0 {
		return nil, nil
	}
	value := []byte{}
	if err := loadData(cs.store, dkey, &value); err != nil {
		return nil, err
	}
	return value, nil
}

// GetInitialData returns the value corresponding to the key from the contract storage.
func (cs *ContractState) GetInitialData(key []byte) ([]byte, error) {
	id := types.GetHashID(key)
	return cs.getInitialData(id[:])
}

// DeleteData remove key and value pair from the storage.
func (cs *ContractState) DeleteData(key []byte) error {
	cs.storage.put(newValueEntryDelete(types.GetHashID(key)))
	return nil
}

// Snapshot returns revision number of storage buffer
func (cs *ContractState) Snapshot() Snapshot {
	return Snapshot(cs.storage.Buffer.snapshot())
}

// Rollback discards changes of storage buffer to revision number
func (cs *ContractState) Rollback(revision Snapshot) error {
	return cs.storage.Buffer.rollback(int(revision))
}

// Hash implements types.ImplHashBytes
func (cs *ContractState) Hash() []byte {
	return getHashBytes(cs.State)
}

// Marshal implements types.ImplMarshal
func (cs *ContractState) Marshal() ([]byte, error) {
	return proto.Encode(cs.State)
}

func (cs *ContractState) cache() *stateBuffer {
	return cs.storage.Buffer
}

//---------------------------------------------------------------//
// global functions

func OpenContractStateAccount(id []byte, states *StateDB) (*ContractState, error) {
	aid := types.ToAccountID(id)
	st, err := states.GetAccountState(aid)
	if err != nil {
		return nil, err
	}
	return OpenContractState(id, st, states)
}

func OpenContractState(id []byte, st *types.State, states *StateDB) (*ContractState, error) {
	aid := types.ToAccountID(id)
	storage := states.Cache.get(aid)
	if storage == nil {
		root := common.Compactz(st.StorageRoot)
		storage = newBufferedStorage(root, states.Store)
	}
	res := &ContractState{
		State:   st,
		id:      id,
		account: aid,
		storage: storage,
		store:   states.Store,
	}
	return res, nil
}

func StageContractState(st *ContractState, states *StateDB) error {
	states.Cache.put(st.account, st.storage)
	st.storage = nil
	return nil
}

// GetSystemAccountState returns the ContractState of the AERGO system account.
func GetSystemAccountState(states *StateDB) (*ContractState, error) {
	return OpenContractStateAccount([]byte(types.AergoSystem), states)
}

// GetNameAccountState returns the ContractState of the AERGO name account.
func GetNameAccountState(states *StateDB) (*ContractState, error) {
	return OpenContractStateAccount([]byte(types.AergoName), states)
}

// GetEnterpriseAccountState returns the ContractState of the AERGO enterprise account.
func GetEnterpriseAccountState(states *StateDB) (*ContractState, error) {
	return OpenContractStateAccount([]byte(types.AergoEnterprise), states)
}
