package p2putil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestItem interface {
	ID() int
}
type testItem struct {
	id int
}

func (ti testItem) ID() int {
	return ti.id
}
func TestPressableQueue_Offer(t *testing.T) {
	const arrSize = 30
	var mos [arrSize]interface{}
	for i := 0; i < arrSize; i++ {
		mos[i] = &testItem{i}
	}
	tests := []struct {
		name string
		cap  int
		want bool
	}{
		{"T10", 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target1 := NewPressableQueue(tt.cap)
			target2 := NewPressableQueue(tt.cap)

			assert.True(t, target1.Empty())
			assert.True(t, target2.Empty())
			assert.False(t, target1.Full())
			assert.False(t, target2.Full())

			for i, mo := range mos {
				expected := i < tt.cap
				assert.Equal(t, expected, target1.Offer(mo))
				if expected {
					assert.Nil(t, target2.Press(mo))
				} else {
					assert.NotNil(t, target2.Press(mo))
				}
			}
			assert.Equal(t, tt.cap, target1.Size())
			assert.Equal(t, tt.cap, target2.Size())
			assert.False(t, target1.Empty())
			assert.False(t, target2.Empty())
			assert.True(t, target1.Full())
			assert.True(t, target2.Full())

			for i := 0; i < tt.cap; i++ {
				mo := target1.Poll().(TestItem)
				assert.NotNil(t, mo)
				assert.Equal(t, i, mo.ID())

				mo2 := target2.Poll().(TestItem)
				assert.NotNil(t, mo2)
				assert.Equal(t, arrSize-tt.cap+i, mo2.ID())
			}

			assert.True(t, target1.Empty())
			assert.True(t, target2.Empty())
			assert.False(t, target1.Full())
			assert.False(t, target2.Full())

		})
	}
}

func TestPressableQueue_Peek(t *testing.T) {
	const arrSize = 10
	var mos [arrSize]interface{}
	for i := 0; i < arrSize; i++ {
		mos[i] = &testItem{i}
	}
	tests := []struct {
		name string
		cap  int
		want bool
	}{
		{"T10", 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target1 := NewPressableQueue(tt.cap)
			for _, mo := range mos {
				target1.Offer(mo)
			}
			size := target1.Size()
			for size > 0 {

				actual1 := target1.Peek()
				assert.Equal(t, size, target1.Size())
				actual2 := target1.Poll()
				assert.Equal(t, size-1, target1.Size())
				assert.Equal(t, actual1, actual2)
				size = target1.Size()
			}
		})
	}
}
