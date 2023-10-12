/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/internal/common"
)

func TestTrieEmpty(t *testing.T) {
	smt := NewTrie(nil, common.Hasher, nil)
	if len(smt.Root) != 0 {
		t.Fatal("empty trie root hash not correct")
	}
}

func TestTrieUpdateAndGet(t *testing.T) {
	smt := NewTrie(nil, common.Hasher, nil)
	smt.atomicUpdate = false

	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	ch := make(chan mresult, 1)
	smt.update(smt.Root, keys, values, nil, 0, smt.TrieHeight, ch)
	res := <-ch
	root := res.update

	// Check all keys have been stored
	for i, key := range keys {
		value, _ := smt.get(root, key, nil, 0, smt.TrieHeight)
		if !bytes.Equal(values[i], value) {
			t.Fatal("value not updated")
		}
	}

	// Append to the trie
	newKeys := getFreshData(5, 32)
	newValues := getFreshData(5, 32)
	ch = make(chan mresult, 1)
	smt.update(root, newKeys, newValues, nil, 0, smt.TrieHeight, ch)
	res = <-ch
	newRoot := res.update
	if bytes.Equal(root, newRoot) {
		t.Fatal("trie not updated")
	}
	for i, newKey := range newKeys {
		newValue, _ := smt.get(newRoot, newKey, nil, 0, smt.TrieHeight)
		if !bytes.Equal(newValues[i], newValue) {
			t.Fatal("failed to get value")
		}
	}
}

func TestTrieAtomicUpdate(t *testing.T) {
	smt := NewTrie(nil, common.Hasher, nil)
	smt.CacheHeightLimit = 0
	keys := getFreshData(1, 32)
	values := getFreshData(1, 32)
	root, _ := smt.AtomicUpdate(keys, values)
	updatedNb := len(smt.db.updatedNodes)
	cacheNb := len(smt.db.liveCache)
	newvalues := getFreshData(1, 32)
	smt.AtomicUpdate(keys, newvalues)
	if len(smt.db.updatedNodes) != 2*updatedNb {
		t.Fatal("Atomic update doesnt store all tries")
	}
	if len(smt.db.liveCache) != cacheNb {
		t.Fatal("Cache size should remain the same")
	}

	// check keys of previous atomic update are accessible in
	// updated nodes with root.
	smt.atomicUpdate = false
	for i, key := range keys {
		value, _ := smt.get(root, key, nil, 0, smt.TrieHeight)
		if !bytes.Equal(values[i], value) {
			t.Fatal("failed to get value")
		}
	}
}

func TestTriePublicUpdateAndGet(t *testing.T) {
	smt := NewTrie(nil, common.Hasher, nil)
	smt.CacheHeightLimit = 0
	// Add data to empty trie
	keys := getFreshData(20, 32)
	values := getFreshData(20, 32)
	root, _ := smt.Update(keys, values)
	updatedNb := len(smt.db.updatedNodes)
	cacheNb := len(smt.db.liveCache)

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

	newValues := getFreshData(20, 32)
	smt.Update(keys, newValues)

	if len(smt.db.updatedNodes) != updatedNb {
		t.Fatal("multiple updates don't actualise updated nodes")
	}
	if len(smt.db.liveCache) != cacheNb {
		t.Fatal("multiple updates don't actualise liveCache")
	}
	// Check all keys have been modified
	for i, key := range keys {
		value, _ := smt.Get(key)
		if !bytes.Equal(newValues[i], value) {
			t.Fatal("trie not updated")
		}
	}
}

