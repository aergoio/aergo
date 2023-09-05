/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"testing"

	"github.com/aergoio/aergo/v2/types"
)

func Test_readGenesis(t *testing.T) {
	sampleGenesis := &types.Genesis{}

	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		wantRet *types.Genesis
		wantErr bool
	}{
		{"Tsucc", args{"../../examples/genesis.json"}, sampleGenesis, false},
		{"TNotExist", args{"../../examples/genesis.notjson"}, nil, true},
		{"TNotGenesis", args{"../../examples/component/main.go"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRet, err := readGenesis(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("readGenesis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			//if !reflect.DeepEqual(gotRet, tt.wantRet) {
			//	t.Errorf("readGenesis() = %v, want %v", gotRet, tt.wantRet)
			//}
			if (tt.wantRet == nil) != (gotRet == nil) {
				t.Errorf("readGenesis() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}
