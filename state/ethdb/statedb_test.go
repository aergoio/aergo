package ethdb

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	db, err := NewDB(testPath, "memorydb")
	require.NoError(t, err)

	sdbOld, err := NewStateDB(0, nil, db)
	require.NoError(t, err)

	expectRoot := []byte{86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33}
	require.Equal(t, sdbOld.Root(), expectRoot, "root mismatch with expect")
	require.Equal(t, sdbOld.Root(), sdbOld.evmStateDB.IntermediateRoot(false).Bytes(), "root mismatch with IntermediateRoot(false)")
	require.Equal(t, sdbOld.Root(), sdbOld.evmStateDB.IntermediateRoot(true).Bytes(), "root mismatch with IntermediateRoot(true)")

	// put and commit
	addr := common.BigToAddress(big.NewInt(1))
	balance := big.NewInt(100)
	nonce := uint64(1)
	sdbOld.PutState(addr, balance, nonce)

	newRoot, err := sdbOld.Commit()
	require.NoError(t, err)
	fmt.Println("oldRoot", expectRoot)
	fmt.Println("newRoot", newRoot)
	sdbOld.Close()

	sdbNew, err := NewStateDB(1, newRoot, db)
	require.NoError(t, err)

	fmt.Println(sdbNew.evmStateDB.GetBalance(addr).String())

	dump := sdbOld.evmStateDB.RawDump(nil)
	for k, v := range dump.Accounts {
		fmt.Println("addr :", k, "balance :", v.Balance)
	}
	// iter := memoryDB.NewIterator(nil, nil)
	// fmt.Println("key", iter.Key(), "value", iter.Value())
}
