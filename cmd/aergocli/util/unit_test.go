package util

import (
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestParseTrimUnit(t *testing.T) {
	amount, err := ParseUnit("1 aergo")
	assert.NoError(t, err, "parsing aergo")
	assert.Equalf(t, types.NewAmount(1, types.Aergo).String(), amount.String(), "parse is failed")

	amount, err = ParseUnit("101 Aergo")
	assert.NoError(t, err, "parsing Aergo")
	assert.Equalf(t, types.NewAmount(101, types.Aergo).String(), amount.String(), "parse is failed")

	amount, err = ParseUnit("101 AERGO")
	assert.NoError(t, err, "parsing AERGO")
	assert.Equalf(t, types.NewAmount(101, types.Aergo).String(), amount.String(), "parse is failed")

	amount, err = ParseUnit("123 aer")
	assert.NoError(t, err, "parsing aer")
	assert.Equalf(t, types.NewAmount(123, types.Aer).String(), amount.String(), "parse is failed")

	amount, err = ParseUnit("4567 aer")
	assert.NoError(t, err, "parsing aer")
	assert.Equalf(t, types.NewAmount(4567, types.Aer).String(), amount.String(), "parse is failed")

	amount, err = ParseUnit("4567 Aer")
	assert.NoError(t, err, "parsing aer")
	assert.Equalf(t, types.NewAmount(4567, types.Aer).String(), amount.String(), "parse is failed")

	amount, err = ParseUnit("101 AER")
	assert.NoError(t, err, "parsing aer")
	assert.Equalf(t, types.NewAmount(101, types.Aer).String(), amount.String(), "parse is failed")

	amount, err = ParseUnit("1 gaer")
	assert.NoError(t, err, "parsing gaer")
	assert.Equalf(t, types.NewAmount(1, types.Gaer).String(), amount.String(), "parse is failed")

	amount, err = ParseUnit("1010")
	assert.NoError(t, err, "parsing implicit unit")
	assert.Equalf(t, types.NewAmount(1010, types.Aer).String(), amount.String(), "parse is failed")
}

func TestParseDecimalUnit(t *testing.T) {
	amount, err := ParseUnit("1.01 aergo")
	assert.NoError(t, err, "parsing point aergo")
	assert.Equal(t, types.NewAmount(1010000000000000000, types.Aer), amount, "converting result")

	amount, err = ParseUnit("1.01 gaer")
	assert.NoError(t, err, "parsing point gaer")
	assert.Equalf(t, types.NewAmount(1010000000, types.Aer), amount, "converting result")

	amount, err = ParseUnit("0.123456789012345678 aergo")
	assert.NoError(t, err, "parsing point")
	assert.Equalf(t, types.NewAmount(123456789012345678, types.Aer), amount, "converting result")

	amount, err = ParseUnit("0.100000000000000001 aergo")
	assert.NoError(t, err, "parsing point max length of decimal")
	assert.Equalf(t, types.NewAmount(100000000000000001, types.Aer), amount, "converting result")

	amount, err = ParseUnit("499999999.100000000000000001 aergo")
	assert.NoError(t, err, "parsing point max length of decimal")
	assert.Equalf(t, new(big.Int).Add(types.NewAmount(499999999, types.Aergo), types.NewAmount(100000000000000001, types.Aer)), amount, "converting result")

	amount, err = ParseUnit("499999999100000000000000001 aer")
	assert.NoError(t, err, "parsing point max length of decimal")
	assert.Equalf(t, new(big.Int).Add(types.NewAmount(499999999, types.Aergo), types.NewAmount(100000000000000001, types.Aer)), amount, "converting result")
}

func TestFailParseUnit(t *testing.T) {
	amount, err := ParseUnit("0.0000000000000000001 aergo")
	assert.Error(t, err, "exceed max length of decimal")
	t.Log(amount)
	amount, err = ParseUnit("499999999100000000000000001.1 aer")
	assert.Error(t, err, "parsing point max length of decimal")
	amount, err = ParseUnit("1 aergoa")
	assert.Error(t, err, "parsing aergoa")
	amount, err = ParseUnit("1 aerg")
	assert.Error(t, err, "parsing aerg")
	amount, err = ParseUnit("1 aaergo")
	assert.Error(t, err, "parsing aaergo")
	amount, err = ParseUnit("1 aergo ")
	assert.Error(t, err, "check fail")
	amount, err = ParseUnit("1aergo.1aer")
	assert.Error(t, err, "check fail")
	amount, err = ParseUnit("0.1")
	assert.Error(t, err, "default unit assumed aergo")
	amount, err = ParseUnit("0.1.1")
	assert.Error(t, err, "only one dot is allowed")
}

func TestConvertUnit(t *testing.T) {
	result, err := ConvertUnit(new(big.Int).SetUint64(1000000000000000000), "aergo")
	assert.NoError(t, err, "convert 1 aergo")
	assert.Equalf(t, "1 aergo", result, "converting result")

	result, err = ConvertUnit(new(big.Int).SetUint64(1020300000000000000), "aergo")
	assert.NoError(t, err, "convert 1.0203 aergo")
	assert.Equalf(t, "1.0203 aergo", result, "converting result")

	result, err = ConvertUnit(new(big.Int).SetUint64(1000000000), "gaer")
	assert.NoError(t, err, "convert 1 gaer")
	assert.Equal(t, "1 gaer", result)

	result, err = ConvertUnit(new(big.Int).SetUint64(1), "gaer")
	assert.NoError(t, err, "convert 0.000000001 gaer")
	assert.Equal(t, "0.000000001 gaer", result)

	result, err = ConvertUnit(new(big.Int).SetUint64(10), "gaer")
	assert.NoError(t, err, "convert 0.00000001 gaer")
	assert.Equal(t, "0.00000001 gaer", result)

	result, err = ConvertUnit(new(big.Int).SetUint64(0), "gaer")
	assert.NoError(t, err, "convert 0 gaer")
	assert.Equal(t, "0 gaer", result)

	result, err = ConvertUnit(new(big.Int).SetUint64(1), "aer")
	assert.NoError(t, err, "convert 1 aer")
	assert.Equal(t, "1 aer", result)

	result, err = ConvertUnit(new(big.Int).SetUint64(1000000000000000000), "gaer")
	assert.NoError(t, err, "convert 1000000000 gaer")
	assert.Equal(t, "1000000000 gaer", result)
}

func TestParseUnit(t *testing.T) {
	n100 := types.NewAmount(100, types.Aer)
	n1000 := types.NewAmount(1000, types.Aer)
	OneGaer := types.NewAmount(1, types.Gaer)
	OneAergo := types.NewAmount(1, types.Aergo)

	tests := []struct {
		name    string
		args    string
		want    *big.Int
		wantErr bool
	}{
		{"TNum", "10000", big.NewInt(10000), false},
		{"TNumDot", "1000.5", new(big.Int), true},
		{"TNum20pow", "100000000000000000000", new(big.Int).Mul(OneAergo, n100), false},
		{"TAer", "10000aer", big.NewInt(10000), false},
		{"TAerDot", "1000.5aer", new(big.Int), true},
		{"TAer20pow", "100000000000000000000", new(big.Int).Mul(OneAergo, n100), false},
		{"TGaer", "1000gaer", new(big.Int).Mul(OneGaer, n1000), false},
		{"TGaerDot", "1000.21245gaer", new(big.Int).Add(new(big.Int).Mul(OneGaer, n1000), big.NewInt(212450000)), false},
		{"TGaerDot2", "0.21245gaer", big.NewInt(212450000), false},
		{"TGaerDot3", ".21245gaer", big.NewInt(212450000), false},
		{"TGaer11pow", "100000000000gaer", new(big.Int).Mul(OneAergo, n100), false},
		{"TAergo", "100aergo", new(big.Int).Mul(OneAergo, n100), false},
		{"TAergoDot", "1000.00000000021245aergo", new(big.Int).Add(new(big.Int).Mul(OneAergo, n1000), big.NewInt(212450000)), false},
		{"TWrongNum", "100d0.321245", new(big.Int), true},
		{"TWrongNum2", "100d0.321245aergo", new(big.Int), true},
		{"TWrongUnit", "100d0.321245argo", new(big.Int), true},
		{"TWrongUnit", "100d0.321245ear", new(big.Int), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUnit(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && (tt.want.Cmp(got) != 0) {
				t.Errorf("ParseUnit() got = %v, want %v", got, tt.want)
			}
		})
	}
}
