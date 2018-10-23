package dpos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockInfo(t *testing.T) {
	var bi *blockInfo
	a := assert.New(t)
	a.Equal("(nil)", bi.Hash())
}
