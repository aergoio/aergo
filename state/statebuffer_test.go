package state

import (
	"encoding/hex"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestBufferStack(t *testing.T) {
	stk := newStack()
	assert.Equal(t, -1, stk.peek())
	t.Log("stk", stk)

	stk.push(1)
	assert.Equal(t, 1, stk.peek())
	t.Log("stk", stk)

	stk.push(2).push(3).push(4)
	stk.push(5)
	t.Log("stk", stk)

	assert.Equal(t, 5, stk.peek())
	t.Log("stk", stk)

	assert.Equal(t, 5, stk.pop())
	assert.Equal(t, 4, stk.peek())
	t.Log("stk", stk)

	stk.push(6)
	t.Log("stk", stk)

	assert.Equal(t, 6, stk.peek())
	assert.Equal(t, 6, stk.pop())
	assert.Equal(t, 4, stk.pop())
	assert.Equal(t, 3, stk.pop())
	t.Log("stk", stk)

	assert.Equal(t, 2, stk.peek())
	assert.Equal(t, 2, stk.pop())
	assert.Equal(t, 1, stk.pop())
	t.Log("stk", stk)

	assert.Equal(t, -1, stk.peek())
	assert.Equal(t, -1, stk.pop())
	assert.Equal(t, -1, stk.pop())
	t.Log("stk", stk)

	stk.push(7).push(8, 9, 10)
	t.Log("stk", stk)

	assert.Equal(t, 10, stk.peek())
	assert.Equal(t, 10, stk.pop())
	assert.Equal(t, 9, stk.pop())
	assert.Equal(t, 8, stk.pop())
	t.Log("stk", stk)
}

// func printBufferIndex(t *testing.T, kset []types.HashID, idxs bufferIndex) {
// 	for i, v := range kset {
// 		t.Logf("(%d)[%v]%v", i, hex.EncodeToString(v[:]), idxs[v])
// 	}
// }

func TestBufferIndexStack(t *testing.T) {
	kset := []types.HashID{
		types.ToHashID([]byte{0x01}),
		types.ToHashID([]byte{0x02}),
	}
	k0, k1 := kset[0], kset[1]

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
	kset := []types.HashID{
		types.ToHashID([]byte{0x01}),
		types.ToHashID([]byte{0x02}),
	}
	k0, k1 := kset[0], kset[1]

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
