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

	luaRootHash := []byte{92, 90, 150, 219, 141, 234, 143, 68, 14, 116, 35, 101, 161, 220, 7, 20, 93, 34, 87, 84, 223, 170, 25, 217, 214, 33, 188, 229, 154, 158, 232, 0}
	evmRootHash := []byte{17, 123, 75, 250, 212, 173, 6, 117, 223, 85, 45, 148, 76, 133, 7, 211, 5, 29, 15, 144, 176, 232, 255, 0, 171, 9, 7, 83, 6, 58, 55, 202}
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
