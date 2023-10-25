package contract

import (
	"errors"
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func bigIntFromString(str string) *big.Int {
	bigInt, success := new(big.Int).SetString(str, 10)
	if !success {
		panic("bigIntFromString: invalid number: " + str)
	}
	return bigInt
}

func TestTransformAmount(t *testing.T) {
	// Define the test cases
	tests := []struct {
		amountStr      string
		expectedAmount *big.Int
		expectedError  error
	}{
		// Empty Input String
		{"", big.NewInt(0), nil},
		// Valid Amount without Unit
		{"1", big.NewInt(1), nil},
		{"10", big.NewInt(10), nil},
		{"123", big.NewInt(123), nil},
		{"123000000", big.NewInt(123000000), nil},
		// Valid Amount with Unit
		{"100aergo", types.NewAmount(100, types.Aergo), nil},
		{"100 aergo", types.NewAmount(100, types.Aergo), nil},
		{"123gaer", types.NewAmount(123, types.Gaer), nil},
		{"123 gaer", types.NewAmount(123, types.Gaer), nil},
		{"123aer", types.NewAmount(123, types.Aer), nil},
		{"123 aer", types.NewAmount(123, types.Aer), nil},
		// Multipart Amount
		{"100aergo 200gaer", bigIntFromString("100000000200000000000"), nil},
		{"100 aergo 123 gaer", bigIntFromString("100000000123000000000"), nil},
		{"123aergo 456aer", bigIntFromString("123000000000000000456"), nil},
		{"123 aergo 456 aer", bigIntFromString("123000000000000000456"), nil},
		{"123aergo 456gaer 789aer", bigIntFromString("123000000456000000789"), nil},
		{"123 aergo 456 gaer 789 aer", bigIntFromString("123000000456000000789"), nil},
		// Invalid Order
		{"789aer 456gaer 123aergo", nil, errors.New("converting error for BigNum: 789aer 456gaer 123aergo")},
		{"789 aer 456 gaer 123 aergo", nil, errors.New("converting error for BigNum: 789 aer 456 gaer 123 aergo")},
		{"789aer 123aergo 456gaer", nil, errors.New("converting error for BigNum: 789aer 123aergo 456gaer")},
		{"789 aer 123 aergo 456 gaer", nil, errors.New("converting error for BigNum: 789 aer 123 aergo 456 gaer")},
		{"456gaer 789aer 123aergo", nil, errors.New("converting error for BigNum: 456gaer 789aer 123aergo")},
		{"123aergo 789aer 456gaer", nil, errors.New("converting error for BigNum: 123aergo 789aer 456gaer")},
		// Repeated Units
		{"123aergo 456aergo", nil, errors.New("converting error for Integer: 123aergo 456aergo")},
		{"123gaer 456gaer", nil, errors.New("converting error for BigNum: 123gaer 456gaer")},
		{"123aer 456aer", nil, errors.New("converting error for Integer: 123aer 456aer")},
		{"123 aergo 456 aergo", nil, errors.New("converting error for Integer: 123 aergo 456 aergo")},
		{"123 gaer 456 gaer", nil, errors.New("converting error for BigNum: 123 gaer 456 gaer")},
		{"123 aer 456 aer", nil, errors.New("converting error for Integer: 123 aer 456 aer")},
		{"123aergo 456aergo 789aer", nil, errors.New("converting error for Integer: 123aergo 456aergo 789aer")},
		{"123aergo 456aergo 789gaer", nil, errors.New("converting error for BigNum: 123aergo 456aergo 789gaer")},
		{"123aergo 456gaer 789gaer", nil, errors.New("converting error for BigNum: 123aergo 456gaer 789gaer")},
		{"123aergo 456aer 789aer", nil, errors.New("converting error for Integer: 123aergo 456aer 789aer")},
		{"123 aergo 456 aergo 789 aer", nil, errors.New("converting error for Integer: 123 aergo 456 aergo 789 aer")},
		{"123 aergo 456 aergo 789 gaer", nil, errors.New("converting error for BigNum: 123 aergo 456 aergo 789 gaer")},
		{"123 aergo 456 gaer 789 gaer", nil, errors.New("converting error for BigNum: 123 aergo 456 gaer 789 gaer")},
		{"123 aergo 456 aer 789 aer", nil, errors.New("converting error for Integer: 123 aergo 456 aer 789 aer")},
		// Invalid Amount String
		{"notanumber", nil, errors.New("converting error for Integer: notanumber")},
		{"e123", nil, errors.New("converting error for Integer: e123")},
		{"123e", nil, errors.New("converting error for Integer: 123e")},
		{"123 456", nil, errors.New("converting error for Integer: 123 456")},
		// Negative Amount
		{"-100", nil, errors.New("negative amount not allowed")},
		{"-100aergo", nil, errors.New("negative amount not allowed")},
		{"-100 aergo", nil, errors.New("negative amount not allowed")},
		{"-100  aergo", nil, errors.New("negative amount not allowed")},
		{"-100aer", nil, errors.New("negative amount not allowed")},
		{"-100 aer", nil, errors.New("negative amount not allowed")},
		{"-100  aer", nil, errors.New("negative amount not allowed")},
		// Large Number
		{"99999999999999999999999999", bigIntFromString("99999999999999999999999999"), nil},
		// Zero Value
		{"0", big.NewInt(0), nil},
		{"0aergo", big.NewInt(0), nil},
		{"0 aergo", big.NewInt(0), nil},
		{"0gaer", big.NewInt(0), nil},
		{"0 gaer", big.NewInt(0), nil},
		{"0aer", big.NewInt(0), nil},
		{"0 aer", big.NewInt(0), nil},
		// Only Unit
		{"aergo", nil, errors.New("converting error for BigNum: aergo")},
		{"gaer", nil, errors.New("converting error for BigNum: gaer")},
		{"aer", nil, errors.New("converting error for BigNum: aer")},
		// Invalid Content
		{"100 invalid 200", nil, errors.New("converting error for Integer: 100 invalid 200")},
		{"invalid 200", nil, errors.New("converting error for Integer: invalid 200")},
		{"100 invalid", nil, errors.New("converting error for Integer: 100 invalid")},
		// Non-Integer Values
		{"123.456", nil, errors.New("converting error for Integer: 123.456")},
		{"123.456 aergo", nil, errors.New("converting error for BigNum: 123.456 aergo")},
		{".1", nil, errors.New("converting error for Integer: .1")},
		{".1aergo", nil, errors.New("converting error for BigNum: .1aergo")},
		{".1 aergo", nil, errors.New("converting error for BigNum: .1 aergo")},
		{".10", nil, errors.New("converting error for Integer: .10")},
		// Exponents
		{"1e+18", nil, errors.New("converting error for Integer: 1e+18")},
		{"2e18", nil, errors.New("converting error for Integer: 2e18")},
		{"3e08", nil, errors.New("converting error for Integer: 3e08")},
		{"1e+18 aer", nil, errors.New("converting error for BigNum: 1e+18 aer")},
		{"2e+18 aer", nil, errors.New("converting error for BigNum: 2e+18 aer")},
		{"3e18 aer", nil, errors.New("converting error for BigNum: 3e18 aer")},
		{"1e+18aer", nil, errors.New("converting error for BigNum: 1e+18aer")},
		{"2e+18aer", nil, errors.New("converting error for BigNum: 2e+18aer")},
		{"3e18aer", nil, errors.New("converting error for BigNum: 3e18aer")},
		{"3e+5 aergo", nil, errors.New("converting error for BigNum: 3e+5 aergo")},
		{"3e5 aergo", nil, errors.New("converting error for BigNum: 3e5 aergo")},
		{"3e05 aergo", nil, errors.New("converting error for BigNum: 3e05 aergo")},
		{"5e+3aergo", nil, errors.New("converting error for BigNum: 5e+3aergo")},
	}

	for _, tt := range tests {
		result, err := transformAmount(tt.amountStr)

		if tt.expectedError != nil {
			if assert.Error(t, err, "Expected error: %s", tt.expectedError.Error()) {
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			}
		} else {
			if assert.NoError(t, err) && tt.expectedAmount != nil {
				assert.Equal(t, tt.expectedAmount, result)
			}
		}
	}

}
