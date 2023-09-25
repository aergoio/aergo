package types

import "math/big"

func (s *Staking) GetAmountBigInt() *big.Int {
	return new(big.Int).SetBytes(s.GetAmount())
}

// Add adds amount to s.Amount.
func (s *Staking) Add(amount *big.Int) {
	s.Amount = new(big.Int).Add(s.GetAmountBigInt(), amount).Bytes()
}

// Sub substracts amount from s.Amount and returns the actual adjustment.
func (s *Staking) Sub(amount *big.Int) *big.Int {
	var (
		staked           = s.GetAmountBigInt()
		actualAdjustment = amount
	)

	// Cannot be a negative value.
	if staked.Cmp(amount) < 0 {
		actualAdjustment = staked
	}

	s.Amount = new(big.Int).Sub(s.GetAmountBigInt(), actualAdjustment).Bytes()

	return actualAdjustment
}

func (s *Staking) SetWhen(blockNo BlockNo) {
	s.When = uint64(blockNo)
}
