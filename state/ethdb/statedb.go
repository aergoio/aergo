package ethdb

import (
	"bytes"
	"math/big"

	"github.com/aergoio/aergo/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/trie"
)

const (
	StateName = "state_evm"
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

func (sdb *StateDB) PutState(id []byte, addr common.Address, balance *big.Int, nonce uint64, code []byte) {
	sdb.evmStateDB.SetNonce(addr, nonce)
	sdb.evmStateDB.SetBalance(addr, balance)

	// id must be 33 bytes
	idWithCode := make([]byte, types.AddressLength+len(code))
	copy(idWithCode, id)
	copy(idWithCode[types.AddressLength:], code)

	sdb.evmStateDB.SetCode(addr, idWithCode)
}

func (sdb *StateDB) GetState(addr common.Address) (id []byte, balance *big.Int, nonce uint64, code []byte) {
	idWithCode := sdb.evmStateDB.GetCode(addr)
	id = idWithCode[:types.AddressLength]
	if bytes.Contains(idWithCode, []byte("aergo.")) { // FIXME : check governance id
		id = bytes.TrimRight(idWithCode, "\x00")
	}
	balance = sdb.evmStateDB.GetBalance(addr)
	nonce = sdb.evmStateDB.GetNonce(addr)
	code = idWithCode[types.AddressLength:]
	return id, balance, nonce, code
}

func (sdb *StateDB) GetId(addr common.Address) (id []byte) {
	idWithCode := sdb.evmStateDB.GetCode(addr)
	id = idWithCode[:types.AddressLength]
	if bytes.Contains(idWithCode, []byte("aergo.")) { // FIXME : check governance id
		id = bytes.TrimRight(idWithCode, "\x00")
	}
	return id
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
