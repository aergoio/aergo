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
