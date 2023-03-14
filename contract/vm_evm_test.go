package contract

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
)

func TestEVM(t *testing.T) {
	// load eth DB and state
	testState, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	if testState == nil {
		t.Errorf("eth state not created")
	}
	t.Log("created eth state")

	// create contract
	testAddress := common.HexToAddress("0x0a")
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

	// execute contract
	ret, _, err := runtime.Call(testAddress, nil, &runtime.Config{State: testState})
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	// test result
	num := new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(10)) != 0 {
		t.Error("Expected 10, got", num)
	}

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

	// execute contract
	ret, _, err = runtime.Call(testAddress, nil, &runtime.Config{State: testState})
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	// test result
	num = new(big.Int).SetBytes(ret)
	if num.Cmp(big.NewInt(20)) != 0 {
		t.Error("Expected 20, got", num)
	}
}
