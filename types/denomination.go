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

func NewAer(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(Aer), big.NewInt(n))
}

func NewGaer(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(Gaer), big.NewInt(n))
}

func NewAergo(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(Aergo), big.NewInt(n))
}
