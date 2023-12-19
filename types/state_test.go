package types

import (
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
