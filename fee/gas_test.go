package fee

import (
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
