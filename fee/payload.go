package fee

import (
	"math/big"
)

func PayloadFee(payloadSize int) *big.Int {
	if IsZeroFee() {
		return NewZeroFee()
	}

	// set data fee
	dataFee := paymentDataSize(int64(payloadSize))
	if dataFee > payloadMaxSize {
		dataFee = payloadMaxSize
	}
	// return base fee + data fee
	return new(big.Int).Add(baseTxAergo, CalcFee(aerPerByte, uint64(dataFee)))
}

func MaxPayloadFee(payloadSize int) *big.Int {
	if IsZeroFee() {
		return NewZeroFee()
	}
	if payloadSize == 0 {
		return new(big.Int).Set(baseTxAergo)
	}
	return new(big.Int).Add(PayloadFee(payloadSize), stateDbMaxFee)
}

func paymentDataSize(dataSize int64) int64 {
	pSize := dataSize - freeByteSize
	if pSize < 0 {
		pSize = 0
	}
	return pSize
}

func StateDataFee(dataSize int64) *big.Int {
	return CalcFee(aerPerByte, uint64(paymentDataSize(dataSize)))
}
