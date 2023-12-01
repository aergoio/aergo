package ethdb

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	db, err := NewDB(testPath, "memorydb")
	require.NoError(t, err)

	sdbOld, err := NewStateDB(nil, db)
	require.NoError(t, err)

	require.Equal(t, sdbOld.Root(), types.EmptyRootHash.Bytes(), "root mismatch with expect")
	require.Equal(t, sdbOld.Root(), sdbOld.evmStateDB.IntermediateRoot(false).Bytes(), "root mismatch with IntermediateRoot(false)")
	require.Equal(t, sdbOld.Root(), sdbOld.evmStateDB.IntermediateRoot(true).Bytes(), "root mismatch with IntermediateRoot(true)")

	// put and commit
	addr := common.BigToAddress(big.NewInt(1))
	balance := big.NewInt(100)
	nonce := uint64(0)
	code := []byte("code")
	sdbOld.PutState(addr, balance, nonce, code)

	newRoot, err := sdbOld.Commit(0)
	require.NoError(t, err)
	fmt.Println("newRoot", newRoot)

	sdbNew, err := NewStateDB(newRoot, db)
	require.NoError(t, err)

	fmt.Println("blaance :", sdbNew.evmStateDB.GetBalance(addr).String())
	fmt.Println("code :", sdbNew.evmStateDB.GetCode(addr))
	fmt.Println("nonde :", sdbNew.evmStateDB.GetNonce(addr))

	iter := db.Store.NewIterator(nil, nil)
	for {
		fmt.Println("key", iter.Key(), "value", iter.Value())
		if !iter.Next() {
			break
		}
	}
}
