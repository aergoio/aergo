package statedb

import (
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

// func TestCacheBasic(t *testing.T) {
// 	t.Log("StorageCache")
// }

func TestStorageBasic(t *testing.T) {
	storage := NewBufferedStorage(nil, nil)
	v1 := types.GetHashID([]byte("v1"))
	v2 := types.GetHashID([]byte("v2"))

	storage.Checkpoint(0)
	storage.Put(NewValueEntry(v1, []byte{1})) // rev 1

	storage.Checkpoint(1)
	storage.Put(NewValueEntry(v2, []byte{2})) // rev 3

	storage.Checkpoint(2)
	storage.Put(NewValueEntry(v1, []byte{3})) // rev 5
	storage.Put(NewValueEntry(v2, []byte{4})) // rev 6
	storage.Put(NewValueEntry(v2, []byte{5})) // rev 7

	storage.Checkpoint(3)
	storage.Put(NewValueEntry(v1, []byte{6})) // rev 9
	storage.Put(NewValueEntry(v2, []byte{7})) // rev 10

	t.Log("v1", storage.Get(v1).Value(), "v2", storage.Get(v2).Value())
	assert.Equal(t, []byte{6}, storage.Get(v1).Value())
	assert.Equal(t, []byte{7}, storage.Get(v2).Value())

	storage.Rollback(3)
	t.Log("v1", storage.Get(v1).Value(), "v2", storage.Get(v2).Value())
	assert.Equal(t, []byte{3}, storage.Get(v1).Value())
	assert.Equal(t, []byte{5}, storage.Get(v2).Value())

	storage.Rollback(1)
	t.Log("v1", storage.Get(v1).Value(), "v2", storage.Get(v2))
	assert.Equal(t, []byte{1}, storage.Get(v1).Value())
	assert.Nil(t, storage.Get(v2))

	storage.Rollback(0)
	t.Log("v1", storage.Get(v1), "v2", storage.Get(v2))
	assert.Nil(t, storage.Get(v1))
	assert.Nil(t, storage.Get(v2))
}

func TestStorageDelete(t *testing.T) {
	storage := NewBufferedStorage(nil, nil)
	v1 := types.GetHashID([]byte("v1"))
	v2 := types.GetHashID([]byte("v2"))

	storage.Put(NewValueEntry(v1, []byte{1}))
	storage.Put(NewValueEntry(v2, []byte{2}))

	storage.Checkpoint(0)
	storage.Put(NewValueEntry(v1, []byte{3}))
	storage.Put(NewValueEntry(v2, []byte{4}))
	storage.Put(NewValueEntry(v2, []byte{5}))

	storage.Checkpoint(1)
	storage.Put(NewValueEntry(v1, []byte{6}))
	storage.Put(NewValueEntry(v2, []byte{7}))
	storage.Put(NewValueEntryDelete(v2))
	t.Log("v1", storage.Get(v1).Value(), "v2", storage.Get(v2).Value())
	assert.Equal(t, []byte{6}, storage.Get(v1).Value())
	assert.Equal(t, []byte{0}, storage.Get(v2).Hash())
	assert.Nil(t, storage.Get(v2).Value())

	storage.Rollback(1)
	t.Log("v1", storage.Get(v1).Value(), "v2", storage.Get(v2).Value())
	assert.Equal(t, []byte{3}, storage.Get(v1).Value())
	assert.Equal(t, []byte{5}, storage.Get(v2).Value())

	storage.Put(NewValueEntryDelete(v2))
	t.Log("v1", storage.Get(v1).Value(), "v2", storage.Get(v2).Value())
	assert.Equal(t, []byte{3}, storage.Get(v1).Value())
	assert.Equal(t, []byte{0}, storage.Get(v2).Hash())
	assert.Nil(t, storage.Get(v2).Value())
}

func TestStorageHasKey(t *testing.T) {
	storage := NewBufferedStorage(nil, nil)
	v1 := types.GetHashID([]byte("v1"))

	assert.False(t, storage.Has(v1, false)) // check buffer only
	assert.False(t, storage.Has(v1, true))  // check buffer and trie

	// put entry
	storage.Put(NewValueEntry(v1, []byte{1}))
	assert.True(t, storage.Has(v1, false)) // buffer has key
	assert.True(t, storage.Has(v1, true))  // buffer has key

	// update storage and reset buffer
	err := storage.Update()
	assert.NoError(t, err, "failed to update storage")
	err = storage.Buffer.Reset()
	assert.NoError(t, err, "failed to reset buffer")
	// after update and reset
	assert.False(t, storage.Has(v1, false)) // buffer doesn't have key
	assert.True(t, storage.Has(v1, true))   // buffer doesn't have, but trie has key

	// delete entry
	storage.Put(NewValueEntryDelete(v1))
	assert.True(t, storage.Has(v1, false)) // buffer has key
	assert.True(t, storage.Has(v1, true))  // buffer has key

	// update storage and reset buffer
	err = storage.Update()
	assert.NoError(t, err, "failed to update storage")
	err = storage.Buffer.Reset()
	assert.NoError(t, err, "failed to reset buffer")
	// after update and reset
	assert.False(t, storage.Has(v1, false)) // buffer doesn't have key
	assert.False(t, storage.Has(v1, true))  // buffer and trie don't have key
}
