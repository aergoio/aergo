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
	key = convNilToBytes(key)
	value = convNilToBytes(value)
	err := db.store.Set(key, value, pebble.Sync)
	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}
}

func (db *PebbleDB) Delete(key []byte) {
	key = convNilToBytes(key)
	err := db.store.Delete(key, pebble.Sync)
	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}
}

func (db *PebbleDB) Get(key []byte) []byte {
	key = convNilToBytes(key)
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
	batch     *pebble.Batch
	setCount  uint
	delCount  uint
	keySize   uint64
	valueSize uint64
}

func (tx *pebbleTransaction) Set(key, value []byte) {
	key = convNilToBytes(key)
	value = convNilToBytes(value)

	err := tx.batch.Set(key, value, pebble.Sync)
	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}

	tx.setCount++
	tx.keySize += uint64(len(key))
	tx.valueSize += uint64(len(value))
}

func (tx *pebbleTransaction) Delete(key []byte) {
	key = convNilToBytes(key)

	err := tx.batch.Delete(key, pebble.Sync)
	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}

	tx.delCount++
}

func (tx *pebbleTransaction) Commit() {
	err := tx.batch.Commit(pebble.Sync)
	if err != nil {
		panic(err)
	}
}

func (tx *pebbleTransaction) Discard() {
	tx.batch.Close()
}

//=========================================================
// Bulk Implementation
//=========================================================

type pebbleBulk struct {
	batch     *pebble.Batch
	setCount  uint
	delCount  uint
	keySize   uint64
	valueSize uint64
}

func (bulk *pebbleBulk) Set(key, value []byte) {
	key = convNilToBytes(key)
	value = convNilToBytes(value)

	err := bulk.batch.Set(key, value, pebble.Sync)
	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}

	bulk.setCount++
	bulk.keySize += uint64(len(key))
	bulk.valueSize += uint64(len(value))
}

func (bulk *pebbleBulk) Delete(key []byte) {
	key = convNilToBytes(key)

	err := bulk.batch.Delete(key, pebble.Sync)
	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}

	bulk.delCount++
}

func (bulk *pebbleBulk) Flush() {
	err := bulk.batch.Commit(pebble.Sync)
	if err != nil {
		panic(err)
	}
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

func convNilToBytes(byteArray []byte) []byte {
	if byteArray == nil {
		return []byte{}
	}
	return byteArray
}
