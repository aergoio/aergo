package contract

import (
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
)

func TestEVM(t *testing.T) {
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
	prevRoot, _ := testState.Commit(true)
	testState2, _ := state.New(prevRoot, testDB, nil)

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

}
