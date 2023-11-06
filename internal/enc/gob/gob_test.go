package gob

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGobCodec(t *testing.T) {
	a := assert.New(t)

	x := []int{1, 2, 3}
	b, err := Encode(x)
	a.Nil(err)

	y := []int{0, 0, 0}
	err = Decode(b, &y)
	a.Nil(err)

	for i, v := range x {
		a.Equal(v, y[i])
	}
}
