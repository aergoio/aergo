/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	//"runtime"
	//"io/ioutil"
	"os"
	"path"
	//"time"
	//"encoding/hex"
	//"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	//"github.com/dgraph-io/badger"
	//"github.com/dgraph-io/badger/options"
)

func TestSmtEmptyTrie(t *testing.T) {
	smt := NewSMT(32, hash, nil)
	if !bytes.Equal(smt.DefaultHash(256), smt.Root) {
		t.Fatal("empty trie root hash not correct")
	}
}

func TestSmtUpdateAndGet(t *testing.T) {
	smt := NewSMT(32, hash, nil)

	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	ch := make(chan result, 1)
	smt.update(smt.Root, keys, values, smt.TrieHeight, false, true, ch)
	res := <-ch
	root := res.update

	// Check all keys have been stored
	for i, key := range keys {
		value, _ := smt.get(root, key, smt.TrieHeight)
		if !bytes.Equal(values[i], value) {
			t.Fatal("value not updated")
		}
	}

	// Append to the trie
	newKeys := getFreshData(5, 32)
	newValues := getFreshData(5, 32)
	ch = make(chan result, 1)
	smt.update(root, newKeys, newValues, smt.TrieHeight, false, true, ch)
	res = <-ch
	newRoot := res.update
	if bytes.Equal(root, newRoot) {
		t.Fatal("trie not updated")
	}
	for i, newKey := range newKeys {
		newValue, _ := smt.get(newRoot, newKey, smt.TrieHeight)
		if !bytes.Equal(newValues[i], newValue) {
			t.Fatal("failed to get value")
		}
	}
	// Check old keys are still stored
	for i, key := range keys {
		value, _ := smt.get(root, key, smt.TrieHeight)
		if !bytes.Equal(values[i], value) {
			t.Fatal("failed to get value")
		}
	}
}

