package main

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types/dbkey"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEVM(t *testing.T) {
	t.Skip()

	ethrdb, err := ethdb.NewDB("./data/state_eth", "leveldb")
	require.NoError(t, err)

	rootHash, err := ethrdb.Store.Get(dbkey.EthRootHash())
	require.NoError(t, err)

	luaRootHash := []byte{21, 219, 27, 74, 192, 35, 249, 191, 91, 111, 1, 79, 10, 61, 234, 206, 120, 180, 91, 225, 253, 221, 12, 151, 103, 73, 29, 188, 34, 35, 152, 245}
	evmRootHash := []byte{37, 72, 66, 15, 51, 221, 78, 68, 29, 177, 38, 252, 116, 25, 53, 176, 6, 119, 172, 222, 252, 130, 90, 149, 189, 98, 249, 47, 246, 82, 253, 5}
	require.Equal(t, evmRootHash, rootHash)

	ethsdb, err := ethdb.NewStateDB(rootHash, ethrdb)
	require.NoError(t, err)
	luardb := db.NewDB(db.BadgerImpl, "./data/state")
	luasdb := statedb.NewStateDB(luardb, luaRootHash, false)

	ethDump := ethsdb.GetStateDB().RawDump(&state.DumpConfig{SkipStorage: true})

	for address, state := range ethDump.Accounts {
		if address == ethdb.IdManager.String() {
			continue
		}

		aid := ethsdb.GetAid(common.HexToAddress(address))
		luaState, err := luasdb.GetAccountState(aid)
		require.NoError(t, err)
		evmBalance := state.Balance
		luaBalance := big.NewInt(0).SetBytes(luaState.Balance).String()

		fmt.Printf("bal: (%v, %v) | nonce: (%v, %v) | addr: (%v, %v) \n", evmBalance, luaBalance, state.Nonce, luaState.Nonce, address, aid.String())
		assert.Equalf(t, evmBalance, luaBalance, address, aid.String())
		assert.Equalf(t, state.Nonce, luaState.Nonce, address, aid.String())
	}
	fmt.Println("total addr :", len(ethDump.Accounts))

}
