package contract

import (
	"fmt"
	"testing"

	"github.com/aergoio/aergo/config"
	"github.com/stretchr/testify/require"
)

// TODO: move to hardfork_test.go
func TestHardForkVersion(t *testing.T) {
	for _, test := range []struct {
		blockNumV2Fork  uint64
		blockNumV3Fork  uint64
		blockNumCurrent uint64
		expectVersion   int32
	}{
		{0, 0, 0, 3}, // version 3

		{0, 100, 0, 2},   // version 2
		{0, 100, 99, 2},  // version 2
		{0, 100, 100, 3}, // version 3

		{100, 200, 0, 0},   // version 0
		{100, 200, 99, 0},  // version 0
		{100, 200, 100, 2}, // version 2
		{100, 200, 199, 2}, // version 2
		{100, 200, 200, 3}, // version 3
	} {
		// set hardfork config
		HardforkConfig = &config.HardforkConfig{
			V2: test.blockNumV2Fork,
			V3: test.blockNumV3Fork,
		}
		// check fork version
		version := HardforkConfig.Version(test.blockNumCurrent)
		require.Equal(t, test.expectVersion, version, "failed to check fork version")
	}
}

func TestMaxCallDepth(t *testing.T) {
	for _, test := range []struct {
		blockNumV2Fork  uint64
		blockNumV3Fork  uint64
		blockNumCurrent uint64

		expectMaxCallDepth int32
		isExceed           bool
	}{
		// version 3
		{10, 20, 20, maxCallDepth, false},
		{10, 20, 20, maxCallDepth + 1, true},

		// version 2
		{10, 20, 10, maxCallDepthOld, false},
		{10, 20, 10, maxCallDepthOld + 1, true},
		{10, 20, 19, maxCallDepthOld, false},
		{10, 20, 19, maxCallDepthOld + 1, true},

		// version 0
		{10, 20, 9, maxCallDepthOld, false},
		{10, 20, 9, maxCallDepthOld + 1, true},
	} {
		bc, err := LoadDummyChain()
		require.NoError(t, err, "failed to create test database")
		defer bc.Release()

		// set hardfork config
		HardforkConfig = &config.HardforkConfig{
			V2: test.blockNumV2Fork,
			V3: test.blockNumV3Fork,
		}

		// update dummy block
		for i := uint64(0); i < test.blockNumCurrent-1; i++ {
			err = bc.ConnectBlock()
			require.NoError(t, err, "failed to connect block")
		}

		err = bc.ConnectBlock(
			NewLuaTxAccount("kch", 1e18),
			NewLuaTxDef("kch", "make_call", 0, `
function make_call(remaining)
	remaining = remaining - 1
	if remaining >= 0 then
		contract.call(system.getContractID(), "make_call", remaining)
	end
end
abi.register(make_call)
`),
			NewLuaTxCall("kch", "make_call", 0, fmt.Sprintf(`{"Name":"make_call", "Args":[%d]}`, test.expectMaxCallDepth)),
		)
		require.EqualValues(t, test.isExceed, err != nil, "failed to check max call depth")
	}
}
