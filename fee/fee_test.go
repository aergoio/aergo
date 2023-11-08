package fee

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.EqualValues(t, test.expectFee, resultFee, "TxBaseFee(forkVersion:%d, payloadSize:%d, gasPrice:%s)", test.forkVersion, test.payloadSize, test.gasPrice)
	}
}

func TestTxExecuteFee(t *testing.T) {
	DisableZeroFee()
	defer EnableZeroFee()

	for _, test := range []struct {
		forkVersion       int32
		dbUpdateTotalSize int64
		gasPrice          *big.Int
		usedGas           uint64
		expectFee         *big.Int
	}{
		// v1 - fee by dbUpdateTotalSize
		{1, 0, nil, 0, Gaer(0)},
		{1, 200, nil, 0, Gaer(0)},
		{1, 201, nil, 0, Gaer(5000)},
		{1, 300, nil, 0, Gaer(500000)},
		{1, 1200, nil, 0, Gaer(5000000)},
		{1, 10200, nil, 0, Gaer(50000000)},

		// after v2 - fee by gas * gasPrice
		{2, 0, Gaer(1), 0, Gaer(0)},
		{2, 0, Gaer(1), 200, Gaer(200)},
		{3, 0, Gaer(5), 200, Gaer(1000)},
		{3, 0, Gaer(5), 100000, Gaer(500000)},
	} {
		resultFee := TxExecuteFee(test.forkVersion, test.gasPrice, test.usedGas, test.dbUpdateTotalSize)
		assert.EqualValues(t, test.expectFee, resultFee, "TxExecuteFee(forkVersion:%d, gasPrice:%s, usedGas:%d, dbUpdateTotalSize:%d)", test.forkVersion, test.gasPrice, test.usedGas, test.dbUpdateTotalSize)
	}
}

func TestTxMaxFee(t *testing.T) {
	DisableZeroFee()
	defer EnableZeroFee()

	for _, test := range []struct {
		forkVersion int32
		lenPayload  int
		gasLimit    uint64
		gasPrice    *big.Int
		balance     *big.Int
		expectErr   bool
		expectFee   *big.Int
	}{
		// v1 - max fee by payload
		{1, 0, 0, nil, nil, false, Gaer(2000000)}, // base 0.002 AERGO
		{1, 1, 0, nil, nil, false, Gaer(1025000000)},
		{1, 200, 0, nil, nil, false, Gaer(1025000000)},
		{1, 201, 0, nil, nil, false, Gaer(1025005000)},
		{1, 300, 0, nil, nil, false, Gaer(1025500000)},
		{1, 1200, 0, nil, nil, false, Gaer(1030000000)},

		// after v2 - max fee by gas limit
		{2, 0, 1000, Gaer(5), nil, true, nil}, // smaller than base fee
		{2, 0, 100000, Gaer(1), nil, false, Gaer(100000)},
		{2, 0, 100000, Gaer(5), nil, false, Gaer(500000)},

		{2, 0, 0, Gaer(5), Gaer(500000), false, Gaer(500000)}, // no gas limit - get from balance
		{2, 0, 0, Gaer(5), Gaer(1000000), false, Gaer(1000000)},
		{2, 0, 0, Gaer(5), Gaer(-1), false, new(big.Int).Mul(new(big.Int).SetUint64(math.MaxUint64), Gaer(5))}, // no gas limit and no balance = max gas
	} {
		resultFee, err := TxMaxFee(test.forkVersion, test.lenPayload, test.gasLimit, test.balance, test.gasPrice)
		require.Equal(t, test.expectErr, err != nil, "TxMaxFee(forkVersion:%d, lenPayload:%d, gasLimit:%d, balance:%s, gasPrice:%s)", test.forkVersion, test.lenPayload, test.gasLimit, test.balance, test.gasPrice)
		require.EqualValues(t, test.expectFee, resultFee, "TxMaxFee(forkVersion:%d, lenPayload:%d, gasLimit:%d, balance:%s, gasPrice:%s)", test.forkVersion, test.lenPayload, test.gasLimit, test.balance, test.gasPrice)
	}
}

func Gaer(n int) *big.Int {
	return big.NewInt(0).Mul(big.NewInt(int64(n)), big.NewInt(int64(1e9)))
}
