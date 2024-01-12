package statedb

import (
	"fmt"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	var rootHash []byte
	rawdb := db.NewDB(db.MemoryImpl, "./testdb")
	sdb := NewStateDB(rawdb, rootHash, false)

	dump, err := sdb.RawDump()
	require.NoError(t, err)
	require.Equal(t, rootHash, dump.Root)

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
