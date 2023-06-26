package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// validate set for testing
const (
	aergoNum = 1000000000000000000
	gaerNum  = 1000000000
	aerNum   = 1

	aergoStr = "1000000000000000000"
	gaerStr  = "1000000000"
	aerStr   = "1"
)

func TestDenomination(t *testing.T) {
	// aer
	assert.Equalf(t, int64(aerNum), NewAmount(1, Aer).Int64(), "Aer is not valid. check types/denomination.go")
	// gaer
	assert.Equalf(t, int64(gaerNum), NewAmount(1, Gaer).Int64(), "Gaer is not valid. check types/denomination.go")
	// aergo
	assert.Equalf(t, int64(aergoNum), NewAmount(1, Aergo).Int64(), "Aergo not valid. check types/denomination.go")

	// aer
	assert.Equalf(t, aerStr, NewAmount(1, Aer).String(), "Aer is not valid. check types/denomination.go")
	// gaer
	assert.Equalf(t, gaerStr, NewAmount(1, Gaer).String(), "Gaer is not valid. check types/denomination.go")
	// aergo
	assert.Equalf(t, aergoStr, NewAmount(1, Aergo).String(), "Aergo is not valid. check types/denomination.go")
}

func TestUnitConversion(t *testing.T) {
	// aer to gaer
	assert.Equalf(t, int64(gaerNum), NewAmount(1e9, Aer).Int64(), "Aer to Gaer is not valid. check types/denomination.go")
	// gaer to aergo
	assert.Equalf(t, int64(aergoNum), NewAmount(1e9, Gaer).Int64(), "Gaer to Aergo is not valid. check types/denomination.go")
	// aer to aergo
	assert.Equalf(t, int64(aergoNum), NewAmount(1e18, Aer).Int64(), "Aer to Aergo is not valid. check types/denomination.go")

	// aer to gaer
	assert.Equalf(t, gaerStr, NewAmount(1e9, Aer).String(), "Aer to Gaer is not valid. check types/denomination.go")
	// gaer to aergo
	assert.Equalf(t, aergoStr, NewAmount(1e9, Gaer).String(), "Gaer to Aergo is not valid. check types/denomination.go")
	// aer to aergo
	assert.Equalf(t, aergoStr, NewAmount(1e18, Aer).String(), "Aer to Aergo is not valid. check types/denomination.go")
}
