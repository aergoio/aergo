package contract

import (
	"math"
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

//----------------------------------------------------------------------------------------//
// tests for tx Execute functions

func TestTxFee(t *testing.T) {
	initContractTest(t)
	defer deinitContractTest(t)

	for _, test := range []struct {
		forkVersion int32
		payloadSize int
		gasPrice    *big.Int
		expectFee   *big.Int
	}{
		// v1
		{1, 0, types.NewAmount(1, types.Gaer), types.NewAmount(2000000, types.Gaer)},          // gas price not affect in v1
		{1, 200, types.NewAmount(5, types.Gaer), types.NewAmount(2000000, types.Gaer)},        // max freeByteSize
		{1, 201, types.NewAmount(5, types.Gaer), types.NewAmount(2005000, types.Gaer)},        // 2000000+5000
		{1, 2047800, types.NewAmount(5, types.Gaer), types.NewAmount(1026000000, types.Gaer)}, // 2000000+5000*2048000 ( 2047800 + freeByteSize )
		{1, 2048000, types.NewAmount(5, types.Gaer), types.NewAmount(1026000000, types.Gaer)}, // exceed payload max size
		{1, 20480000, types.NewAmount(5, types.Gaer), types.NewAmount(1026000000, types.Gaer)},

		// v2 - 1 gaer
		{2, 0, types.NewAmount(1, types.Gaer), types.NewAmount(100000, types.Gaer)},
		{2, 200, types.NewAmount(1, types.Gaer), types.NewAmount(100000, types.Gaer)},      // max freeByteSize
		{2, 201, types.NewAmount(1, types.Gaer), types.NewAmount(100005, types.Gaer)},      // 100000+5
		{2, 2047800, types.NewAmount(1, types.Gaer), types.NewAmount(1124000, types.Gaer)}, // 100000+5*204800 ( 2047800 + freeByteSize )
		{2, 2048000, types.NewAmount(1, types.Gaer), types.NewAmount(1124000, types.Gaer)}, // exceed payload max size
		{2, 20480000, types.NewAmount(1, types.Gaer), types.NewAmount(1124000, types.Gaer)},

		// v2 - 5 gaer
		{2, 0, types.NewAmount(5, types.Gaer), types.NewAmount(500000, types.Gaer)},
		{2, 200, types.NewAmount(5, types.Gaer), types.NewAmount(500000, types.Gaer)}, // max freeByteSize
		{2, 201, types.NewAmount(5, types.Gaer), types.NewAmount(500025, types.Gaer)},
		{2, 700, types.NewAmount(5, types.Gaer), types.NewAmount(512500, types.Gaer)},
		{2, 2047800, types.NewAmount(5, types.Gaer), types.NewAmount(5620000, types.Gaer)},
		{2, 2048000, types.NewAmount(5, types.Gaer), types.NewAmount(5620000, types.Gaer)}, // exceed payload max size

		// v3 is same as v2
		{3, 100, types.NewAmount(5, types.Gaer), types.NewAmount(500000, types.Gaer)},
	} {
		resultFee := TxFee(test.payloadSize, test.gasPrice, test.forkVersion)
		assert.EqualValues(t, test.expectFee, resultFee, "TxFee(forkVersion:%d, payloadSize:%d, gasPrice:%s)", test.forkVersion, test.payloadSize, test.gasPrice)
	}
}

func TestGasLimit(t *testing.T) {
	initContractTest(t)
	defer deinitContractTest(t)

	for _, test := range []struct {
		version        int32
		feeDelegation  bool
		txGasLimit     uint64
		payloadSize    int
		gasPrice       *big.Int
		usedFee        *big.Int
		sender         *big.Int
		receiver       *big.Int
		expectErr      error
		expectGasLimit uint64
	}{
		// no gas limit
		{version: 1, expectErr: nil, expectGasLimit: 0},

		// fee delegation
		{version: 2, feeDelegation: true, gasPrice: types.NewAmount(1, types.Gaer), receiver: types.NewAmount(5, types.Gaer), usedFee: types.NewAmount(10, types.Gaer), expectErr: nil, expectGasLimit: math.MaxUint64},                 // max
		{version: 2, feeDelegation: true, gasPrice: types.NewAmount(1, types.Gaer), receiver: types.NewAmount(5, types.Gaer), usedFee: types.NewAmount(5, types.Gaer), expectErr: newVmError(types.ErrNotEnoughGas), expectGasLimit: 0}, // not enough error
		{version: 2, feeDelegation: true, gasPrice: types.NewAmount(1, types.Gaer), receiver: types.NewAmount(10, types.Gaer), usedFee: types.NewAmount(5, types.Gaer), expectErr: nil, expectGasLimit: 5},

		// no gas limit specified in tx, the limit is the sender's balance
		{version: 2, gasPrice: types.NewAmount(1, types.Gaer), sender: types.NewAmount(5, types.Gaer), usedFee: types.NewAmount(10, types.Gaer), expectErr: nil, expectGasLimit: math.MaxUint64},                 // max
		{version: 2, gasPrice: types.NewAmount(1, types.Gaer), sender: types.NewAmount(5, types.Gaer), usedFee: types.NewAmount(5, types.Gaer), expectErr: newVmError(types.ErrNotEnoughGas), expectGasLimit: 0}, // not enough error
		{version: 2, gasPrice: types.NewAmount(1, types.Gaer), sender: types.NewAmount(10, types.Gaer), usedFee: types.NewAmount(5, types.Gaer), expectErr: nil, expectGasLimit: 5},

		// if gas limit specified in tx, check if the sender has enough balance for gas
		{version: 2, txGasLimit: 100000, payloadSize: 100, expectErr: newVmError(types.ErrNotEnoughGas), expectGasLimit: 100000},
		{version: 2, txGasLimit: 150000, payloadSize: 100, expectErr: nil, expectGasLimit: 50000},
		{version: 2, txGasLimit: 200000, payloadSize: 100, expectErr: nil, expectGasLimit: 100000},
	} {
		gasLimit, resultErr := GasLimit(test.version, test.feeDelegation, test.txGasLimit, test.payloadSize, test.gasPrice, test.usedFee, test.sender, test.receiver)
		assert.EqualValues(t, test.expectErr, resultErr, "GasLimit(forkVersion:%d, isFeeDelegation:%t, txGasLimit:%d, payloadSize:%d, gasPrice:%s, usedFee:%s, senderBalance:%s, receiverBalance:%s)", test.version, test.feeDelegation, test.txGasLimit, test.payloadSize, test.gasPrice, test.usedFee, test.sender, test.receiver)
		assert.EqualValues(t, test.expectGasLimit, gasLimit, "GasLimit(forkVersion:%d, isFeeDelegation:%t, txGasLimit:%d, payloadSize:%d, gasPrice:%s, usedFee:%s, senderBalance:%s, receiverBalance:%s)", test.version, test.feeDelegation, test.txGasLimit, test.payloadSize, test.gasPrice, test.usedFee, test.sender, test.receiver)
	}
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

//----------------------------------------------------------------------------------------//
// tests for chain / block factory functions

func TestCreateContractID(t *testing.T) {
	initContractTest(t)
	defer deinitContractTest(t)

	for _, test := range []struct {
		account          []byte
		nonce            uint64
		expectContractID []byte
	}{
		// purpose to detect logic change
		{[]byte{0x01}, 0, []byte{0xc, 0x44, 0xc8, 0x8, 0xfd, 0x16, 0x6d, 0xbb, 0x89, 0x61, 0xfb, 0x79, 0x55, 0x87, 0xe5, 0xee, 0x0, 0x82, 0xf8, 0xa2, 0xdc, 0x78, 0x1f, 0xf0, 0x6a, 0x3f, 0x2, 0x22, 0x3d, 0xcc, 0x6, 0xa7, 0xda}},
		{[]byte{0x01}, 1, []byte{0xc, 0xf1, 0x6b, 0xa6, 0xfa, 0x61, 0xda, 0x33, 0x98, 0x81, 0x5b, 0xe2, 0xa6, 0xc0, 0xf7, 0xcb, 0x13, 0x51, 0x98, 0x2d, 0xbc, 0xc6, 0xc6, 0x4b, 0xbe, 0xb9, 0xb6, 0x5f, 0x67, 0x2a, 0x8b, 0x10, 0x2a}},
		{[]byte{0xFF}, 0, []byte{0xc, 0x65, 0x1c, 0xb3, 0x16, 0x99, 0xd4, 0xd, 0xd0, 0xd0, 0x94, 0x44, 0xc7, 0xd7, 0x41, 0x87, 0xa0, 0xee, 0xcb, 0x4c, 0xbc, 0x2b, 0x1b, 0x4, 0x61, 0xbc, 0x4a, 0x3f, 0x1a, 0x5f, 0x97, 0x2e, 0xdb}},
		{[]byte{0xFF}, 1, []byte{0xc, 0x71, 0x36, 0x9f, 0x7c, 0x97, 0x5c, 0xf, 0x86, 0x19, 0x57, 0xbc, 0x6, 0x4, 0x28, 0x1e, 0x86, 0x37, 0x6a, 0x12, 0xd7, 0x1e, 0xe7, 0xf6, 0x2f, 0x98, 0xab, 0x14, 0xbe, 0x4d, 0xf5, 0xd4, 0x56}},
	} {
		resultContractID := CreateContractID(test.account, test.nonce)
		assert.Equal(t, test.expectContractID, resultContractID, "CreateContractID(account:%x, nonce:%d)", test.account, test.nonce)
	}
}

func TestGasUsed(t *testing.T) {
	initContractTest(t)
	defer deinitContractTest(t)

	for _, test := range []struct {
		version       int32
		txType        types.TxType
		txFee         *big.Int
		gasPrice      *big.Int
		expectGasUsed uint64
	}{
		// no gas used
		{1, types.TxType_NORMAL, types.NewAmount(100, types.Gaer), types.NewAmount(5, types.Gaer), 0},     // v1
		{2, types.TxType_GOVERNANCE, types.NewAmount(100, types.Gaer), types.NewAmount(5, types.Gaer), 0}, // governance

		// gas used
		{2, types.TxType_NORMAL, types.NewAmount(10, types.Gaer), types.NewAmount(1, types.Gaer), 10},
		{2, types.TxType_NORMAL, types.NewAmount(10, types.Gaer), types.NewAmount(5, types.Gaer), 2},
		{2, types.TxType_NORMAL, types.NewAmount(100, types.Gaer), types.NewAmount(1, types.Gaer), 100},
		{2, types.TxType_NORMAL, types.NewAmount(100, types.Gaer), types.NewAmount(5, types.Gaer), 20},
	} {
		resultGasUsed := GasUsed(test.txFee, test.gasPrice, test.txType, test.version)
		assert.Equal(t, test.expectGasUsed, resultGasUsed, "GasUsed(txFee:%s, gasPrice:%s, txType:%d, version:%d)", test.txFee, test.gasPrice, test.txType, test.version)
	}
}
