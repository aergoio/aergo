package ethdb

import (
	"github.com/aergoio/aergo/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/trie"
)

const (
	StateName = "state_eth"
)

type StateDB struct {
	db         *DB
	trieDB     *trie.Database
	ethStateDB *state.StateDB
}

func NewStateDB(root []byte, db *DB) (*StateDB, error) {
	triedb := trie.NewDatabase(db.Store, &trie.Config{Preimages: true})
	statedb := state.NewDatabaseWithNodeDB(db.Store, triedb)

	sdb, err := state.New(common.BytesToHash(root), statedb, nil)
	if err != nil {
		return nil, err
	}

	return &StateDB{
		db:         db,
		trieDB:     triedb,
		ethStateDB: sdb,
	}, nil
}

func (sdb *StateDB) Copy() *StateDB {
	return &StateDB{
		trieDB:     sdb.trieDB,
		ethStateDB: sdb.ethStateDB.Copy(),
	}
}

func (sdb *StateDB) GetStateDB() *state.StateDB {
	return sdb.ethStateDB
}

func (sdb *StateDB) PutState(addr common.Address, state *types.State) {
	sdb.ethStateDB.SetBalance(addr, state.GetBalanceBigInt())
	if state.GetNonce() != 0 {
		sdb.ethStateDB.SetNonce(addr, state.GetNonce())
	}
	if state.GetCodeHash() != nil {
		sdb.ethStateDB.SetCode(addr, state.GetCodeHash())
	}
}

func (sdb *StateDB) GetState(addr common.Address) (state *types.State) {
	if !sdb.ethStateDB.Exist(addr) {
		return nil
	}
	code := sdb.ethStateDB.GetCode(addr)
	balance := sdb.ethStateDB.GetBalance(addr)
	nonce := sdb.ethStateDB.GetNonce(addr)
	return &types.State{
		Balance:  balance.Bytes(),
		Nonce:    nonce,
		CodeHash: code,
	}
}

func (sdb *StateDB) PutId(addr common.Address, id []byte) {
	sdb.db.PutAddrId(addr, id)
}

func (sdb *StateDB) GetId(addr common.Address) []byte {
	return sdb.db.GetAddrId(addr)
}

func (sdb *StateDB) Root() []byte {
	return sdb.ethStateDB.IntermediateRoot(false).Bytes()
}

func (sdb *StateDB) Snapshot() int {
	return sdb.ethStateDB.Snapshot()
}

func (sdb *StateDB) Rollback(snapshot int) {
	sdb.ethStateDB.RevertToSnapshot(snapshot)
}

func (sdb *StateDB) Commit(blockNo uint64) (root []byte, err error) {
	newRoot, err := sdb.ethStateDB.Commit(blockNo, false)
	if err != nil {
		return nil, err
	}
	err = sdb.trieDB.Commit(newRoot, false)
	if err != nil {
		return nil, err
	}

	return newRoot.Bytes(), nil
}
