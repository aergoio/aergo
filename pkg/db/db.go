/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package db

import "fmt"

var dbImpls = map[DBImplType]dbConstructor{}

func registorDBConstructor(dbimpl DBImplType, constructor dbConstructor) {
	dbImpls[dbimpl] = constructor
}

// NewDB creates new database or load existing database in the directory
func NewDB(dbimpltype DBImplType, dir string) DB {
	db, err := dbImpls[dbimpltype](dir)

	if err != nil {
		panic(fmt.Sprintf("Fail to Create New DB: %v", err))
	}

	//TODO launch a garbage collector

	return db
}

func convNilToBytes(byteArray []byte) []byte {
	if byteArray == nil {
		return []byte{}
	}
	return byteArray
}
