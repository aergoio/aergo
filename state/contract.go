package state

import (
	"crypto/sha256"

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
		State: st,
	}
	return res, nil
}

func (sdb *ChainStateDB) CommitContractState(st *ContractState) error {
	err := st.storage.Commit()
	if err != nil {
		return err
	}
	st.State.StorageRoot = st.storage.Root
	return nil
}

type ContractState struct {
	*types.State
	code    []byte
	storage *trie.Trie
	dbstore *db.DB
}

func (st *ContractState) loadStorage() error {
	hasher := types.GetTrieHasher()
	storage := trie.NewTrie(32, hasher, *st.dbstore)
	root := st.State.StorageRoot
	if root != nil {
		err := storage.LoadCache(root)
		if err != nil {
			return err
		}
	}
	st.storage = storage
	return nil
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
	err := saveData(st.dbstore, st.State.CodeHash, code)
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
	err := loadData(st.dbstore, st.State.CodeHash, st.code)
	if err != nil {
		return nil, err
	}
	return st.code, nil
}

func (st *ContractState) SetData(key, value []byte) error {
	if st.storage == nil {
		err := st.loadStorage()
		if err != nil {
			return err
		}
	}
	hkey := sha256.Sum256(key)
	_, err := st.storage.Update(trie.DataArray{hkey[:]}, trie.DataArray{value})
	if err != nil {
		return err
	}
	st.State.StorageRoot = st.storage.Root
	return nil
}
func (st *ContractState) GetData(key []byte) ([]byte, error) {
	if st.storage == nil {
		err := st.loadStorage()
		if err != nil {
			return nil, err
		}
	}
	hkey := sha256.Sum256(key)
	value, err := st.storage.Get(hkey[:])
	if err != nil {
		return nil, err
	}
	return value, nil
}
