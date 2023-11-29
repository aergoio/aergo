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

	sdbOld, err := NewStateDB(0, nil, db)
	require.NoError(t, err)

	require.Equal(t, sdbOld.Root(), types.EmptyRootHash.Bytes(), "root mismatch with expect")
	require.Equal(t, sdbOld.Root(), sdbOld.evmStateDB.IntermediateRoot(false).Bytes(), "root mismatch with IntermediateRoot(false)")
	require.Equal(t, sdbOld.Root(), sdbOld.evmStateDB.IntermediateRoot(true).Bytes(), "root mismatch with IntermediateRoot(true)")

	// put and commit
	addr := common.BigToAddress(big.NewInt(1))
	balance := big.NewInt(100)
	nonce := uint64(1)
	sdbOld.PutState(addr, balance, nonce)

	newRoot, err := sdbOld.Commit()
	require.NoError(t, err)
	fmt.Println("newRoot", newRoot)
	sdbOld.Close()

	sdbNew, err := NewStateDB(1, newRoot, db)
	require.NoError(t, err)

	fmt.Println(sdbNew.evmStateDB.GetBalance(addr).String())

	iter := db.Store.NewIterator(nil, nil)
	fmt.Println("key", iter.Key(), "value", iter.Value())
}
