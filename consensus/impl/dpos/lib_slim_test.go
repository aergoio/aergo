package dpos

import (
	"fmt"
	"testing"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

type slimTestCDB struct {
	consensus.ChainDB
	chain *testChain
}

func (d *slimTestCDB) GetBlockByNo(no types.BlockNo) (*types.Block, error) {
	if no >= types.BlockNo(len(d.chain.chain)) {
		return nil, fmt.Errorf("block %d not found", no)
	}
	return d.chain.chain[no], nil
}

func loadLibStatusForTest(t *testing.T, tc *testChain, encoded []byte) *libStatus {
	t.Helper()

	oldLoader := bsLoader
	t.Cleanup(func() { bsLoader = oldLoader })

	best := tc.chain[len(tc.chain)-1]
	confirmsRequired := tc.status.libState.confirmsRequired
	bsLoader = &bootLoader{
		ls:               newLibStatus(confirmsRequired),
		best:             best,
		genesis:          tc.chain[0],
		cdb:              &slimTestCDB{chain: tc},
		confirmsRequired: confirmsRequired,
	}

	pls := newLibStatus(confirmsRequired)
	if err := common.GobDecode(encoded, pls); err != nil {
		t.Fatalf("failed to decode lib status: %v", err)
	}
	pls.load(best.BlockNo())
	return pls
}

func assertSameLibStatus(t *testing.T, want, got *libStatus) {
	t.Helper()

	assert.Equal(t, want.Lib.BlockNo, got.Lib.BlockNo)
	assert.Equal(t, want.Lib.BlockHash, got.Lib.BlockHash)
	assert.Equal(t, want.LpbNo, got.LpbNo)
	assert.Equal(t, len(want.Prpsd), len(got.Prpsd))
	for bpID, pl := range want.Prpsd {
		gpl, ok := got.Prpsd[bpID]
		if !assert.True(t, ok, "missing Prpsd entry for %s", bpID) {
			continue
		}
		assert.Equal(t, pl.Plib, gpl.Plib)
		assert.Equal(t, pl.PlibBy, gpl.PlibBy)
	}

	assert.Equal(t, want.confirms.Len(), got.confirms.Len())
	we, ge := want.confirms.Front(), got.confirms.Front()
	for we != nil && ge != nil {
		wc, gc := cInfo(we), cInfo(ge)
		assert.Equal(t, wc.BlockNo, gc.BlockNo)
		assert.Equal(t, wc.bpid, gc.bpid)
		assert.Equal(t, wc.confirmsLeft, gc.confirmsLeft)
		we, ge = we.Next(), ge.Next()
	}
}

func TestSlimLibStatusCompatibility(t *testing.T) {
	const clusterSize = 3
	const maxBlockNo = 40

	tc, err := newTestChain(clusterSize)
	assert.NoError(t, err)
	for i := types.BlockNo(1); i <= maxBlockNo; i++ {
		assert.NoError(t, tc.addBlock(i))
	}

	bestNo := types.BlockNo(len(tc.chain) - 1)
	full, err := common.GobEncode(tc.status.libState)
	assert.NoError(t, err)

	slim := tc.status.libState.slim(bestNo)
	slimEncoded, err := common.GobEncode(slim)
	assert.NoError(t, err)

	// In a healthy round-robin cluster every BP produced inside the replay
	// window, so the persisted Prpsd map is empty.
	assert.Empty(t, slim.Prpsd)
	assert.Less(t, len(slimEncoded), len(full))

	fullLoaded := loadLibStatusForTest(t, tc, full)
	slimLoaded := loadLibStatusForTest(t, tc, slimEncoded)

	assertSameLibStatus(t, fullLoaded, slimLoaded)
	assert.Equal(t, tc.status.libState.Lib.BlockNo, slimLoaded.Lib.BlockNo)
}

func TestSlimLibStatusKeepsInactiveEntries(t *testing.T) {
	const clusterSize = 3
	const maxBlockNo = 40

	tc, err := newTestChain(clusterSize)
	assert.NoError(t, err)
	for i := types.BlockNo(1); i <= maxBlockNo; i++ {
		assert.NoError(t, tc.addBlock(i))
	}

	bestNo := types.BlockNo(len(tc.chain) - 1)

	// Simulate a BP whose last proposal is older than the replay window.
	ls := *tc.status.libState
	ls.Prpsd = make(proposed, len(tc.status.libState.Prpsd)+1)
	for bpID, p := range tc.status.libState.Prpsd {
		ls.Prpsd[bpID] = p
	}
	old := tc.chain[1]
	ls.Prpsd["inactive-bp"] = &plInfo{
		Plib:   newBlockInfo(old),
		PlibBy: newBlockInfo(old),
	}

	full, err := common.GobEncode(&ls)
	assert.NoError(t, err)

	slim := ls.slim(bestNo)
	_, exists := slim.Prpsd["inactive-bp"]
	assert.True(t, exists, "inactive proposal must survive the replay window filter")

	slimEncoded, err := common.GobEncode(slim)
	assert.NoError(t, err)
	assert.Less(t, len(slimEncoded), len(full))

	fullLoaded := loadLibStatusForTest(t, tc, full)
	slimLoaded := loadLibStatusForTest(t, tc, slimEncoded)

	assertSameLibStatus(t, fullLoaded, slimLoaded)
	assert.Contains(t, slimLoaded.Prpsd, "inactive-bp")
}
