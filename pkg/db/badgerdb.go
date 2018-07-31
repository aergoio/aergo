/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package db

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
)

const (
	badgerDbDiscardRatio = 0.5
	badgerDbGcInterval   = 10 * time.Minute
)

// This function is always called first
func init() {
	dbConstructor := func(dir string) (DB, error) {
		return NewBadgerDB(dir)
	}
	registorDBConstructor(BadgerImpl, dbConstructor)
}

func (db *badgerDB) runBadgerGC() {
	ticker := time.NewTicker(badgerDbGcInterval)
	for {
		select {
		case <-ticker.C:
			err := db.db.RunValueLogGC(badgerDbDiscardRatio)

			if err != badger.ErrNoRewrite {
				panic(err)
			}

		case <-db.ctx.Done():
			return
		}
	}
}

func NewBadgerDB(dir string) (DB, error) {
	// set option file
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	// TODO : options tuning.
	// Quick fix to prevent RAM usage from going to the roof when adding 10Million new keys during tests
	opts.ValueLogLoadingMode = options.FileIO
	opts.TableLoadingMode = options.FileIO

	// open badger db
	db, err := badger.Open(opts)

	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	database := &badgerDB{
		db:         db,
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}

	go database.runBadgerGC()

	return database, nil
}

//=========================================================
// DB Implementation
//=========================================================

// Enforce database and transaction implements interfaces
var _ DB = (*badgerDB)(nil)

type badgerDB struct {
	db         *badger.DB
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// Type function returns a database type name
func (db *badgerDB) Type() string {
	return "badgerdb"
}

func (db *badgerDB) Set(key, value []byte) {
	key = convNilToBytes(key)
	value = convNilToBytes(value)

	err := db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})

	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}
}

func (db *badgerDB) Delete(key []byte) {
	key = convNilToBytes(key)

	err := db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})

	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}
}

func (db *badgerDB) Get(key []byte) []byte {
	key = convNilToBytes(key)

	var val []byte
	err := db.db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		getVal, err := item.Value()
		if err != nil {
			return err
		}

		val = getVal

		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return []byte{}
		}
		panic(fmt.Sprintf("Database Error: %v", err))
	}

	return val
}

func (db *badgerDB) Exist(key []byte) bool {
	key = convNilToBytes(key)

	var isExist bool

	err := db.db.View(func(txn *badger.Txn) error {

		_, err := txn.Get(key)
		if err != nil {
			return err
		}

		isExist = true

		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false
		}
	}

	return isExist
}

func (db *badgerDB) Close() {

	db.cancelFunc() // wait until gc goroutine is finished

	err := db.db.Close()
	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}
}

func (db *badgerDB) NewTx(writable bool) Transaction {
	badgerTx := db.db.NewTransaction(writable)

	retTransaction := &badgerTransaction{
		db: db,
		tx: badgerTx,
	}

	return retTransaction
}

//=========================================================
// Transaction Implementation
//=========================================================

type badgerTransaction struct {
	db *badgerDB
	tx *badger.Txn
}

func (transaction *badgerTransaction) Get(key []byte) []byte {
	key = convNilToBytes(key)

	getVal, err := transaction.tx.Get(key)

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return []byte{}
		}
		panic(fmt.Sprintf("Database Error: %v", err))
	}

	val, err := getVal.Value()
	if err != nil {
		//TODO handle retry error??
		panic(fmt.Sprintf("Database Error: %v", err))
	}

	return val
}

func (transaction *badgerTransaction) Set(key, value []byte) {
	// TODO Updating trie nodes may require many updates but ErrTxnTooBig is not handled
	key = convNilToBytes(key)
	value = convNilToBytes(value)

	err := transaction.tx.Set(key, value)
	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}
}

func (transaction *badgerTransaction) Delete(key []byte) {
	// TODO Reverting trie may require many updates but ErrTxnTooBig is not handled
	key = convNilToBytes(key)

	err := transaction.tx.Delete(key)
	if err != nil {
		panic(fmt.Sprintf("Database Error: %v", err))
	}
}

func (transaction *badgerTransaction) Commit() {
	err := transaction.tx.Commit(nil)

	if err != nil {
		//TODO if there is conflict during commit, this panic will occurs
		panic(err)
	}
}

//=========================================================
// Iterator Implementation
//=========================================================

type badgerIterator struct {
	start   []byte
	end     []byte
	reverse bool
	iter    *badger.Iterator
}

func (db *badgerDB) Iterator(start, end []byte) Iterator {
	badgerTx := db.db.NewTransaction(true)

	var reverse bool

	// if end is bigger then start, then reverse order
	if bytes.Compare(start, end) == 1 {
		reverse = true
	} else {
		reverse = false
	}

	opt := badger.DefaultIteratorOptions
	opt.PrefetchValues = false
	opt.Reverse = reverse

	badgerIter := badgerTx.NewIterator(opt)

	badgerIter.Seek(start)

	retIter := &badgerIterator{
		start:   start,
		end:     end,
		reverse: reverse,
		iter:    badgerIter,
	}
	return retIter
}

func (iter *badgerIterator) Next() {
	if iter.Valid() {
		iter.iter.Next()
	}
}

func (iter *badgerIterator) Valid() bool {

	if !iter.iter.Valid() {
		return false
	}

	if iter.end != nil {
		if iter.reverse == false {
			if bytes.Compare(iter.end, iter.iter.Item().Key()) <= 0 {
				return false
			}
		} else {
			if bytes.Compare(iter.iter.Item().Key(), iter.end) <= 0 {
				return false
			}
		}
	}

	return true
}

func (iter *badgerIterator) Key() (key []byte) {
	return iter.iter.Item().Key()
}

func (iter *badgerIterator) Value() (value []byte) {
	retVal, err := iter.iter.Item().Value()

	if err != nil {
		//FIXME: test and handle errs
		panic(err)
	}

	return retVal
}