func TestTrieDelete(t *testing.T) {
	smt := NewTrie(nil, common.Hasher, nil)
	// Add data to empty trie
	keys := getFreshData(20, 32)
	values := getFreshData(20, 32)
	ch := make(chan mresult, 1)
	smt.update(smt.Root, keys, values, nil, 0, smt.TrieHeight, ch)
	result := <-ch
	root := result.update
	value, _ := smt.get(root, keys[0], nil, 0, smt.TrieHeight)
	if !bytes.Equal(values[0], value) {
		t.Fatal("trie not updated")
	}

	// Delete from trie
	// To delete a key, just set it's value to Default leaf hash.
	ch = make(chan mresult, 1)
	smt.update(root, keys[0:1], [][]byte{DefaultLeaf}, nil, 0, smt.TrieHeight, ch)
	result = <-ch
	updatedNb := len(smt.db.updatedNodes)
	newRoot := result.update
	newValue, _ := smt.get(newRoot, keys[0], nil, 0, smt.TrieHeight)
	if len(newValue) != 0 {
		t.Fatal("Failed to delete from trie")
	}
	// Remove deleted key from keys and check root with a clean trie.
	smt2 := NewTrie(nil, common.Hasher, nil)
	ch = make(chan mresult, 1)
	smt2.update(smt.Root, keys[1:], values[1:], nil, 0, smt.TrieHeight, ch)
	result = <-ch
	cleanRoot := result.update
	if !bytes.Equal(newRoot, cleanRoot) {
		t.Fatal("roots mismatch")
	}

	if len(smt2.db.updatedNodes) != updatedNb {
		t.Fatal("deleting doesn't actualise updated nodes")
	}

	//Empty the trie
	var newValues [][]byte
	for i := 0; i < 20; i++ {
		newValues = append(newValues, DefaultLeaf)
	}
	ch = make(chan mresult, 1)
	smt.update(root, keys, newValues, nil, 0, smt.TrieHeight, ch)
	result = <-ch
	root = result.update
	//if !bytes.Equal(smt.DefaultHash(256), root) {
	if len(root) != 0 {
		t.Fatal("empty trie root hash not correct")
	}

	// Test deleting an already empty key
	smt = NewTrie(nil, common.Hasher, nil)
	keys = getFreshData(2, 32)
	values = getFreshData(2, 32)
	root, _ = smt.Update(keys, values)
	key0 := make([]byte, 32, 32)
	key1 := make([]byte, 32, 32)
	smt.Update([][]byte{key0, key1}, [][]byte{DefaultLeaf, DefaultLeaf})
	if !bytes.Equal(root, smt.Root) {
		t.Fatal("deleting a default key shouldnt' modify the tree")
	}
}

// test updating and deleting at the same time
func TestTrieUpdateAndDelete(t *testing.T) {
	smt := NewTrie(nil, common.Hasher, nil)
	smt.CacheHeightLimit = 0
	key0 := make([]byte, 32, 32)
	values := getFreshData(1, 32)
	root, _ := smt.Update([][]byte{key0}, values)
	cacheNb := len(smt.db.liveCache)
	updatedNb := len(smt.db.updatedNodes)
	smt.atomicUpdate = false
	_, _, k, v, isShortcut, _ := smt.loadChildren(root, smt.TrieHeight, 0, nil)
	if !isShortcut || !bytes.Equal(k[:HashLength], key0) || !bytes.Equal(v[:HashLength], values[0]) {
		t.Fatal("leaf shortcut didn't move up to root")
	}

	key1 := make([]byte, 32, 32)
	// set the last bit
	bitSet(key1, 255)
	keys := [][]byte{key0, key1}
	values = [][]byte{DefaultLeaf, getFreshData(1, 32)[0]}
	root, _ = smt.Update(keys, values)

	if len(smt.db.liveCache) != cacheNb {
		t.Fatal("number of cache nodes not correct after delete")
	}
	if len(smt.db.updatedNodes) != updatedNb {
		t.Fatal("number of cache nodes not correct after delete")
	}

	smt.atomicUpdate = false
	_, _, k, v, isShortcut, _ = smt.loadChildren(root, smt.TrieHeight, 0, nil)
	if !isShortcut || !bytes.Equal(k[:HashLength], key1) || !bytes.Equal(v[:HashLength], values[1]) {
		t.Fatal("leaf shortcut didn't move up to root")
	}
}

func TestTrieMerkleProof(t *testing.T) {
	smt := NewTrie(nil, common.Hasher, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)

	for i, key := range keys {
		ap, _, k, v, _ := smt.MerkleProof(key)
		if !smt.VerifyInclusion(ap, key, values[i]) {
			t.Fatalf("failed to verify inclusion proof")
		}
		if !bytes.Equal(key, k) && !bytes.Equal(values[i], v) {
			t.Fatalf("merkle proof didnt return the correct key-value pair")
		}
	}
	emptyKey := common.Hasher([]byte("non-member"))
	ap, included, proofKey, proofValue, _ := smt.MerkleProof(emptyKey)
	if included {
		t.Fatalf("failed to verify non inclusion proof")
	}
	if !smt.VerifyNonInclusion(ap, emptyKey, proofValue, proofKey) {
		t.Fatalf("failed to verify non inclusion proof")
	}
}

