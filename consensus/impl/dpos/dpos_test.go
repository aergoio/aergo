package dpos

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/consensus/impl/dpos/slot"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

const (
	nSlots     = 5
	bpInterval = 1
)

func TestDposFutureBlock(t *testing.T) {
	slot.Init(bpInterval)

	dpos := &DPoS{}

	block := types.NewBlock(nil, nil, nil, nil, nil, time.Now().Add(3*time.Second).UnixNano())
	assert.True(t, !dpos.VerifyTimestamp(block), "future block check failed")

	block = types.NewBlock(nil, nil, nil, nil, nil, time.Now().UnixNano())
	assert.True(t, dpos.VerifyTimestamp(block), "future block check failed")

	block = types.NewBlock(nil, nil, nil, nil, nil, time.Now().Add(-time.Second).UnixNano())
	assert.True(t, dpos.VerifyTimestamp(block), "future block check failed")

}

func TestDposPastBlock(t *testing.T) {
	slot.Init(bpInterval)

	dpos := &DPoS{}

	block0 := types.NewBlock(nil, nil, nil, nil, nil, time.Now().UnixNano())
	assert.True(t, dpos.VerifyTimestamp(block0), "invalid timestamp")

	time.Sleep(time.Second)
	now := time.Now().UnixNano()
	block1 := types.NewBlock(block0, nil, nil, nil, nil, now)
	assert.True(t, dpos.VerifyTimestamp(block1), "invalid timestamp")

	// Add LIB, manually.
	dpos.Status = &Status{libState: &libStatus{}}
	dpos.Status.libState.Lib = newBlockInfo(block1)
	block2 := types.NewBlock(block0, nil, nil, nil, nil, now)
	// Test whether a block number error is raised or not by checking the
	// return value.
	assert.True(t, !dpos.VerifyTimestamp(block2), "block number error must be raised")
}
