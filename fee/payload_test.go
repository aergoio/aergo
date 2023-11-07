package fee

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPayloadTxFee(t *testing.T) {
	DisableZeroFee()
	defer EnableZeroFee()

	for _, tt := range []struct {
		name        string
		payloadSize int
		want        *big.Int
	}{
		{"zero", 0, baseTxAergo},
		{"under200", 198, baseTxAergo},
		{"exact200", 198, baseTxAergo},
		{"over200", 265, new(big.Int).Add(baseTxAergo, new(big.Int).Mul(new(big.Int).SetUint64(65), aerPerByte))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := PayloadFee(tt.payloadSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PayloadTxFee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxPayloadTxFee(t *testing.T) {
	DisableZeroFee()
	defer EnableZeroFee()

	for _, test := range []struct {
		payloadSize int
		expectFee   *big.Int
	}{
		{0, PayloadFee(0)},
		{1, new(big.Int).Add(PayloadFee(1), stateDbMaxFee)},
		{200, new(big.Int).Add(PayloadFee(200), stateDbMaxFee)},
		{201, new(big.Int).Add(PayloadFee(201), stateDbMaxFee)},
		{1000, new(big.Int).Add(PayloadFee(1000), stateDbMaxFee)},
	} {
		resultTxFee := MaxPayloadFee(test.payloadSize)
		assert.EqualValues(t, test.expectFee, resultTxFee, "MaxPayloadTxFee(payloadSize:%d)", test.payloadSize)
	}
}
