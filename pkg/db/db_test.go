/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package db

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	tmpDbTestKey1    = "tempkey1"
	tmpDbTestKey2    = "tempkey2"
	tmpDbTestStrVal1 = "val1"
	tmpDbTestStrVal2 = "val2"
	tmpDbTestIntVal1 = 1
	tmpDbTestIntVal2 = 2
)

func createTmpDB(key DBImplType) (dir string, db DB) {
	dir, err := ioutil.TempDir("", string(key))
	if err != nil {
		log.Fatal(err)
	}

	db = NewDB(key, dir)

	return
}

func setInitData(db DB) {
	tx := db.NewTx(true)

	tx.Set([]byte("1"), []byte("1"))
	tx.Set([]byte("2"), []byte("2"))
	tx.Set([]byte("3"), []byte("3"))
	tx.Set([]byte("4"), []byte("4"))
	tx.Set([]byte("5"), []byte("5"))
	tx.Set([]byte("6"), []byte("6"))
	tx.Set([]byte("7"), []byte("7"))

	tx.Commit()
}

func TestGetSet(t *testing.T) {
	// for each db implementation
	for key := range dbImpls {
		dir, db := createTmpDB(key)

		//initial value of empty key  must be empty byte
		assert.Empty(t, db.Get([]byte(tmpDbTestKey1)), db.Type())

		// set value
		db.Set([]byte(tmpDbTestKey1), []byte(tmpDbTestStrVal1))

		assert.Equal(t, tmpDbTestStrVal1, string(db.Get([]byte(tmpDbTestKey1))), db.Type())

		db.Close()
		os.RemoveAll(dir)
	}
}

func TestTransactionSet(t *testing.T) {

	for key := range dbImpls {
		dir, db := createTmpDB(key)

		// create a new writable tx
		tx := db.NewTx(true)
		// in the tx, an unsetted value must be empty byte
		assert.Empty(t, tx.Get([]byte(tmpDbTestKey1)), db.Type())

		// set the value in the tx
		tx.Set([]byte(tmpDbTestKey1), []byte(tmpDbTestStrVal1))
		// the value now has a value in the tx context
		assert.Equal(t, tmpDbTestStrVal1, string(tx.Get([]byte(tmpDbTestKey1))), db.Type())
		// the value will not visible at a db
		assert.Empty(t, db.Get([]byte(tmpDbTestKey1)), db.Type())

		tx.Commit()

		// after commit, the value visible from the db
		assert.Equal(t, tmpDbTestStrVal1, string(db.Get([]byte(tmpDbTestKey1))), db.Type())

		db.Close()
		os.RemoveAll(dir)
	}
}

func TestTransactionCommitTwice(t *testing.T) {
	for key := range dbImpls {
		dir, db := createTmpDB(key)

		// create a new writable tx
		tx := db.NewTx(true)

		// a first commit will success
		tx.Commit()

		// a second commit will cause panic
		assert.Panics(t, func() { tx.Commit() })

		db.Close()
		os.RemoveAll(dir)
	}
}

func TestIter(t *testing.T) {

	for key := range dbImpls {
		dir, db := createTmpDB(key)

		setInitData(db)

		i := 1

		for iter := db.Iterator(nil, nil); iter.Valid(); iter.Next() {
			assert.EqualValues(t, strconv.Itoa(i), string(iter.Key()))
			i++
		}

		db.Close()
		os.RemoveAll(dir)
	}
}

func TestRangeIter(t *testing.T) {

	for key := range dbImpls {
		dir, db := createTmpDB(key)

		setInitData(db)

		// test iteration 2 -> 5
		i := 2
		for iter := db.Iterator([]byte("2"), []byte("5")); iter.Valid(); iter.Next() {
			assert.EqualValues(t, strconv.Itoa(i), string(iter.Key()))
			i++
		}
		assert.EqualValues(t, i, 5)

		// nil sames with []byte("0")
		// test iteration 0 -> 5
		i = 1
		for iter := db.Iterator(nil, []byte("5")); iter.Valid(); iter.Next() {
			assert.EqualValues(t, strconv.Itoa(i), string(iter.Key()))
			i++
		}
		assert.EqualValues(t, i, 5)

		db.Close()
		os.RemoveAll(dir)
	}

}

func TestReverseIter(t *testing.T) {

	for key := range dbImpls {
		dir, db := createTmpDB(key)

		setInitData(db)

		// test reverse iteration 5 <- 2
		i := 5
		for iter := db.Iterator([]byte("5"), []byte("2")); iter.Valid(); iter.Next() {
			assert.EqualValues(t, strconv.Itoa(i), string(iter.Key()))
			i--
		}
		assert.EqualValues(t, i, 2)

		// nil sames with []byte("0")
		// test reverse iteration 5 -> 0
		i = 5
		for iter := db.Iterator([]byte("5"), nil); iter.Valid(); iter.Next() {
			assert.EqualValues(t, strconv.Itoa(i), string(iter.Key()))
			i--
		}
		assert.EqualValues(t, i, 0)

		db.Close()
		os.RemoveAll(dir)
	}
}
