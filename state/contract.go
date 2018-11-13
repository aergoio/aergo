package state

import (
	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	sha256 "github.com/minio/sha256-simd"
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
		storage = newBufferedStorage(root, *states.store)
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

type ContractState struct {
	*types.State
	account types.AccountID
	code    []byte
	storage *bufferedStorage
	store   *db.DB
}

func (st *ContractState) SetNonce(nonce uint64) {
	st.State.Nonce = nonce
}
func (st *ContractState) GetNonce() uint64 {
	return st.State.GetNonce()
}

func (st *ContractState) SetBalance(balance uint64) {
	st.State.Balance = balance
}
func (st *ContractState) GetBalance() uint64 {
	return st.State.GetBalance()
}

func (st *ContractState) SetCode(code []byte) error {
	codeHash := sha256.Sum256(code)
	err := saveData(st.store, codeHash[:], &code)
	if err != nil {
		return err
	}
	st.State.CodeHash = codeHash[:]
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

func (st *ContractState) SetData(key, value []byte) error {
	st.storage.put(newValueEntry(types.GetHashID(st.account[:], key), value))
	return nil
}

func (st *ContractState) GetData(key []byte) ([]byte, error) {
	id := types.GetHashID(st.account[:], key)
	if entry := st.storage.get(id); entry != nil {
		return entry.Value().([]byte), nil
	}
	dkey, err := st.storage.trie.Get(id[:])
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

// Snapshot returns revision number of storage buffer
func (st *ContractState) Snapshot() Snapshot {
	return Snapshot(st.storage.buffer.snapshot())
}

// Rollback discards changes of storage buffer to revision number
func (st *ContractState) Rollback(revision Snapshot) error {
	return st.storage.buffer.rollback(int(revision))
}

// HashID implements types.ImplHashID
func (st *ContractState) HashID() types.HashID {
	return getHash(st.State)
}

// Marshal implements types.ImplMarshal
func (st *ContractState) Marshal() ([]byte, error) {
	return proto.Marshal(st.State)
}

func (st *ContractState) cache() *stateBuffer {
	return st.storage.buffer
}
