package fee

import "math/big"

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

// fee = gas price * gas
func CalcFee(gasPrice *big.Int, gas uint64) *big.Int {
	return new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gas))
}

// compute the base fee for a transaction
func TxBaseFee(version int32, gasPrice *big.Int, payloadSize int) *big.Int {
	if IsZeroFee() {
		return NewZeroFee()
	}

	if IsUseTxGas(version) {
		// get the amount of gas needed for the payload
		txGas := TxGas(payloadSize)
		// multiply the amount of gas with the gas price
		return CalcFee(gasPrice, txGas)
	}
	return PayloadTxFee(payloadSize)
}

// compute the execute fee for a transaction
func TxExecuteFee(version int32, isQuery bool, gasPrice *big.Int, usedGas uint64, dbUpdateTotalSize int64) *big.Int {
	if IsZeroFee() {
		return NewZeroFee()
	}

	if IsVmGasSystem(version, isQuery) {
		return CalcFee(gasPrice, usedGas)
	}
	return PaymentDataFee(dbUpdateTotalSize)
}