func TestSmtPublicUpdateAndGet(t *testing.T) {
	smt := NewSMT(32, hash, nil)
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

func TestSmtDifferentKeySize(t *testing.T) {
	keySize := 20
	smt := NewSMT(uint64(keySize), hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, keySize)
	values := getFreshData(10, 32)
	smt.Update(keys, values)

	// Check all keys have been stored
	for i, key := range keys {
		value, _ := smt.Get(key)
		if !bytes.Equal(values[i], value) {
			t.Fatal("trie not updated")
		}
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
	smt.Update(keys[0:1], [][]byte{DefaultLeaf})
	newValue, _ := smt.Get(keys[0])
	if len(newValue) != 0 {
		t.Fatal("Failed to delete from trie")
	}
	newValue, _ = smt.Get(make([]byte, keySize))
	if len(newValue) != 0 {
		t.Fatal("Failed to delete from trie")
	}
	ap, _ := smt.MerkleProof(keys[8])
	if !smt.VerifyMerkleProof(ap, keys[8], newValues[8]) {
		t.Fatalf("failed to verify inclusion proof")
	}
}

func TestSmtDelete(t *testing.T) {
	smt := NewSMT(32, hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	ch := make(chan result, 1)
	smt.update(smt.Root, keys, values, smt.TrieHeight, false, true, ch)
	res := <-ch
	root := res.update
	value, _ := smt.get(root, keys[0], smt.TrieHeight)
	if !bytes.Equal(values[0], value) {
		t.Fatal("trie not updated")
	}

	// Delete from trie
	// To delete a key, just set it's value to Default leaf hash.
	ch = make(chan result, 1)
	smt.update(root, keys[0:1], [][]byte{DefaultLeaf}, smt.TrieHeight, false, true, ch)
	res = <-ch
	newRoot := res.update
	newValue, _ := smt.get(newRoot, keys[0], smt.TrieHeight)
	if len(newValue) != 0 {
		t.Fatal("Failed to delete from trie")
	}
	// Remove deleted key from keys and check root with a clean trie.
	smt2 := NewSMT(32, hash, nil)
	ch = make(chan result, 1)
	smt2.update(smt.Root, keys[1:], values[1:], smt.TrieHeight, false, true, ch)
	res = <-ch
	cleanRoot := res.update
	if !bytes.Equal(newRoot, cleanRoot) {
		t.Fatal("roots mismatch")
	}

	//Empty the trie
	var newValues [][]byte
	for i := 0; i < 10; i++ {
		newValues = append(newValues, DefaultLeaf)
	}
	ch = make(chan result, 1)
	smt.update(root, keys, newValues, smt.TrieHeight, false, true, ch)
	res = <-ch
	root = res.update
	if !bytes.Equal(smt.DefaultHash(256), root) {
		t.Fatal("empty trie root hash not correct")
	}

}

func TestSmtMerkleProof(t *testing.T) {
	smt := NewSMT(32, hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)

	for i, key := range keys {
		ap, _ := smt.MerkleProof(key)
		if !smt.VerifyMerkleProof(ap, key, values[i]) {
			t.Fatalf("failed to verify inclusion proof")
		}
	}
	emptyKey := hash([]byte("non-member"))
	ap, _ := smt.MerkleProof(emptyKey)
	if !smt.VerifyMerkleProof(ap, emptyKey, DefaultLeaf) {
		t.Fatalf("failed to verify non inclusion proof")
	}
}

func TestSmtMerkleProofCompressed(t *testing.T) {
	smt := NewSMT(32, hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)

	for i, key := range keys {
		bitmap, ap, _ := smt.MerkleProofCompressed(key)
		bitmap2, ap2, _ := smt.MerkleProofCompressed2(key)
		if !smt.VerifyMerkleProofCompressed(bitmap, ap, key, values[i]) {
			t.Fatalf("failed to verify inclusion proof")
		}
		if !smt.VerifyMerkleProofCompressed(bitmap2, ap2, key, values[i]) {
			t.Fatalf("failed to verify inclusion proof")
		}
		if !bytes.Equal(bitmap, bitmap2) {
			t.Fatal("the 2 versions of compressed merkle proofs don't match")
		}
		for i, a := range ap {
			if !bytes.Equal(a, ap2[i]) {
				t.Fatal("the 2 versions of compressed merkle proofs don't match")
			}
		}
	}
	emptyKey := hash([]byte("non-member"))
	bitmap, ap, _ := smt.MerkleProofCompressed(emptyKey)
	if !smt.VerifyMerkleProofCompressed(bitmap, ap, emptyKey, DefaultLeaf) {
		t.Fatalf("failed to verify non inclusion proof")
	}
}

func TestSmtMerkleProofCompressed2(t *testing.T) {
	smt := NewSMT(32, hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)

	for i, key := range keys {
		bitmap2, ap2, _ := smt.MerkleProofCompressed2(key)
		if !smt.VerifyMerkleProofCompressed(bitmap2, ap2, key, values[i]) {
			t.Fatalf("failed to verify inclusion proof")
		}
	}
}

func TestSmtCommit(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewSMT(32, hash, st)
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)
	smt.Commit()
	// liveCache is deleted so the key is fetched in badger db
	smt.db.liveCache = make(map[Hash][]byte)
	value, _ := smt.Get(keys[0])
	if !bytes.Equal(values[0], value) {
		t.Fatal("failed to get value in committed db")
	}
	st.Close()
	os.RemoveAll(".aergo")
}

func TestSmtRevert(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewSMT(32, hash, st)
	smt.Commit()
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	root, _ := smt.Update(keys, values)
	smt.Commit()

	newValues := getFreshData(10, 32)
	smt.Update(keys, newValues)
	smt.Commit()
	newValues = getFreshData(10, 32)
	newRoot, _ := smt.Update(keys, newValues)
	smt.Commit()

	smt.Revert(root)

	if !bytes.Equal(smt.Root, root) {
		t.Fatal("revert failed")
	}
	if len(smt.pastTries) != 2 { // contains empty trie + reverted trie
		t.Fatal("past tries not updated after revert")
	}
	// Check all keys have been reverted
	for i, key := range keys {
		value, _ := smt.Get(key)
		if !bytes.Equal(values[i], value) {
			t.Fatal("revert failed, values not updated")
		}
	}
	if len(smt.db.liveCache) != 256 {
		t.Fatal("live cache not reset after revert")
	}
	if len(smt.db.store.Get(newRoot)) != 0 {
		t.Fatal("nodes not deleted from database")
	}
	st.Close()
	os.RemoveAll(".aergo")
}

func TestSmtRaisesError(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewSMT(20, hash, st)
	// Add data to empty trie
	keys := getFreshData(10, 20)
	values := getFreshData(10, 20)
	smt.Update(keys, values)
	smt.db.liveCache = make(map[Hash][]byte)
	smt.db.updatedNodes = make(map[Hash][]byte)
	smt.loadDefaultHashes()

	// Check errors are raised is a keys is not in cache nore db
	for _, key := range keys {
		_, err := smt.Get(key)
		if err == nil {
			t.Fatal("Error not created if database doesnt have a node")
		}
	}
	_, _, err := smt.MerkleProofCompressed(keys[0])
	if err == nil {
		t.Fatal("Error not created if database doesnt have a node")
	}
	_, err = smt.Update(keys, values)
	if err == nil {
		t.Fatal("Error not created if database doesnt have a node")
	}
	st.Close()
	os.RemoveAll(".aergo")

	smt = NewSMT(20, hash, nil)
	err = smt.Commit()
	if err == nil {
		t.Fatal("Error not created if database not connected")
	}
	smt.db.liveCache = make(map[Hash][]byte)
	_, _, _, err = smt.loadChildren(make([]byte, 32, 32))
	if err == nil {
		t.Fatal("Error not created if database not connected")
	}

}

func getFreshData(size, length int) [][]byte {
	var data [][]byte
	for i := 0; i < size; i++ {
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			panic(err)
		}
		data = append(data, hash(key)[:length])
	}
	sort.Sort(DataArray(data))
	return data
}

/*
func TestSmtLiveCache(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	//st, _ := db.NewBadgerDB(dbPath)
	smt := NewSMT(32, hash, st)
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
		for i, key := range newkeys {
			val, _ := smt.Get(key)
			if !bytes.Equal(newvalues[i], val) {
				t.Fatal("new keys were not updated")
			}
		}
		//start = time.Now()
		//smt.Update(keys, newvalues)
		//end = time.Now()
		//elapsed = end.Sub(start)
		//fmt.Println("elapsed2 : ", elapsed)

		// deleted keys are not getting collected from liveCache by GC
		// this doesnt seem to work well if liveCache is over 1Million keys
		/*
			if i%100 == 0 {
				temp := smt.db.liveCache
				smt.db.liveCache = make(map[Hash][]byte, len(temp)*3)
				for k, v := range temp {
					smt.db.liveCache[k] = v
				}
			}

		//runtime.GC()
	}

	start = time.Now()
	smt.Update(keys, newvalues)
	end = time.Now()
	fmt.Println("\n Updating 1k accounts in a 10M account tree")
	fmt.Println("elapsed : ", elapsed)
	fmt.Println("db read : ", smt.LoadDbCounter, "    cache read : ", smt.LoadCacheCounter)
	fmt.Println("cache size : ", len(smt.db.liveCache))

	st.Close()
	os.RemoveAll(".aergo")
}
*/

/*
func TestDB(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)
	j := 0
	for {
		fmt.Println(j)
		kv := getFreshData(10000, 32)
		txn := st.NewTx(true)
		for i := 0; i < 10000; i++ {
			txn.Set(kv[i], kv[i])
		}
		txn.Commit()
		j++
	}
}

func TestDB(t *testing.T) {
	dir, err := ioutil.TempDir(".", "badger")
	if err != nil {
		fmt.Println(err)
	}
	defer os.RemoveAll(dir)
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	opts.TableLoadingMode = options.FileIO
	opts.ValueLogLoadingMode = options.FileIO
	//opts.ValueLogMaxEntries = 10
	//opts.Truncate = true
	db, err := badger.Open(opts)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	err = db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("key"))
		// We expect ErrKeyNotFound
		fmt.Println(err)
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

	txn := db.NewTransaction(true) // Read-write txn
	err = txn.Set([]byte("key"), []byte("value"))
	if err != nil {
		fmt.Println(err)
	}
	err = txn.Commit(nil)
	if err != nil {
		fmt.Println(err)
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("key"))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(val))
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
	j := 0
	k0 := getFreshData(1, 32)
	for {
		fmt.Println(j)
		txn := db.NewTransaction(true) // Read-write txn
		for i := 0; i < 10000; i++ {
			kv := getFreshData(1, 32)
			err = txn.Set(k0[0], kv[0])
			if err != nil {
				fmt.Println(err)
			}
		}
		err = txn.Commit(nil)
		if err != nil {
			fmt.Println(err)
		}
		j++
	}
}
*/
