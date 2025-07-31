package utils

import "testing"

func TestToUint64(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{"nil", args{nil}, 0, true},
		{"short", args{[]byte{1, 0, 0, 0}}, 0, true},
		{"257", args{[]byte{1, 1, 0, 0, 0, 0, 0, 0}}, 257, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToUint64(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToUint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToUint64() got = %v, want %v", got, tt.want)
			}
		})
	}
}
