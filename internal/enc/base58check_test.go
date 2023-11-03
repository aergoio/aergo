package enc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestB58CheckEncode(t *testing.T) {
	for _, test := range []struct {
		name    string
		version string
		data    string
		expect  string
	}{
		{"T1", HexEncode([]byte{0}), HexEncode([]byte("Hello")), "1vSxRbq6DSYXc"},
		{"T2", HexEncode([]byte{1}), HexEncode([]byte("Hello")), "5BShidwAu2ieX"},
		{"T3", HexEncode([]byte{5}), HexEncode([]byte("abcdefghijklmnopqrstuvwxyz1234567890")), "2BSSzM1LQHgVeyiCPn5bEfgWY3HmiC3cbjGYFhTs1bVv5GTT7nJ8ajSE"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := B58CheckEncode(test.version, test.data)
			require.NoErrorf(t, err, "B58CheckEncode() error = %v", err)
			require.Equalf(t, test.expect, got, "B58CheckEncode() = %v, want %v", got, test.expect)

			recover, err := B58CheckDecode(got)
			require.NoErrorf(t, err, "B58CheckDecode() error = %v", err)
			require.Equalf(t, test.version+test.data, recover, "B58CheckDecode() = %v, want %v", recover, test.version+test.data)
		})
	}
}
