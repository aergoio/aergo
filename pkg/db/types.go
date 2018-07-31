/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package db

type DBImplType string

const (
	BadgerImpl DBImplType = "badgerdb"
)

type dbConstructor func(dir string) (DB, error)

type DB interface {
	Type() string
	Set(key, value []byte)
	Delete(key []byte)
	Get(key []byte) []byte
	Exist(key []byte) bool
	Iterator(start, end []byte) Iterator
	NewTx(writable bool) Transaction
	Close()
	//Print()
	//Stats() map[string]string
}

type Transaction interface {
	Get(key []byte) []byte
	Set(key, value []byte)
	Delete(key []byte)
	Commit()
}

type Iterator interface {
	Next()
	Valid() bool
	Key() []byte
	Value() []byte
}
