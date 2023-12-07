package main

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEVM(t *testing.T) {
	t.Skip()
	ethrdb, err := ethdb.NewDB("./data/state_evm", "leveldb")
	require.NoError(t, err)
	rootHash, _ := ethrdb.Store.Get(dbkey.EthRootHash())

	luaRootHash := []byte{44, 187, 127, 8, 150, 64, 67, 28, 6, 32, 223, 231, 89, 246, 155, 54, 190, 217, 27, 184, 222, 48, 88, 98, 165, 44, 184, 221, 145, 169, 187, 175}
	evmRootHash := []byte{23, 112, 84, 112, 102, 122, 167, 23, 97, 21, 7, 89, 190, 111, 98, 146, 108, 124, 117, 78, 200, 56, 2, 249, 152, 113, 57, 243, 200, 170, 7, 210}
	require.Equal(t, evmRootHash, rootHash)

	ethsdb, err := ethdb.NewStateDB(rootHash, ethrdb)
	require.NoError(t, err)
	luardb := db.NewDB(db.BadgerImpl, "./data/state")
	luasdb := statedb.NewStateDB(luardb, luaRootHash, false)

	ethDump := ethsdb.GetStateDB().RawDump(&state.DumpConfig{SkipStorage: true})

	for address, state := range ethDump.Accounts {
		id := ethsdb.GetId(address)
		luaState, err := luasdb.GetAccountState(types.ToAccountID(id))
		require.NoError(t, err)
		evmBalance := state.Balance
		luaBalance := big.NewInt(0).SetBytes(luaState.Balance).String()

		fmt.Printf("bal: (%v, %v) | nonce: (%v, %v) | addr: (%v, %v) \n", evmBalance, luaBalance, state.Nonce, luaState.Nonce, address.Hex(), types.EncodeAddress(id))
		assert.Equal(t, evmBalance, luaBalance)
		assert.Equal(t, state.Nonce, luaState.Nonce)
	}
	fmt.Println("total addr :", len(ethDump.Accounts))

}
