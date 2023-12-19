package evm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/evm/compiled"
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

func TestHello(t *testing.T) {
	initTest(t)
	defer deinitTest(t)

	sender := types.GetSpecialAccountEth([]byte("aergo.system"))
	contract := compiled.CompiledContract{}
	err := json.Unmarshal(compiled.HelloWorldJSON, &contract)
	require.NoError(t, err)

	payload, err := contract.Data()
	require.NoError(t, err)

	evm := NewEVM(nil, 10000000, testBlockState)
	ret, contractAddr, gasPrice, err := evm.Create(sender, payload)
	fmt.Println("contract :", contractAddr)
	fmt.Println("ret :", string(ret))
	fmt.Println("gas price :", gasPrice.String())
	fmt.Println("error :", err)

	_ = evm
}

// id 가 정의되지 않은 address 로 aergo 를 보냈을 경우, address 를 통해 id 를 가져올 수가 없는데, 그럴 때는 어떡하지?
func TestErc20(t *testing.T) {
	initTest(t)
	defer deinitTest(t)

	sender := types.GetSpecialAccountEth([]byte("aergo.system"))

	data, err := compiled.ERC20Contract.Data(sender, big.NewInt(0))
	require.NoError(t, err)
	// ethtypes.NewTransaction()
	evm := NewEVM(nil, 10000000, testBlockState)
	ret, contractAddr, gasPrice, err := evm.Create(sender, data)
	fmt.Println("contract :", contractAddr)
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
