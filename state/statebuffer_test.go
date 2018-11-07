package state

import (
	"encoding/hex"
	"testing"

	"github.com/aergoio/aergo/types"
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

// func TestBufferBufferedEntryRollback(t *testing.T) {
// 	h10 := types.ToHashID([]byte{0x10})
// 	h20 := types.ToHashID([]byte{0x20})

// 	stb := newStateBuffer()

// 	assert.Equal(t, 0, stb.snapshot())
// 	stb.put(newValueEntry(k0, []byte{1})) // rev 1: k0=1
// 	stb.put(newValueEntry(k0, []byte{2})) // rev 2: k0=2
// 	stb.put(newValueEntry(k0, []byte{3})) // rev 3: k0=3
// 	stb.put(newValueEntry(k0, []byte{4})) // rev 4: k0=4

// 	d1 := newStateBuffer()
// 	d1.put(newValueEntry(h10, []byte{11}))
// 	d1.put(newValueEntry(h10, []byte{12}))
// 	d1.put(newValueEntry(h20, []byte{21}))
// 	stb.put(newBufferedEntry(k1, []byte{1}, d1, nil)) // rev 5: k0=4, k1=1{h10:12,h20:21}

// 	// snapshot revision 5
// 	revision := stb.snapshot()
// 	assert.Equal(t, 5, stb.snapshot())

// 	stb.put(newValueEntry(k0, []byte{5})) // rev 6: k0=5, k1=1{h10:12,h20:21}

// 	d1.put(newValueEntry(h10, []byte{13}))
// 	d1.put(newValueEntry(h10, []byte{14}))
// 	d1.put(newValueEntry(h20, []byte{22}))
// 	stb.put(newBufferedEntry(k1, []byte{2}, d1, nil)) // rev 7: k0=5, k1=2{h10:14,h20:22}

// 	d1.put(newValueEntry(h10, []byte{15}))
// 	d1.put(newValueEntry(h10, []byte{16}))
// 	d1.put(newValueEntry(h20, []byte{23}))
// 	stb.put(newBufferedEntry(k1, []byte{3}, d1, nil)) // rev 8: k0=5, k1=3{h10:16,h20:23}

// 	be := getBufferedEntry(stb, k1)
// 	t.Logf("k0=%v, k1=%v{h10:%v,h20:%v}", stb.get(k0).Value(),
// 		be.Value(), be.buffer.get(h10).Value(), be.buffer.get(h20).Value())
// 	assert.Equal(t, []byte{5}, stb.get(k0).Value())
// 	assert.Equal(t, []byte{3}, be.Value())
// 	assert.Equal(t, []byte{16}, be.buffer.get(h10).Value())
// 	assert.Equal(t, []byte{23}, be.buffer.get(h20).Value())

// 	// rollback to snapshot 5 // rev 5: k0=4, k1=1{h10:12,h20:21}
// 	stb.rollback(revision)

// 	be = getBufferedEntry(stb, k1)
// 	t.Logf("k0=%v, k1=%v{h10:%v,h20:%v}", stb.get(k0).Value(),
// 		be.Value(), be.buffer.get(h10).Value(), be.buffer.get(h20).Value())
// 	assert.Equal(t, []byte{4}, stb.get(k0).Value())
// 	assert.Equal(t, []byte{1}, be.Value())
// 	assert.Equal(t, []byte{12}, be.buffer.get(h10).Value())
// 	assert.Equal(t, []byte{21}, be.buffer.get(h20).Value())

// }

// func TestBufferLookupCached(t *testing.T) {
// 	v0 := types.ToHashID([]byte{0x10})
// 	stb := newStateBuffer()

// 	cs0 := &ContractState{
// 		State:  &types.State{Nonce: 0},
// 		buffer: newStateBuffer(),
// 	}

// 	cs1 := ContractState(*cs0)
// 	cs1.revision = cs0.buffer.snapshot()
// 	cs1.State.Nonce = 1
// 	t.Log("1.getHash", cs1.HashID())
// 	stb.put(newValueEntry(k0, &cs1))

// 	cs2 := ContractState(*cs0)
// 	cs2.buffer.put(newValueEntry(v0, []byte{0x01}))
// 	cs2.buffer.put(newValueEntry(v0, []byte{0x02}))
// 	cs2.revision = cs0.buffer.snapshot()
// 	cs2.State.Nonce = 2
// 	t.Log("2.getHash", cs2.HashID())
// 	stb.put(newValueEntry(k0, &cs2))

// 	t.Log("getHash", stb.get(k0).HashID())
// 	t.Log("getNonce", stb.get(k0).Value().(*ContractState).Nonce)
// 	t.Log("lookupCache.v0", stb.lookupCache(k0).get(v0).Value())

// 	revision := stb.snapshot() // revision 2
// 	t.Log("snapshot", revision)

// 	cs3 := new(ContractState)
// 	*cs3 = ContractState(*cs0)
// 	//cs3 := ContractState(*cs0)
// 	cs3.buffer.put(newValueEntry(v0, []byte{0x03}))
// 	cs3.buffer.put(newValueEntry(v0, []byte{0x04}))
// 	cs3.buffer.put(newValueEntry(v0, []byte{0x05}))
// 	cs3.revision = cs3.buffer.snapshot()
// 	cs3.State.Nonce = 3
// 	t.Log("3.getHash", cs3.HashID())
// 	stb.put(newValueEntry(k0, cs3))

// 	st4 := new(types.State)
// 	*st4 = types.State(*cs0.State)
// 	//st4 := types.State(*cs0.State)
// 	st4.Nonce = 4
// 	t.Log("4.getHash", getHash(st4))
// 	stb.put(newValueEntry(k0, st4))

// 	t.Log("getHash", stb.get(k0).HashID())
// 	t.Log("getNonce", stb.get(k0).Value().(*types.State).Nonce)
// 	t.Log("lookupCache.v0", stb.lookupCache(k0).get(v0).Value())

// 	// rollback
// 	t.Log("rollback", revision)
// 	stb.rollback(revision) // revision 2

// 	// after rollback
// 	cs2c := stb.get(k0).Value().(*ContractState)
// 	t.Log("cs2c", cs2c)
// 	t.Log("state", stb.get(k0).Value().(*ContractState))
// 	t.Log("getHash", stb.get(k0).HashID())
// 	t.Log("getNonce", stb.get(k0).Value().(*ContractState).Nonce)
// 	t.Log("lookupCache.v0", stb.lookupCache(k0).get(v0).Value())

// 	t.Log("cs0", cs0)
// 	t.Log("cs1", cs1)
// 	t.Log("cs2", cs2)
// 	t.Log("cs3", cs3)
// 	t.Log("st4", st4)
// }

// func TestBufferClone(t *testing.T) {
// 	st0 := &types.State{Nonce: 0}
// 	t.Log("st0", st0)

// 	st1 := new(types.State)
// 	*st1 = types.State(*st0)
// 	st1.Nonce = 1
// 	t.Log("st0", st0)
// 	t.Log("st1", st1)

// 	st2 := types.State(*st0)
// 	st2.Nonce = 2
// 	t.Log("st0", st0)
// 	t.Log("st2", &st2)

// 	st3 := types.State(*st0)
// 	st3.Nonce = 3
// 	st0.Nonce = 4
// 	t.Log("st0", st0)
// 	t.Log("st1", st1)
// 	t.Log("st2", &st2)
// 	t.Log("st3", &st3)
// }
