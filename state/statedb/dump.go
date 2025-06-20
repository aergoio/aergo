package statedb

import (
	"encoding/json"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

type DumpAccount struct {
	State   *types.State
	Code    []byte
	Storage map[types.AccountID][]byte
}

func (d DumpAccount) MarshalJSON() ([]byte, error) {
	mapState := make(map[string]interface{})
	mapState["nonce"] = d.State.Nonce
	mapState["balance"] = d.State.GetBalanceBigInt().String()
	mapState["codeHash"] = base58.Encode(d.State.CodeHash)
	mapState["storageRoot"] = base58.Encode(d.State.StorageRoot)
	mapState["sqlRecoveryPoint"] = d.State.SqlRecoveryPoint

	mapStorage := make(map[string]string)
	for k, v := range d.Storage {
		mapStorage[k.String()] = base58.Encode(v)
	}

	return json.Marshal(map[string]interface{}{
		"state":   mapState,
		"code":    string(d.Code),
		"storage": mapStorage,
	})
}

type Dump struct {
	Root     []byte                          `json:"root"`
	Accounts map[types.AccountID]DumpAccount `json:"accounts"`
}

func (d Dump) MarshalJSON() ([]byte, error) {
	mapAccounts := make(map[string]DumpAccount)
	for k, v := range d.Accounts {
		mapAccounts[k.String()] = v
	}

	return json.Marshal(map[string]interface{}{
		"root":     base58.Encode(d.Root),
		"accounts": mapAccounts,
	})
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
		storage := make(map[types.AccountID][]byte)

		// load account state
		aid := types.AccountID(types.ToHashID(key))
		st, err = sdb.getState(aid)
		if err != nil {
			return dump, err
		}
		if len(st.GetCodeHash()) > 0 {
			// load code
			loadData(self.Store, st.GetCodeHash(), &code)

			// load contract state
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

type RawDumpConfig struct {
	ShowCode bool
}

func DefaultConfig() RawDumpConfig {
	return RawDumpConfig{
		ShowCode: false,
	}
}

type Processor func(idx int64, accountId types.AccountID, account *DumpAccount) error

// RawDumpWith iterate accounts information and let the processor handle them.
func (sdb *StateDB) RawDumpWith(config RawDumpConfig, processor Processor) error {
	var err error
	skipCode := []byte("(skipped)")

	self := sdb.Clone()

	// write accounts
	for idx, key := range self.Trie.GetKeys() {
		var st *types.State
		var code []byte
		storage := make(map[types.AccountID][]byte)

		// load account state
		aid := types.AccountID(types.ToHashID(key))
		st, err = sdb.getState(aid)
		if err != nil {
			return err
		}
		if len(st.GetCodeHash()) > 0 {
			if config.ShowCode {
				// load code
				loadData(self.Store, st.GetCodeHash(), &code)
			} else {
				code = skipCode
			}

			// load contract state
			cs, err := OpenContractState(key, st, self)
			if err != nil {
				return err
			}

			// load storage
			for _, key := range cs.storage.Trie.GetKeys() {
				data, _ := cs.getInitialData(key)
				aid := types.AccountID(types.ToHashID(key))
				storage[aid] = data
			}
		}

		dumpAccount := DumpAccount{
			State:   st,
			Code:    code,
			Storage: storage,
		}
		err = processor(int64(idx), aid, &dumpAccount)
		if err != nil {
			return err
		}
	}

	return nil
}
