package contract

import (
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func initContractTest(t *testing.T) {
	PubNet = true
}

func deinitContractTest(t *testing.T) {
	PubNet = false
}

func TestCheckExecution(t *testing.T) {
	initContractTest(t)
	defer deinitContractTest(t)

	for _, test := range []struct {
		version     int32
		txType      types.TxType
		amount      *big.Int
		payloadSize int
		isDeploy    bool
		isContract  bool

		expectErr  error
		expectExec bool
	}{
		// deploy
		{version: 2, txType: types.TxType_NORMAL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: true, isContract: false, expectErr: nil, expectExec: true},
		{version: 2, txType: types.TxType_DEPLOY, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: true, isContract: false, expectErr: nil, expectExec: true},
		{version: 2, txType: types.TxType_REDEPLOY, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: true, isContract: false, expectErr: nil, expectExec: true},
		{version: 3, txType: types.TxType_NORMAL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: true, isContract: false, expectErr: nil, expectExec: true},
		{version: 3, txType: types.TxType_DEPLOY, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: true, isContract: false, expectErr: nil, expectExec: true},
		{version: 3, txType: types.TxType_REDEPLOY, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: true, isContract: false, expectErr: nil, expectExec: true},
		// recipient is contract
		{version: 2, txType: types.TxType_NORMAL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: true, expectErr: nil, expectExec: true},
		{version: 2, txType: types.TxType_TRANSFER, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: true, expectErr: nil, expectExec: true},
		{version: 2, txType: types.TxType_CALL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: true, expectErr: nil, expectExec: true},
		{version: 2, txType: types.TxType_FEEDELEGATION, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: true, expectErr: nil, expectExec: true},
		{version: 3, txType: types.TxType_NORMAL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: true, expectErr: nil, expectExec: true},
		{version: 3, txType: types.TxType_TRANSFER, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: true, expectErr: nil, expectExec: true},
		{version: 3, txType: types.TxType_CALL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: true, expectErr: nil, expectExec: true},
		{version: 3, txType: types.TxType_FEEDELEGATION, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: true, expectErr: nil, expectExec: true},
		// recipient is not a contract
		{version: 2, txType: types.TxType_NORMAL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: false, expectErr: nil, expectExec: false},
		{version: 2, txType: types.TxType_TRANSFER, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: false, expectErr: nil, expectExec: false},
		{version: 2, txType: types.TxType_CALL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: false, expectErr: nil, expectExec: false},
		{version: 2, txType: types.TxType_FEEDELEGATION, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: false, expectErr: nil, expectExec: false},
		{version: 3, txType: types.TxType_NORMAL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: false, expectErr: nil, expectExec: false},
		{version: 3, txType: types.TxType_TRANSFER, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: false, expectErr: nil, expectExec: false},
		{version: 3, txType: types.TxType_CALL, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: false, expectErr: nil, expectExec: true},
		{version: 3, txType: types.TxType_FEEDELEGATION, amount: types.NewAmount(1, types.Aergo), payloadSize: 1000, isDeploy: false, isContract: false, expectErr: nil, expectExec: false},
	} {
		do_execute, err := checkExecution(test.txType, test.amount, test.payloadSize, test.version, test.isDeploy, test.isContract)
		assert.Equal(t, test.expectErr, err, "checkExecution(version:%d, txType:%d, amount:%s, payloadSize:%d)", test.version, test.txType, test.amount, test.payloadSize)
		assert.Equal(t, test.expectExec, do_execute, "checkExecution(version:%d, txType:%d, amount:%s, payloadSize:%d)", test.version, test.txType, test.amount, test.payloadSize)
	}
}
