package statedb

import (
	"github.com/aergoio/aergo/v2/types"
)

type DumpAccount struct {
	State   *types.State
	Code    []byte
	Storage map[types.AccountID][]byte
}

type Dump struct {
	Root     []byte                          `json:"root"`
	Accounts map[types.AccountID]DumpAccount `json:"accounts"`
}

func (sdb *StateDB) RawDump() (Dump, error) {
	var err error

	self := sdb.Clone()
	dump := Dump{
		Root:     self.GetRoot(),
		Accounts: make(map[types.AccountID]DumpAccount),
	}

	for _, key := range self.Trie.GetKeys() {
		var st *types.State
		var code []byte
		var storage map[types.AccountID][]byte = make(map[types.AccountID][]byte)

		// load account state
		aid := types.AccountID(types.ToHashID(key))
		st, err = sdb.getState(aid)
		if err != nil {
			return dump, err
		}
		if len(st.GetCodeHash()) > 0 {
			// load code
			loadData(self.Store, st.GetCodeHash(), &code)

			// load contract
			cs, err := OpenContractState(key, st, self)
			if err != nil {
				return dump, err
			}

			// load storage
			for _, key := range cs.storage.Trie.GetKeys() {
				data, _ := cs.getInitialData(key)
				aid := types.AccountID(types.ToHashID(key))
				storage[aid] = data
			}
		}

		dump.Accounts[aid] = DumpAccount{
			State:   st,
			Code:    code,
			Storage: storage,
		}
	}

	return dump, nil
}
