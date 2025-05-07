package raftv2

import (
	"fmt"
	"github.com/aergoio/aergo/v2/types"
	"testing"
)

func TestLoop(t *testing.T) {
	// simulate loop in BlockFactory#getHardStateOfBlock()
	initialHeight := types.BlockNo(1000000000)
	steps := types.BlockNo(1)
	cnt := 0
	for i := initialHeight - 1; true; i -= steps {
		cnt++
		fmt.Printf("i: %d, height: %d, steps: 0x%x\n", cnt, i, steps)
		steps <<= 1
		if i < steps || steps >= overflowBarrier {
			break
		}
	}
}
