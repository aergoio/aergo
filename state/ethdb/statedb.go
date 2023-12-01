package ethdb

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/trie"
)

const (
	StateName = "evm_state"
)

type StateDB struct {
	trieDB     *trie.Database
	evmStateDB *state.StateDB
}

func NewStateDB(root []byte, db *DB) (*StateDB, error) {
	triedb := trie.NewDatabase(db.Store, &trie.Config{Preimages: true})
	statedb := state.NewDatabaseWithNodeDB(db.Store, triedb)

	sdb, err := state.New(common.BytesToHash(root), statedb, nil)
	if err != nil {
		return nil, err
	}

	return &StateDB{
		trieDB:     triedb,
		evmStateDB: sdb,
	}, nil
}

func (sdb *StateDB) Copy() *StateDB {
	return &StateDB{
		trieDB:     sdb.trieDB,
		evmStateDB: sdb.evmStateDB.Copy(),
	}
}

func (sdb *StateDB) GetStateDB() *state.StateDB {
	return sdb.evmStateDB
}

func (sdb *StateDB) PutState(addr common.Address, balance *big.Int, nonce uint64, code []byte) {
	sdb.evmStateDB.SetNonce(addr, nonce)
	sdb.evmStateDB.SetBalance(addr, balance)
	if len(code) > 0 {
		sdb.evmStateDB.SetCode(addr, code)
	}
}

func (sdb *StateDB) GetState(addr common.Address) (balance *big.Int, nonce uint64, code []byte) {
	return sdb.evmStateDB.GetBalance(addr), sdb.evmStateDB.GetNonce(addr), sdb.evmStateDB.GetCode(addr)
}

func (sdb *StateDB) Root() []byte {
	return sdb.evmStateDB.IntermediateRoot(false).Bytes()
}

func (sdb *StateDB) Snapshot() int {
	return sdb.evmStateDB.Snapshot()
}

func (sdb *StateDB) Rollback(snapshot int) {
	sdb.evmStateDB.RevertToSnapshot(snapshot)
}

func (sdb *StateDB) Commit(blockNo uint64) (root []byte, err error) {
	newRoot, err := sdb.evmStateDB.Commit(blockNo, false)
	if err != nil {
		return nil, err
	}
	err = sdb.trieDB.Commit(newRoot, false)
	if err != nil {
		return nil, err
	}

	return newRoot.Bytes(), nil
}
