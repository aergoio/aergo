package dpos

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/consensus/impl/dpos/slot"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

const (
	nSlots     = 5
	bpInterval = 1
)

func newBlock(ts int64) *types.Block {
	b, _ := types.NewChainID().Bytes()
	return types.NewBlock(
		&types.BlockHeaderInfo{Ts: ts, ChainId: b},
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func newBlockFromPrev(prev *types.Block, ts int64, bv types.BlockVersionner) *types.Block {
	return types.NewBlock(
		types.NewBlockHeaderInfoFromPrevBlock(prev, ts, bv),
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func TestDposFutureBlock(t *testing.T) {
	slot.Init(bpInterval)

	dpos := &DPoS{}

	block := newBlock(time.Now().Add(3 * time.Second).UnixNano())
	assert.True(t, !dpos.VerifyTimestamp(block), "future block check failed")

	block = newBlock(time.Now().UnixNano())
	assert.True(t, dpos.VerifyTimestamp(block), "future block check failed")

	block = newBlock(time.Now().Add(-time.Second).UnixNano())
	assert.True(t, dpos.VerifyTimestamp(block), "future block check failed")
}

func TestDposPastBlock(t *testing.T) {
	slot.Init(bpInterval)

	bv := types.DummyBlockVersionner(0)

	dpos := &DPoS{}

	block0 := newBlock(time.Now().UnixNano())
	assert.True(t, dpos.VerifyTimestamp(block0), "invalid timestamp")

	time.Sleep(time.Second)
	now := time.Now().UnixNano()
	block1 := newBlockFromPrev(block0, now, bv)
	assert.True(t, dpos.VerifyTimestamp(block1), "invalid timestamp")

	// Add LIB, manually.
	dpos.Status = &Status{libState: &libStatus{}}
	dpos.Status.libState.Lib = newBlockInfo(block1)
	block2 := newBlockFromPrev(block0, now, bv)
	// Test whether a block number error is raised or not by checking the
	// return value.
	assert.True(t, !dpos.VerifyTimestamp(block2), "block number error must be raised")
}
