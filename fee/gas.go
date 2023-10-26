package fee

import (
	"math"
	"math/big"
)

const (
	txGasSize      = uint64(100000)
	payloadGasSize = uint64(5)
)

func IsUseTxGas(version int32) bool {
	return !IsZeroFee() && version >= 2
}

func IsVmGasSystem(version int32, isQuery bool) bool {
	return IsUseTxGas(version) && !isQuery
}

func IsReceiptGasUsed(version int32, isGovernance bool) bool {
	return IsUseTxGas(version) && !isGovernance
}

//---------------------------------------------------------------//
// calc gas

func ReceiptGasUsed(version int32, isGovernance bool, txFee, gasPrice *big.Int) uint64 {
	if IsReceiptGasUsed(version, isGovernance) != true {
		return 0
	}
	return new(big.Int).Div(txFee, gasPrice).Uint64()
}

func TxGas(payloadSize int) uint64 {
	if IsZeroFee() {
		return 0
	}
	size := paymentDataSize(int64(payloadSize))
	if size > payloadMaxSize {
		size = payloadMaxSize
	}
	txGas := txGasSize
	payloadGas := uint64(size) * payloadGasSize
	return txGas + payloadGas
}

func MaxGasLimit(balance, gasPrice *big.Int) uint64 {
	gasLimit := uint64(math.MaxUint64)
	n := new(big.Int).Div(balance, gasPrice)
	if n.IsUint64() {
		gasLimit = n.Uint64()
	}
	return gasLimit
}
