package ethdb

import (
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
	require.Equal(t, sdbOld.Root(), sdbOld.ethStateDB.IntermediateRoot(false).Bytes(), "root mismatch with IntermediateRoot(false)")
	require.Equal(t, sdbOld.Root(), sdbOld.ethStateDB.IntermediateRoot(true).Bytes(), "root mismatch with IntermediateRoot(true)")

	// put and commit
	addr := common.BigToAddress(big.NewInt(1))
	balance := big.NewInt(100)
	nonce := uint64(0)
	code := []byte("code")
	sdbOld.PutState(nil, addr, balance, nonce, code)

	_, err = sdbOld.Commit(0)
	require.NoError(t, err)

}
