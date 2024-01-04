package ethdb

import (
	hash "github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/trie"
)

const (
	StateName = "state_eth"
)

type StateDB struct {
	db     *DB
	trieDB *trie.Database
	*state.StateDB

	// These maps hold the state id that occurred in this **block**.
	accountId map[common.Address][]byte
}

func NewStateDB(root []byte, db *DB) (*StateDB, error) {
	triedb := trie.NewDatabase(db.Store, &trie.Config{Preimages: true})
	statedb := state.NewDatabaseWithNodeDB(db.Store, triedb)

	sdb, err := state.New(common.BytesToHash(root), statedb, nil)
	if err != nil {
		return nil, err
	}

	return &StateDB{
		db:      db,
		trieDB:  triedb,
		StateDB: sdb,
	}, nil
}

func (sdb *StateDB) Copy() *StateDB {
	return &StateDB{
		trieDB:  sdb.trieDB,
		StateDB: sdb.StateDB.Copy(),
	}
}

func (sdb *StateDB) GetStateDB() *state.StateDB {
	return sdb.StateDB
}

func (sdb *StateDB) PutCode(addr common.Address, code []byte) {
	sdb.StateDB.SetCode(addr, code)
}

func (sdb *StateDB) Put(addr common.Address, state *types.State) {
	sdb.StateDB.SetBalance(addr, state.GetBalanceBigInt())
	if state.GetNonce() != 0 {
		sdb.StateDB.SetNonce(addr, state.GetNonce())
	}
}

func (sdb *StateDB) Get(addr common.Address) (state *types.State) {
	if !sdb.StateDB.Exist(addr) {
		return nil
	}
	balance := sdb.StateDB.GetBalance(addr)
	nonce := sdb.StateDB.GetNonce(addr)
	var codeHash []byte
	if code := sdb.StateDB.GetCode(addr); len(code) != 0 {
		codeHash = hash.Hasher(code)
	}

	return &types.State{
		Balance:  balance.Bytes(),
		Nonce:    nonce,
		CodeHash: codeHash,
	}
}

func (sdb *StateDB) PutId(addr common.Address, id []byte) {
	sdb.db.PutAddrId(addr, id)
}

func (sdb *StateDB) GetId(addr common.Address) []byte {
	return sdb.db.GetAddrId(addr)
}

func (sdb *StateDB) Root() []byte {
	return sdb.StateDB.IntermediateRoot(false).Bytes()
}

func (sdb *StateDB) Snapshot() int {
	return sdb.StateDB.Snapshot()
}

func (sdb *StateDB) Rollback(snapshot int) {
	sdb.StateDB.RevertToSnapshot(snapshot)
}

func (sdb *StateDB) Commit(blockNo uint64) (root []byte, err error) {
	newRoot, err := sdb.StateDB.Commit(blockNo, false)
	if err != nil {
		return nil, err
	}
	err = sdb.trieDB.Commit(newRoot, false)
	if err != nil {
		return nil, err
	}

	/* 	for account, id := range sdb.accountId {

	   	}
	*/
	return newRoot.Bytes(), nil
}