func TestTrieMerkleProofCompressed(t *testing.T) {
	smt := NewTrie(nil, common.Hasher, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)

	for i, key := range keys {
		bitmap, ap, length, _, k, v, _ := smt.MerkleProofCompressed(key)
		if !smt.VerifyInclusionC(bitmap, key, values[i], ap, length) {
			t.Fatalf("failed to verify inclusion proof")
		}
		if !bytes.Equal(key, k) && !bytes.Equal(values[i], v) {
			t.Fatalf("merkle proof didnt return the correct key-value pair")
		}
	}
	emptyKey := common.Hasher([]byte("non-member"))
	bitmap, ap, length, included, proofKey, proofValue, _ := smt.MerkleProofCompressed(emptyKey)
	if included {
		t.Fatalf("failed to verify non inclusion proof")
	}
	if !smt.VerifyNonInclusionC(ap, length, bitmap, emptyKey, proofValue, proofKey) {
		t.Fatalf("failed to verify non inclusion proof")
	}
}

func TestTrieCommit(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewTrie(nil, common.Hasher, st)
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)
	smt.Commit()
	// liveCache is deleted so the key is fetched in badger db
	smt.db.liveCache = make(map[Hash][][]byte)
	for i, key := range keys {
		value, _ := smt.Get(key)
		if !bytes.Equal(value, values[i]) {
			t.Fatal("failed to get value in committed db")
		}
	}
	st.Close()
	os.RemoveAll(".aergo")
}

