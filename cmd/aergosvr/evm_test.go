package main

import (
	"fmt"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types/dbkey"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/stretchr/testify/require"
)

func TestEVM(t *testing.T) {
	ethrdb, err := ethdb.NewDB("./data/state_evm", "leveldb")
	require.NoError(t, err)
	rootHash, _ := ethrdb.Store.Get(dbkey.EthRootHash())

	ethsdb, err := ethdb.NewStateDB(rootHash, ethrdb)
	require.NoError(t, err)

	ethDump := ethsdb.GetStateDB().RawDump(&state.DumpConfig{SkipStorage: true})

	for address, state := range ethDump.Accounts {
		fmt.Printf("addr: %v, bal: %v, nonce: %v\n", address, state.Balance, state.Nonce)
	}

	luardb := db.NewDB("./data/state", "badgerdb")
	luasdb := statedb.NewStateDB(luardb, nil, false)
	_ = luasdb
	// luasdb.GetState()

}
