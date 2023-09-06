package state

import (
	"encoding/hex"
	"testing"

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
		t.Logf("(%d)[%v]%v", i, hex.EncodeToString(v[:]), idxs[v])
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
		t.Logf("(%d)[%v]%v", i, hex.EncodeToString(v[:]), idxs[v])
	}

	idxs.push(k0, 6, 8, 12)
	idxs.push(k1, 7, 9, 10, 11)
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.EncodeToString(v[:]), idxs[v])
	}

	assert.Equal(t, 12, idxs[k0].peek())
	assert.Equal(t, 11, idxs[k1].peek())
}
func TestBufferIndexRollback(t *testing.T) {
	idxs := bufferIndex{}

	idxs.push(k0, 0, 1, 3, 4, 6, 7, 8)
	idxs.push(k1, 2, 5, 9)
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.EncodeToString(v[:]), idxs[v])
	}

	assert.Equal(t, 8, idxs[k0].peek())
	assert.Equal(t, 9, idxs[k1].peek())

	idxs.rollback(5)
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.EncodeToString(v[:]), idxs[v])
	}

	assert.Equal(t, 4, idxs[k0].peek())
	assert.Equal(t, 2, idxs[k1].peek())

	idxs.rollback(0)
	for i, v := range kset {
		t.Logf("(%d)[%v]%v", i, hex.EncodeToString(v[:]), idxs[v])
	}

	assert.Equal(t, -1, idxs[k0].peek())
	assert.Equal(t, -1, idxs[k1].peek())
}

func TestBufferRollback(t *testing.T) {
	stb := newStateBuffer()

	assert.Equal(t, 0, stb.snapshot())
	stb.put(newValueEntry(k0, []byte{1})) // rev 1: k0=1
	stb.put(newValueEntry(k0, []byte{2})) // rev 2: k0=2
	stb.put(newValueEntry(k0, []byte{3})) // rev 3: k0=3
	stb.put(newValueEntry(k0, []byte{4})) // rev 4: k0=4
	stb.put(newValueEntry(k1, []byte{1})) // rev 5: k0=4, k1=1
	stb.put(newValueEntry(k1, []byte{2})) // rev 6: k0=4, k1=2
	stb.put(newValueEntry(k1, []byte{3})) // rev 7: k0=4, k1=3

	// snapshot revision 7
	revision := stb.snapshot() // 7
	assert.Equal(t, 7, stb.snapshot())

	stb.put(newValueEntry(k0, []byte{5})) // rev 8: k0=5, k1=3
	stb.put(newValueEntry(k0, []byte{6})) // rev 9: k0=6, k1=3
	assert.Equal(t, []byte{6}, stb.get(k0).Value())
	assert.Equal(t, []byte{3}, stb.get(k1).Value())
	t.Logf("k0: %v, k1: %v", stb.get(k0).Value(), stb.get(k1).Value())

	// rollback to revision 7
	stb.rollback(revision)
	assert.Equal(t, 7, stb.snapshot())

	assert.Equal(t, []byte{4}, stb.get(k0).Value())
	assert.Equal(t, []byte{3}, stb.get(k1).Value())
	t.Logf("k0: %v, k1: %v", stb.get(k0).Value(), stb.get(k1).Value())

	stb.put(newValueEntry(k1, []byte{4})) // rev 8: k0=4, k1=4
	stb.put(newValueEntry(k1, []byte{5})) // rev 9: k0=4, k1=5
	stb.put(newValueEntry(k0, []byte{7})) // rev 10: k0=7, k1=5

	// snapshot revision 10
	revision = stb.snapshot() // 10
	assert.Equal(t, 10, stb.snapshot())

	stb.put(newValueEntry(k0, []byte{8})) // rev 11: k0=8, k1=5
	stb.put(newValueEntry(k1, []byte{6})) // rev 12: k0=8, k1=6
	assert.Equal(t, []byte{8}, stb.get(k0).Value())
	assert.Equal(t, []byte{6}, stb.get(k1).Value())
	t.Logf("k0: %v, k1: %v", stb.get(k0).Value(), stb.get(k1).Value())

	// rollback to revision 10
	stb.rollback(revision) // 10
	assert.Equal(t, 10, stb.snapshot())

	assert.Equal(t, []byte{7}, stb.get(k0).Value())
	assert.Equal(t, []byte{5}, stb.get(k1).Value())
	t.Logf("k0: %v, k1: %v", stb.get(k0).Value(), stb.get(k1).Value())
}

func TestBufferHasKey(t *testing.T) {
	stb := newStateBuffer()
	assert.False(t, stb.has(k0))

	stb.put(newValueEntry(k0, []byte{1}))
	assert.True(t, stb.has(k0)) // buffer has key

	stb.put(newValueEntryDelete(k0))
	assert.True(t, stb.has(k0)) // buffer has key for ValueEntryDelete

	stb.put(newValueEntry(k0, []byte{2}))
	assert.True(t, stb.has(k0)) // buffer has key

	stb.reset()
	assert.False(t, stb.has(k0)) // buffer doesn't have key
}
