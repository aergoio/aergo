package statedb

import (
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	kset = []types.HashID{
		types.ToHashID([]byte{0x01}),
		types.ToHashID([]byte{0x02}),
	}
	k0, k1 = kset[0], kset[1]
)

func TestBufferIndexStack(t *testing.T) {
	idxs := bufferIndex{}

	idxs.push(k0, 0)
	idxs.push(k0, 1)
	idxs.push(k1, 2)
	idxs.push(k0, 3)
	idxs.push(k0, 4)
	idxs.push(k1, 5)
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.Encode(v[:]), idxs[v])
	}

	assert.Equal(t, 4, idxs.pop(k0))
	assert.Equal(t, 3, idxs.pop(k0))
	assert.Equal(t, 1, idxs.pop(k0))
	assert.Equal(t, 0, idxs.pop(k0))
	assert.Equal(t, 5, idxs.peek(k1))
	assert.Equal(t, 5, idxs.peek(k1))
	assert.Equal(t, 5, idxs.pop(k1))
	assert.Equal(t, 2, idxs.peek(k1))
	assert.Equal(t, 2, idxs.pop(k1))
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.Encode(v[:]), idxs[v])
	}

	idxs.push(k0, 6, 8, 12)
	idxs.push(k1, 7, 9, 10, 11)
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.Encode(v[:]), idxs[v])
	}

	assert.Equal(t, 12, idxs[k0].peek())
	assert.Equal(t, 11, idxs[k1].peek())
}
func TestBufferIndexRollback(t *testing.T) {
	idxs := bufferIndex{}

	idxs.push(k0, 0, 1, 3, 4, 6, 7, 8)
	idxs.push(k1, 2, 5, 9)
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.Encode(v[:]), idxs[v])
	}

	assert.Equal(t, 8, idxs[k0].peek())
	assert.Equal(t, 9, idxs[k1].peek())

	idxs.rollback(5)
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.Encode(v[:]), idxs[v])
	}

	assert.Equal(t, 4, idxs[k0].peek())
	assert.Equal(t, 2, idxs[k1].peek())

	idxs.rollback(0)
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.Encode(v[:]), idxs[v])
	}

	assert.Equal(t, -1, idxs[k0].peek())
	assert.Equal(t, -1, idxs[k1].peek())
}

func TestBufferRollback(t *testing.T) {
	stb := NewStateBuffer()

	assert.Equal(t, 0, stb.Snapshot())
	stb.Put(NewValueEntry(k0, []byte{1})) // rev 1: k0=1
	stb.Put(NewValueEntry(k0, []byte{2})) // rev 2: k0=2
	stb.Put(NewValueEntry(k0, []byte{3})) // rev 3: k0=3
	stb.Put(NewValueEntry(k0, []byte{4})) // rev 4: k0=4
	stb.Put(NewValueEntry(k1, []byte{1})) // rev 5: k0=4, k1=1
	stb.Put(NewValueEntry(k1, []byte{2})) // rev 6: k0=4, k1=2
	stb.Put(NewValueEntry(k1, []byte{3})) // rev 7: k0=4, k1=3

	// snapshot revision 7
	revision := stb.Snapshot() // 7
	assert.Equal(t, 7, stb.Snapshot())

	stb.Put(NewValueEntry(k0, []byte{5})) // rev 8: k0=5, k1=3
	stb.Put(NewValueEntry(k0, []byte{6})) // rev 9: k0=6, k1=3
	assert.Equal(t, []byte{6}, stb.Get(k0).Value())
	assert.Equal(t, []byte{3}, stb.Get(k1).Value())
	t.Logf("k0: %v, k1: %v", stb.Get(k0).Value(), stb.Get(k1).Value())

	// rollback to revision 7
	stb.Rollback(revision)
	assert.Equal(t, 7, stb.Snapshot())

	assert.Equal(t, []byte{4}, stb.Get(k0).Value())
	assert.Equal(t, []byte{3}, stb.Get(k1).Value())
	t.Logf("k0: %v, k1: %v", stb.Get(k0).Value(), stb.Get(k1).Value())

	stb.Put(NewValueEntry(k1, []byte{4})) // rev 8: k0=4, k1=4
	stb.Put(NewValueEntry(k1, []byte{5})) // rev 9: k0=4, k1=5
	stb.Put(NewValueEntry(k0, []byte{7})) // rev 10: k0=7, k1=5

	// snapshot revision 10
	revision = stb.Snapshot() // 10
	assert.Equal(t, 10, stb.Snapshot())

	stb.Put(NewValueEntry(k0, []byte{8})) // rev 11: k0=8, k1=5
	stb.Put(NewValueEntry(k1, []byte{6})) // rev 12: k0=8, k1=6
	assert.Equal(t, []byte{8}, stb.Get(k0).Value())
	assert.Equal(t, []byte{6}, stb.Get(k1).Value())
	t.Logf("k0: %v, k1: %v", stb.Get(k0).Value(), stb.Get(k1).Value())

	// rollback to revision 10
	stb.Rollback(revision) // 10
	assert.Equal(t, 10, stb.Snapshot())

	assert.Equal(t, []byte{7}, stb.Get(k0).Value())
	assert.Equal(t, []byte{5}, stb.Get(k1).Value())
	t.Logf("k0: %v, k1: %v", stb.Get(k0).Value(), stb.Get(k1).Value())
}

func TestBufferHasKey(t *testing.T) {
	stb := NewStateBuffer()
	assert.False(t, stb.Has(k0))

	stb.Put(NewValueEntry(k0, []byte{1}))
	assert.True(t, stb.Has(k0)) // buffer has key

	stb.Put(NewValueEntryDelete(k0))
	assert.True(t, stb.Has(k0)) // buffer has key for ValueEntryDelete

	stb.Put(NewValueEntry(k0, []byte{2}))
	assert.True(t, stb.Has(k0)) // buffer has key

	stb.Reset()
	assert.False(t, stb.Has(k0)) // buffer doesn't have key
}
