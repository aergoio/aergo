package ethdb

import (
	"bytes"
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

func TestId(t *testing.T) {
	db, err := NewDB(testPath, "memorydb")
	require.NoError(t, err)
	sdb, err := NewStateDB(nil, db)
	require.NoError(t, err)

	for _, test := range []struct {
		eid common.Address
		aid types.AccountID
	}{
		{common.BytesToAddress([]byte{0x01}), types.ToAccountID([]byte{0x01})},
		{common.BytesToAddress(bytes.Repeat([]byte{0x01}, 20)), types.ToAccountID(bytes.Repeat([]byte{0x01}, 20))},
		{common.BytesToAddress(bytes.Repeat([]byte{0x01}, 32)), types.ToAccountID(bytes.Repeat([]byte{0x01}, 32))},
		{common.BytesToAddress(bytes.Repeat([]byte{0x01}, 64)), types.ToAccountID(bytes.Repeat([]byte{0x01}, 64))},
	} {
		sdb.PutId(test.eid, test.aid)
		newAid := sdb.GetAid(test.eid)
		require.Equal(t, test.aid, newAid, "aid mismatch with expect")
		newEid := sdb.GetEid(test.aid)
		require.Equal(t, test.eid, newEid, "eid mismatch with expect")
	}
}
