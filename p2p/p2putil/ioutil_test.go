/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"bytes"
	"testing"
)

func TestReadToLen(t *testing.T) {
	sample := []byte("0123456789ABCDEFGHIJabcdefghij")

	type args struct {
		bfLen int
	}
	tests := []struct {
		name string

		args   args
		repeat int

		want int
	}{
		{"TExact", args{4}, 0, 4},
		{"TBigBuf", args{8}, 0, 8},
		{"TRepeat", args{4}, 4, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := bytes.NewReader(sample)
			bf := make([]byte, 100)
			prev := make([]byte, 0, tt.args.bfLen)
			for i := 0; i <= tt.repeat; i++ {
				got, _ := ReadToLen(rd, bf[:tt.args.bfLen])
				if got != tt.want {
					t.Errorf("baseWireHandshaker.readToLen() = %v, want %v", got, tt.want)
				}
				if bytes.Equal(prev, bf) {
					t.Errorf("baseWireHandshaker.readToLen() wrong, same as prev %v", bf)
				}
				copy(prev, bf)
			}
		})
	}
}
