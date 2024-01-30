package state

import (
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/require"
)

func TestAidWithPadding(t *testing.T) {
	for _, test := range []struct {
		padding   bool
		id        []byte
		expectAid string
	}{
		{false, []byte(types.AergoName), "55EYudcSHptibgAWjhJGEFXzti7ggpgQTCCBgyeCa2hR"},
		{true, []byte(types.AergoName), "DkAm9mQvmRDCqAPARNhwjyrvGPgrGYmErpov8S9sbkc3"},
		{false, base58.DecodeOrNil("AmLrV7tg69KE5ZQehkjGiA7yJyDxJ25uyF96PMRptfzwozhAJbaK"), "6eULcXBWJMgjCMeGybbpfoPWVwyuidMFFjsrJJvXmF7z"},
		{true, base58.DecodeOrNil("AmLrV7tg69KE5ZQehkjGiA7yJyDxJ25uyF96PMRptfzwozhAJbaK"), "6eULcXBWJMgjCMeGybbpfoPWVwyuidMFFjsrJJvXmF7z"},
	} {
		as := &AccountState{
			id: test.id,
		}
		var aid types.AccountID
		if test.padding {
			aid = types.ToAccountID(as.ID())
		} else {
			aid = types.ToAccountID(as.IDNoPadding())
		}
		require.Equal(t, test.expectAid, aid.String())
	}
}
