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
	assert.Equalf(t, int64(aerNum), NewAer(1).Int64(), "Aer is not valid. check types/denomination.go")
	// gaer
	assert.Equalf(t, int64(gaerNum), NewGaer(1).Int64(), "Gaer is not valid. check types/denomination.go")
	// aergo
	assert.Equalf(t, int64(aergoNum), NewAergo(1).Int64(), "Aergo not valid. check types/denomination.go")

	// aer
	assert.Equalf(t, aerStr, NewAer(1).String(), "Aer is not valid. check types/denomination.go")
	// gaer
	assert.Equalf(t, gaerStr, NewGaer(1).String(), "Gaer is not valid. check types/denomination.go")
	// aergo
	assert.Equalf(t, aergoStr, NewAergo(1).String(), "Aergo is not valid. check types/denomination.go")
}

func TestUnitConversion(t *testing.T) {
	// aer to gaer
	assert.Equalf(t, int64(gaerNum), NewAer(1e9).Int64(), "Aer to Gaer is not valid. check types/denomination.go")
	// gaer to aergo
	assert.Equalf(t, int64(aergoNum), NewGaer(1e9).Int64(), "Gaer to Aergo is not valid. check types/denomination.go")
	// aer to aergo
	assert.Equalf(t, int64(aergoNum), NewAer(1e18).Int64(), "Aer to Aergo is not valid. check types/denomination.go")

	// aer to gaer
	assert.Equalf(t, gaerStr, NewAer(1e9).String(), "Aer to Gaer is not valid. check types/denomination.go")
	// gaer to aergo
	assert.Equalf(t, aergoStr, NewGaer(1e9).String(), "Gaer to Aergo is not valid. check types/denomination.go")
	// aer to aergo
	assert.Equalf(t, aergoStr, NewAer(1e18).String(), "Aer to Aergo is not valid. check types/denomination.go")
}
