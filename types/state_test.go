package types

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/stretchr/testify/require"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

func TestStateClone(t *testing.T) {
	for _, test := range []struct {
		state *State
	}{
		{&State{
			Nonce:            1,
			Balance:          NewAmount(1, Aergo).Bytes(),
			CodeHash:         []byte{1},
			StorageRoot:      []byte{1},
			SqlRecoveryPoint: 1,
		}},
		{&State{
			state:         protoimpl.MessageState{},
			sizeCache:     1,
			unknownFields: []byte{1, 2, 3},

			Nonce:            1,
			Balance:          NewAmount(1, Aergo).Bytes(),
			CodeHash:         []byte{1},
			StorageRoot:      []byte{1},
			SqlRecoveryPoint: 1,
		}},
	} {
		stateRaw, err := proto.Encode(test.state)
		require.NoError(t, err)

		clone := test.state.Clone()
		cloneRaw, err := proto.Encode(clone)
		require.NoError(t, err)

		require.Equal(t, stateRaw, cloneRaw)
	}

}

var sampleAddress = "AmQHDFX45YMxrhN35S6csRweFK3WEAUCRWP78bHJguKCu28NjHzb"
var emptyAddress = "1111111111111111111111111111111111111111111111111111"

var (
	sample1KeyB58 = "47TFXngxf7PcjMZTXqNAUoPYiPLra52pEwVJ9r94uRhwQE33s9kicaUpkiP9vYhhmbwKfH3M7"
	sample1Key    = "aa529b171489a18125a054e6661aa57593be3c1e6416974746f3172c25bcba6959dd93ba013b5e3726ab60e2db5c2a767968907176"
	sample1Addr   = "AmMLMVLzjUQEm16HDWd5QfxdRj45kLgNSYKTuLG3su3H6ngNf9QQ"
)

func TestToAccountID(t *testing.T) {
	zeroBytes := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	tests := []struct {
		name    string
		account []byte
		want    AccountID
	}{
		{"sample", ToAddress(sampleAddress), ToAccountID(ToAddress(sampleAddress))},
		{"zeros", Address(zeroBytes),
			ToAccountID(zeroBytes)},
		{"empty", ToAddress(emptyAddress), ToAccountID([]byte{})},
		{"nil", ToAddress(emptyAddress), ToAccountID(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address := Address(tt.account)
			fmt.Printf("Address %s , account %s", EncodeAddress(address), tt.want.String())
			assert.Equalf(t, tt.want, ToAccountID(tt.account), "ToAccountID(%v)", tt.account)
		})
	}
}
