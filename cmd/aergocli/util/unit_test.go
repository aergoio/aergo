package util

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasreUnit(t *testing.T) {
	amount, err := ParseUnit("1 aergo")
	assert.NoError(t, err, "parsing aergo")
	t.Log(amount)
	amount, err = ParseUnit("101 Aergo")
	assert.NoError(t, err, "parsing Aergo")
	t.Log(amount)
	amount, err = ParseUnit("101 AERGO")
	assert.NoError(t, err, "parsing AERGO")
	t.Log(amount)
	amount, err = ParseUnit("123 aer")
	assert.NoError(t, err, "parsing aer")
	assert.Equal(t, new(big.Int).SetUint64(123), amount, "123 aer")
	amount, err = ParseUnit("4567 Aer")
	assert.NoError(t, err, "parsing Aer")
	assert.Equal(t, new(big.Int).SetUint64(4567), amount, "4567 Aer")
	amount, err = ParseUnit("101 AER")
	assert.NoError(t, err, "parsing AER")
	assert.Equal(t, new(big.Int).SetUint64(101), amount, "101 AER")
	amount, err = ParseUnit("1 gaer")
	assert.NoError(t, err, "parsing gaer")
	assert.Equal(t, new(big.Int).SetUint64(1000000000), amount, "1 gaer")
	amount, err = ParseUnit("1010")
	assert.Equal(t, new(big.Int).SetUint64(1010), amount, "1010")
	assert.NoError(t, err, "parsing implicit unit")
	t.Log(amount)
}

func TestPasreDecimalUnit(t *testing.T) {
	amount, err := ParseUnit("1.01 aergo")
	assert.NoError(t, err, "parsing point aergo")
	assert.Equal(t, new(big.Int).SetUint64(1010000000000000000), amount, "converting result")
	amount, err = ParseUnit("1.01 gaer")
	assert.NoError(t, err, "parsing point gaer")
	assert.Equal(t, new(big.Int).SetUint64(1010000000), amount, "converting result")
	amount, err = ParseUnit("0.123456789012345678 aergo")
	assert.NoError(t, err, "parsing point")
	t.Log(amount)
	amount, err = ParseUnit("0.100000000000000001 aergo")
	assert.NoError(t, err, "parsing point max length of decimal")
	t.Log(amount)
	amount, err = ParseUnit("499999999.100000000000000001 aergo")
	assert.NoError(t, err, "parsing point max length of decimal")
	t.Log(amount)
	amount, err = ParseUnit("499999999100000000000000001 aer")
	assert.NoError(t, err, "parsing point max length of decimal")
	t.Log(amount)
}
func TestFailPasreUnit(t *testing.T) {
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
	t.Log(result)
	result, err = ConvertUnit(new(big.Int).SetUint64(1020300000000000000), "aergo")
	assert.NoError(t, err, "convert 1.0203 aergo")
	t.Log(result)
	result, err = ConvertUnit(new(big.Int).SetUint64(1000000000), "gaer")
	assert.NoError(t, err, "convert 1 gaer")
	t.Log(result)
	result, err = ConvertUnit(new(big.Int).SetUint64(1), "gaer")
	assert.NoError(t, err, "convert 0.000000001 gaer")
	assert.Equal(t, "0.000000001 gaer", result)
	result, err = ConvertUnit(new(big.Int).SetUint64(10), "gaer")
	assert.NoError(t, err, "convert 0.00000001 gaer")
	assert.Equal(t, "0.00000001 gaer", result)
	t.Log(result)
	result, err = ConvertUnit(new(big.Int).SetUint64(0), "gaer")
	assert.NoError(t, err, "convert 0 gaer")
	assert.Equal(t, "0 gaer", result)
	t.Log(result)
	result, err = ConvertUnit(new(big.Int).SetUint64(1), "aer")
	assert.NoError(t, err, "convert 1 aer")
	assert.Equal(t, "1 aer", result)
	t.Log(result)
	result, err = ConvertUnit(new(big.Int).SetUint64(1000000000000000000), "gaer")
	assert.NoError(t, err, "convert 1000000000 gaer")
	t.Log(result)
}

func TestParseUnit(t *testing.T) {
	n100 := big.NewInt(100)
	n1000 := big.NewInt(1000)
	OneGaer := big.NewInt(1000000000)
	OneAergo := big.NewInt(1000000000).Mul(OneGaer, OneGaer)

	tests := []struct {
		name    string
		args    string
		want    *big.Int
		wantErr bool
	}{
		{"TNum", "10000", big.NewInt(10000), false},
		{"TNumDot", "1000.5", big.NewInt(0), true},
		{"TNum20pow", "100000000000000000000", big.NewInt(1).Mul(OneAergo, n100), false},
		{"TAer", "10000aer", big.NewInt(10000), false},
		{"TAerDot", "1000.5aer", big.NewInt(0), true},
		{"TAer20pow", "100000000000000000000", big.NewInt(1).Mul(OneAergo, n100), false},
		{"TGaer", "1000gaer", big.NewInt(1).Mul(OneGaer, n1000), false},
		{"TGaerDot", "1000.21245gaer", big.NewInt(1).Add(big.NewInt(0).Mul(OneGaer, n1000), big.NewInt(212450000)), false},
		{"TGaerDot2", "0.21245gaer", big.NewInt(212450000), false},
		{"TGaerDot3", ".21245gaer", big.NewInt(212450000), false},
		{"TGaer11pow", "100000000000gaer", big.NewInt(1).Mul(OneAergo, n100), false},
		{"TAergo", "100aergo", big.NewInt(1).Mul(OneAergo, n100), false},
		{"TAergoDot", "1000.00000000021245aergo", big.NewInt(1).Add(big.NewInt(0).Mul(OneAergo, n1000), big.NewInt(212450000)), false},
		{"TWrongNum", "100d0.321245", big.NewInt(1), true},
		{"TWrongNum2", "100d0.321245aergo", big.NewInt(1), true},
		{"TWrongUnit", "100d0.321245argo", big.NewInt(1), true},
		{"TWrongUnit", "100d0.321245ear", big.NewInt(1), true},
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
