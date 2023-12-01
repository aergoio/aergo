package ethdb

import (
	"fmt"

	"github.com/aergoio/aergo/v2/types/dbkey"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

type DB struct {
	dbType string
	Store  ethdb.Database
	Triedb *trie.Database
}

func NewDB(path string, dbType string) (*DB, error) {
	ethdb := &DB{}

	var err error
	ethdb.dbType = dbType
	switch dbType {
	case "memorydb":
		ethdb.Store = rawdb.NewMemoryDatabase()
	case "leveldb", "badgerdb": // FIXME : badgerdb is not supported yet
		ethdb.Store, err = rawdb.NewLevelDBDatabase(path, 128, 1024, "", false)
	case "pebbledb":
		ethdb.Store, err = rawdb.NewPebbleDBDatabase(path, 128, 1024, "", false, false)
	default:
		err = fmt.Errorf("unsupported db type: %s", dbType)
	}
	if err != nil {
		return nil, err
	}
	ethdb.Triedb = trie.NewDatabase(ethdb.Store, trie.HashDefaults)

	return ethdb, nil
}

func (db *DB) Close() error {
	return db.Store.Close()
}

// TODO : before put eth root in block, it can saved in eth db
func (db *DB) SetEthRoot(root []byte) {
	db.Store.Put(dbkey.EthRootHash(), root)
}

func (db *DB) GetEthRoot() []byte {
	root, _ := db.Store.Get(dbkey.EthRootHash())
	return root
}
