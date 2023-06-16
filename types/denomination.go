package types

import "math/big"

// Denomination of Aergo
const (
	Aergo = 1e18
	Gaer  = 1e9
	Aer   = 1
)

func NewZero() *big.Int {
	return new(big.Int)
}

func NewAer(n uint64) *big.Int {
	return new(big.Int).Mul(big.NewInt(Aer), new(big.Int).SetUint64(n))
}

func NewGaer(n uint64) *big.Int {
	return new(big.Int).Mul(big.NewInt(Gaer), new(big.Int).SetUint64(n))
}

func NewAergo(n uint64) *big.Int {
	return new(big.Int).Mul(big.NewInt(Aergo), new(big.Int).SetUint64(n))
}
