package jsonrpc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseEncodingType(t *testing.T) {
	tests := []struct {
		name string
		args string
		want EncodingType
	}{
		{"raw", "raW", Raw},
		{"raw2", "raw", Raw},
		{"obj", "OBJ", Obj},
		{"base58", "base58", Base58},
		{"nope", "base58", Base58},
		{"", "", Base58},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ParseEncodingType(tt.args), "ParseEncodingType(%v)", tt.args)
		})
	}
}
