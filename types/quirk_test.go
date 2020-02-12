package types

import "testing"

func TestIsQuirkTx(t *testing.T) {
	tx := NewTx()
	tx.Body.Nonce = 111
	tx.Hash = tx.CalculateTxHash()

	type args struct {
	}
	tests := []struct {
		name string
		args []byte
		want bool
	}{
		{"TQuirk", DecodeB58(B23994084_001), true},
		{"TOther", tx.Hash, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsQuirkTx(tt.args); got != tt.want {
				t.Errorf("IsQuirkTx() = %v, want %v", got, tt.want)
			}
		})
	}
}
