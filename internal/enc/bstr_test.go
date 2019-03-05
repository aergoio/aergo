package enc

import (
	"bytes"
	"testing"
)

func TestToString(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want string
		wantInverse bool
	}{
		{"TSingle", args{[]byte("A")},"28", true},
		{"TEmpty", args{make([]byte,0)},"", false},
		{"TNil", args{nil},"", false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToString(tt.args.b)
			if got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}

			if tt.wantInverse {
				got2, err := ToBytes(got)
				if err != nil {
					t.Errorf("ToBytes() = <err> %s, want no err", err.Error())
				}
				if !bytes.Equal(tt.args.b, got2) {
					t.Errorf("ToString() = %v, want %v", got2, tt.args.b)
				}
			}
		})
	}
}
