package types

import "testing"

func TestIsSpecialAccount(t *testing.T) {
	tests := []struct {
		name string
		args []byte
		want bool
	}{
		{"TAergoName", []byte("aergo.name"), true},
		{"TAergoEnterprise", []byte("aergo.enterprise"), true},
		{"TAergoVault", []byte("aergo.vault"), true},
		{"TNormal", []byte("kslee0000007"), false},
		{"TInvalid", []byte("ks.lee"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSpecialAccount(tt.args); got != tt.want {
				t.Errorf("IsSpecialAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
