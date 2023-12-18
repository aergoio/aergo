package evm

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/require"
)

var (
	testBlockState *state.BlockState
	blockContext   *types.BlockHeader
)

func initTest(t *testing.T) {
	blockContext = &types.BlockHeader{
		BlockNo: 1,
	}

	os.MkdirAll("test_db/lua", os.ModePerm)
	luaStore := db.NewDB(db.ImplType(db.BadgerImpl), "test_db/lua")
	os.MkdirAll("test_db/eth", os.ModePerm)
	ethStore, _ := ethdb.NewDB("test_db/eth", string(db.BadgerImpl))

	luaState := statedb.NewStateDB(luaStore, nil, false)
	ethState, _ := ethdb.NewStateDB(nil, ethStore)
	testBlockState = state.NewBlockState(luaState, ethState, state.SetGasPrice(system.GetGasPrice()), state.SetBlock(blockContext))
}

func deinitTest(t *testing.T) {
	testBlockState.LuaStateDB.Store.Close()
	testBlockState.EthStateDB.GetStateDB().Database().DiskDB().Close()
	testBlockState = nil
	os.RemoveAll("test_db")
}

// id 가 정의되지 않은 address 로 aergo 를 보냈을 경우, address 를 통해 id 를 가져올 수가 없는데, 그럴 때는 어떡하지?
func TestHello(t *testing.T) {
	t.Skip()

	initTest(t)
	defer deinitTest(t)

	sender := types.GetSpecialAccountEth([]byte("aergo.system"))
	// payload :=

	// ethtypes.NewTransaction()
	evm := NewEVM(nil, 1000000, testBlockState)
	ret, _, gasPrice, err := evm.Create(sender, []byte("608060405234801561000f575f80fd5b5060043610610029575f3560e01c8063cfae32171461002d575b5f80fd5b61003561004b565b60405161004291906100d6565b60405180910390f35b5f805461005790610122565b80601f016020809104026020016040519081016040528092919081815260200182805461008390610122565b80156100ce5780601f106100a5576101008083540402835291602001916100ce565b820191905f5260205f20905b8154815290600101906020018083116100b157829003601f168201915b505050505081565b5f602080835283518060208501525f5b81811015610102578581018301518582016040015282016100e6565b505f604082860101526040601f19601f8301168501019250505092915050565b600181811c9082168061013657607f821691505b60208210810361015457634e487b7160e01b5f52602260045260245ffd5b5091905056fea2646970667358221220f5cf3d077ae43a818b9ce2fe81f238857611699e988d432db068008c8f7b337c64736f6c63430008170033"))
	fmt.Println("ret :", string(ret))
	fmt.Println("gas price :", gasPrice.String())
	fmt.Println("error :", err)

	_ = evm
}

// utility function for tests
func readSolidity(t *testing.T, file string) (luaCode string) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if ok != true {
		return ""
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "test_files", file))
	require.NoErrorf(t, err, "failed to read "+file)
	require.NotEmpty(t, raw, "failed to read "+file)
	return string(raw)
}
