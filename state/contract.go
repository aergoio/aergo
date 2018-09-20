package state

import (
	"bytes"

	sha256 "github.com/minio/sha256-simd"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/pkg/trie"
	"github.com/aergoio/aergo/types"
)

func (sdb *ChainStateDB) OpenContractStateAccount(aid types.AccountID) (*ContractState, error) {
	st, err := sdb.GetAccountStateClone(aid)
	if err != nil {
		return nil, err
	}
	return sdb.OpenContractState(st)
}
func (sdb *ChainStateDB) OpenContractState(st *types.State) (*ContractState, error) {
	res := &ContractState{
		State:   st,
		storage: trie.NewTrie(nil, types.TrieHasher, sdb.store),
		buffer:  newStateBuffer(),
		store:   &sdb.store,
	}
	if st.StorageRoot != nil && !emptyHashID.Equal(types.ToHashID(st.StorageRoot)) {
		res.storage.Root = st.StorageRoot
	}
	return res, nil
}

func (sdb *ChainStateDB) CommitContractState(st *ContractState) error {
	defer func() {
		if bytes.Compare(st.State.StorageRoot, st.storage.Root) != 0 {
			st.State.StorageRoot = st.storage.Root
		}
		st.storage = nil
	}()

	if st.buffer.isEmpty() {
		// do nothing
		return nil
	}

	keys, vals := st.buffer.export()
	_, err := st.storage.Update(keys, vals)
	if err != nil {
		return err
	}
	st.buffer.commit(st.store)

	err = st.storage.Commit()
	if err != nil {
		return err
	}
	return st.buffer.reset()
}

type ContractState struct {
	*types.State
	code    []byte
	storage *trie.Trie
	buffer  *stateBuffer
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
	return st.buffer.put(types.GetHashID(key), value)
}

func (st *ContractState) GetData(key []byte) ([]byte, error) {
	id := types.GetHashID(key)
	entry := st.buffer.get(id)
	if entry != nil {
		return entry.data.([]byte), nil
	}
	dkey, err := st.storage.Get(id[:])
	if err != nil {
		return nil, err
	}
	if len(dkey) == 0 {
		return nil, nil
	}
	value := []byte{}
	err = loadData(st.store, dkey, &value)
	if err != nil {
		return nil, err
	}
	return value, nil
}
