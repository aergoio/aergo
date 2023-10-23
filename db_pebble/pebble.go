package db_pebble

import (
	"bytes"
	"fmt"

	"github.com/aergoio/aergo-lib/db"
	"github.com/cockroachdb/pebble"
)

func NewPebbleDB(dir string) db.DB {
	db := &PebbleDB{}
	err := db.Open(dir, nil)
	if err != nil {
		panic(err)
	}
	return db
}

//=========================================================
// DB Implementation
//=========================================================

// Enforce database and transaction implements interfaces
var _ db.DB = (*PebbleDB)(nil)

type PebbleDB struct {
	store *pebble.DB
}

func (db *PebbleDB) Type() string {
	return "pebbledb"
}

func (db *PebbleDB) Open(path string, opt *pebble.Options) error {
	var err error
	db.store, err = pebble.Open(path, opt)
	if err != nil {
		return err
	}
	return nil
}

func (db *PebbleDB) Set(key, value []byte) {
	db.store.Set(key, value, pebble.Sync)
}

func (db *PebbleDB) Delete(key []byte) {
	db.store.Delete(key, pebble.Sync)
}

func (db *PebbleDB) Get(key []byte) []byte {
	if key == nil {
		key = []byte{}
	}
	value, closer, err := db.store.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return []byte{}
		}
		panic(fmt.Sprintf("Database Error: %v", err))
	}
	defer closer.Close()

	return value
}

func (db *PebbleDB) Exist(key []byte) bool {
	return db.Get(key) != nil
}

func (db *PebbleDB) NewTx() db.Transaction {
	batch := db.store.NewBatch()

	return &pebbleTransaction{
		batch: batch,
	}
}

func (db *PebbleDB) NewBulk() db.Bulk {
	batch := db.store.NewBatch()

	return &pebbleBulk{
		batch: batch,
	}
}

func (db *PebbleDB) Iterator(start, end []byte) db.Iterator {
	var reverse bool
	if bytes.Compare(start, end) == 1 {
		reverse = true
	} else {
		reverse = false
	}

	iter, err := db.store.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		return nil
	}

	if reverse {
		iter.First()
	} else {
		iter.Last()
	}
	return &pebbleIterator{
		start:   start,
		end:     end,
		reverse: reverse,
		iter:    iter,
	}
}

func (db *PebbleDB) Close() {
	db.store.Close()
}

//=========================================================
// Transaction Implementation
//=========================================================

type pebbleTransaction struct {
	batch *pebble.Batch
}

func (tx *pebbleTransaction) Set(key, value []byte) {
	tx.batch.Set(key, value, pebble.Sync)
}

func (tx *pebbleTransaction) Delete(key []byte) {
	tx.batch.Delete(key, pebble.Sync)
}

func (tx *pebbleTransaction) Commit() {
	tx.batch.Commit(pebble.Sync)
}

func (tx *pebbleTransaction) Discard() {
	tx.batch.Close()
}

//=========================================================
// Bulk Implementation
//=========================================================

type pebbleBulk struct {
	batch *pebble.Batch
}

func (bulk *pebbleBulk) Set(key, value []byte) {
	bulk.batch.Set(key, value, pebble.Sync)
}

func (bulk *pebbleBulk) Delete(key []byte) {
	bulk.batch.Delete(key, pebble.Sync)
}

func (bulk *pebbleBulk) Flush() {
	bulk.batch.Commit(pebble.Sync)
}

func (bulk *pebbleBulk) DiscardLast() {
	bulk.batch.Close()
}

//=========================================================
// Iterator Implementation
//=========================================================

type pebbleIterator struct {
	start   []byte
	end     []byte
	reverse bool
	iter    *pebble.Iterator
}

func (i *pebbleIterator) Next() {
	if i.reverse {
		i.iter.Prev()
	} else {
		i.iter.Next()
	}
}

func (i *pebbleIterator) Valid() bool {
	return i.iter.Valid()
}

func (i *pebbleIterator) Key() []byte {
	return i.iter.Key()
}

func (i *pebbleIterator) Value() []byte {
	return i.iter.Value()
}
