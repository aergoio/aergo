package evm

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aergoio/aergo/v2/evm/compiled"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var (
	// testBlockState *state.BlockState
	blockContext   *types.BlockHeader
	testChainState *state.ChainStateDB
)

func initTest(t *testing.T) {
	blockContext = &types.BlockHeader{
		BlockNo: 1,
	}

	os.MkdirAll("test_db/lua", os.ModePerm)
	testChainState = state.NewChainStateDB()
	testChainState.Init("badgerdb", "test_db", nil, false)
}

func deinitTest(t *testing.T) {
	testChainState.Close()
	testChainState = nil
	os.RemoveAll("test_db")
}

// id 가 정의되지 않은 address 로 aergo 를 보냈을 경우, address 를 통해 id 를 가져올 수가 없는데, 그럴 때는 어떡하지?
// 아마 evm 내부에서 code 를 가져올 때 id 를 합쳐서 가져와서 오류가 나는듯. id 를 지워야 함
func TestHello(t *testing.T) {
	initTest(t)
	defer deinitTest(t)

	senderId := []byte(types.AergoSystem)
	receiverId := []byte(types.AergoName)

	// deploy
	var contractAddr []byte
	{
		bs := testChainState.NewBlockState(nil, nil, state.SetBlock(blockContext))
		// set sender state
		accStateSender, err := state.GetAccountState(senderId, bs)
		require.NoError(t, err)
		accStateSender.AddBalance(types.NewAmount(100, types.Aergo))
		err = accStateSender.PutState()
		require.NoError(t, err)
		accStateReceiver, err := state.GetAccountState(receiverId, bs)
		require.NoError(t, err)
		accStateReceiver.AddBalance(types.NewAmount(100, types.Aergo))
		err = accStateReceiver.PutState()
		require.NoError(t, err)

		bytecode, err := GetPayloadDeploy("HelloWorld.json")
		require.NoError(t, err)
		evm := NewEVM(nil, bs)

		var ret []byte
		var gasPrice *big.Int
		ret, contractAddr, gasPrice, err = evm.Create(accStateSender.EthID(), bytecode, 1000000)
		require.NoError(t, err)
		_ = ret
		_ = gasPrice
		var contract common.Address
		contract.SetBytes(contractAddr)

		// apply
		err = testChainState.Apply(bs)
		require.NoError(t, err)
	}

	// call
	{
		blockContext.BlockNo = 2
		bs := testChainState.NewBlockState(nil, nil, state.SetBlock(blockContext))
		accStateSender, err := state.GetAccountState(senderId, bs)
		require.NoError(t, err)

		bytecodeCall, err := GetPayloadCall("HelloWorld.json", "hello")
		require.NoError(t, err)

		evmCall := NewEVM(nil, bs)
		ret, gasPrice, err := evmCall.Call(accStateSender.EthID(), common.BytesToAddress(contractAddr), bytecodeCall, types.NewAmount(0, types.Aergo), 1000000)
		require.NoError(t, err)
		fmt.Println("ret :", string(ret))
		_ = gasPrice
	}
}

func TestSendAergo(t *testing.T) {
	initTest(t)
	defer deinitTest(t)

	senderId := []byte(types.AergoSystem)
	receiverId := []byte(types.AergoName)

	// deploy
	var contractAddr []byte
	{
		bs := testChainState.NewBlockState(nil, nil, state.SetBlock(blockContext))
		// set sender state
		accStateSender, err := state.GetAccountState(senderId, bs)
		require.NoError(t, err)
		accStateSender.AddBalance(types.NewAmount(10, types.Aergo))
		err = accStateSender.PutState()
		require.NoError(t, err)

		bytecode, err := GetPayloadDeploy("Send.json")
		require.NoError(t, err)
		evm := NewEVM(nil, bs)

		var ret []byte
		var gasPrice *big.Int
		ret, contractAddr, gasPrice, err = evm.Create(accStateSender.EthID(), bytecode, 10000000)
		require.NoError(t, err)
		_ = ret
		_ = gasPrice

		// apply
		err = testChainState.Apply(bs)
		require.NoError(t, err)

		newSenderState, err := state.GetAccountState(senderId, bs)
		require.NoError(t, err)
		fmt.Println("before sender balance   :", newSenderState.Balance().String())
		newReceiverState, err := state.GetAccountState(receiverId, bs)
		require.NoError(t, err)
		fmt.Println("before receiver balance :", newReceiverState.Balance().String())
		fmt.Println()
	}

	// call
	{
		blockContext.BlockNo = 2
		bs := testChainState.NewBlockState(nil, nil, state.SetBlock(blockContext))
		accStateSender, err := state.GetAccountState(senderId, bs)
		require.NoError(t, err)
		accStateReceiver, err := state.GetAccountState(receiverId, bs)
		require.NoError(t, err)
		bytecodeCall, err := GetPayloadCall("Send.json", "transferAergo", accStateReceiver.EthID(), types.NewAmount(1, types.Aergo))
		require.NoError(t, err)

		evmCall := NewEVM(nil, bs)
		ret, gasPrice, err := evmCall.Call(accStateSender.EthID(), common.BytesToAddress(contractAddr), bytecodeCall, types.NewAmount(1, types.Aergo), 10000000)
		require.NoError(t, err)
		_ = ret
		_ = gasPrice

		// apply
		err = testChainState.Apply(bs)
		require.NoError(t, err)

		newSenderState, err := state.GetAccountState(senderId, bs)
		require.NoError(t, err)
		fmt.Println("after sender balance    :", newSenderState.Balance().String())
		newReceiverState, err := state.GetAccountState(receiverId, bs)
		require.NoError(t, err)
		fmt.Println("after receiver balance  :", newReceiverState.Balance().String())
	}
}

func TestErc20(t *testing.T) {
	initTest(t)
	defer deinitTest(t)

	sender := types.GetSpecialAccountEth([]byte("aergo.system"))
	data, err := GetPayloadDeploy("ERC20.json", sender, big.NewInt(0))
	require.NoError(t, err)

	bs := testChainState.NewBlockState(nil, nil, state.SetBlock(blockContext))
	evm := NewEVM(nil, bs)
	ret, contractAddr, gasPrice, err := evm.Create(sender, data, 10000000)
	fmt.Println("contract :", contractAddr)
	fmt.Println("ret :", string(ret))
	fmt.Println("gas price :", gasPrice.String())
	fmt.Println("error :", err)

	err = testChainState.Apply(bs)
	require.NoError(t, err)

}

//---------------------------------------------------------------------------//
// utility function for tests

func GetPayloadDeploy(fileName string, args ...interface{}) ([]byte, error) {
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
	return contract.DeployData(args...)
}

func GetPayloadCall(fileName string, funcName string, args ...interface{}) ([]byte, error) {
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
