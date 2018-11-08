package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtilStackBasic(t *testing.T) {
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

func TestUtilStackExport(t *testing.T) {
	stk := newStack()
	stk.push(1, 2, 3, 4, 5)
	t.Log("stk", stk)

	assert.Equal(t, []int{5, 4, 3, 2, 1}, stk.export())
	t.Log("export", stk.export())

	assert.Equal(t, []int{1, 2, 3, 4, 5}, []int(*stk))
}

func TestUtilStackIter(t *testing.T) {
	cmp := newStack().push(1, 2, 3, 4, 5)
	stk := newStack().push(1, 2, 3, 4, 5)
	t.Log("stk", stk)

	it := stk.iter()
	for v := it(); v >= 0; v = it() {
		assert.Equal(t, cmp.pop(), v)
		t.Log("it", v)
	}

	it = stk.iter()
	assert.Equal(t, 5, it())
	assert.Equal(t, 4, it())
	assert.Equal(t, 3, it())
	assert.Equal(t, 2, it())
	assert.Equal(t, 1, it())
	assert.Equal(t, -1, it())
	assert.Equal(t, -1, it())

	assert.Equal(t, 5, stk.peek())
	assert.Equal(t, []int{1, 2, 3, 4, 5}, []int(*stk))
}

func TestUtilStackGet(t *testing.T) {
	cmp := newStack().push(1, 2, 3, 4, 5)
	stk := newStack().push(1, 2, 3, 4, 5)
	t.Log("stk", stk)

	for i := stk.last(); i >= 0; i-- {
		assert.Equal(t, cmp.pop(), stk.get(i))
		t.Logf("get(%d) %v", i, stk.get(i))
	}

	assert.Equal(t, 5, stk.get(4))
	assert.Equal(t, 4, stk.get(3))
	assert.Equal(t, 3, stk.get(2))
	assert.Equal(t, 2, stk.get(1))
	assert.Equal(t, 1, stk.get(0))
	assert.Equal(t, -1, stk.get(-1))
	assert.Equal(t, -1, stk.get(-2))

	assert.Equal(t, 5, stk.peek())
	assert.Equal(t, []int{1, 2, 3, 4, 5}, []int(*stk))
}
