package fee

import (
	"fmt"
	"math/big"
)

const (
	baseTxFee            = "2000000000000000" // 0.002 AERGO
	payloadMaxSize       = 200 * 1024
	StateDbMaxUpdateSize = payloadMaxSize
	freeByteSize         = 200
)

var (
	baseTxAergo   *big.Int
	zeroFee       bool
	stateDbMaxFee *big.Int
	aerPerByte    *big.Int
)

func init() {
	baseTxAergo, _ = new(big.Int).SetString(baseTxFee, 10)
	zeroFee = false
	aerPerByte = big.NewInt(5000000000000) // 5,000 GAER, feePerBytes * PayloadMaxBytes = 1 AERGO
	stateDbMaxFee = new(big.Int).Mul(aerPerByte, big.NewInt(StateDbMaxUpdateSize-freeByteSize))
}

//---------------------------------------------------------------//
// zerofee

func EnableZeroFee() {
	zeroFee = true
}

func DisableZeroFee() {
	zeroFee = false
}

func IsZeroFee() bool {
	return zeroFee
}

func NewZeroFee() *big.Int {
	return big.NewInt(0)
}

//---------------------------------------------------------------//
// calc fee

// fee = used gas * gas price
func CalcFee(gasPrice *big.Int, gas uint64) *big.Int {
	return new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gas))
}

// compute the base fee for a transaction
func TxBaseFee(version int32, gasPrice *big.Int, payloadSize int) *big.Int {
	if version < 2 {
		return PayloadFee(payloadSize)
	}

	// after v2
	txGas := TxGas(payloadSize)
	return CalcFee(gasPrice, txGas)
}

// compute the execute fee for a transaction
func TxExecuteFee(version int32, gasPrice *big.Int, usedGas uint64, dbUpdateTotalSize int64) *big.Int {
	if IsZeroFee() {
		return NewZeroFee()
	}

	if version < 2 {
		return StateDataFee(dbUpdateTotalSize)
	}

	// after v2
	return CalcFee(gasPrice, usedGas)
}

// estimate the max fee for a transaction
func TxMaxFee(version int32, lenPayload int, gasLimit uint64, balance, gasPrice *big.Int) (*big.Int, error) {
	if IsZeroFee() {
		return NewZeroFee(), nil
	}

	if version < 2 {
		return MaxPayloadFee(lenPayload), nil
	}

	// after v2
	minGasLimit := TxGas(lenPayload)
	if gasLimit == 0 {
		gasLimit = MaxGasLimit(balance, gasPrice)
	}
	if minGasLimit > gasLimit {
		return nil, fmt.Errorf("the minimum required amount of gas: %d", minGasLimit)
	}
	return CalcFee(gasPrice, gasLimit), nil
}
