package types

import "math/big"

type TokenUnit uint64

const (
	Aergo TokenUnit = 1e18
	Gaer  TokenUnit = 1e9
	Aer   TokenUnit = 1
)

func NewAmount(amount uint64, unit TokenUnit) *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(uint64(unit)))
}

func NewZeroAmount() *big.Int {
	return new(big.Int)
}
