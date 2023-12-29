package ethdb

import (
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	db, err := NewDB(testPath, "memorydb")
	require.NoError(t, err)

	sdbOld, err := NewStateDB(nil, db)
	require.NoError(t, err)

	require.Equal(t, sdbOld.Root(), ethtypes.EmptyRootHash.Bytes(), "root mismatch with expect")
	require.Equal(t, sdbOld.Root(), sdbOld.StateDB.IntermediateRoot(false).Bytes(), "root mismatch with IntermediateRoot(false)")
	require.Equal(t, sdbOld.Root(), sdbOld.StateDB.IntermediateRoot(true).Bytes(), "root mismatch with IntermediateRoot(true)")

	// put and commit
	addr := common.BigToAddress(big.NewInt(1))
	st := &types.State{
		Balance:  big.NewInt(100).Bytes(),
		Nonce:    0,
		CodeHash: []byte("code"),
	}
	sdbOld.Put(addr, st)

	_, err = sdbOld.Commit(0)
	require.NoError(t, err)

}
