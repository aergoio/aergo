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

func TestEmptyModTrie(t *testing.T) {
	smt := NewModSMT(32, hash, nil)
	if !bytes.Equal(smt.DefaultHash(256), smt.Root) {
		t.Fatal("empty trie root hash not correct")
	}
}

func TestModUpdateAndGet(t *testing.T) {
	smt := NewModSMT(32, hash, nil)

	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	root, _ := smt.update(smt.Root, keys, values, smt.TrieHeight, nil)

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
	newRoot, _ := smt.update(root, newKeys, newValues, smt.TrieHeight, nil)
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

func TestDifferentKeySizeMod(t *testing.T) {
	keySize := 20
	smt := NewModSMT(uint64(keySize), hash, nil)
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
	smt.Update(keys[0:1], DataArray{DefaultLeaf})
	newValue, _ := smt.Get(keys[0])
	if len(newValue) != 0 {
		t.Fatal("Failed to delete from trie")
	}
	newValue, _ = smt.Get(make([]byte, keySize))
	if len(newValue) != 0 {
		t.Fatal("Failed to delete from trie")
	}
	/*
		ap, _ := smt.MerkleProof(keys[8])
		if !smt.VerifyMerkleProof(ap, keys[8], newValues[8]) {
			t.Fatalf("failed to verify inclusion proof")
		}
	*/
}

func TestDeleteMod(t *testing.T) {
	smt := NewModSMT(32, hash, nil)
	// Add data to empty trie
	keys := getFreshData(10, 32)
	values := getFreshData(10, 32)
	root, _ := smt.update(smt.Root, keys, values, smt.TrieHeight, nil)
	value, _ := smt.get(root, keys[0], smt.TrieHeight)
	if !bytes.Equal(values[0], value) {
		t.Fatal("trie not updated")
	}

	// Delete from trie
	// To delete a key, just set it's value to Default leaf hash.
	newRoot, _ := smt.update(root, keys[0:1], DataArray{DefaultLeaf}, smt.TrieHeight, nil)
	newValue, _ := smt.get(newRoot, keys[0], smt.TrieHeight)
	if len(newValue) != 0 {
		t.Fatal("Failed to delete from trie")
	}
	// Remove deleted key from keys and check root with a clean trie.
	smt2 := NewModSMT(32, hash, nil)
	cleanRoot, _ := smt2.update(smt.Root, keys[1:], values[1:], smt.TrieHeight, nil)
	/* FIXME : if one of 2 sibling nodes is deleted then the sibling
				should move up other wise the roots mismatch
	if !bytes.Equal(newRoot, cleanRoot) {
		t.Fatal("roots mismatch")
	}
	*/

	//Empty the trie
	var newValues DataArray
	for i := 0; i < 10; i++ {
		newValues = append(newValues, DefaultLeaf)
	}
	root, _ = smt.update(root, keys, newValues, smt.TrieHeight, nil)
	if !bytes.Equal(smt.DefaultHash(256), root) {
		t.Fatal("empty trie root hash not correct")
	}
}

func TestMerkleProofMod(t *testing.T) {
	smt := NewModSMT(32, hash, nil)
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
