package types

import "math/big"

func (s *Staking) GetAmountBigInt() *big.Int {
	return new(big.Int).SetBytes(s.GetAmount())
}
