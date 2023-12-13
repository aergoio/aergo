package state

import (
	"testing"

	key "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/require"
)

func TestEthAccount(t *testing.T) {
	for _, test := range []struct {
		account    []byte
		luaAccount string
		evmAccount string
	}{
		{[]byte{3, 112, 238, 158, 33, 17, 101, 89, 68, 95, 155, 198, 187, 127, 233, 38, 180, 46, 175, 134, 228, 158, 146, 64, 79, 134, 218, 92, 224, 53, 224, 191, 124},
			"AmPJht1vphFthrzB5TTHBRsqJ3yezmPgb9P8V8v4aCbCLpEg2WYC", // base58check encoded from compressed pubkey ( 33 length )
			"0xDe00eFa73EF966c5a28DbD6E2C9A2830a3453207"},          // hex encoded from uncompressed pubkey ( 65 length )
	} {
		acc := key.NewAddressEth(test.account)
		require.Equal(t, test.evmAccount, acc.String())
	}
}

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
