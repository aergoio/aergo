package fee

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReceiptGasUsed(t *testing.T) {
	DisableZeroFee()
	defer EnableZeroFee()

	for _, test := range []struct {
		version       int32
		isGovernance  bool
		txFee         *big.Int
		gasPrice      *big.Int
		expectGasUsed uint64
	}{
		// no gas used
		{1, false, big.NewInt(100), big.NewInt(5), 0}, // v1
		{2, true, big.NewInt(100), big.NewInt(5), 0},  // governance

		// gas used
		{2, false, big.NewInt(10), big.NewInt(1), 10},
		{2, false, big.NewInt(10), big.NewInt(5), 2},
		{2, false, big.NewInt(100), big.NewInt(1), 100},
		{2, false, big.NewInt(100), big.NewInt(5), 20},
	} {
		resultGasUsed := ReceiptGasUsed(test.version, test.isGovernance, test.txFee, test.gasPrice)
		assert.Equal(t, test.expectGasUsed, resultGasUsed, "GasUsed(txFee:%s, gasPrice:%s, isGovernance:%d, version:%d)", test.txFee, test.gasPrice, test.isGovernance, test.version)
	}
}

func TestTxGas(t *testing.T) {
	DisableZeroFee()
	defer EnableZeroFee()

	for _, test := range []struct {
		payloadSize   int
		expectGasUsed uint64
	}{
		// less than freeByteSize
		{0, 100000},
		{200, 100000},

		{201, 100005},
		{1000, 104000},
		{204800, 1123000},

		// more than payloadMaxSize + freeByteSize
		{205000, 1124000},
		{2048000, 1124000},
		{20480000, 1124000},
	} {
		resultGasUsed := TxGas(test.payloadSize)
		assert.Equal(t, int(test.expectGasUsed), int(resultGasUsed), "GasUsed(payloadSize:%d)", test.payloadSize)
	}
}

func TestMaxGasLimit(t *testing.T) {
	for _, test := range []struct {
		balance   *big.Int
		gasPrice  *big.Int
		expectGas uint64
	}{
		{big.NewInt(100), big.NewInt(5), 20},
		{big.NewInt(100), big.NewInt(1), 100},
		{big.NewInt(0), big.NewInt(5), 0},
		{big.NewInt(-100), big.NewInt(1), math.MaxUint64},
		{big.NewInt(-100), big.NewInt(5), math.MaxUint64},
	} {
		resultGas := MaxGasLimit(test.balance, test.gasPrice)
		assert.Equal(t, test.expectGas, resultGas, "MaxGasLimit(balance:%s, gasPrice:%s)", test.balance, test.gasPrice)
	}
}
