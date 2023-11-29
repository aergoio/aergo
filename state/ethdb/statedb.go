package ethdb

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
)

const (
	StateName = "evm_state"
)

type StateDB struct {
	blockNo    uint64
	evmStateDB *state.StateDB
}

func NewStateDB(blockNo uint64, evmRoot []byte, db *DB) (*StateDB, error) {
	sdb, err := state.New(
		common.BytesToHash(evmRoot),
		state.NewDatabaseWithNodeDB(db.Store, db.Triedb),
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &StateDB{
		blockNo:    blockNo,
		evmStateDB: sdb,
	}, nil
}

func (sdb *StateDB) Close() {
	sdb.evmStateDB.StopPrefetcher()
}

func (sdb *StateDB) PutState(addr common.Address, balance *big.Int, nonce uint64) int {
	sdb.evmStateDB.CreateAccount(addr)
	sdb.evmStateDB.SetNonce(addr, nonce)
	sdb.evmStateDB.SetBalance(addr, balance)
	return 0
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

func (sdb *StateDB) Commit() (root []byte, err error) {
	newRoot, err := sdb.evmStateDB.Commit(sdb.blockNo, false)
	if err != nil {
		return nil, err
	}
	return newRoot.Bytes(), nil
}
