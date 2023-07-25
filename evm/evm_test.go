package evm

import (
	"encoding/hex"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestGETH(t *testing.T) {
	// set up levelDB
	testFile, _ := ioutil.TempFile(os.TempDir(), "temp")
	testPath, _ := filepath.Abs(filepath.Dir(testFile.Name()))
	testPath = testPath + "/ethleveldb/test"
	t.Log("creating temp DB at", testPath)
	testLevelDB, _ := rawdb.NewLevelDBDatabase(testPath, 128, 1024, "", false)
	defer testLevelDB.Close()

	// set up state
	testDB := state.NewDatabase(testLevelDB)
	stateRoot := common.Hash{}
	testState, _ := state.New(stateRoot, testDB, nil)
	if testState == nil {
		t.Errorf("eth state not created")
	}
	t.Log("created eth state")

	testAddress := common.HexToAddress("0x0a")
	// create evmCfg
	evmCfg := vm.Config{
		Debug: false,
	}

	// create call cfg
	testCfg := &runtime.Config{
		State:     testState,
		EVMConfig: evmCfg,
	}

	// configure state #1
	testState.SetCode(testAddress, []byte{
		byte(vm.PUSH1), 10,
		byte(vm.PUSH1), 0,
		byte(vm.MSTORE),
		byte(vm.PUSH1), 20,
		byte(vm.PUSH1), 32,
		byte(vm.MSTORE),
		byte(vm.PUSH1), 32,
		byte(vm.PUSH1), 0,
		byte(vm.RETURN),
	})

	// execute contract #1
	ret, _, err := runtime.Call(testAddress, nil, testCfg)
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	// test result #1
	num := new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(10)) != 0 {
		t.Error("Expected 10, got", num)
	}

	// configure state #2
	testState.SetCode(testAddress, []byte{
		byte(vm.PUSH1), 10,
		byte(vm.PUSH1), 0,
		byte(vm.MSTORE),
		byte(vm.PUSH1), 20,
		byte(vm.PUSH1), 32,
		byte(vm.MSTORE),
		byte(vm.PUSH1), 32,
		byte(vm.PUSH1), 32,
		byte(vm.RETURN),
	})

	// execute contract #2
	ret, _, err = runtime.Call(testAddress, nil, &runtime.Config{State: testState})
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	// test result #3
	num = new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(20)) != 0 {
		t.Error("Expected 20, got", num)
	}

	// try creating new state (new block)
	nextRoot, _ := testState.Commit(true)
	testState2, _ := state.New(nextRoot, testDB, nil)

	if testState2 == nil {
		t.Error("Failed to create a new state")
	}

	// configure new state
	testState2.SetCode(testAddress, []byte{
		byte(vm.PUSH1), 31,
		byte(vm.PUSH1), 0,
		byte(vm.MSTORE),
		byte(vm.PUSH1), 32,
		byte(vm.PUSH1), 0,
		byte(vm.RETURN),
	})

	// execute contract again
	testCfg2 := &runtime.Config{
		State:     testState2,
		EVMConfig: evmCfg,
	}

	ret, _, err = runtime.Call(testAddress, nil, testCfg2)
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	// test result again
	num = new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(31)) != 0 {
		t.Error("Expected 31, got", num)
	}

	// test result again prev state
	ret, _, err = runtime.Call(testAddress, nil, testCfg)
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	// test result
	num = new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(20)) != 0 {
		t.Error("Expected 10, got", num)
	}

	t.Log("evm test complete")
}

func TestEVM(t *testing.T) {

	testFile, _ := ioutil.TempFile(os.TempDir(), "temp")
	testPath, _ := filepath.Abs(filepath.Dir(testFile.Name()))
	testPath = testPath + "/ethleveldb/test"
	t.Log("creating temp DB at", testPath)

	LoadDatabase(testPath)

	testAddress, _ := hex.DecodeString("0x0a")

	ret, _, err := Call(testAddress, testAddress, []byte{
		byte(vm.PUSH1), 10,
		byte(vm.PUSH1), 0,
		byte(vm.MSTORE),
		byte(vm.PUSH1), 20,
		byte(vm.PUSH1), 32,
		byte(vm.MSTORE),
		byte(vm.PUSH1), 32,
		byte(vm.PUSH1), 0,
		byte(vm.RETURN),
	})

	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	// test result #1
	num := new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(10)) != 0 {
		t.Error("Expected 10, got", num)
	}
	t.Log("Invoked EVM in-memory")
	t.Log("ret length", len(ret))

	// test creating a contract
	// pragma solidity ^0.5.8;
	// contract  SampleContract {
	// function testFunc() public pure returns (int) {
	// 	return 1;
	// }
	// }

	contractByteCode, _ := hex.DecodeString("6080604052348015600f57600080fd5b5060878061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063037a417c14602d575b600080fd5b60336049565b6040518082815260200191505060405180910390f35b6000600190509056fea265627a7a7230582050d33093e20eb388eec760ca84ba30ec42dadbdeb8edf5cd8b261e89b8d4279264736f6c634300050a0032")
	ret, contractAddress, _, err := Create(testAddress, contractByteCode)
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	t.Log("created a contract with byte length", len(contractByteCode))
	t.Log("ret length", len(ret))
	t.Log("deployed at", hex.EncodeToString(contractAddress))

	// test creating contract again from different address
	testAddress2, _ := hex.DecodeString("0x0b")
	ret, contractAddress, _, err = Create(testAddress2, contractByteCode)
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	t.Log("created a contract with byte length", len(contractByteCode))
	t.Log("deployed at", hex.EncodeToString(contractAddress))
	t.Log("ret length", len(ret))
	// ret is the actual deployed contrat code, while contractByteCode includes deployment code
	t.Log("ret value", hex.EncodeToString(ret))

	Commit()

	// try calling the contract
	// need to call testFunc
	methodHash := crypto.Keccak256([]byte("testFunc()"))[:4]
	ret, gas, err := Call(testAddress, contractAddress, methodHash)
	if err != nil {
		t.Log("code", gas)
		t.Fatal("didn't expect error:", err)
	}

	t.Log("called the contract")
	t.Log("ret length", len(ret))
	t.Log("ret value", hex.EncodeToString(ret))
}
