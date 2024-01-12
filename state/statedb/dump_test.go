package statedb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	initTest(t)
	defer deinitTest()

	dump, err := stateDB.RawDump()
	require.NoError(t, err)
	require.Equal(t, stateDB.Trie.Root, dump.Root)

	for aid, account := range dump.Accounts {
		fmt.Printf("accountId : [%v]\n", aid)
		fmt.Printf("balance : [%v]\n", account.State.Balance)
		fmt.Printf("nonce : [%v]\n", account.State.Nonce)
		if len(account.State.CodeHash) > 0 {
			fmt.Printf("codeHash : [%v]\n", account.State.CodeHash)
			fmt.Printf("code : [%v]\n", account.Code)
		}
	}
}
