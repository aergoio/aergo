package state

import (
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

// func TestCacheBasic(t *testing.T) {
// 	t.Log("StorageCache")
// }

func TestStorageBasic(t *testing.T) {
	storage := newBufferedStorage(nil, nil)
	v1 := types.GetHashID([]byte("v1"))
	v2 := types.GetHashID([]byte("v2"))

	storage.checkpoint(0)
	storage.put(newValueEntry(v1, []byte{1})) // rev 1

	storage.checkpoint(1)
	storage.put(newValueEntry(v2, []byte{2})) // rev 3

	storage.checkpoint(2)
	storage.put(newValueEntry(v1, []byte{3})) // rev 5
	storage.put(newValueEntry(v2, []byte{4})) // rev 6
	storage.put(newValueEntry(v2, []byte{5})) // rev 7

	storage.checkpoint(3)
	storage.put(newValueEntry(v1, []byte{6})) // rev 9
	storage.put(newValueEntry(v2, []byte{7})) // rev 10

	t.Log("v1", storage.get(v1).Value(), "v2", storage.get(v2).Value())
	assert.Equal(t, []byte{6}, storage.get(v1).Value())
	assert.Equal(t, []byte{7}, storage.get(v2).Value())

	storage.rollback(3)
	t.Log("v1", storage.get(v1).Value(), "v2", storage.get(v2).Value())
	assert.Equal(t, []byte{3}, storage.get(v1).Value())
	assert.Equal(t, []byte{5}, storage.get(v2).Value())

	storage.rollback(1)
	t.Log("v1", storage.get(v1).Value(), "v2", storage.get(v2))
	assert.Equal(t, []byte{1}, storage.get(v1).Value())
	assert.Nil(t, storage.get(v2))

	storage.rollback(0)
	t.Log("v1", storage.get(v1), "v2", storage.get(v2))
	assert.Nil(t, storage.get(v1))
	assert.Nil(t, storage.get(v2))
}

func TestStorageDelete(t *testing.T) {
	storage := newBufferedStorage(nil, nil)
	v1 := types.GetHashID([]byte("v1"))
	v2 := types.GetHashID([]byte("v2"))

	storage.put(newValueEntry(v1, []byte{1}))
	storage.put(newValueEntry(v2, []byte{2}))

	storage.checkpoint(0)
	storage.put(newValueEntry(v1, []byte{3}))
	storage.put(newValueEntry(v2, []byte{4}))
	storage.put(newValueEntry(v2, []byte{5}))

	storage.checkpoint(1)
	storage.put(newValueEntry(v1, []byte{6}))
	storage.put(newValueEntry(v2, []byte{7}))
	storage.put(newValueEntryDelete(v2))
	t.Log("v1", storage.get(v1).Value(), "v2", storage.get(v2).Value())
	assert.Equal(t, []byte{6}, storage.get(v1).Value())
	assert.Equal(t, []byte{0}, storage.get(v2).Hash())
	assert.Nil(t, storage.get(v2).Value())

	storage.rollback(1)
	t.Log("v1", storage.get(v1).Value(), "v2", storage.get(v2).Value())
	assert.Equal(t, []byte{3}, storage.get(v1).Value())
	assert.Equal(t, []byte{5}, storage.get(v2).Value())

	storage.put(newValueEntryDelete(v2))
	t.Log("v1", storage.get(v1).Value(), "v2", storage.get(v2).Value())
	assert.Equal(t, []byte{3}, storage.get(v1).Value())
	assert.Equal(t, []byte{0}, storage.get(v2).Hash())
	assert.Nil(t, storage.get(v2).Value())
}

func TestStorageHasKey(t *testing.T) {
	storage := newBufferedStorage(nil, nil)
	v1 := types.GetHashID([]byte("v1"))

	assert.False(t, storage.has(v1, false)) // check buffer only
	assert.False(t, storage.has(v1, true))  // check buffer and trie

	// put entry
	storage.put(newValueEntry(v1, []byte{1}))
	assert.True(t, storage.has(v1, false)) // buffer has key
	assert.True(t, storage.has(v1, true))  // buffer has key

	// update storage and reset buffer
	err := storage.update()
	assert.NoError(t, err, "failed to update storage")
	err = storage.buffer.reset()
	assert.NoError(t, err, "failed to reset buffer")
	// after update and reset
	assert.False(t, storage.has(v1, false)) // buffer doesn't have key
	assert.True(t, storage.has(v1, true))   // buffer doesn't have, but trie has key

	// delete entry
	storage.put(newValueEntryDelete(v1))
	assert.True(t, storage.has(v1, false)) // buffer has key
	assert.True(t, storage.has(v1, true))  // buffer has key

	// update storage and reset buffer
	err = storage.update()
	assert.NoError(t, err, "failed to update storage")
	err = storage.buffer.reset()
	assert.NoError(t, err, "failed to reset buffer")
	// after update and reset
	assert.False(t, storage.has(v1, false)) // buffer doesn't have key
	assert.False(t, storage.has(v1, true))  // buffer and trie don't have key
}
