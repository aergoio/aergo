package evm

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	key "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
)

func TestLevelDB(t *testing.T) {
	// set up levelDB
	testFile, _ := ioutil.TempFile(os.TempDir(), "temp")
	testPath, _ := filepath.Abs(filepath.Dir(testFile.Name()))
	testPath = testPath + "/ethleveldb/test"
	t.Log("creating temp DB at", testPath)
	testLevelDB, _ := rawdb.NewLevelDBDatabase(testPath, 128, 1024, "", false)
	defer os.RemoveAll(testPath)

	res, _ := testLevelDB.Get([]byte("foo"))
	if res != nil {
		t.Errorf("fetching non-existant yields non nil item")
		t.Log(string(res))
	}

	testLevelDB.Put([]byte("foo"), []byte("bar"))

	res, _ = testLevelDB.Get([]byte("foo"))
	if !bytes.Equal(res, []byte("bar")) {
		t.Errorf("retrieved value does not match")
	}

	testLevelDB.Close()

	// open level DB against
	testLevelDB, _ = rawdb.NewLevelDBDatabase(testPath, 128, 1024, "", false)
	res, _ = testLevelDB.Get([]byte("foo"))
	if !bytes.Equal(res, []byte("bar")) {
		t.Errorf("retrieved value does not match")
	}

	t.Log("foo value", string(res))

	res, _ = testLevelDB.Get([]byte("foo2"))
	if res != nil {
		t.Errorf("fetching non-existant yields non nil item")
		t.Log(string(res))
	}

}

