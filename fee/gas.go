package fee

import (
	"errors"
	"math"
	"math/big"
)

const (
	txGasSize      = uint64(100000)
	payloadGasSize = uint64(5)
)

func GasEnabled(version int32) bool {
	return !IsZeroFee() && version >= 2
}

//---------------------------------------------------------------//
// calc gas

// gas = fee / gas price
func CalcGas(fee, gasPrice *big.Int) *big.Int {
	return new(big.Int).Div(fee, gasPrice)
}

func ReceiptGasUsed(version int32, isGovernance bool, txFee, gasPrice *big.Int) uint64 {
	if GasEnabled(version) && !isGovernance {
		return CalcGas(txFee, gasPrice).Uint64()
	}
	return 0
}

func TxGas(payloadSize int) uint64 {
	if IsZeroFee() {
		return 0
	}
	size := min(paymentDataSize(int64(payloadSize)), payloadMaxSize)
	txGas := txGasSize
	payloadGas := uint64(size) * payloadGasSize
	return txGas + payloadGas
}

func GasLimit(version int32, isFeeDelegation bool, txGasLimit uint64, payloadSize int, gasPrice, usedFee, senderBalance, receiverBalance *big.Int) (gasLimit uint64, err error) {
	// 1. no gas limit
	if GasEnabled(version) != true {
		return
	}

	// 2. fee delegation
	if isFeeDelegation {
		// check if the contract has enough balance for fee
		balance := new(big.Int).Sub(receiverBalance, usedFee)
		gasLimit = MaxGasLimit(balance, gasPrice)
		if gasLimit == 0 {
			err = errors.New("not enough gas")
		}
		return
	}

	// read the gas limit from the tx
	gasLimit = txGasLimit
	// 3. no gas limit specified, the limit is the sender's balance
	if gasLimit == 0 {
		balance := new(big.Int).Sub(senderBalance, usedFee)
		gasLimit = MaxGasLimit(balance, gasPrice)
		if gasLimit == 0 {
			err = errors.New("not enough gas")
		}
		return
	}

	// 4. check if the sender has enough balance for gas
	usedGas := TxGas(payloadSize)
	if gasLimit <= usedGas {
		err = errors.New("not enough gas")
		return
	}
	// subtract the used gas from the gas limit
	gasLimit -= usedGas

	return gasLimit, nil
}

func MaxGasLimit(balance, gasPrice *big.Int) uint64 {
	gasLimit := uint64(math.MaxUint64)
	if n := CalcGas(balance, gasPrice); n.IsUint64() {
		gasLimit = n.Uint64()
	}
	return gasLimit
}
