package fee

import (
	"math/big"
)

func PayloadTxFee(payloadSize int) *big.Int {
	if IsZeroFee() {
		return NewZeroFee()
	}
	if payloadSize == 0 {
		return new(big.Int).Set(baseTxAergo)
	}
	size := paymentDataSize(int64(payloadSize))
	if size > payloadMaxSize {
		size = payloadMaxSize
	}
	return new(big.Int).Add(
		baseTxAergo,
		new(big.Int).Mul(
			aerPerByte,
			big.NewInt(size),
		),
	)
}

func MaxPayloadTxFee(payloadSize int) *big.Int {
	if IsZeroFee() {
		return NewZeroFee()
	}
	if payloadSize == 0 {
		return new(big.Int).Set(baseTxAergo)
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
	return new(big.Int).Mul(big.NewInt(paymentDataSize(dataSize)), aerPerByte)
}
