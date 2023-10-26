package fee

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxBaseFee(t *testing.T) {
	DisableZeroFee()
	defer EnableZeroFee()

	for _, test := range []struct {
		forkVersion int32
		gasPrice    *big.Int
		payloadSize int
		expectFee   *big.Int
	}{
		// v1
		{1, Gaer(1), 0, Gaer(2000000)},          // gas price not affect in v1
		{1, Gaer(5), 200, Gaer(2000000)},        // max freeByteSize
		{1, Gaer(5), 201, Gaer(2005000)},        // 2000000+5000
		{1, Gaer(5), 2047800, Gaer(1026000000)}, // 2000000+5000*2048000 ( 2047800 + freeByteSize )
		{1, Gaer(5), 2048000, Gaer(1026000000)}, // exceed payload max size
		{1, Gaer(5), 20480000, Gaer(1026000000)},

		// v2 - 1
		{2, Gaer(1), 0, Gaer(100000)},
		{2, Gaer(1), 200, Gaer(100000)},      // max freeByteSize
		{2, Gaer(1), 201, Gaer(100005)},      // 100000+5
		{2, Gaer(1), 2047800, Gaer(1124000)}, // 100000+5*204800 ( 2047800 + freeByteSize )
		{2, Gaer(1), 2048000, Gaer(1124000)}, // exceed payload max size
		{2, Gaer(1), 20480000, Gaer(1124000)},

		// v2 - 5
		{2, Gaer(5), 0, Gaer(500000)},
		{2, Gaer(5), 200, Gaer(500000)}, // max freeByteSize
		{2, Gaer(5), 201, Gaer(500025)},
		{2, Gaer(5), 700, Gaer(512500)},
		{2, Gaer(5), 2047800, Gaer(5620000)},
		{2, Gaer(5), 2048000, Gaer(5620000)}, // exceed payload max size

		// v3 is same as v2
		{3, Gaer(5), 100, Gaer(500000)},
	} {
		resultFee := TxBaseFee(test.forkVersion, test.gasPrice, test.payloadSize)
		assert.EqualValues(t, test.expectFee, resultFee, "TxFee(forkVersion:%d, payloadSize:%d, gasPrice:%s)", test.forkVersion, test.payloadSize, test.gasPrice)
	}
}

// TODO : replace to types.NewAmount after resolve cycling import
func Gaer(n int) *big.Int {
	return big.NewInt(0).Mul(big.NewInt(int64(n)), big.NewInt(int64(1e9)))
}
