package fee

import (
	"math/big"
)

func PayloadTxFee(payloadSize int) *big.Int {
	if IsZeroFee() {
		return NewZeroFee()
	}

	// set data fee
	dataFee := PaymentDataFee(int64(payloadSize))

	// return base fee + data fee
	return new(big.Int).Add(baseTxAergo, dataFee)
}

func MaxPayloadTxFee(payloadSize int) *big.Int {
	if IsZeroFee() {
		return NewZeroFee()
	}
	if payloadSize == 0 {
		return PayloadTxFee(payloadSize)
	}
	return new(big.Int).Add(PayloadTxFee(payloadSize), stateDbMaxFee)
}

func paymentDataSize(dataSize int64) int64 {
	pSize := dataSize - freeByteSize
	if pSize < 0 {
		pSize = 0
	}
	return pSize
}

func PaymentDataFee(dataSize int64) *big.Int {
	return CalcFee(aerPerByte, uint64(paymentDataSize(dataSize)))
}
