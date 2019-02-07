package dpos

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/consensus/impl/dpos/slot"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestDposFutureBlock(t *testing.T) {
	const (
		nSlots     = 5
		bpInterval = 1
	)

	slot.Init(bpInterval, nSlots)

	dpos := &DPoS{}

	block := types.NewBlock(nil, nil, nil, nil, nil, time.Now().Add(3*time.Second).UnixNano())
	assert.True(t, !dpos.VerifyTimestamp(block), "future block check failed")

	block = types.NewBlock(nil, nil, nil, nil, nil, time.Now().UnixNano())
	assert.True(t, dpos.VerifyTimestamp(block), "future block check failed")

	block = types.NewBlock(nil, nil, nil, nil, nil, time.Now().Add(-time.Second).UnixNano())
	assert.True(t, dpos.VerifyTimestamp(block), "future block check failed")

}
