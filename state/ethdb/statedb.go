package ethdb

import (
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

func (sdb *StateDB) PutState(id []byte, addr common.Address, balance *big.Int, nonce uint64, code []byte) {
	sdb.ethStateDB.SetNonce(addr, nonce)
	sdb.ethStateDB.SetBalance(addr, balance)

	// id must be 33 bytes
	idWithCode := make([]byte, types.AddressLength+len(code))
	copy(idWithCode, id)
	copy(idWithCode[types.AddressLength:], code)

	sdb.ethStateDB.SetCode(addr, idWithCode)
}

func (sdb *StateDB) GetState(addr common.Address) (id []byte, balance *big.Int, nonce uint64, code []byte) {
	idWithCode := sdb.ethStateDB.GetCode(addr)
	id = sdb.GetId(addr)
	balance = sdb.ethStateDB.GetBalance(addr)
	nonce = sdb.ethStateDB.GetNonce(addr)
	code = idWithCode[types.AddressLength:]
	return id, balance, nonce, code
}

func (sdb *StateDB) GetId(addr common.Address) []byte {
	if id := types.GetSpecialAccountEthReverse(addr); id != "" {
		return []byte(id)
	}

	idWithCode := sdb.ethStateDB.GetCode(addr)
	return idWithCode[:types.AddressLength]
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
