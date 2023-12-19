package evm

import (
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

func TestSendAergo(t *testing.T) {
	initTest(t)
	defer deinitTest(t)

	senderId := []byte(types.AergoSystem)
	receiverId := []byte(types.AergoName)

	accStateSender, err := state.GetAccountState(senderId, testBlockState)
	require.NoError(t, err)
	accStateSender.AddBalance(types.NewAmount(100, types.Aergo))
	err = accStateSender.PutState()
	require.NoError(t, err)

	accStateReceiver, err := state.GetAccountState(receiverId, testBlockState)
	require.NoError(t, err)

	bytecode, err := getPayloadDeploy("Send.json")
	require.NoError(t, err)

	// deploy
	evm := NewEVM(nil, 10000000, testBlockState)
	ret, contractAddr, gasPrice, err := evm.Create(accStateSender.EthID(), bytecode)
	require.NoError(t, err)
	fmt.Println("contract :", contractAddr)
	fmt.Println("ret :", string(ret))
	fmt.Println("gas price :", gasPrice.String())
	testBlockState.Commit()

	accStateSender.AddBalance(types.NewAmount(100, types.Aergo))
	err = accStateSender.PutState()

	// call
	bytecodeCall, err := getPayloadCall("Send.json", "transferEther", accStateReceiver.EthID(), types.NewAmount(1, types.Aergo))
	require.NoError(t, err)

	evmCall := NewEVM(nil, 10000000, testBlockState)
	ret, gasPrice, err = evmCall.Call(accStateSender.EthID(), contractAddr, bytecodeCall)
	require.NoError(t, err)
	fmt.Println("ret :", string(ret))
	fmt.Println("gas price :", gasPrice.String())

	fmt.Println("new balance sender :", accStateSender.Balance())
	fmt.Println("new balance receiver :", accStateReceiver.Balance())

	newSenderState, err := state.GetAccountState(senderId, testBlockState)
	require.NoError(t, err)
	fmt.Println("\nsender balance :", newSenderState.Balance().String())
}

func TestHello(t *testing.T) {
	initTest(t)
	defer deinitTest(t)

	sender := types.GetSpecialAccountEth([]byte("aergo.system"))
	data, err := compiled.HelloWorldContract.Data()
	require.NoError(t, err)

	evm := NewEVM(nil, 10000000, testBlockState)
	ret, contractAddr, gasPrice, err := evm.Create(sender, data)
	fmt.Println("contract :", contractAddr)
	fmt.Println("ret :", string(ret))
	fmt.Println("gas price :", gasPrice.String())
	fmt.Println("error :", err)
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
}

func TestDeploy(t *testing.T) {
	initTest(t)
	defer deinitTest(t)

	sender := types.GetSpecialAccountEth([]byte("aergo.system"))
	data, err := getPayloadDeploy("ERC20.json", sender, big.NewInt(0))
	require.NoError(t, err)

	evm := NewEVM(nil, 10000000, testBlockState)
	ret, contractAddr, gasPrice, err := evm.Create(sender, data)
	fmt.Println("contract :", contractAddr)
	fmt.Println("ret :", string(ret))
	fmt.Println("gas price :", gasPrice.String())
	fmt.Println("error :", err)
}

// utility function for tests
func getPayloadDeploy(fileName string, args ...interface{}) ([]byte, error) {
	_, filename, _, ok := runtime.Caller(0)
	if ok != true {
		return nil, fmt.Errorf("failed to get caller info")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "compiled", fileName))
	if err != nil {
		return nil, err
	}

	contract := compiled.CompiledContract{}
	err = contract.UnmarshalJSON(raw)
	if err != nil {
		return nil, err
	}
	return contract.Data(args...)
}

func getPayloadCall(fileName string, funcName string, args ...interface{}) ([]byte, error) {
	_, filename, _, ok := runtime.Caller(0)
	if ok != true {
		return nil, fmt.Errorf("failed to get caller info")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "compiled", fileName))
	if err != nil {
		return nil, err
	}

	contract := compiled.CompiledContract{}
	err = contract.UnmarshalJSON(raw)
	if err != nil {
		return nil, err
	}
	return contract.CallData(funcName, args...)
}
