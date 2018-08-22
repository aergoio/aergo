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
	//"math/rand"
	//"sort"
	"github.com/aergoio/aergo/pkg/db"
	"testing"
	//"github.com/dgraph-io/badger"
	//"github.com/dgraph-io/badger/options"
)

func TestTrieEmpty(t *testing.T) {
	smt := NewTrie(32, hash, nil)
	if !bytes.Equal(smt.DefaultHash(256), smt.Root) {
		t.Fatal("empty trie root hash not correct")
	}
}

func TestTrieUpdateAndGet(t *testing.T) {
	smt := NewTrie(32, hash, nil)

	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	root, _, _ := smt.update(smt.Root, keys, values, smt.TrieHeight, nil)

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
	newRoot, _, _ := smt.update(root, newKeys, newValues, smt.TrieHeight, nil)
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

func TestTriePublicUpdateAndGet(t *testing.T) {
	smt := NewTrie(32, hash, nil)
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

func TestTrieDifferentKeySize(t *testing.T) {
	keySize := 20
	smt := NewTrie(uint64(keySize), hash, nil)
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
	ap, _, _, _, _ := smt.MerkleProof(keys[8])
	if !smt.VerifyMerkleProof(ap, keys[8], newValues[8]) {
		t.Fatalf("failed to verify inclusion proof")
	}
}

func TestTrieDelete(t *testing.T) {
	smt := NewTrie(32, hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	root, _, _ := smt.update(smt.Root, keys, values, smt.TrieHeight, nil)
	value, _ := smt.get(root, keys[0], smt.TrieHeight)
	if !bytes.Equal(values[0], value) {
		t.Fatal("trie not updated")
	}

	// Delete from trie
	// To delete a key, just set it's value to Default leaf hash.
	newRoot, _, _ := smt.update(root, keys[0:1], [][]byte{DefaultLeaf}, smt.TrieHeight, nil)
	newValue, _ := smt.get(newRoot, keys[0], smt.TrieHeight)
	if len(newValue) != 0 {
		t.Fatal("Failed to delete from trie")
	}
	// Remove deleted key from keys and check root with a clean trie.
	smt2 := NewTrie(32, hash, nil)
	cleanRoot, _, _ := smt2.update(smt.Root, keys[1:], values[1:], smt.TrieHeight, nil)
	if !bytes.Equal(newRoot, cleanRoot) {
		t.Fatal("roots mismatch")
	}

	//Empty the trie
	var newValues [][]byte
	for i := 0; i < 10; i++ {
		newValues = append(newValues, DefaultLeaf)
	}
	root, _, _ = smt.update(root, keys, newValues, smt.TrieHeight, nil)
	if !bytes.Equal(smt.DefaultHash(256), root) {
		t.Fatal("empty trie root hash not correct")
	}
}

// test updating and deleting at the same time
func TestTrieUpdateAndDelete(t *testing.T) {
	smt := NewTrie(32, hash, nil)
	key0 := make([]byte, 32, 32)
	values := getFreshData(1, 32)
	root, _ := smt.Update([][]byte{key0}, values)
	k, v, isShortcut, _ := smt.loadChildren(root)
	if isShortcut != 1 || !bytes.Equal(k, key0) || !bytes.Equal(v, values[0]) {
		t.Fatal("leaf shortcut didn't move up to root")
	}

	key1 := make([]byte, 32, 32)
	bitSet(key1, 255)
	keys := [][]byte{key0, key1}
	values = [][]byte{DefaultLeaf, getFreshData(1, 32)[0]}
	root, _ = smt.Update(keys, values)
	k, v, isShortcut, _ = smt.loadChildren(root)
	if isShortcut != 1 || !bytes.Equal(k, key1) || !bytes.Equal(v, values[1]) {
		t.Fatal("leaf shortcut didn't move up to root")
	}
}

func TestTrieMerkleProof(t *testing.T) {
	smt := NewTrie(32, hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)

	for i, key := range keys {
		ap, _, _, _, _ := smt.MerkleProof(key)
		if !smt.VerifyMerkleProof(ap, key, values[i]) {
			t.Fatalf("failed to verify inclusion proof")
		}
	}
	emptyKey := hash([]byte("non-member"))
	ap, included, proofKey, proofValue, _ := smt.MerkleProof(emptyKey)
	if included {
		t.Fatalf("failed to verify non inclusion proof")
	}
	if !smt.VerifyMerkleProofEmpty(ap, emptyKey, proofKey, proofValue) {
		t.Fatalf("failed to verify non inclusion proof")
	}
}

func TestTrieMerkleProofCompressed(t *testing.T) {
	smt := NewTrie(32, hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)

	for i, key := range keys {
		bitmap, ap, length, _, _, _, _ := smt.MerkleProofCompressed(key)
		if !smt.VerifyMerkleProofCompressed(bitmap, ap, length, key, values[i]) {
			t.Fatalf("failed to verify inclusion proof")
		}
	}
	emptyKey := hash([]byte("non-member"))
	bitmap, ap, length, included, proofKey, proofValue, _ := smt.MerkleProofCompressed(emptyKey)
	if included {
		t.Fatalf("failed to verify non inclusion proof")
	}
	if !smt.VerifyMerkleProofCompressedEmpty(bitmap, ap, length, emptyKey, proofKey, proofValue) {
		t.Fatalf("failed to verify non inclusion proof")
	}
}

func TestTrieCommit(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewTrie(32, hash, st)
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)
	smt.Commit()
	// liveCache is deleted so the key is fetched in badger db
	smt.db.liveCache = make(map[Hash][]byte)
	value, _ := smt.Get(keys[0])
	if !bytes.Equal(values[0], value) {
		t.Fatal("failed to get value in commited db")
	}
	st.Close()
	os.RemoveAll(".aergo")
}

func TestTrieRevert(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewTrie(32, hash, st)
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

func TestTrieRaisesError(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewTrie(20, hash, st)
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
	_, _, _, _, _, _, err := smt.MerkleProofCompressed(keys[0])
	if err == nil {
		t.Fatal("Error not created if database doesnt have a node")
	}
	_, err = smt.Update(keys, values)
	if err == nil {
		t.Fatal("Error not created if database doesnt have a node")
	}
	st.Close()
	os.RemoveAll(".aergo")
}

func TestTrieLoadCache(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewTrie(32, hash, st)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)
	smt.Commit()

	// Simulate node restart by deleting and loading cache
	cacheSize := len(smt.db.liveCache)
	smt.db.liveCache = make(map[Hash][]byte)

	err := smt.LoadCache(smt.Root)

	if err != nil {
		t.Fatal(err)
	}
	if cacheSize != len(smt.db.liveCache) {
		t.Fatal("Cache loading from db incorrect")
	}
	st.Close()
	os.RemoveAll(".aergo")
}

func TestHeight0LeafShortcut(t *testing.T) {
	keySize := uint64(20)
	smt := NewTrie(keySize, hash, nil)
	// Add 2 sibling keys that will be stored at height 0
	key0 := make([]byte, keySize, keySize)
	key1 := make([]byte, keySize, keySize)
	bitSet(key1, keySize*8-1)
	keys := [][]byte{key0, key1}
	values := getFreshData(2, 32)
	smt.Update(keys, values)

	// Check all keys have been stored
	for i, key := range keys {
		value, _ := smt.Get(key)
		if !bytes.Equal(values[i], value) {
			t.Fatal("trie not updated")
		}
	}
	bitmap, ap, length, _, _, _, _ := smt.MerkleProofCompressed(key1)
	if length != smt.TrieHeight {
		t.Fatal("proof should have length equal to trie height for a leaf shortcut")
	}
	if !smt.VerifyMerkleProofCompressed(bitmap, ap, length, key1, values[1]) {
		t.Fatal("failed to verify inclusion proof")
	}

	// Delete one key and check that the remaining one moved up to the root of the tree
	newRoot, _ := smt.Update(keys[0:1], [][]byte{DefaultLeaf})
	k, v, isShortcut, _ := smt.loadChildren(newRoot)
	if isShortcut != 1 || !bytes.Equal(k, key1) || !bytes.Equal(v, values[1]) {
		t.Fatal("leaf shortcut didn't move up to root")
	}

	_, _, length, _, _, _, _ = smt.MerkleProofCompressed(key1)
	if length != 0 {
		t.Fatal("proof should have length equal to trie height for a leaf shortcut")
	}
}

func benchmark10MAccounts10Ktps(smt *Trie, b *testing.B) {
	//b.ReportAllocs()
	newvalues := getFreshData(1000, 32)
	//fmt.Println("\nLoading b.N x 1000 accounts")
	for i := 0; i < b.N; i++ {
		//start := time.Now()
		newkeys := getFreshData(1000, 32)
		smt.Update(newkeys, newvalues)
		smt.Commit()
		/*
			end := time.Now()
			elapsed := end.Sub(start)
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			fmt.Println(i, " : elapsed : ", elapsed,
				"\ndb read : ", smt.LoadDbCounter, "    cache read : ", smt.LoadCacheCounter,
				"\ncache size : ", len(smt.db.liveCache),
				"\nRAM : ", m.Sys/1024/1024, " MiB")
		*/
	}
}

//go test -run=xxx -bench=. -benchmem -test.benchtime=20s
func BenchmarkCacheHeightLimit233(b *testing.B) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)
	smt := NewTrie(32, hash, st)
	smt.CacheHeightLimit = 233
	benchmark10MAccounts10Ktps(smt, b)
	st.Close()
	os.RemoveAll(".aergo")
}
func BenchmarkCacheHeightLimit238(b *testing.B) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)
	smt := NewTrie(32, hash, st)
	smt.CacheHeightLimit = 238
	benchmark10MAccounts10Ktps(smt, b)
	st.Close()
	os.RemoveAll(".aergo")
}
func BenchmarkCacheHeightLimit245(b *testing.B) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)
	smt := NewTrie(32, hash, st)
	smt.CacheHeightLimit = 245
	benchmark10MAccounts10Ktps(smt, b)
	st.Close()
	os.RemoveAll(".aergo")
}