func TestTrieStageUpdates(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewTrie(nil, common.Hasher, st)
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)
	txn := st.NewTx()
	smt.StageUpdates(txn.(DbTx))
	txn.Commit()
	// liveCache is deleted so the key is fetched in badger db
	smt.db.liveCache = make(map[Hash][][]byte)
	for i, key := range keys {
		value, _ := smt.Get(key)
		if !bytes.Equal(value, values[i]) {
			t.Fatal("failed to get value in committed db")
		}
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

	smt := NewTrie(nil, common.Hasher, st)

	// Edge case : Test that revert doesnt delete shortcut nodes
	// when moved to a different position in tree
	key0 := make([]byte, 32, 32)
	key1 := make([]byte, 32, 32)
	// setting the bit at 251 creates 2 shortcut batches at height 252
	bitSet(key1, 251)
	values := getFreshData(2, 32)
	keys := [][]byte{key0, key1}
	root, _ := smt.Update([][]byte{key0}, [][]byte{values[0]})
	smt.Commit()
	root2, _ := smt.Update([][]byte{key1}, [][]byte{values[1]})
	smt.Commit()
	smt.Revert(root)
	if len(smt.db.Store.Get(root)) == 0 {
		t.Fatal("shortcut node shouldnt be deleted by revert")
	}
	if len(smt.db.Store.Get(root2)) != 0 {
		t.Fatal("reverted root should have been deleted")
	}
	key1 = make([]byte, 32, 32)
	// setting the bit at 255 stores the keys as the tip
	bitSet(key1, 255)
	smt.Update([][]byte{key1}, [][]byte{values[1]})
	smt.Commit()
	smt.Revert(root)
	if len(smt.db.Store.Get(root)) == 0 {
		t.Fatal("shortcut node shouldnt be deleted by revert")
	}

	// Test all nodes are reverted in the usual case
	// Add data to empty trie
	keys = getFreshData(10, 32)
	values = getFreshData(10, 32)
	root, _ = smt.Update(keys, values)
	smt.Commit()

	// Update the values
	newValues := getFreshData(10, 32)
	smt.Update(keys, newValues)
	updatedNodes1 := smt.db.updatedNodes
	smt.Commit()
	newKeys := getFreshData(10, 32)
	newValues = getFreshData(10, 32)
	smt.Update(newKeys, newValues)
	updatedNodes2 := smt.db.updatedNodes
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
	if len(smt.db.liveCache) != 0 {
		t.Fatal("live cache not reset after revert")
	}
	// Check all reverted nodes have been deleted
	for node := range updatedNodes2 {
		if len(smt.db.Store.Get(node[:])) != 0 {
			t.Fatal("nodes not deleted from database", node)
		}
	}
	for node := range updatedNodes1 {
		if len(smt.db.Store.Get(node[:])) != 0 {
			t.Fatal("nodes not deleted from database", node)
		}
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

	smt := NewTrie(nil, common.Hasher, st)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	smt.Update(keys, values)
	smt.db.liveCache = make(map[Hash][][]byte)
	smt.db.updatedNodes = make(map[Hash][][]byte)

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

	smt = NewTrie(nil, common.Hasher, nil)
	err = smt.Commit()
	if err == nil {
		t.Fatal("Error not created if database not connected")
	}
	smt.db.liveCache = make(map[Hash][][]byte)
	smt.atomicUpdate = false
	_, _, _, _, _, err = smt.loadChildren(make([]byte, 32, 32), smt.TrieHeight, 0, nil)
	if err == nil {
		t.Fatal("Error not created if database not connected")
	}
	err = smt.LoadCache(make([]byte, 32))
	if err == nil {
		t.Fatal("Error not created if database not connected")
	}
}

func TestTrieLoadCache(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)

	smt := NewTrie(nil, common.Hasher, st)
	// Test size of cache
	smt.CacheHeightLimit = 0
	key0 := make([]byte, 32, 32)
	key1 := make([]byte, 32, 32)
	bitSet(key1, 255)
	values := getFreshData(2, 32)
	smt.Update([][]byte{key0, key1}, values)
	if len(smt.db.liveCache) != 66 {
		// the nodes are at the tip, so 64 + 2 = 66
		t.Fatal("cache size incorrect")
	}

	// Add data to empty trie
	keys := getFreshData(10, 32)
	values = getFreshData(10, 32)
	smt.Update(keys, values)
	smt.Commit()

	// Simulate node restart by deleting and loading cache
	cacheSize := len(smt.db.liveCache)
	smt.db.liveCache = make(map[Hash][][]byte)

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
	keySize := 32
	smt := NewTrie(nil, common.Hasher, nil)
	// Add 2 sibling keys that will be stored at height 0
	key0 := make([]byte, keySize, keySize)
	key1 := make([]byte, keySize, keySize)
	bitSet(key1, keySize*8-1)
	keys := [][]byte{key0, key1}
	values := getFreshData(2, 32)
	smt.Update(keys, values)
	updatedNb := len(smt.db.updatedNodes)

	// Check all keys have been stored
	for i, key := range keys {
		value, _ := smt.Get(key)
		if !bytes.Equal(values[i], value) {
			t.Fatal("trie not updated")
		}
	}
	bitmap, ap, length, _, k, v, err := smt.MerkleProofCompressed(key1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(key1, k) && !bytes.Equal(values[1], v) {
		t.Fatalf("merkle proof didnt return the correct key-value pair")
	}
	if length != smt.TrieHeight {
		t.Fatal("proof should have length equal to trie height for a leaf shortcut")
	}
	if !smt.VerifyInclusionC(bitmap, key1, values[1], ap, length) {
		t.Fatal("failed to verify inclusion proof")
	}

	// Delete one key and check that the remaining one moved up to the root of the tree
	newRoot, _ := smt.AtomicUpdate(keys[0:1], [][]byte{DefaultLeaf})

	// Nb of updated nodes remains same because the new shortcut root was already stored at height 0.
	if len(smt.db.updatedNodes) != updatedNb {
		fmt.Println(len(smt.db.updatedNodes), updatedNb)
		t.Fatal("number of cache nodes not correct after delete")
	}
	smt.atomicUpdate = false
	_, _, k, v, isShortcut, err := smt.loadChildren(newRoot, smt.TrieHeight, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !isShortcut || !bytes.Equal(k[:HashLength], key1) || !bytes.Equal(v[:HashLength], values[1]) {
		t.Fatal("leaf shortcut didn't move up to root")
	}

	_, _, length, _, k, v, _ = smt.MerkleProofCompressed(key1)
	if length != 0 {
		t.Fatal("proof should have length equal to trie height for a leaf shortcut")
	}
	if !bytes.Equal(key1, k) && !bytes.Equal(values[1], v) {
		t.Fatalf("merkle proof didnt return the correct key-value pair")
	}
}

func TestStash(t *testing.T) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)
	smt := NewTrie(nil, common.Hasher, st)
	// Add data to empty trie
	keys := getFreshData(20, 32)
	values := getFreshData(20, 32)
	root, _ := smt.Update(keys, values)
	cacheSize := len(smt.db.liveCache)
	smt.Commit()
	if len(smt.pastTries) != 1 {
		t.Fatal("Past tries not updated after commit")
	}
	values = getFreshData(20, 32)
	smt.Update(keys, values)
	smt.Stash(true)
	if len(smt.pastTries) != 1 {
		t.Fatal("Past tries not updated after commit")
	}
	if !bytes.Equal(smt.Root, root) {
		t.Fatal("Trie not rolled back")
	}
	if len(smt.db.updatedNodes) != 0 {
		t.Fatal("Trie not rolled back")
	}
	if len(smt.db.liveCache) != cacheSize {
		t.Fatal("Trie not rolled back")
	}
	keys = getFreshData(20, 32)
	values = getFreshData(20, 32)
	smt.AtomicUpdate(keys, values)
	values = getFreshData(20, 32)
	smt.AtomicUpdate(keys, values)
	if len(smt.pastTries) != 3 {
		t.Fatal("Past tries not updated after commit")
	}
	smt.Stash(true)
	if !bytes.Equal(smt.Root, root) {
		t.Fatal("Trie not rolled back")
	}
	if len(smt.db.updatedNodes) != 0 {
		t.Fatal("Trie not rolled back")
	}
	if len(smt.db.liveCache) != cacheSize {
		t.Fatal("Trie not rolled back")
	}
	if len(smt.pastTries) != 1 {
		t.Fatal("Past tries not updated after commit")
	}
	st.Close()
	os.RemoveAll(".aergo")
}

