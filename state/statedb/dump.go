package statedb

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

type DumpAccount struct {
	State   *types.State
	Code    []byte
	Storage map[types.AccountID][]byte
}

type Dump struct {
	Root     string                          `json:"root"`
	Accounts map[types.AccountID]DumpAccount `json:"accounts"`
}

func (sdb *StateDB) RawDump() (Dump, error) {
	self := sdb.Clone()
	dump := Dump{
		Root:     base58.Encode(self.GetRoot()),
		Accounts: make(map[types.AccountID]DumpAccount),
	}

	keys := self.Trie.GetKeys()
	for _, key := range keys {
		aid := types.AccountID(types.ToHashID(key))

		// load accoutn state
		st, err := sdb.getState(aid)
		if err != nil {
			return dump, err
		}
		var code []byte
		var storage map[types.AccountID][]byte = make(map[types.AccountID][]byte)
		if len(st.GetCodeHash()) > 0 {
			// load code
			loadData(self.Store, st.GetCodeHash(), &code)

			// load contract
			cs, err := OpenContractState(key, st, self)
			if err != nil {
				return dump, err
			}
			keys := cs.storage.Trie.GetKeys()
			for _, key := range keys {
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
