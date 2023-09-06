package state

import (
	"bytes"
	"math/big"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
)

func (states *StateDB) OpenContractStateAccount(aid types.AccountID) (*ContractState, error) {
	st, err := states.GetAccountState(aid)
	if err != nil {
		return nil, err
	}
	return states.OpenContractState(aid, st)
}
func (states *StateDB) OpenContractState(aid types.AccountID, st *types.State) (*ContractState, error) {
	storage := states.cache.get(aid)
	if storage == nil {
		root := common.Compactz(st.StorageRoot)
		storage = newBufferedStorage(root, states.store)
	}
	res := &ContractState{
		State:   st,
		account: aid,
		storage: storage,
		store:   states.store,
	}
	return res, nil
}

func (states *StateDB) StageContractState(st *ContractState) error {
	states.cache.put(st.account, st.storage)
	st.storage = nil
	return nil
}

// GetSystemAccountState returns the ContractState of the AERGO system account.
func (states *StateDB) GetSystemAccountState() (*ContractState, error) {
	return states.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
}

// GetNameAccountState returns the ContractState of the AERGO name account.
func (states *StateDB) GetNameAccountState() (*ContractState, error) {
	return states.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoName)))
}

// GetEnterpriseAccountState returns the ContractState of the AERGO enterprise account.
func (states *StateDB) GetEnterpriseAccountState() (*ContractState, error) {
	return states.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoEnterprise)))
}

type ContractState struct {
	*types.State
	account types.AccountID
	code    []byte
	storage *bufferedStorage
	store   db.DB
}

func (st *ContractState) SetNonce(nonce uint64) {
	st.State.Nonce = nonce
}
func (st *ContractState) GetNonce() uint64 {
	return st.State.GetNonce()
}

func (st *ContractState) SetBalance(balance *big.Int) {
	st.State.Balance = balance.Bytes()
}
func (st *ContractState) GetBalance() *big.Int {
	return new(big.Int).SetBytes(st.State.GetBalance())
}

func (st *ContractState) SetCode(code []byte) error {
	codeHash := common.Hasher(code)
	storedCode, err := st.GetRawKV(codeHash[:])
	if err == nil && !bytes.Equal(code, storedCode) {
		err = st.SetRawKV(codeHash[:], code)
	}
	if err != nil {
		return err
	}
	st.State.CodeHash = codeHash[:]
	st.code = code
	return nil
}

func (st *ContractState) GetCode() ([]byte, error) {
	if st.code != nil {
		// already loaded.
		return st.code, nil
	}
	codeHash := st.State.GetCodeHash()
	if codeHash == nil {
		// not defined. do nothing.
		return nil, nil
	}
	err := loadData(st.store, st.State.CodeHash, &st.code)
	if err != nil {
		return nil, err
	}
	return st.code, nil
}

func (st *ContractState) GetAccountID() types.AccountID {
	return st.account
}

// SetRawKV saves (key, value) to st.store without any kind of encoding.
func (st *ContractState) SetRawKV(key []byte, value []byte) error {
	return saveData(st.store, key, value)
}

// GetRawKV loads (key, value) from st.store.
func (st *ContractState) GetRawKV(key []byte) ([]byte, error) {
	var b []byte
	if err := loadData(st.store, key, &b); err != nil {
		return nil, err
	}
	return b, nil
}

// HasKey returns existence of the key
func (st *ContractState) HasKey(key []byte) bool {
	return st.storage.has(types.GetHashID(key), true)
}

// SetData store key and value pair to the storage.
func (st *ContractState) SetData(key, value []byte) error {
	st.storage.put(newValueEntry(types.GetHashID(key), value))
	return nil
}

// GetData returns the value corresponding to the key from the buffered storage.
func (st *ContractState) GetData(key []byte) ([]byte, error) {
	id := types.GetHashID(key)
	if entry := st.storage.get(id); entry != nil {
		if value := entry.Value(); value != nil {
			return value.([]byte), nil
		}
		return nil, nil
	}
	return st.getInitialData(id[:])
}

func (st *ContractState) getInitialData(id []byte) ([]byte, error) {
	dkey, err := st.storage.trie.Get(id)
	if err != nil {
		return nil, err
	}
	if len(dkey) == 0 {
		return nil, nil
	}
	value := []byte{}
	if err := loadData(st.store, dkey, &value); err != nil {
		return nil, err
	}
	return value, nil
}

// GetInitialData returns the value corresponding to the key from the contract storage.
func (st *ContractState) GetInitialData(key []byte) ([]byte, error) {
	id := types.GetHashID(key)
	return st.getInitialData(id[:])
}

// DeleteData remove key and value pair from the storage.
func (st *ContractState) DeleteData(key []byte) error {
	st.storage.put(newValueEntryDelete(types.GetHashID(key)))
	return nil
}

// Snapshot returns revision number of storage buffer
func (st *ContractState) Snapshot() Snapshot {
	return Snapshot(st.storage.buffer.snapshot())
}

// Rollback discards changes of storage buffer to revision number
func (st *ContractState) Rollback(revision Snapshot) error {
	return st.storage.buffer.rollback(int(revision))
}

// Hash implements types.ImplHashBytes
func (st *ContractState) Hash() []byte {
	return getHashBytes(st.State)
}

// Marshal implements types.ImplMarshal
func (st *ContractState) Marshal() ([]byte, error) {
	return proto.Marshal(st.State)
}

func (st *ContractState) cache() *stateBuffer {
	return st.storage.buffer
}