func benchmark10MAccounts10Ktps(smt *Trie, b *testing.B) {
	//b.ReportAllocs()
	keys := getFreshData(100, 32)
	values := getFreshData(100, 32)
	smt.Update(keys, values)
	fmt.Println("\nLoading b.N x 1000 accounts")
	for i := 0; i < b.N; i++ {
		newkeys := getFreshData(1000, 32)
		newvalues := getFreshData(1000, 32)
		start := time.Now()
		smt.Update(newkeys, newvalues)
		end := time.Now()
		smt.Commit()
		end2 := time.Now()
		for j, key := range newkeys {
			val, _ := smt.Get(key)
			if !bytes.Equal(val, newvalues[j]) {
				b.Fatal("new key not included")
			}
		}
		end3 := time.Now()
		elapsed := end.Sub(start)
		elapsed2 := end2.Sub(end)
		elapsed3 := end3.Sub(end2)
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Println(i, " : update time : ", elapsed, "commit time : ", elapsed2,
			"\n1000 Get time : ", elapsed3,
			"\ndb read : ", smt.LoadDbCounter, "    cache read : ", smt.LoadCacheCounter,
			"\ncache size : ", len(smt.db.liveCache),
			"\nRAM : ", m.Sys/1024/1024, " MiB")
	}
}

// go test -run=xxx -bench=. -benchmem -test.benchtime=20s
func BenchmarkCacheHeightLimit233(b *testing.B) {
	dbPath := path.Join(".aergo", "db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)
	smt := NewTrie(nil, common.Hasher, st)
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
	smt := NewTrie(nil, common.Hasher, st)
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
	smt := NewTrie(nil, common.Hasher, st)
	smt.CacheHeightLimit = 245
	benchmark10MAccounts10Ktps(smt, b)
	st.Close()
	os.RemoveAll(".aergo")
}

func getFreshData(size, length int) [][]byte {
	var data [][]byte
	for i := 0; i < size; i++ {
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			panic(err)
		}
		data = append(data, common.Hasher(key)[:length])
	}
	sort.Sort(DataArray(data))
	return data
}
