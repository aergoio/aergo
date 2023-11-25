package account

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountFormat(t *testing.T) {
	fmt.Println(len([]byte{3, 112, 238, 158, 33, 17, 101, 89, 68, 95, 155, 198, 187, 127, 233, 38, 180, 46, 175, 134, 228, 158, 146, 64, 79, 134, 218, 92, 224, 53, 224, 191, 124}))
	for _, test := range []struct {
		account    []byte
		luaAccount string
		evmAccount string
	}{
		{[]byte{3, 112, 238, 158, 33, 17, 101, 89, 68, 95, 155, 198, 187, 127, 233, 38, 180, 46, 175, 134, 228, 158, 146, 64, 79, 134, 218, 92, 224, 53, 224, 191, 124},
			"AmPJht1vphFthrzB5TTHBRsqJ3yezmPgb9P8V8v4aCbCLpEg2WYC", // base58check encoded from compressed pubkey ( 33 length )
			"0x17cC364f7b86772b7bbd40e23e0217eDC7edbCCF"},          // hex encoded from uncompressed pubkey ( 65 length )
	} {
		acc, err := NewAccount(test.account, nil, nil)
		require.NoError(t, err)

		require.Equal(t, test.luaAccount, acc.LuaAccount)
		require.Equal(t, test.evmAccount, acc.EthAccount.String())
	}
}
