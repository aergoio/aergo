/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	//"runtime"
	//"io/ioutil"
	//"os"
	//"path"
	//"time"
	//"encoding/hex"
	//"fmt"
	//"math/rand"
	//"sort"
	//"github.com/aergoio/aergo/pkg/db"
	"testing"
	//"github.com/dgraph-io/badger"
	//"github.com/dgraph-io/badger/options"
)

func TestModPublicUpdateAndGet(t *testing.T) {
	smt := NewModSMT(32, hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	root, _ := smt.Update(keys, values)

	// Check all keys have been stored
	for i, key := range keys {
		value, _ := smt.Get(key)
		if !bytes.Equal(values[i], value) {
			t.Fatal("trie not updated")
		}
	}
	if !bytes.Equal(root, smt.Root) {
		t.Fatal("Root not stored")
	}

	newValues := getFreshData(10, 32)
	smt.Update(keys, newValues)
	// Check all keys have been modified
	for i, key := range keys {
		value, _ := smt.Get(key)
		if !bytes.Equal(newValues[i], value) {
			t.Fatal("trie not updated")
		}
	}
}

/*
func TestLiveCache(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	//st, _ := db.NewBadgerDB(dbPath)
	smt := NewModSMT(32, hash, st)
	keys := getFreshData(1000, 32)
	values := getFreshData(1000, 32)
	fmt.Println("db read : ", smt.LoadDbCounter, "    cache read : ", smt.LoadCacheCounter)
	fmt.Println("cache size : ", len(smt.db.liveCache))
	fmt.Println("db size : ", len(smt.db.updatedNodes))

	start := time.Now()
	smt.Update(keys, values)
	end := time.Now()
	fmt.Println("\nLoad all accounts")
	fmt.Println("db read : ", smt.LoadDbCounter, "    cache read : ", smt.LoadCacheCounter)
	fmt.Println("cache size : ", len(smt.db.liveCache))
	fmt.Println("db size : ", len(smt.db.updatedNodes))
	elapsed := end.Sub(start)
	fmt.Println("elapsed : ", elapsed)

	smt.Commit()

	newvalues := getFreshData(1000, 32)
	start = time.Now()
	smt.Update(keys, newvalues)
	end = time.Now()
	fmt.Println("\n2nd update")
	fmt.Println("db read : ", smt.LoadDbCounter, "    cache read : ", smt.LoadCacheCounter)
	fmt.Println("cache size : ", len(smt.db.liveCache))
	fmt.Println("db size : ", len(smt.db.updatedNodes))
	elapsed = end.Sub(start)
	fmt.Println("elapsed : ", elapsed)

	smt.Commit()

	fmt.Println("\nLoading 10M accounts")
	for i := 0; i < 10000; i++ {
		newkeys := getFreshData(1000, 32)
		start = time.Now()
		smt.Update(newkeys, newvalues)
		end = time.Now()
		elapsed = end.Sub(start)
		fmt.Println(i, " : elapsed : ", elapsed)
		fmt.Println("db read : ", smt.LoadDbCounter, "    cache read : ", smt.LoadCacheCounter)
		fmt.Println("cache size : ", len(smt.db.liveCache))
		smt.Commit()
		/*
			for i, key := range newkeys {
				val, _ := smt.Get(key)
				if !bytes.Equal(newvalues[i], val) {
					t.Fatal("new keys were not updated")
				}
			}
	}
	start = time.Now()
	smt.Update(keys, newvalues)
	end = time.Now()
	fmt.Println("\n Updating 1k accounts in a 1M account tree")
	fmt.Println("elapsed : ", elapsed)
	fmt.Println("db read : ", smt.LoadDbCounter, "    cache read : ", smt.LoadCacheCounter)
	fmt.Println("cache size : ", len(smt.db.liveCache))

	st.Close()
	os.RemoveAll(".aergo")
}
*/