func TestGETH(t *testing.T) {
	// set up levelDB
	testFile, _ := ioutil.TempFile(os.TempDir(), "temp")
	testPath, _ := filepath.Abs(filepath.Dir(testFile.Name()))
	testPath = testPath + "/ethleveldb/test"
	t.Log("creating temp DB at", testPath)
	testLevelDB, _ := rawdb.NewLevelDBDatabase(testPath, 128, 1024, "", false)
	defer testLevelDB.Close()
	defer os.RemoveAll(testPath)

	// set up state
	testDB := state.NewDatabase(testLevelDB)
	stateRoot := common.Hash{}
	testState, _ := state.New(stateRoot, testDB, nil)
	if testState == nil {
		t.Errorf("eth state not created")
	}
	t.Log("created eth state")

	testLevelDB.Put([]byte("foo"), []byte("bar"))
	res, _ := testLevelDB.Get([]byte("foo"))
	if !bytes.Equal(res, []byte("bar")) {
		t.Errorf("retrieved value does not match")
	}
	t.Log("fetching foo:", string(res))

	res, _ = testLevelDB.Get([]byte("foo2"))
	if res != nil {
		t.Errorf("fetching non-existant yields non nil item")
	}

	testAddress := common.HexToAddress("0x0a")
	// create evmCfg
	evmCfg := vm.Config{}

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
	nextRoot, _ := testState.Commit(0, true)
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

/*
	func TestEVM(t *testing.T) {
		testFile, _ := os.CreateTemp(os.TempDir(), "temp")
		testPath, _ := filepath.Abs(filepath.Dir(testFile.Name()))
		testPath = testPath + "/ethleveldb/test"

		t.Log("creating temp DB at", testPath)
		defer os.RemoveAll(testPath)
		evm := NewEVM()
		evm.LoadDatabase(testPath)
		if evm.prevStateRoot.String() != nullRootHash {
			t.Error("state root should be nil")
		}

		t.Log("root hash", evm.prevStateRoot.String())

		evm.Commit()

		if evm.prevStateRoot.String() == nullRootHash {
			t.Error("state root should not be nil")
		}

		t.Log("root hash", evm.prevStateRoot.String())

		evm.Commit()
		t.Log("root hash", evm.prevStateRoot.String())

		testAddress, _ := hex.DecodeString("0x000000000000000000000000000000000000000a")

		ret, _, err := evm.Call(testAddress, testAddress, []byte{
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
		// contractByteCode, _ := hex.DecodeString("6080604052348015600f57600080fd5b5060878061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063037a417c14602d575b600080fd5b60336049565b6040518082815260200191505060405180910390f35b6000600190509056fea265627a7a7230582050d33093e20eb388eec760ca84ba30ec42dadbdeb8edf5cd8b261e89b8d4279264736f6c634300050a0032")
		contractByteCode, _ := hex.DecodeString("6060604052341561000f57600080fd5b60405160208061020183398101604052808051906020019091905050806000806101000a81548163ffffffff021916908363ffffffff160217905550506101a68061005b6000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680630e1c06e414610051578063203dfd5614610086575b600080fd5b341561005c57600080fd5b610064610112565b604051808263ffffffff1663ffffffff16815260200191505060405180910390f35b341561009157600080fd5b6100bd600480803573ffffffffffffffffffffffffffffffffffffffff16906020019091905050610127565b604051808363ffffffff1663ffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390f35b6000809054906101000a900463ffffffff1681565b60008060016000809054906101000a900463ffffffff16016000806101000a81548163ffffffff021916908363ffffffff1602179055506000809054906101000a900463ffffffff1683915091509150915600a165627a7a72305820e0fce0096a05377bd5e28fb780a6b9e5c8cdd7ca708ecf3d47a99808761247b10029")
		ret, contractAddress, _, err := evm.Create(testAddress, contractByteCode)
		if err != nil {
			t.Fatal("didn't expect error", err)
		}

		t.Log("created a contract with byte length", len(contractByteCode))
		t.Log("ret length", len(ret))
		t.Log("deployed at", hex.EncodeToString(contractAddress))

		evm.Commit()
		t.Log("root hash", evm.prevStateRoot.String())

		// test creating contract again from different address
		testAddress2, _ := hex.DecodeString("0x0b")
		ret, contractAddress, _, err = evm.Create(testAddress2, contractByteCode)
		if err != nil {
			t.Fatal("didn't expect error", err)
		}

		t.Log("created a contract with byte length", len(contractByteCode))
		t.Log("deployed at", hex.EncodeToString(contractAddress))
		t.Log("ret length", len(ret))
		// ret is the actual deployed contrat code, while contractByteCode includes deployment code
		t.Log("ret value", hex.EncodeToString(ret))

		evm.Commit()
		t.Log("root hash", evm.prevStateRoot.String())

		// try calling the contract
		// need to call testFunc
		methodHash := crypto.Keccak256([]byte("storageTest(address)"))[:4]
		arg1, _ := hex.DecodeString("88e726de6cbadc47159c6ccd4f7868ae7a037730")
		payload := append(methodHash, arg1...) // ABI layout: hash("storageTest(address") + address

		t.Log("payload", hex.EncodeToString(arg1))
		t.Log("payload", hex.EncodeToString(payload))

		ret, gas, err := evm.Call(testAddress, contractAddress, payload)
		if err != nil {
			t.Log("code", gas)
			t.Fatal("didn't expect error:", err)
		}

		t.Log("called the contract")
		t.Log("ret length", len(ret))
		t.Log("ret value", hex.EncodeToString(ret))

		evm.Commit()
		t.Log("root hash", evm.prevStateRoot.String())

		// call again
		ret, gas, err = evm.Call(testAddress, contractAddress, payload)
		if err != nil {
			t.Log("code", gas)
			t.Fatal("didn't expect error:", err)
		}

		t.Log("called the contract")
		t.Log("ret length", len(ret))
		t.Log("ret value", hex.EncodeToString(ret))

		// try query
		ret, gas, err = evm.Query(testAddress, contractAddress, payload)
		if err != nil {
			t.Log("code", gas)
			t.Fatal("didn't expect error:", err)
		}

		t.Log("queried the contract")
		t.Log("ret length", len(ret))
		t.Log("ret value", hex.EncodeToString(ret))

		// call again
		ret, gas, err = evm.Call(testAddress, contractAddress, payload)
		if err != nil {
			t.Log("code", gas)
			t.Fatal("didn't expect error:", err)
		}

		t.Log("called the contract")
		t.Log("ret length", len(ret))
		t.Log("ret value", hex.EncodeToString(ret))

		// try query
		ret, gas, err = evm.Query(testAddress, contractAddress, payload)
		if err != nil {
			t.Log("code", gas)
			t.Fatal("didn't expect error:", err)
		}

		t.Log("queried the contract")
		t.Log("ret length", len(ret))
		t.Log("ret value", hex.EncodeToString(ret))

		evm.Commit()

		// try query
		ret, gas, err = evm.Query(testAddress, contractAddress, payload)
		if err != nil {
			t.Log("code", gas)
			t.Fatal("didn't expect error:", err)
		}

		t.Log("queried the contract")
		t.Log("ret length", len(ret))
		t.Log("ret value", hex.EncodeToString(ret))

		lastRootHash := evm.prevStateRoot.String()

		evm.CloseDatabase()
		// try loading EVM again
		evm.LoadDatabase(testPath)

		currentRootHash := evm.prevStateRoot.String()
		t.Log("last root", lastRootHash, "current root", currentRootHash)
	}
*/
func TestAddressConversion(t *testing.T) {
	// 02d9cff15387da27a1df7e8144d3e04133b0aabc2cc5b2830dfb9020b33d897bc4
	// -> 65ab92e68e2f79d4541367184d3c52b9bf733d0f
	original, _ := hex.DecodeString("02d9cff15387da27a1df7e8144d3e04133b0aabc2cc5b2830dfb9020b33d897bc4")
	convertedAddress := key.NewAccountEth(original)
	t.Log(convertedAddress.Hex())
}
