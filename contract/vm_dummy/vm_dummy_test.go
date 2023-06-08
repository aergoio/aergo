package vm_dummy

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if runtime.GOARCH != "amd64" {
		fmt.Println("skip architecture dependent test")
	} else {
		os.Exit(m.Run())
	}
}

func readLuaCode(file string) (luaCode string) {
	_, filename, _, ok := runtime.Caller(0)
	if ok != true {
		return ""
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "test_sample", file))
	if err != nil {
		return ""
	}

	return string(raw)
}

func TestContractSystem(t *testing.T) {
	code := readLuaCode("contract_system.lua")
	require.NotEmpty(t, code, "failed to read contract_system.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000))
	require.NoErrorf(t, err, "failed to new account")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "system", 0, code))
	require.NoErrorf(t, err, "failed to deploy contract")

	tx := NewLuaTxCall("ktlee", "system", 0, `{"Name":"testState", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	receipt := bc.GetReceipt(tx.Hash())
	exRv := fmt.Sprintf(`["Amg6nZWXKB6YpNgBPv9atcjdm6hnFvs5wMdRgb2e9DmaF5g9muF2","6FbDRScGruVdATaNWzD51xJkTfYCVwxSZDb7gzqCLzwf","AmhNNBNY7XFk4p5ym4CJf8nTcRTEHjWzAeXJfhP71244CjBCAQU3",%d,3,999]`, bc.cBlock.Header.Timestamp/1e9)
	assert.Equal(t, exRv, receipt.GetRet(), "receipt ret error")
}

func TestContractHello(t *testing.T) {
	code := readLuaCode("contract_hello.lua")
	require.NotEmpty(t, code, "failed to read contract_hello.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create test database")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000))
	require.NoErrorf(t, err, "failed to new account")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "hello", 0, code))
	require.NoErrorf(t, err, "failed to deploy contract")

	tx := NewLuaTxCall("ktlee", "hello", 0, `{"Name":"hello", "Args":["World"]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	receipt := bc.GetReceipt(tx.Hash())
	assert.Equal(t, `"Hello World"`, receipt.GetRet(), "receipt ret error")
}

func TestContractSend(t *testing.T) {
	code := readLuaCode("contract_send_1.lua")
	require.NotEmpty(t, code, "failed to read contract_send_1.lua")
	code2 := readLuaCode("contract_send_2.lua")
	require.NotEmpty(t, code2, "failed to read contract_send_2.lua")
	code3 := readLuaCode("contract_send_3.lua")
	require.NotEmpty(t, code3, "failed to read contract_send_3.lua")
	code4 := readLuaCode("contract_send_4.lua")
	require.NotEmpty(t, code4, "failed to read contract_send_4.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "test1", 50, code),
		NewLuaTxDeploy("ktlee", "test2", 0, code2),
		NewLuaTxDeploy("ktlee", "test3", 0, code3),
		NewLuaTxDeploy("ktlee", "test4", 0, code4),
	)
	assert.NoErrorf(t, err, "failed to deploy contract")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("test2")))))
	assert.NoErrorf(t, err, "failed to call tx")

	state, err := bc.GetAccountState("test2")
	assert.Equalf(t, uint64(2), state.GetBalanceBigInt().Uint64(), "balance error")

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("test3")))).Fail(`[Contract.LuaSendAmount] call err: not found function: default`),
	)
	assert.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("test4")))).Fail(`[Contract.LuaSendAmount] call err: 'default' is not payable`),
	)
	assert.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("ktlee")))),
	)
	assert.NoErrorf(t, err, "failed to connect new block")
}

func TestContractSendF(t *testing.T) {
	code := readLuaCode("contract_sendf_1.lua")
	require.NotEmpty(t, code, "failed to read contract_sendf_1.lua")
	code2 := readLuaCode("contract_sendf_2.lua")
	require.NotEmpty(t, code2, "failed to read contract_sendf_2.lua")

	bc, err := LoadDummyChain(SetPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "test1", 50000000000000000, code),
		NewLuaTxDeploy("ktlee", "test2", 0, code2),
	)
	require.NoErrorf(t, err, "failed to connect new block")

	tx := NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("test2"))))
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")

	r := bc.GetReceipt(tx.Hash())
	assert.Equalf(t, uint64(105087), r.GetGasUsed(), "gas used not equal")

	state, err := bc.GetAccountState("test2")
	assert.Equalf(t, uint64(2), state.GetBalanceBigInt().Uint64(), "balance state not equal")

	tx = NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send2", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("test2"))))
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")

	r = bc.GetReceipt(tx.Hash())
	assert.Equalf(t, uint64(105179), r.GetGasUsed(), "gas used not equal")

	state, err = bc.GetAccountState("test2")
	assert.Equalf(t, uint64(6), state.GetBalanceBigInt().Uint64(), "balance state not equal")
}

func TestContractQuery(t *testing.T) {
	code := readLuaCode("contract_query.lua")
	require.NotEmpty(t, code, "failed to read query.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "query", 0, code), NewLuaTxCall("ktlee", "query", 2, `{"Name":"inc", "Args":[]}`))
	require.NoErrorf(t, err, "failed to connect new block")

	query, err := bc.GetAccountState("query")
	require.NoErrorf(t, err, "failed to get account state")
	assert.Equalf(t, uint64(2), query.GetBalanceBigInt().Uint64(), "not equal balance")

	err = bc.Query("query", `{"Name":"inc", "Args":[]}`, "set not permitted in query", "")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
	require.NoErrorf(t, err, "failed to query")
}

func TestContractCall(t *testing.T) {
	code := readLuaCode("contract_call_1.lua")
	require.NotEmpty(t, code, "failed to read contract_call_1.lua")
	code2 := readLuaCode("contract_call_2.lua")
	require.NotEmpty(t, code2, "failed to read contract_call_2.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "counter", 0, code).Constructor("[1]"),
		NewLuaTxCall("ktlee", "counter", 0, `{"Name":"inc", "Args":[]}`),
	)
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "2")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(
		NewLuaTxDeploy("ktlee", "caller", 0, code2).Constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(contract.StrHash("counter")))),
		NewLuaTxCall("ktlee", "caller", 0, `{"Name":"add", "Args":[]}`),
	)
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "3")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("caller", `{"Name":"dget", "Args":[]}`, "", "99")
	require.NoErrorf(t, err, "failed to query")

	tx := NewLuaTxCall("ktlee", "caller", 0, `{"Name":"dadd", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")

	receipt := bc.GetReceipt(tx.Hash())
	assert.Equalf(t, `99`, receipt.GetRet(), "contract Call ret error")

	tx = NewLuaTxCall("ktlee", "caller", 0, `{"Name":"dadd", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")

	receipt = bc.GetReceipt(tx.Hash())
	assert.Equalf(t, `100`, receipt.GetRet(), "contract Call ret error")

	err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "3")
	require.NoErrorf(t, err, "failed to query")
}

func TestContractPingpongCall(t *testing.T) {
	code := readLuaCode("contract_pingpongcall_1.lua")
	require.NotEmpty(t, code, "failed to read contract_pingpongcall_1.lua")
	code2 := readLuaCode("contract_pingpongcall_2.lua")
	require.NotEmpty(t, code2, "failed to read contract_pingpongcall_2.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "a", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "b", 0, code2).Constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(contract.StrHash("a")))))
	require.NoErrorf(t, err, "failed to connect new block")

	tx := NewLuaTxCall("ktlee", "a", 0, fmt.Sprintf(`{"Name":"start", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("b"))))
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("a", `{"Name":"get", "Args":[]}`, "", `"callback"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("b", `{"Name":"get", "Args":[]}`, "", `"called"`)
	require.NoErrorf(t, err, "failed to query")
}

func TestRollback(t *testing.T) {
	code := readLuaCode("rollback.lua")
	require.NotEmpty(t, code, "failed to read rollback.lua")
	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000))
	require.NoErrorf(t, err, "failed to connect new block")
	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "query", 0, code), NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`))
	require.NoErrorf(t, err, "failed to connect new block")
	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`), NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`))
	require.NoErrorf(t, err, "failed to connect new block")
	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`), NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "5")
	require.NoErrorf(t, err, "failed to query")

	err = bc.DisConnectBlock()
	require.NoErrorf(t, err, "failed to disconnect block")

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "3")
	require.NoErrorf(t, err, "failed to query")

	err = bc.DisConnectBlock()
	require.NoErrorf(t, err, "failed to disconnect block")

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "2")
	require.NoErrorf(t, err, "failed to query")
}

func TestAbi(t *testing.T) {
	codeNoAbi := readLuaCode("abi_no.lua")
	require.NotEmpty(t, codeNoAbi, "failed to read abi_no.lua")
	codeEmpty := readLuaCode("abi_empty.lua")
	require.NotEmpty(t, codeEmpty, "failed to read abi_empty.lua")
	codeLocalFunc := readLuaCode("abi_localfunc.lua")
	require.NotEmpty(t, codeLocalFunc, "failed to read abi_localfunc.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "a", 0, codeNoAbi))
	require.Errorf(t, err, fmt.Sprintf("expected err : %s, buf got nil", "no exported functions"))
	require.Containsf(t, err.Error(), "no exported functions", "not contains error message")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "a", 0, codeEmpty))
	require.Errorf(t, err, fmt.Sprintf("expected err : %s, buf got nil", "no exported functions."))
	require.Containsf(t, err.Error(), "no exported functions.", "not contains error message")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "a", 0, codeLocalFunc))
	require.Errorf(t, err, fmt.Sprintf("expected err : %s, buf got nil", "global function expected"))
	require.Containsf(t, err.Error(), "global function expected", "not contains error message")
}

func TestGetABI(t *testing.T) {
	code := readLuaCode("getabi.lua")
	require.NotEmpty(t, code, "failed to read getabi.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "hello", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")

	abi, err := bc.GetABI("hello")
	require.NoErrorf(t, err, "failed to get abi")

	jsonAbi, err := json.Marshal(abi)
	require.NoErrorf(t, err, "failed to marshal abi")
	require.Equalf(t, `{"version":"0.2","language":"lua","functions":[{"name":"hello","arguments":[{"name":"say"}]}],"state_variables":[{"name":"Say","type":"value"}]}`, string(jsonAbi), "not equal abi")
}

func TestPayable(t *testing.T) {
	code := readLuaCode("payable.lua")
	require.NotEmpty(t, code, "failed to read payable.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "payable", 1, code))
	require.Errorf(t, err, "expected: 'constructor' is not payable")
	require.Containsf(t, err.Error(), "'constructor' is not payable", "not contains error message")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`).Fail("not found contract"))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "payable", 0, code), NewLuaTxCall("ktlee", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("payable", `{"Name":"load"}`, "", `"blahblah"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "payable", 1, `{"Name":"save", "Args": ["payed"]}`))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("payable", `{"Name":"load"}`, "", `"payed"`)
	require.NoErrorf(t, err, "failed to query")
}

func TestDefault(t *testing.T) {
	code := readLuaCode("default.lua")
	require.NotEmpty(t, code, "failed to read default.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "default", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")

	tx := NewLuaTxCall("ktlee", "default", 0, "")
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")

	receipt := bc.GetReceipt(tx.Hash())
	require.Equalf(t, `"default"`, receipt.GetRet(), "contract Call ret error")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "default", 1, "").Fail(`'default' is not payable`))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("default", `{"Name":"a"}`, "not found function: a", "")
	require.NoErrorf(t, err, "failed to query")
}

func TestReturn(t *testing.T) {
	code := readLuaCode("return_1.lua")
	require.NotEmpty(t, code, "failed to read return_1.lua")
	code2 := readLuaCode("return_2.lua")
	require.NotEmpty(t, code, "failed to read return_2.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "return_num", 0, code),
		NewLuaTxCall("ktlee", "return_num", 0, `{"Name":"return_num", "Args":[]}`),
	)
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("return_num", `{"Name":"return_num", "Args":[]}`, "", "10")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "foo", 0, code2))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("foo", `{"Name":"foo", "Args":[]}`, "", "[1,2,3]")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("foo", `{"Name":"foo2", "Args":["foo314"]}`, "", `"foo314"`)
	require.NoErrorf(t, err, "failed to query")
}

func TestReturnUData(t *testing.T) {
	code := readLuaCode("return_udata.lua")
	require.NotEmpty(t, code, "failed to read return_udata.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "rs-return", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "rs-return", 0, `{"Name": "test_die", "Args":[]}`).Fail(`unsupport type: userdata`))
	require.NoErrorf(t, err, "failed to connect new block")
}

func TestEvent(t *testing.T) {
	code := readLuaCode("event.lua")
	require.NotEmpty(t, code, "failed to read event.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "event", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "event", 0, `{"Name": "test_ev", "Args":[]}`))
	require.NoErrorf(t, err, "failed to connect new block")
}

func TestView(t *testing.T) {
	code := readLuaCode("view.lua")
	require.NotEmpty(t, code, "failed to read view.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "view", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "view", 0, `{"Name": "test_view", "Args":[]}`).Fail("[Contract.Event] event not permitted in query"))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("view", `{"Name":"k", "Args":[10]}`, "", "10")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "view", 0, `{"Name": "tx_in_view_function", "Args":[]}`).Fail("[Contract.Event] event not permitted in query"))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "view", 0, `{"Name": "tx_after_view_function", "Args":[]}`))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "view", 0, `{"Name": "k2", "Args":[]}`).Fail("[Contract.Event] event not permitted in query"))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "view", 0, `{"Name": "k3", "Args":[]}`))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "view", 0, `{"Name": "sqltest", "Args":[]}`).Fail("not permitted in view function"))
	require.NoErrorf(t, err, "failed to connect new block")
}

func TestDeploy(t *testing.T) {
	code := readLuaCode("deploy.lua")
	require.NotEmpty(t, code, "failed to read deploy.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "deploy", 50000000000, code))
	require.NoErrorf(t, err, "failed to connect new block")

	tx := NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"hello"}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")

	receipt := bc.GetReceipt(tx.Hash())
	assert.Equalf(t, `["AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf","Hello world"]`, receipt.GetRet(), "contract Call ret error")

	err = bc.Query("deploy", `{"Name":"helloQuery", "Args":["AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf"]}`, "", `"Hello world"`)
	require.NoErrorf(t, err, "failed to query")

	tx = NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"testConst"}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")

	receipt = bc.GetReceipt(tx.Hash())
	assert.Equalf(t, `["Amhmj6kKZz7mPstBAPJWRe1e8RHP7bZ5pV35XatqTHMWeAVSyMkc","Hello world2"]`, receipt.GetRet(), "contract Call ret error")

	deployAcc, err := bc.GetAccountState("deploy")
	require.NoErrorf(t, err, "failed to get account state")
	assert.Equalf(t, uint64(49999999900), deployAcc.GetBalanceBigInt().Uint64(), "not same balance")

	deployAcc, err = bc.GetAccountState("deploy")
	require.NoErrorf(t, err, "failed to get account state")

	tx = NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"testFail"}`)
	err = bc.ConnectBlock(tx)
	require.Errorf(t, err, "expect err : `constructor` is not payable")

	deployAcc, err = bc.GetAccountState("deploy")
	require.NoErrorf(t, err, "failed to get account state")
	assert.Equalf(t, uint64(2), deployAcc.Nonce, "not same nonce")

	tx = NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"testPcall"}`)
	err = bc.ConnectBlock(tx)
	require.Errorf(t, err, "expect err : cannot find contract Amhs9v8EeAAWrrvEFrvMng4UksHRsR7wN1iLqKkXw5bqMV18JP3h")

	deployAcc, err = bc.GetAccountState("deploy")
	require.NoErrorf(t, err, "failed to get account state")
	assert.Equalf(t, uint64(2), deployAcc.Nonce, "nonce rollback failed")

	receipt = bc.GetReceipt(tx.Hash())
	assert.Containsf(t, receipt.GetRet(), "cannot find contract", "contract Call ret error")
}

func TestDeploy2(t *testing.T) {
	code := readLuaCode("deploy2.lua")
	require.NotEmpty(t, code, "failed to read deploy2.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "deploy", 50000000000, code))
	require.NoErrorf(t, err, "failed to connect new block")

	tx := NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"hello"}`).Fail(`not permitted state referencing at global scope`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")
}

func TestNDeploy(t *testing.T) {
	code := readLuaCode("deployn.lua")
	require.NotEmpty(t, code, "failed to read deployn.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "n-deploy", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")
}

func TestDeployFee(t *testing.T) {
	code := readLuaCode("deployfee.lua")
	require.NotEmpty(t, code, "failed to read deployfee.lua")

	bc, err := LoadDummyChain(SetTimeout(50), SetPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "deploy", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")

	var before, use, after uint64
	use = 117861

	state, err := bc.GetAccountState("ktlee")
	require.NoErrorf(t, err, "failed to get account state")
	before = state.GetBalanceBigInt().Uint64()

	tx := NewLuaTxCall("ktlee", "deploy", 0, `{"Name": "testPcall"}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to connect new block")

	state, err = bc.GetAccountState("ktlee")
	require.NoErrorf(t, err, "failed to get account state")
	after = state.GetBalanceBigInt().Uint64()

	r := bc.GetReceipt(tx.Hash())
	assert.Equalf(t, use, r.GetGasUsed(), "not same gas used")
	assert.Equalf(t, before-use, after, "not same after balance")
}

func xestInfiniteLoop(t *testing.T) {
	code := readLuaCode("infiniteloop.lua")
	require.NotEmpty(t, code, "failed to read infiniteloop.lua")

	bc, err := LoadDummyChain(SetTimeout(50))
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "loop", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")

	errTimeout := "exceeded the maximum instruction count"

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "loop", 0, `{"Name":"infiniteLoop"}`))
	require.Errorf(t, err, "expected: %v", errTimeout)
	require.Containsf(t, err.Error(), errTimeout, "not contain timeout error")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "loop", 0, `{"Name":"catch"}`))
	require.Errorf(t, err, "expected: %v", errTimeout)
	require.Containsf(t, err.Error(), errTimeout, "not contain timeout error")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "loop", 0, `{"Name":"contract_catch"}`))
	require.Errorf(t, err, "expected: %v", errTimeout)
	require.Containsf(t, err.Error(), errTimeout, "not contain timeout error")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "loop", 0, `{"Name":"infiniteCall"}`).Fail("stack overflow"))
	require.NoErrorf(t, err, "failed to connect new block")
}

func TestInfiniteLoopOnPubNet(t *testing.T) {
	code := readLuaCode("infiniteloop.lua")
	require.NotEmpty(t, code, "failed to read infiniteloop.lua")

	bc, err := LoadDummyChain(SetTimeout(50), SetPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "loop", 0, code))
	require.NoErrorf(t, err, "failed to connect new block")

	errTimeout := contract.VmTimeoutError{}

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "loop", 0, `{"Name":"infiniteLoop"}`))
	require.Errorf(t, err, "expected: %v", errTimeout)
	require.Containsf(t, err.Error(), errTimeout.Error(), "not contain timeout error")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "loop", 0, `{"Name":"catch"}`))
	require.Errorf(t, err, "expected: %v", errTimeout)
	require.Containsf(t, err.Error(), errTimeout.Error(), "not contain timeout error")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "loop", 0, `{"Name":"contract_catch"}`))
	require.Errorf(t, err, "expected: %v", errTimeout)
	require.Containsf(t, err.Error(), errTimeout.Error(), "not contain timeout error")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "loop", 0, `{"Name":"infiniteCall"}`).Fail("stack overflow"))
	require.NoErrorf(t, err, "failed to connect new block")

}

func TestUpdateSize(t *testing.T) {
	code := readLuaCode("updatesize.lua")
	require.NotEmpty(t, code, "failed to read updatesize.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "loop", 0, code),
		NewLuaTxCall("ktlee", "loop", 0, `{"Name":"infiniteLoop"}`),
	)
	errMsg := "exceeded size of updates in the state database"
	require.Errorf(t, err, "expected: %s", errMsg)
	require.Containsf(t, err.Error(), errMsg, "error message not same as expected")
}

func TestTimeoutCnt(t *testing.T) {
	code := readLuaCode("timeout_1.lua")
	require.NotEmpty(t, code, "failed to read timeout_1.lua")

	code2 := readLuaCode("timeout_2.lua")
	require.NotEmpty(t, code, "failed to read timeout_2.lua")

	bc, err := LoadDummyChain(SetTimeout(500), SetPubNet()) // timeout 500 milliseconds
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "timeout-cnt", 0, code),
	)
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "timeout-cnt", 0, `{"Name": "infinite_loop"}`).Fail("contract timeout"))
	require.NoErrorf(t, err, "failed to connect new block")

	err = bc.Query("timeout-cnt", `{"Name": "infinite_loop"}`, "exceeded the maximum instruction count")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "timeout-cnt2", 0, code2))
	require.NoErrorf(t, err, "failed to deploy new tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "timeout-cnt2", 0, `{"Name": "a"}`).Fail("contract timeout"))
	require.NoErrorf(t, err, "failed to call tx")
}

func TestSnapshot(t *testing.T) {
	code := readLuaCode("snapshot.lua")
	require.NotEmpty(t, code, "failed to read snapshot.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "snap", 0, code))
	require.NoErrorf(t, err, "failed to deploy contract")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "snap", 0, `{"Name": "inc", "Args":[]}`))
	assert.NoErrorf(t, err, "failed to call contract")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "snap", 0, `{"Name": "inc", "Args":[]}`))
	assert.NoErrorf(t, err, "failed to call contract")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "snap", 0, `{"Name": "inc", "Args":[]}`))
	assert.NoErrorf(t, err, "failed to call contract")

	err = bc.Query("snap", `{"Name":"query"}`, "", "[3,3,3,3]")
	assert.NoErrorf(t, err, "failed to query")

	err = bc.Query("snap", `{"Name":"query", "Args":[2]}`, "", "[1,null,null,null]")
	assert.NoErrorf(t, err, "failed to query")

	err = bc.Query("snap", `{"Name":"query", "Args":[3]}`, "", "[2,2,2,2]")
	assert.NoErrorf(t, err, "failed to query")

	err = bc.Query("snap", `{"Name":"query2", "Args":[]}`, "invalid argument at getsnap, need (state.array, index, blockheight)", "")
	assert.NoErrorf(t, err, "failed to query")
}

func TestKvstore(t *testing.T) {
	code := readLuaCode("kvstore.lua")
	require.NotEmpty(t, code, "failed to read kvstore.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "map", 0, code))
	require.NoErrorf(t, err, "failed to deploy contract")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "map", 0, `{"Name":"inc", "Args":["ktlee"]}`),
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"setname", "Args":["eve2adam"]}`))
	require.NoErrorf(t, err, "failed to call contract")

	err = bc.ConnectBlock()
	require.NoErrorf(t, err, "failed to new block")

	err = bc.Query("map", `{"Name":"get", "Args":["ktlee"]}`, "", "1")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("map", `{"Name":"get", "Args":["htwo"]}`, "", "null")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "map", 0, `{"Name":"inc", "Args":["ktlee"]}`),
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"inc", "Args":["htwo"]}`),
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"set", "Args":["wook", 100]}`))
	require.NoErrorf(t, err, "failed to call contract")

	err = bc.Query("map", `{"Name":"get", "Args":["ktlee"]}`, "", "2")
	assert.NoErrorf(t, err, "failed to query")

	err = bc.Query("map", `{"Name":"get", "Args":["htwo"]}`, "", "1")
	assert.NoErrorf(t, err, "failed to query")

	err = bc.Query("map", `{"Name":"get", "Args":["wook"]}`, "", "100")
	assert.NoErrorf(t, err, "failed to query")

	err = bc.Query("map", `{"Name":"getname"}`, "", `"eve2adam"`)
	assert.NoErrorf(t, err, "failed to query")
}

// sql tests
func TestSqlConstrains(t *testing.T) {
	code := readLuaCode("sql_constrains.lua")
	require.NotEmpty(t, code, "failed to read sql_constrains.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "constraint", 0, code),
		NewLuaTxCall("ktlee", "constraint", 0, `{"Name":"init"}`),
		NewLuaTxCall("ktlee", "constraint", 0, `{"Name":"pkFail"}`).Fail("UNIQUE constraint failed: r.id"),
		NewLuaTxCall("ktlee", "constraint", 0, `{"Name":"checkFail"}`).Fail("CHECK constraint failed: r"),
		NewLuaTxCall("ktlee", "constraint", 0, `{"Name":"fkFail"}`).Fail("FOREIGN KEY constraint failed"),
		NewLuaTxCall("ktlee", "constraint", 0, `{"Name":"notNullFail"}`).Fail("NOT NULL constraint failed: r.nonull"),
		NewLuaTxCall("ktlee", "constraint", 0, `{"Name":"uniqueFail"}`).Fail("UNIQUE constraint failed: r.only"))
	require.NoErrorf(t, err, "failed to call contract")
}

func TestSqlAutoincrement(t *testing.T) {
	code := readLuaCode("sql_autoincrement.lua")
	require.NotEmpty(t, code, "failed to read sql_autoincrement.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "auto", 0, code),
		NewLuaTxCall("ktlee", "auto", 0, `{"Name":"init"}`),
	)
	require.NoErrorf(t, err, "failed to deploy")

	tx := NewLuaTxCall("ktlee", "auto", 0, `{"Name":"get"}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")
}

func TestSqlOnConflict(t *testing.T) {
	code := readLuaCode("sql_onconflict.lua")
	require.NotEmpty(t, code, "failed to read sql_onconflict.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "on_conflict", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert into t values (2)"]}`),
		NewLuaTxCall("ktlee", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert into t values (3),(2),(4)"]}`).Fail(`UNIQUE constraint failed: t.col`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("on_conflict", `{"name":"get"}`, "", `[1,2]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "on_conflict", 0, `{"name":"stmt_exec", "args": ["replace into t values (2)"]}`),
		NewLuaTxCall("ktlee", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert or ignore into t values (3),(2),(4)"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("on_conflict", `{"name":"get"}`, "", `[1,2,3,4]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert into t values (5)"]}`),
		NewLuaTxCall("ktlee", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert or rollback into t values (5)"]}`).Fail("syntax error"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("on_conflict", `{"name":"get"}`, "", `[1,2,3,4,5]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "on_conflict", 0, `{"name":"stmt_exec_pcall", "args": ["insert or fail into t values (6),(7),(5),(8),(9)"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("on_conflict", `{"name":"get"}`, "", `[1,2,3,4,5,6,7]`)
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlDupCol(t *testing.T) {
	code := readLuaCode("sql_dupcol.lua")
	require.NotEmpty(t, code, "failed to read sql_dupcol.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "dup_col", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("dup_col", `{"name":"get"}`, `too many duplicate column name "1+1", max: 5`)
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlVmSimple(t *testing.T) {
	code := readLuaCode("sql_vm_simple.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_simple.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "simple-query", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("simple-query", `{"Name": "query", "Args":[]}`, "", `[2,3.1,"X Hello Blockchain",2,3.1,"Y Hello Blockchain",2,3.1,"Z Hello Blockchain"]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("simple-query", `{"Name": "count", "Args":[]}`, "", `3`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("simple-query", `{"Name": "count", "Args":[]}`, "", `6`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.DisConnectBlock()
	require.NoErrorf(t, err, "failed to disconnect block")

	err = bc.Query("simple-query", `{"Name": "count", "Args":[]}`, "", `3`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.DisConnectBlock()
	require.NoErrorf(t, err, "failed to disconnect block")

	err = bc.DisConnectBlock()
	require.NoErrorf(t, err, "failed to disconnect block")

	// there is only a genesis block
	err = bc.Query("simple-query", `{"Name": "count", "Args":[]}`, "not found contract", "")
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlVmFail(t *testing.T) {
	code := readLuaCode("sql_vm_fail.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_fail.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "fail", 0, code),
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"init"}`),
	)
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[1]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[2]}`),
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"addFail", "Args":[3]}`).Fail(`near "set": syntax error`),
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[4]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[5]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("fail", `{"Name":"get"}`, "", "12")
	require.NoErrorf(t, err, "failed to query")

	err = bc.DisConnectBlock()
	require.NoErrorf(t, err, "failed to disconnect block")

	err = bc.Query("fail", `{"Name":"get"}`, "", "7")
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlVmPubNet(t *testing.T) {
	code := readLuaCode("sql_vm_pubnet.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_pubnet.lua")

	bc, err := LoadDummyChain(SetPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "simple-query", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`).Fail(`attempt to index global 'db'`))
	require.NoErrorf(t, err, "failed to call tx")
}

func TestSqlVmDateTime(t *testing.T) {
	code := readLuaCode("sql_vm_datetime.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_datetime.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "datetime", 0, code),
		NewLuaTxCall("ktlee", "datetime", 0, `{"Name":"init"}`))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "datetime", 0, `{"Name":"nowNull"}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "datetime", 0, `{"Name":"localtimeNull"}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("datetime", `{"Name":"get"}`, "", `[{"bool":0},{"bool":1},{"bool":1,"date":"1970-01-01 02:46:40"},{"bool":0,"date":"2004-11-23"}]`)
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlVmCustomer(t *testing.T) {
	code := readLuaCode("sql_vm_customer.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_customer.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "customer", 0, code),
		NewLuaTxCall("ktlee", "customer", 0, `{"Name":"createTable"}`))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "customer", 0, `{"Name":"insert", "Args":["id1","passwd1","name1","20180524","010-1234-5678"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "customer", 0, `{"Name":"insert", "Args":["id2","passwd2","name2","20180524","010-1234-5678"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "customer", 0, `{"Name":"update", "Args":["id2","passwd3"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("customer", `{"Name":"count"}`, "", "2")
	require.NoErrorf(t, err, "failed to query")

	err = bc.DisConnectBlock()
	require.NoErrorf(t, err, "failed to disconnect block")

	err = bc.Query("customer", `{"Name":"query", "Args":["id2"]}`, "", `[{"birth":"20180524","id":"id2","mobile":"010-1234-5678","name":"name2","passwd":"passwd2"}]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "customer", 0, `{"Name":"delete", "Args":["id2"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("customer", `{"Name":"query", "Args":["id2"]}`, "", `{}`)
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlVmDataType(t *testing.T) {
	code := readLuaCode("sql_vm_datatype.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_datatype.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "datatype", 0, code),
		NewLuaTxCall("ktlee", "datatype", 0, `{"Name":"createDataTypeTable"}`))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "datatype", 0, `{"Name":"insertDataTypeTable"}`),
		NewLuaTxCall("ktlee", "datatype", 0, `{"Name":"insertDataTypeTable"}`),
		NewLuaTxCall("ktlee", "datatype", 0, `{"Name":"insertDataTypeTable"}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "datatype", 0, `{"Name":"insertDataTypeTable"}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("datatype", `{"Name":"queryOrderByDesc"}`, "", `[{"blockheight1":3,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"}]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("datatype", `{"Name":"queryGroupByBlockheight1"}`, "", `[{"avg_float1":3.14,"blockheight1":2,"count1":3,"sum_int1":3},{"avg_float1":3.14,"blockheight1":3,"count1":1,"sum_int1":1}]`)
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlVmFunction(t *testing.T) {
	code := readLuaCode("sql_vm_function.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_function.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "fns", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("fns", `{"Name":"sql_func"}`, "", `[3,1,6]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("fns", `{"Name":"abs_func"}`, "", `[1,0,1]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("fns", `{"Name":"typeof_func"}`, "", `["integer","text","real","null"]`)
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlVmBook(t *testing.T) {
	code := readLuaCode("sql_vm_book.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_book.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "book", 0, code),
		NewLuaTxCall("ktlee", "book", 0, `{"Name":"createTable"}`))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "book", 0, `{"Name":"makeBook"}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "book", 0, `{"Name":"copyBook"}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("book", `{"Name":"viewCopyBook"}`, "", `[100,"value=1"]`)
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlVmDateformat(t *testing.T) {
	code := readLuaCode("sql_vm_dateformat.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_dateformat.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "data_format", 0, code),
		NewLuaTxCall("ktlee", "data_format", 0, `{"Name":"init"}`))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("data_format", `{"Name":"get"}`, "", `[["2004-10-24","2004-10-24 11:11:11","20041024111111"],["2018-05-28","2018-05-28 10:45:38","20180528104538"]]`)
	require.NoErrorf(t, err, "failed to query")
}

func TestSqlVmRecursiveData(t *testing.T) {
	code := readLuaCode("sql_vm_recursivedata.lua")
	require.NotEmpty(t, code, "failed to read sql_vm_recursivedata.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	tx := NewLuaTxCall("ktlee", "r", 0, `{"Name":"r"}`)
	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "r", 0, code), tx)
	require.Errorf(t, err, "expect err")
	require.Equalf(t, "nested table error", err.Error(), "expect err")
}

func TestSqlJdbc(t *testing.T) {
	code := readLuaCode("sql_jdbc.lua")
	require.NotEmpty(t, code, "failed to read sql_jdbc.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "jdbc", 0, code),
		NewLuaTxCall("ktlee", "jdbc", 0, `{"Name":"init"}`),
	)
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("jdbc", `{"Name":"query", "Args":["select a,b,c from total"]}`, "",
		`{"colcnt":3,"colmetas":{"colcnt":3,"decltypes":["int","int","text"],"names":["a","b","c"]},"data":[[1,{},"2"],[2,2,"3"],[3,2,"3"],[4,2,"3"],[5,2,"3"],[6,2,"3"],[7,2,"3"]],"rowcnt":7,"snap":"2"}`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("jdbc", `{"Name":"getmeta", "Args":["select a,b,?+1 from total"]}`, "",
		`[{"colcnt":3,"decltypes":["int","int",""],"names":["a","b","?+1"]},1]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "jdbc", 0, `{"Name": "exec", "Args":["insert into total values (3,4,5)"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("jdbc", `{"Name":"query", "Args":["select a,b,c from total"]}`, "",
		`{"colcnt":3,"colmetas":{"colcnt":3,"decltypes":["int","int","text"],"names":["a","b","c"]},"data":[[1,{},"2"],[2,2,"3"],[3,2,"3"],[4,2,"3"],[5,2,"3"],[6,2,"3"],[7,2,"3"],[3,4,"5"]],"rowcnt":8,"snap":"3"}`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("jdbc", `{"Name":"queryS", "Args":["2", "select a,b,c from total"]}`, "",
		`{"colcnt":3,"colmetas":{"colcnt":3,"decltypes":["int","int","text"],"names":["a","b","c"]},"data":[[1,{},"2"],[2,2,"3"],[3,2,"3"],[4,2,"3"],[5,2,"3"],[6,2,"3"],[7,2,"3"]],"rowcnt":7,"snap":"3"}`)
	require.NoErrorf(t, err, "failed to query")
}

// type Tests
func TestTypeOP(t *testing.T) {
	code := readLuaCode("op.lua")
	require.NotEmpty(t, code, "failed to read op.lua")

	bc, err := LoadDummyChain(SetPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	balance, _ := new(big.Int).SetString("10000000000000000", 10)
	err = bc.ConnectBlock(NewLuaTxAccountBig("ktlee", balance), NewLuaTxDeploy("ktlee", "op", 0, string(code)))
	require.NoErrorf(t, err, "failed to deploy")

	state, err := bc.GetAccountState("ktlee")
	require.NoErrorf(t, err, "failed to get account state")

	var before, use, after uint64

	before = state.GetBalanceBigInt().Uint64()
	tx := NewLuaTxCall("ktlee", "op", 0, `{"Name": "main"}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	r := bc.GetReceipt(tx.Hash())
	state, err = bc.GetAccountState("ktlee")
	require.NoErrorf(t, err, "failed to get account state")
	after = state.GetBalanceBigInt().Uint64()

	use = 117610
	require.Equalf(t, use, r.GetGasUsed(), "used gas not equal")
	require.Equalf(t, before-use, after, "balance not equal")
}

func TestTypeBF(t *testing.T) {
	code := readLuaCode("bf.lua")
	require.NotEmpty(t, code, "failed to read bf.lua")

	bc, err := LoadDummyChain(SetPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	balance, _ := new(big.Int).SetString("10000000000000000", 10)
	err = bc.ConnectBlock(NewLuaTxAccountBig("ktlee", balance), NewLuaTxDeploy("ktlee", "op", 0, string(code)))
	require.NoErrorf(t, err, "failed to deploy")

	state, err := bc.GetAccountState("ktlee")
	require.NoErrorf(t, err, "failed to get account state")

	feeTest := func(expectedFee uint64) {
		var before, after uint64

		before = state.GetBalanceBigInt().Uint64()
		tx := NewLuaTxCall("ktlee", "op", 0, `{"Name": "main"}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		r := bc.GetReceipt(tx.Hash())
		state, err = bc.GetAccountState("ktlee")
		require.NoErrorf(t, err, "failed to get account state")
		after = state.GetBalanceBigInt().Uint64()

		require.Equalf(t, expectedFee, r.GetGasUsed(), "used gas not equal")
		require.Equalf(t, before-expectedFee, after, "balance not equal")
	}

	feeTest(47456244)
	OldVersion := bc.HardforkVersion
	bc.HardforkVersion = 3
	feeTest(47513803)
	bc.HardforkVersion = OldVersion
}

func TestTypeMaxString(t *testing.T) {
	code := readLuaCode("type_maxstring.lua")
	require.NotEmpty(t, code, "failed to read type_maxstring.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "oom", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	errMsg := "not enough memory"
	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "oom", 0, `{"Name":"oom"}`).Fail(errMsg))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "oom", 0, `{"Name":"p"}`).Fail(errMsg))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "oom", 0, `{"Name":"cp"}`).Fail(errMsg))
	require.NoErrorf(t, err, "failed to call tx")
}

func TestTypeMaxStringOnPubNet(t *testing.T) {
	code := readLuaCode("type_maxstring.lua")
	require.NotEmpty(t, code, "failed to read type_maxstring.lua")

	bc, err := LoadDummyChain(SetPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "oom", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	errMsg := "string length overflow"
	errMsg1 := "not enough memory"
	var travis bool
	if os.Getenv("TRAVIS") == "true" {
		travis = true
	}
	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "oom", 0, `{"Name":"oom"}`))
	require.Errorf(t, err, "expected: %s", errMsg)
	if !strings.Contains(err.Error(), errMsg) && !strings.Contains(err.Error(), errMsg1) {
		t.Error(err)
	}
	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "oom", 0, `{"Name":"p"}`))
	if err != nil && (!travis || !strings.Contains(err.Error(), errMsg1)) {
		t.Error(err)
	}
	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "oom", 0, `{"Name":"cp"}`))
	if err != nil && (!travis || !strings.Contains(err.Error(), errMsg1)) {
		t.Error(err)
	}
}

func TestTypeNsec(t *testing.T) {
	code := readLuaCode("type_nsec.lua")
	require.NotEmpty(t, code, "failed to read type_nsec.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "nsec", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "nsec", 0, `{"Name": "test_nsec"}`).Fail(`attempt to call global 'nsec' (a nil value)`))
	require.NoErrorf(t, err, "failed to call tx")
}

func TestTypeUtf(t *testing.T) {
	code := readLuaCode("type_utf.lua")
	require.NotEmpty(t, code, "failed to read type_utf.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "utf", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("utf", `{"Name":"query"}`, "", "")
	assert.NoErrorf(t, err, "failed to query")

	err = bc.Query("utf", `{"Name":"query2"}`, "", `["E8D4A51000","00"]`)
	assert.NoErrorf(t, err, "failed to query")

	err = bc.Query("utf", `{"Name":"query3"}`, "bignum not allowed negative value", "")
	assert.NoErrorf(t, err, "failed to query")
}

func TestTypeDupVar(t *testing.T) {
	code := readLuaCode("type_dupvar_1.lua")
	require.NotEmpty(t, code, "failed to read type_dupvar_1.lua")
	code2 := readLuaCode("type_dupvar_2.lua")
	require.NotEmpty(t, code, "failed to read type_dupvar_2.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 1000000000000000000))
	require.NoErrorf(t, err, "failed to new tx")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "dupVar", 0, code))
	require.Errorf(t, err, "error expect | duplicated variable: 'Var1'")
	if !strings.Contains(err.Error(), "duplicated variable: 'Var1'") {
		t.Error(err)
	}

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "dupVar1", 0, code2))
	require.NoErrorf(t, err, "failed to deploy")
	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "dupVar1", 0, `{"Name": "Work"}`).Fail("duplicated variable: 'Var1'"))
	require.NoErrorf(t, err, "failed to call tx")
}

func TestTypeInvalidKey(t *testing.T) {
	code := readLuaCode("type_invalidkey.lua")
	require.NotEmpty(t, code, "failed to read type_invalidkey.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "invalidkey", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_table"}`).Fail("cannot use 'table' as a key"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_func"}`).Fail("cannot use 'function' as a key"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_statemap"}`).Fail("cannot use 'userdata' as a key"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_statearray"}`).Fail("cannot use 'userdata' as a key"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_statevalue"}`).Fail("cannot use 'userdata' as a key"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_upval"}`).Fail("cannot use 'table' as a key"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_nil"}`).Fail("invalid key type: 'nil', state.map: 'h'"))
	require.NoErrorf(t, err, "failed to call tx")
}

func TestTypeByteKey(t *testing.T) {
	code := readLuaCode("type_bytekey.lua")
	require.NotEmpty(t, code, "failed to read type_bytekey.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "bk", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("bk", `{"Name":"get"}`, "", `["kk","kk"]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("bk", `{"Name":"getcre"}`, "", `"Amg6nZWXKB6YpNgBPv9atcjdm6hnFvs5wMdRgb2e9DmaF5g9muF2"`)
	require.NoErrorf(t, err, "failed to query")
}

func TestTypeArray(t *testing.T) {
	code := readLuaCode("type_array.lua")
	require.NotEmpty(t, code, "failed to read type_array.lua")

	code2 := readLuaCode("type_array_overflow.lua")
	require.NotEmpty(t, code, "failed to read type_array_overflow.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "array", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[1]}`),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[0]}`).Fail("index out of range"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[1]}`),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[1.00000001]}`).Fail("integer expected, got number"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":["1"]}`).Fail("integer expected, got string)"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[true]}`).Fail("integer expected, got boolean"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[[1, 2]]}`).Fail("integer expected, got table"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[null]}`).Fail("integer expected, got nil)"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[{}]}`).Fail("integer expected, got table)"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[""]}`).Fail("integer expected, got string)"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"set", "Args":[2,"ktlee"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("array", `{"Name":"get", "Args":[11]}`, "index out of range", "")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("array", `{"Name":"get", "Args":[1]}`, "", "2")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("array", `{"Name":"get", "Args":[2]}`, "", `"ktlee"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("array", `{"Name":"len"}`, "", `10`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("array", `{"Name":"iter"}`, "", `[2,"ktlee","nil","nil","nil","nil","nil","nil","nil","nil"]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "overflow", 0, code2))
	errMsg := "integer expected, got number"
	require.Errorf(t, err, "expect no error")
	require.Containsf(t, err.Error(), errMsg, "err not match")
}

func TestTypeMultiArray(t *testing.T) {
	code := readLuaCode("type_multiarray_1.lua")
	require.NotEmpty(t, code, "failed to read type_multiarray_1.lua")

	code2 := readLuaCode("type_multiarray_2.lua")
	require.NotEmpty(t, code, "failed to read type_multiarray_2.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "ma", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "ma", 0, `{"Name": "inc", "Args":[]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "ma", 0, `{"Name": "inc", "Args":[]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("ma", fmt.Sprintf(`{"Name":"query", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("ktlee"))), "", "[2,2,2,null,10,11]")
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "ma", 0, `{"Name": "del", "Args":[]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("ma", fmt.Sprintf(`{"Name":"query", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("ktlee"))), "", "[2,2,null,null,10,11]")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("ma", `{"Name":"iter"}`, "", `{"1,10":"k","10,5":"l"}`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("ma", `{"Name":"seterror"}`, "", ``)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "ma", 0, code2))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("ma", `{"Name":"query", "Args":[]}`, "", `["A","B",null,null,"A","B","v1"]`)
	require.NoErrorf(t, err, "failed to query")

	tx := NewLuaTxCall("ktlee", "ma", 0, `{"Name": "abc", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	receipt := bc.GetReceipt(tx.Hash())
	require.Equalf(t, `["C","D","A","B","v3"]`, receipt.GetRet(), "contract Call ret error")

	err = bc.Query("ma", `{"Name":"query", "Args":[]}`, "", `["A","B","C","D","A","B","v3"]`)
	require.NoErrorf(t, err, "failed to query")
}

func TestTypeArrayArg(t *testing.T) {
	code := readLuaCode("type_arrayarg.lua")
	require.NotEmpty(t, code, "failed to read type_arrayarg.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "a", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("a", `{"Name": "copy", "Args":[1, 2, 3]}`, "table expected", "")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("a", `{"Name": "copy", "Args":[[1, 2, 3]]}`, "", "[1,2,3]")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("a", `{"Name": "two_arr", "Args":[[1, 2, 3],[4, 5]]}`, "", "[3,2]")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("a", `{"Name": "mixed_args", "Args":[[1, 2, 3], {"name": "kslee", "age": 39}, 7]}`, "", `[[1,2,3],{"age":39,"name":"kslee"},7]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("a", `{"Name": "mixed_args", "Args":[
[[1, 2, 3],["first", "second"]],
{"name": "kslee", "age": 39, "address": {"state": "XXX-do", "city": "YYY-si"}},
"end"
]}`, "", `[[[1,2,3],["first","second"]],{"address":{"city":"YYY-si","state":"XXX-do"},"age":39,"name":"kslee"},"end"]`,
	)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("a", `{"Name": "mixed_args", "Args":[
[{"name": "wook", "age": 50}, {"name": "hook", "age": 42}],
{"name": "kslee", "age": 39, "scores": [10, 20, 30, 40, 50]},
"hmm..."
]}`, "", `[[{"age":50,"name":"wook"},{"age":42,"name":"hook"}],{"age":39,"name":"kslee","scores":[10,20,30,40,50]},"hmm..."]`,
	)
	require.NoErrorf(t, err, "failed to query")
}

func TestTypeMapKey(t *testing.T) {
	code := readLuaCode("type_mapkey.lua")
	require.NotEmpty(t, code, "failed to read type_mapkey.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "a", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("a", `{"Name":"getCount", "Args":[1]}`, "", "null")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":[1, 10]}`),
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":["1", 20]}`).Fail("(number expected, got string)"),
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":[1.1, 30]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("a", `{"Name":"getCount", "Args":["1"]}`, "(number expected, got string)", "")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("a", `{"Name":"getCount", "Args":[1]}`, "", "10")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("a", `{"Name":"getCount", "Args":[1.1]}`, "", "30")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":[true, 40]}`).Fail(`invalid key type: 'boolean', state.map: 'counts'`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "a", 0, `{"Name":"delCount", "Args":[1.1]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("a", `{"Name":"getCount", "Args":[1.1]}`, "", "null")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("a", `{"Name":"getCount", "Args":[2]}`, "", "null")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "x", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":["1", 10]}`),
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":[1, 20]}`).Fail("string expected, got number)"),
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":["third", 30]}`),
	)
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("x", `{"Name":"getCount", "Args":["1"]}`, "", "10")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("x", `{"Name":"getCount", "Args":["third"]}`, "", "30")
	require.NoErrorf(t, err, "failed to query")
}

func TestTypeStateVarFieldUpdate(t *testing.T) {
	code := readLuaCode("type_statevarfieldupdate.lua")
	require.NotEmpty(t, code, "failed to read type_statevarfieldupdate.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "c", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "c", 0, `{"Name":"InvalidUpdateAge", "Args":[10]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("c", `{"Name":"GetPerson"}`, "", `{"address":"blahblah...","age":38,"name":"kslee"}`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "c", 0, `{"Name":"ValidUpdateAge", "Args":[10]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("c", `{"Name":"GetPerson"}`, "", `{"address":"blahblah...","age":10,"name":"kslee"}`)
	require.NoErrorf(t, err, "failed to query")
}

func TestTypeDatetime(t *testing.T) {
	code := readLuaCode("type_datetime.lua")
	require.NotEmpty(t, code, "failed to read type_datetime.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "datetime", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("datetime", `{"Name": "CreateDate"}`, "", `"1998-09-17 02:48:10"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%x"]}`, "", `"09/17/98"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%X"]}`, "", `"02:48:10"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%A"]}`, "", `"Thursday"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%I:%M:%S %p"]}`, "", `"02:48:10 AM"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("datetime", `{"Name": "Difftime"}`, "", `2890`)
	require.NoErrorf(t, err, "failed to query")
}

func TestTypeDynamicArray(t *testing.T) {
	code := readLuaCode("type_dynamicarray_zerolen.lua")
	require.NotEmpty(t, code, "failed to read type_dynamicarray_zerolen.lua")

	code2 := readLuaCode("type_dynamicarray.lua")
	require.NotEmpty(t, code, "failed to read type_dynamicarray.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000))
	require.NoErrorf(t, err, "failed to new account")
	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "zeroLen", 0, code))
	require.Errorf(t, err, "no error | expected: the array length must be greater than zero")
	require.Containsf(t, err.Error(), "the array length must be greater than zero", "wrong error message")

	tx := NewLuaTxDeploy("ktlee", "dArr", 0, code2)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("dArr", `{"Name": "Length"}`, "", "0")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [10]}`),
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [20]}`),
	)
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("dArr", `{"Name": "Get", "Args": [1]}`, "", "10")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("dArr", `{"Name": "Get", "Args": [2]}`, "", "20")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("dArr", `{"Name": "Get", "Args": [3]}`, "index out of range", "")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("dArr", `{"Name": "Length"}`, "", "2")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [30]}`),
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [40]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("dArr", `{"Name": "Length"}`, "", "4")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Set", "Args": [3, 50]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("dArr", `{"Name": "Get", "Args": [3]}`, "", "50")
	require.NoErrorf(t, err, "failed to query")
}

func TestTypeCrypto(t *testing.T) {
	code := readLuaCode("type_crypto.lua")
	require.NotEmpty(t, code, "failed to read type_crypto.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "crypto", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("crypto", `{"Name": "get", "Args" : ["ab\u0000\u442a"]}`, "", `"0xc58f6dca13e4bba90a326d8605042862fe87c63a64a9dd0e95608a2ee68dc6f0"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("crypto", `{"Name": "get", "Args" : ["0x616200e490aa"]}`, "", `"0xc58f6dca13e4bba90a326d8605042862fe87c63a64a9dd0e95608a2ee68dc6f0"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("crypto", `{"Name": "checkEther", "Args" : []}`, "", `true`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("crypto", `{"Name": "checkAergo", "Args" : []}`, "", `true`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("crypto", `{"Name": "keccak256", "Args" : ["0x616263"]}`, "", `"0x4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("crypto", `{"Name": "keccak256", "Args" : ["0x616572676F"]}`, "", `"0xe98bb03ab37161f8bbfe131f711dcccf3002a9cd9ec31bbd52edf181f7ab09a0"`)
	require.NoErrorf(t, err, "failed to query")
}

func TestTypeBignum(t *testing.T) {
	bignum := readLuaCode("type_bignum.lua")
	require.NotEmpty(t, bignum, "failed to read type_bignum.lua")
	callee := readLuaCode("type_bignum_callee.lua")
	require.NotEmpty(t, bignum, "failed to read type_bignum_callee.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "bigNum", 50000000000, bignum),
		NewLuaTxDeploy("ktlee", "add", 0, callee))
	require.NoErrorf(t, err, "failed to deploy")

	tx := NewLuaTxCall("ktlee", "bigNum", 0, fmt.Sprintf(`{"Name":"test", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("ktlee"))))
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	receipt := bc.GetReceipt(tx.Hash())
	assert.Equalf(t, `"25000000000"`, receipt.GetRet(), "contract Call ret error")

	tx = NewLuaTxCall("ktlee", "bigNum", 0, fmt.Sprintf(`{"Name":"sendS", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("ktlee"))))
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	receipt = bc.GetReceipt(tx.Hash())
	assert.Equalf(t, `"23999900001"`, receipt.GetRet(), "contract Call ret error")

	tx = NewLuaTxCall("ktlee", "bigNum", 0, `{"Name":"testBignum", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	receipt = bc.GetReceipt(tx.Hash())
	assert.Equalf(t, `"999999999999999999999999999999"`, receipt.GetRet(), "contract Call ret error")

	err = bc.Query("bigNum", `{"Name":"argBignum", "Args":[{"_bignum":"99999999999999999999999999"}]}`, "", `"100000000000000000000000000"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("bigNum", fmt.Sprintf(`{"Name":"calladdBignum", "Args":["%s", {"_bignum":"999999999999999999"}]}`, types.EncodeAddress(contract.StrHash("add"))), "", `"1000000000000000004"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("bigNum", `{"Name":"checkBignum"}`, "", `[false,true,false]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("bigNum", `{"Name":"calcBignum"}`, "bignum divide by zero", "")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("bigNum", `{"Name":"negativeBignum"}`, "bignum not allowed negative value", "")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("bigNum", `{"Name":"byteBignum"}`, "", `{"_bignum":"177"}`)
	require.NoErrorf(t, err, "failed to query")
}

func checkRandomIntValue(v string, min, max int) error {
	n, _ := strconv.Atoi(v)
	if n < min || n > max {
		return errors.New("out of range")
	}
	return nil
}

func TestTypeRandom(t *testing.T) {
	code := readLuaCode("type_random.lua")
	require.NotEmpty(t, code, "failed to read type_random.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "random", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[]}`).Fail("1 or 2 arguments required"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[0]}`).Fail("the maximum value must be greater than zero"))
	require.NoErrorf(t, err, "failed to call tx")

	tx := NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[3]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	receipt := bc.GetReceipt(tx.Hash())
	err = checkRandomIntValue(receipt.GetRet(), 1, 3)
	require.NoErrorf(t, err, "failed to check random value")

	tx = NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[3, 10]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	receipt = bc.GetReceipt(tx.Hash())
	err = checkRandomIntValue(receipt.GetRet(), 3, 10)
	require.NoErrorf(t, err, "failed to check random value")

	err = bc.Query("random", `{"Name": "random", "Args":[1]}`, "", "1")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("random", `{"Name": "random", "Args":[4,4]}`, "", "4")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("random", `{"Name": "random", "Args":[0,4]}`, "system.random: the minimum value must be greater than zero", "")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("random", `{"Name": "random", "Args":[3,1]}`, "system.random: the maximum value must be greater than the minimum value", "")
	require.NoErrorf(t, err, "failed to query")
}

func TestTypeSparseTable(t *testing.T) {
	code := readLuaCode("type_sparsetable.lua")
	require.NotEmpty(t, code, "failed to read type_sparsetable.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	tx := NewLuaTxCall("ktlee", "r", 0, `{"Name":"r"}`)
	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "r", 0, code), tx)
	require.NoErrorf(t, err, "failed to new account, deploy, call")

	receipt := bc.GetReceipt(tx.Hash())
	require.Equalf(t, `1`, receipt.GetRet(), "contract Call ret error")
}

func TestTypeBigTable(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	code := readLuaCode("type_bigtable_1.lua")
	require.NotEmpty(t, code, "failed to read type_bigtable_1.lua")
	code2 := readLuaCode("type_bigtable_2.lua")
	require.NotEmpty(t, code2, "failed to read type_bigtable_2.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "big", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	// About 900MB
	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "big", 0, `{"Name": "inserts", "Args":[25]}`))
	require.NoErrorf(t, err, "failed to call tx")

	contract.SetStateSQLMaxDBSize(20)
	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100), NewLuaTxDeploy("ktlee", "big20", 0, code2))
	require.NoErrorf(t, err, "failed to deploy")

	for i := 0; i < 17; i++ {
		err = bc.ConnectBlock(NewLuaTxCall("ktlee", "big20", 0, `{"Name": "inserts"}`))
		require.NoErrorf(t, err, "failed to call tx")
	}
	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "big20", 0, `{"Name": "inserts"}`).Fail("database or disk is full"))
	require.NoErrorf(t, err, "failed to call tx")
}

func TestTypeJson(t *testing.T) {
	code := readLuaCode("type_json.lua")
	require.NotEmpty(t, code, "failed to read type_json.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "json", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["[1,2,3]"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", "[1,2,3]")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"[1,2,3]"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["{\"key1\":[1,2,3], \"run\", \"key2\":5, [4,5,6]}"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"1":"run","2":[4,5,6],"key1":[1,2,3],"key2":5}`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"{\"1\":\"run\",\"2\":[4,5,6],\"key1\":[1,2,3],\"key2\":5}"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["{\"key1\":{\"arg1\": 1,\"arg2\":null, \"arg3\":[]}, \"key2\":[5,4,3]}"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"key1":{"arg1":1,"arg3":{}},"key2":[5,4,3]}`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"{\"key1\":{\"arg1\":1,\"arg3\":{}},\"key2\":[5,4,3]}"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["{\"key1\":[1,2,3], \"key1\":5}"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"key1":5}`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["[\"\\\"hh\\t\",\"2\",3]"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `["\"hh\u0009","2",3]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("json", `{"Name":"getlen", "Args":[]}`, "", `["\"hh\u0009",4]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"[\"\\\"hh\\u0009\",\"2\",3]"`)
	require.NoErrorf(t, err, "failed to query")

	tx := NewLuaTxCall("ktlee", "json", 100, `{"Name":"getAmount"}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")
	receipt := bc.GetReceipt(tx.Hash())
	require.Equalf(t, `"100"`, receipt.GetRet(), "contract Call ret error")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["{\"key1\":[1,2,3], \"key1\":5}}"]}`).Fail("not proper json format"))
	require.NoErrorf(t, err, "failed to call tx")
}

// feature tests
func TestFeatureVote(t *testing.T) {
	code := readLuaCode("feature_vote.lua")
	require.NotEmpty(t, code, "failed to read feature_vote.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("owner", 100000000000000000),
		NewLuaTxDeploy("owner", "vote", 0, code),
		NewLuaTxAccount("user1", 100000000000000000),
		NewLuaTxAccount("user10", 100000000000000000),
		NewLuaTxAccount("user11", 100000000000000000))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("owner", "vote", 0, `{"Name":"addCandidate", "Args":["candidate1"]}`),
		NewLuaTxCall("owner", "vote", 0, `{"Name":"addCandidate", "Args":["candidate2"]}`),
		NewLuaTxCall("owner", "vote", 0, `{"Name":"addCandidate", "Args":["candidate3"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("vote", `{"Name":"getCandidates"}`, "", `[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("user1", "vote", 0, `{"Name":"addCandidate", "Args":["candidate4"]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("vote", `{"Name":"getCandidates"}`, "", `[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(
		// register voter
		NewLuaTxCall("owner", "vote", 0, fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("user10")))),
		NewLuaTxCall("owner", "vote", 0, fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("user10")))),
		NewLuaTxCall("owner", "vote", 0, fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("user11")))),
		NewLuaTxCall("owner", "vote", 0, fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("user1")))),
		// vote
		NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user1"]}`),
		NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user1"]}`),
		NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user2"]}`),
		NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user2"]}`),
		NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user3"]}`),
	)
	require.NoErrorf(t, err, "failed to call tx | vote error")

	err = bc.Query("vote", `{"Name":"getCandidates"}`, "", `[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxCall("user11", "vote", 0, `{"Name":"vote", "Args":["candidate1"]}`),
		NewLuaTxCall("user10", "vote", 0, `{"Name":"vote", "Args":["candidate1"]}`))
	require.NoErrorf(t, err, "failed to call tx | vote error")

	err = bc.Query("vote", `{"Name":"getCandidates"}`, "", `[{"count":"2","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`)
	require.NoErrorf(t, err, "failed to query")
}

func TestFeatureGovernance(t *testing.T) {
	code := readLuaCode("feature_governance.lua")
	require.NotEmpty(t, code, "failed to read feature_governance.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "gov", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	amount, _ := new(big.Int).SetString("40000000000000000000000", 10)
	err = bc.ConnectBlock(NewLuaTxCallBig("ktlee", "gov", amount, `{"Name": "test_gov", "Args":[]}`))
	require.NoErrorf(t, err, "failed to call tx")

	oldstaking, err := bc.GetStaking("gov")
	require.NoErrorf(t, err, "failed to get staking")

	oldgov, err := bc.GetAccountState("gov")
	require.NoErrorf(t, err, "failed to get gov account state")

	tx := NewLuaTxCall("ktlee", "gov", 0, `{"Name": "test_pcall", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	staking, err := bc.GetStaking("gov")
	require.NoErrorf(t, err, "failed to get staking")

	gov, err := bc.GetAccountState("gov")
	require.NoErrorf(t, err, "failed to get gov account state")
	require.Equalf(t, oldstaking.Amount, staking.Amount, "pcall error, staking amount should be same")
	require.Equalf(t, oldgov.GetBalance(), gov.GetBalance(), "pcall error, gov balance should be same")

	tx = NewLuaTxCall("ktlee", "gov", 0, `{"Name": "error_case", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	require.Errorf(t, err, "expect error | less time has passed")

	newstaking, err := bc.GetStaking("gov")
	require.NoErrorf(t, err, "failed to get staking")

	newgov, err := bc.GetAccountState("gov")
	require.NoErrorf(t, err, "failed to get gov account state")

	require.Equalf(t, oldstaking.Amount, newstaking.Amount, "pcall error, staking amount should be same")
	require.Equalf(t, oldgov.GetBalance(), newgov.GetBalance(), "pcall error, gov balance should be same")
}

func TestFeaturePcallRollback(t *testing.T) {
	code := readLuaCode("feature_pcallrollback_1.lua")
	require.NotEmpty(t, code, "failed to read feature_pcallrollback_1.lua")
	code2 := readLuaCode("feature_pcallrollback_2.lua")
	require.NotEmpty(t, code, "failed to read feature_pcallrollback_2.lua")
	code3 := readLuaCode("feature_pcallrollback_3.lua")
	require.NotEmpty(t, code, "failed to read feature_pcallrollback_3.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDeploy("ktlee", "counter", 10, code).Constructor("[0]"),
		NewLuaTxCall("ktlee", "counter", 15, `{"Name":"inc", "Args":[]}`))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "1")
	require.NoErrorf(t, err, "failed to query")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "caller", 10, code2).Constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(contract.StrHash("counter")))),
		NewLuaTxCall("ktlee", "caller", 15, `{"Name":"add", "Args":[]}`))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "caller", 0, `{"Name":"sql", "Args":[]}`))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "2")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("caller", `{"Name":"sqlget", "Args":[]}`, "", "2")
	require.NoErrorf(t, err, "failed to query")

	tx := NewLuaTxCall("ktlee", "caller", 0, `{"Name":"getOrigin", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	receipt := bc.GetReceipt(tx.Hash())
	require.Equalf(t, "\""+types.EncodeAddress(contract.StrHash("ktlee"))+"\"", receipt.GetRet(), "contract Call ret error")

	bc, err = LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxAccount("bong", 0),
		NewLuaTxDeploy("ktlee", "counter", 0, code3))
	require.NoErrorf(t, err, "failed to deploy")

	tx = NewLuaTxCall("ktlee", "counter", 20, fmt.Sprintf(`{"Name":"set", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("bong"))))
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "1")
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("counter", `{"Name":"getBalance", "Args":[]}`, "", "\"18\"")
	require.NoErrorf(t, err, "failed to query")

	state, err := bc.GetAccountState("bong")
	require.NoErrorf(t, err, "failed to get account state")
	assert.Equal(t, uint64(2), state.GetBalanceBigInt().Uint64(), "balance error")

	tx = NewLuaTxCall("ktlee", "counter", 10, fmt.Sprintf(`{"Name":"set2", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("bong"))))
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "2")
	require.NoErrorf(t, err, "failed to query")

	state, err = bc.GetAccountState("bong")
	require.NoErrorf(t, err, "failed to get account state")
	assert.Equal(t, uint64(3), state.GetBalanceBigInt().Uint64(), "balance error")
}

func TestFeaturePcallNested(t *testing.T) {
	code := readLuaCode("feature_pcallnested.lua")
	require.NotEmpty(t, code, "failed to read feature_pcallnested.lua")

	bc, err := LoadDummyChain()
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxAccount("bong", 0),
		NewLuaTxDeploy("ktlee", "pcall", 10000000000000000000, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCall("ktlee", "pcall", 0, fmt.Sprintf(`{"Name":"pcall1", "Args":["%s", "%s"]}`,
		types.EncodeAddress(contract.StrHash("pcall")), types.EncodeAddress(contract.StrHash("bong")))))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.Query("pcall", fmt.Sprintf(`{"Name":"map", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("pcall"))), "", "2")
	require.NoErrorf(t, err, "failed to query")

	state, err := bc.GetAccountState("bong")
	require.NoErrorf(t, err, "failed to get account state")
	assert.Equal(t, uint64(1000000000000000000), state.GetBalanceBigInt().Uint64(), "balance error")
}

func TestFeatureLuaCryptoVerifyProof(t *testing.T) {
	code := readLuaCode("feature_luacryptoverifyproof.lua")
	require.NotEmpty(t, code, "failed to read feature_luacryptoverifyproof.lua")

	bc, err := LoadDummyChain(SetPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount("ktlee", 100000000000000000), NewLuaTxDeploy("ktlee", "eth", 0, code))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.Query("eth", `{"Name":"verifyProofRaw"}`, "", `true`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("eth", `{"Name":"verifyProofHex"}`, "", `true`)
	require.NoErrorf(t, err, "failed to query")

	state, err := bc.GetAccountState("ktlee")
	require.NoErrorf(t, err, "failed to get account state")

	var before, use, after uint64

	// verify proof raw
	before = state.GetBalanceBigInt().Uint64()
	tx := NewLuaTxCall("ktlee", "eth", 0, `{"Name": "verifyProofRaw"}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")
	r := bc.GetReceipt(tx.Hash())
	state, err = bc.GetAccountState("ktlee")
	require.NoErrorf(t, err, "failed to get account state")
	after = state.GetBalanceBigInt().Uint64()

	use = 154137
	require.Equalf(t, use, r.GetGasUsed(), "not equal gas used")
	require.Equalf(t, before-use, after, "not equal balance")

	// verify proof hex
	before = after
	tx = NewLuaTxCall("ktlee", "eth", 0, `{"Name": "verifyProofHex"}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")
	r = bc.GetReceipt(tx.Hash())
	state, err = bc.GetAccountState("ktlee")
	require.NoErrorf(t, err, "failed to get account state")
	after = state.GetBalanceBigInt().Uint64()

	use = 108404
	require.Equalf(t, use, r.GetGasUsed(), "not equal gas used")
	require.Equalf(t, before-use, after, "not equal balance")
}

func TestFeatureFeeDelegation(t *testing.T) {
	code := readLuaCode("feature_feedelegation_1.lua")
	require.NotEmpty(t, code, "failed to read feature_feedelegation_1.lua")
	code2 := readLuaCode("feature_feedelegation_2.lua")
	require.NotEmpty(t, code, "failed to read feature_feedelegation_2.lua")

	bc, err := LoadDummyChain(SetPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	balance, _ := new(big.Int).SetString("1000000000000000000000", 10)
	send, _ := new(big.Int).SetString("500000000000000000000", 10)

	err = bc.ConnectBlock(NewLuaTxAccountBig("ktlee", balance),
		NewLuaTxAccount("user1", 0),
		NewLuaTxDeploy("ktlee", "fd", 0, code),
		NewLuaTxSendBig("ktlee", "fd", send))
	require.NoErrorf(t, err, "failed to deploy")

	err = bc.ConnectBlock(NewLuaTxCallFeeDelegate("user1", "fd", 0, `{"Name": "check_delegation", "Args":[]}`).
		Fail("check_delegation function is not declared of fee delegation"))
	require.NoErrorf(t, err, "failed to call check_delegation")

	err = bc.ConnectBlock(NewLuaTxCall("user1", "fd", 0, `{"Name": "query", "Args":[]}`).
		Fail("not enough balance"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxCallFeeDelegate("user1", "fd", 0, `{"Name": "query", "Args":[]}`).
		Fail("fee delegation is not allowed"))
	require.NoErrorf(t, err, "failed to call tx")

	contract1, err := bc.GetAccountState("fd")
	require.NoErrorf(t, err, "failed to get contract")

	tx := NewLuaTxCallFeeDelegate("user1", "fd", 0, `{"Name": "query", "Args":["arg"]}`)
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "fd", 0, fmt.Sprintf(`{"Name":"reg", "Args":["%s"]}`, types.EncodeAddress(contract.StrHash("user1")))),
		tx,
	)
	require.NoErrorf(t, err, "failed to call tx")

	contract2, err := bc.GetAccountState("fd")
	require.NoErrorf(t, err, "failed to get contract")
	require.NotEqualf(t, contract1.GetBalanceBigInt().Uint64(), contract2.GetBalanceBigInt().Uint64(), "balance is not changed")

	err = bc.ConnectBlock(tx.Fail("fee delegation is not allowed"))
	require.NoErrorf(t, err, "failed to call tx")

	err = bc.ConnectBlock(NewLuaTxDeploy("ktlee", "fd2", 0, code2))
	require.Errorf(t, err, "expect error")
	require.Containsf(t, err.Error(), "no 'check_delegation' function", "invalid error message")
}

/*
func TestFeatureFeeDelegationLoop(t *testing.T) {
	definition := `
	state.var{
        whitelist = state.map(),
    }

    function query_no(a)
		if (system.isFeeDelegation() == true) then
        	whitelist[system.getSender()] = false
		end
        return 1,2,3,4,5
    end
	function default()
	end
    function check_delegation(fname,k)
		return true
    end
    abi.payable(default)
	abi.fee_delegation(query_no)
`
	bc, err := LoadDummyChain(OnPubNet)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	balance, _ := new(big.Int).SetString("1000000000000000000000", 10)
	send, _ := new(big.Int).SetString("500000000000000000000", 10)

	err = bc.ConnectBlock(
		NewLuaTxAccountBig("ktlee", balance),
		NewLuaTxAccount("user1", 0),
		NewLuaTxDef("ktlee", "fd", 0, definition),
		NewLuaTxSendBig("ktlee", "fd", send),
	)

	err = bc.ConnectBlock(
		NewLuaTxCall("user1", "fd", 0, `{"Name": "query_no", "Args":[]}`).
			Fail("not enough balance"),
	)
	if err != nil {
		t.Error(err)
	}
	txs := make([]LuaTxTester, 10000)

	for i:=0; i < 10000; i++ {
		txs[i] =
			NewLuaTxCallFeeDelegate("user1", "fd", 0, `{"Name": "query_no", "Args":[]}`)
	}
	err = bc.ConnectBlock(txs...)
	if err != nil {
		t.Error(err)
	}
}
*/

// gas test
const (
	DEF_TEST_CONTRACT = "testcontract"
	DEF_TEST_ACCOUNT  = "testaccount"
	DEF_TEST_AMOUNT   = 100000000000000000
)

func expectGas(contractCode string, amount uint64, funcName, funcArgs string, expectGas uint64, opt ...DummyChainOptions) error {
	// append set pubnet
	bc, err := LoadDummyChain(append(opt, SetPubNet())...)
	if err != nil {
		return err
	}
	defer bc.Release()

	err = bc.ConnectBlock(NewLuaTxAccount(DEF_TEST_ACCOUNT, DEF_TEST_AMOUNT),
		NewLuaTxDeploy(DEF_TEST_ACCOUNT, DEF_TEST_CONTRACT, 0, contractCode))
	if err != nil {
		return err
	}

	var code string
	if len(funcArgs) == 0 {
		code = fmt.Sprintf(`{"Name":%s}`, funcName)
	} else {
		code = fmt.Sprintf(`{"Name":%s, "Args":%s}`, funcName, funcArgs)
	}
	tx := NewLuaTxCall(DEF_TEST_ACCOUNT, DEF_TEST_CONTRACT, amount, code)
	err = bc.ConnectBlock(tx)
	if err != nil {
		return err
	}
	r := bc.GetReceipt(tx.Hash())
	if expectGas != r.GetGasUsed() {
		return fmt.Errorf("failed to expect gas, expected: %d, but got: %d", expectGas, r.GetGasUsed())
	}
	return nil
}

func TestGasHello(t *testing.T) {
	var err error
	contract := `function hello() end abi.register(hello)`

	err = expectGas(contract, 0, `"hello"`, ``, 100000, SetHardForkVersion(1))
	assert.NoError(t, err)

	err = expectGas(contract, 0, `"hello"`, ``, 101004, SetHardForkVersion(2))
	assert.NoError(t, err)

	err = expectGas(contract, 0, `"hello"`, ``, 101004, SetHardForkVersion(3))
	assert.NoError(t, err)
}

/*
	func TestGasContractSend(t *testing.T) {
		code := `
		function constructor()
		end
		function send(addr)
			contract.send(addr,1)
			contract.call.value(1)(addr)
		end
		function send2(addr)
			contract.call.value(1)(addr)
			contract.call.value(3)(addr)
		end

		abi.register(send, send2, constructor)
		abi.payable(constructor)

`

	code2 := `
	function default()
		system.print("default called")
	end
	abi.register(default)
	abi.payable(default)

`

		err := expectGas(code, 50000000000000000, `"send"`, fmt.Sprintf(`["%s"]`, types.EncodeAddress(contract.StrHash(DEF_TEST_ACCOUNT))), 105087, SetHardForkVersion(2))
		assert.NoError(t, err)

		err = expectGas(code2, 0, `"send2"`, fmt.Sprintf(`["%s"]`, types.EncodeAddress(contract.StrHash(DEF_TEST_ACCOUNT))), 102179, SetHardForkVersion(2))
		assert.NoError(t, err)
	}
*/
func TestGasOP(t *testing.T) {
	code := readLuaCode("op.lua")
	require.NotEmpty(t, code, "failed to read op.lua")

	err := expectGas(string(code), 0, `"main"`, ``, 100000, SetHardForkVersion(1))
	assert.NoError(t, err)

	err = expectGas(string(code), 0, `"main"`, ``, 117610, SetHardForkVersion(2))
	assert.NoError(t, err)

	err = expectGas(string(code), 0, `"main"`, ``, 120270, SetHardForkVersion(3))
	assert.NoError(t, err)
}

func TestGasBF(t *testing.T) {
	code := readLuaCode("bf.lua")
	require.NotEmpty(t, code, "failed to read bf.lua")

	// err = expectGas(t, string(code), 0, `"main"`, ``, 100000, SetHardForkVersion(1))
	// assert.NoError(t, err)

	err := expectGas(string(code), 0, `"main"`, ``, 47456244, SetHardForkVersion(2))
	assert.NoError(t, err)

	err = expectGas(string(code), 0, `"main"`, ``, 47513803, SetHardForkVersion(3))
	assert.NoError(t, err)
}

/*
// TODO
// ascii
err = expectGas(t, contract, `"hello"`, `[""]`, 101014+96, SetHardForkVersion(2))
assert.NoError(t, err)
err = expectGas(t, contract, `"hello"`, `[""]`, 101014+96, SetHardForkVersion(2))
assert.NoError(t, err)
err = expectGas(t, contract, `"hello"`, `[""]`, 101014+96, SetHardForkVersion(2))
assert.NoError(t, err)

// unicode
err = expectGas(t, contract, `"setItem"`, `["안"]`, 101004+93+9, SetHardForkVersion(2))
assert.NoError(t, err)
err = expectGas(t, contract, `"setItem"`, `["안녕"]`, 101004+93+18, SetHardForkVersion(2))
assert.NoError(t, err)
err = expectGas(t, contract, `"setItem"`, `["안녕하세요"]`, 101004+93+45, SetHardForkVersion(2))
assert.NoError(t, err)

// number
err = expectGas(t, contract, `"hello"`, `[0]`, 101004+189+3, SetHardForkVersion(2))
assert.NoError(t, err)
err = expectGas(t, contract, `"hello"`, `[1]`, 101004+189+3, SetHardForkVersion(2))
assert.NoError(t, err)
err = expectGas(t, contract, `"hello"`, `[10]`, 101004+189+6, SetHardForkVersion(2))
assert.NoError(t, err)
err = expectGas(t, contract, `"hello"`, `[100]`, 101004+189+9, SetHardForkVersion(2))
assert.NoError(t, err)

err = expectGas(t, contract, `"hello"`, `["a","b","c"]`, 101004+93+3, SetHardForkVersion(2))
assert.NoError(t, err)

err = expectGas(t, contract, `"hello"`, `["a","b","aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"]`, 101110+93+3, SetHardForkVersion(2))
assert.NoError(t, err)

err = expectGas(t, contract, `"hello"`, `[1,2,3]`, 101110+189+3, SetHardForkVersion(2))
assert.NoError(t, err)

err = expectGas(t, contract, `"hello"`, `[""]`, 101218, SetHardForkVersion(2))
assert.NoError(t, err)

v3
err = expectGas(t, contract, `"hello"`, `["World"]`, 101218, SetHardForkVersion(3))
assert.NoError(t, err)
*/


func TestMaxCallDepth(t *testing.T) {
  bc, err := LoadDummyChain(OnPubNet)
  if err != nil {
    t.Errorf("failed to create test database: %v", err)
  }
  defer bc.Release()

  err = bc.ConnectBlock(
    NewLuaTxAccount("user", 100000000000000000),
  )
  if err != nil {
    t.Error(err)
  }

/*
  // this contract calls itself recursively
  definition1 := `
state.var {
  call_info = state.map(),
  last_call = state.value(),
  total_calls = state.value()
}

function call_me(call_depth, max_depth)
  call_info[tostring(call_depth)] = call_depth
  last_call:set(call_depth)
  total_calls:set((total_calls:get() or 0) + 1)
  if call_depth == max_depth then
    return call_depth
  else
    return contract.call(system.getContractID(), "call_me", call_depth + 1, max_depth)
  end
end

function get_total_calls()
  return last_call:get(), total_calls:get()
end

function get_call_info(key)
  return call_info[key]
end

function check_state()
  assert(last_call:get() == 64, "last_call")
  assert(total_calls:get() == 64, "total_calls")
  for i = 1, 64 do
    assert(call_info[tostring(i)] == i, "call_info[" .. tostring(i) .. "] = " .. tostring(call_info[tostring(i)]))
  end
  return true
end

abi.register(call_me)
abi.register_view(get_total_calls, get_call_info, check_state)
`
  // deploy 2 identical contracts
  err = bc.ConnectBlock(
    NewLuaTxDef("user", "c1", 0, definition1),
    NewLuaTxDef("user", "c2", 0, definition1),
  )
  if err != nil {
    t.Error(err)
  }

  // call first contract - recursion depth 64
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c1", 0, `{"Name":"call_me", "Args":[1, 64]}`),
  )
  if err != nil {
    t.Error(err)
  }
  // check state
  err = bc.Query("c1", `{"Name":"check_state"}`, "", "true")
  if err != nil {
    t.Error(err)
  }
  // query view
  err = bc.Query("c1", `{"Name":"get_total_calls"}`, "", "[64,64]")
  if err != nil {
    t.Error(err)
  }
  for i := 1; i <= 64; i++ {
    err = bc.Query("c1", fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, i), "", fmt.Sprintf("%d", i))
    if err != nil {
      t.Error(err)
    }
  }

  // call second contract - recursion depth 66
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c2", 0, `{"Name":"call_me", "Args":[1, 66]}`).
      Fail("exceeded the maximum call depth"),
  )
  if err != nil {
    t.Error(err)
  }
  // check state - should fail
  err = bc.Query("c2", `{"Name":"check_state"}`, "", "")
  if err == nil {
    t.Error("should fail")
  }
  // query view - must return nil
  err = bc.Query("c2", `{"Name":"get_total_calls"}`, "", "[null,null]")
  if err != nil {
    t.Error(err)
  }
  for i := 1; i <= 64; i++ {
    err = bc.Query("c2", fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, i), "", "null")
    if err != nil {
      t.Error(err)
    }
  }
*/

  // this contract receives a list of contract IDs to be called
  definition2 := `
state.var {
  call_info = state.map(),
  total_calls = state.value()
}

function call_me(contracts, call_depth, max_depth)
  local last_call = (total_calls:get() or 0) + 1
  total_calls:set(last_call)
  call_info[tostring(last_call)] = call_depth
  if call_depth == max_depth then
    return call_depth
  else
    local next_contract = contracts[call_depth % #contracts + 1]
    return contract.call(next_contract, "call_me", contracts, call_depth + 1, max_depth)
  end
end

function get_total_calls()
  return total_calls:get()
end

function get_call_info(key)
  return call_info[key]
end

abi.register(call_me)
abi.register_view(get_total_calls, get_call_info)
`

  // this contract stores the address of the next contract to be called
  definition3 := `
state.var {
  next_contract = state.value(),
  call_info = state.map(),
  total_calls = state.value()
}

function set_next_contract(next_contract_id)
  next_contract:set(next_contract_id)
end

function call_me(call_depth, max_depth)
  local last_call = (total_calls:get() or 0) + 1
  total_calls:set(last_call)
  call_info[tostring(last_call)] = call_depth
  if call_depth == max_depth then
    return call_depth
  else
    return contract.call(next_contract:get(), "call_me", call_depth + 1, max_depth)
  end
end

function get_total_calls()
  return total_calls:get()
end

function get_call_info(key)
  return call_info[key]
end

abi.register(set_next_contract, call_me)
abi.register_view(get_total_calls, get_call_info)
`

  // deploy 66 identical contracts using definition2
  for i := 1; i <= 66; i++ {
    err = bc.ConnectBlock(
      NewLuaTxDef("user", fmt.Sprintf("c2%d", i), 0, definition2),
    )
    if err != nil {
      t.Error(err)
    }
  }
  // deploy 66 identical contracts using definition3
  for i := 1; i <= 66; i++ {
    err = bc.ConnectBlock(
      NewLuaTxDef("user", fmt.Sprintf("c3%d", i), 0, definition3),
    )
    if err != nil {
      t.Error(err)
    }
  }

  // build a list of contract IDs, used to call the first contract
  contracts := make([]string, 64)
  contracts_str := []byte("")
  for i := 1; i <= 64; i++ {
    contracts[i-1] = types.EncodeAddress(strHash(fmt.Sprintf("c2%d", i)))
  }
  contracts_str, err = json.Marshal(contracts)
	if err != nil {
		t.Error(err)
	}
  // call first contract - recursion depth 64
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c2"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 64]}`, string(contracts_str))),
  )
  if err != nil {
    t.Error(err)
  }
  // check state on all the 64 contracts (query total calls and call info)
  for i := 1; i <= 64; i++ {
    err = bc.Query(fmt.Sprintf("c2%d", i), `{"Name":"get_total_calls"}`, "", "1")
    if err != nil {
      t.Error(err)
    }
    //err = bc.Query(fmt.Sprintf("c2%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, i), "", fmt.Sprintf("%d", i))
    err = bc.Query(fmt.Sprintf("c2%d", i), `{"Name":"get_call_info", "Args":["1"]}`, "", fmt.Sprintf("%d", i))
    if err != nil {
      t.Error(err)
    }
  }

  // add the 66th contract to the list
  contracts = append(contracts, types.EncodeAddress(strHash(fmt.Sprintf("c2%d", 66))))
  contracts_str, err = json.Marshal(contracts)
	if err != nil {
		t.Error(err)
	}
  // call first contract - recursion depth 66
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c2"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 66]}`, string(contracts_str))).Fail("exceeded the maximum call depth"),
  )
  if err != nil {
    t.Error(err)
  }
  // check state on all the 64 contracts (query total calls and call info)
  for i := 1; i <= 64; i++ {
    err = bc.Query(fmt.Sprintf("c2%d", i), `{"Name":"get_total_calls"}`, "", "1")
    if err != nil {
      t.Error(err)
    }
    err = bc.Query(fmt.Sprintf("c2%d", i), `{"Name":"get_call_info", "Args":["1"]}`, "", fmt.Sprintf("%d", i))
    if err != nil {
      t.Error(err)
    }
  }
  // check state on the 66th contract (query total calls and call info)
  err = bc.Query("c2"+fmt.Sprintf("%d", 66), `{"Name":"get_total_calls"}`, "", "null")
  if err != nil {
    t.Error(err)
  }
  err = bc.Query("c2"+fmt.Sprintf("%d", 66), `{"Name":"get_call_info", "Args":["1"]}`, "", "null")
  if err != nil {
    t.Error(err)
  }

  // set next_contract for each contract
  for i := 1; i <= 66; i++ {
    err = bc.ConnectBlock(
      NewLuaTxCall("user", fmt.Sprintf("c3%d", i), 0, fmt.Sprintf(`{"Name":"set_next_contract", "Args":["%s"]}`, types.EncodeAddress(strHash(fmt.Sprintf("c3%d", i+1))))),
    )
    if err != nil {
      t.Error(err)
    }
  }
  // call first contract - recursion depth 64
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c3"+fmt.Sprintf("%d", 1), 0, `{"Name":"call_me", "Args":[1, 64]}`),
  )
  if err != nil {
    t.Error(err)
  }
  // check state on all the 64 contracts (query total calls and call info)
  for i := 1; i <= 64; i++ {
    err = bc.Query(fmt.Sprintf("c3%d", i), `{"Name":"get_total_calls"}`, "", "1")
    if err != nil {
      t.Error(err)
    }
    err = bc.Query(fmt.Sprintf("c3%d", i), `{"Name":"get_call_info", "Args":["1"]}`, "", fmt.Sprintf("%d", i))
    if err != nil {
      t.Error(err)
    }
  }

  // call first contract - recursion depth 66
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c3"+fmt.Sprintf("%d", 1), 0, `{"Name":"call_me", "Args":[1, 66]}`).Fail("exceeded the maximum call depth"),
  )
  if err != nil {
    t.Error(err)
  }
  // check state on all the 64 contracts (query total calls and call info)
  for i := 1; i <= 64; i++ {
    err = bc.Query(fmt.Sprintf("c3%d", i), `{"Name":"get_total_calls"}`, "", "1")
    if err != nil {
      t.Error(err)
    }
    err = bc.Query(fmt.Sprintf("c3%d", i), `{"Name":"get_call_info", "Args":["1"]}`, "", fmt.Sprintf("%d", i))
    if err != nil {
      t.Error(err)
    }
  }
  // check state on the 66th contract (query total calls and call info)
  err = bc.Query("c3"+fmt.Sprintf("%d", 66), `{"Name":"get_total_calls"}`, "", "null")
  if err != nil {
    t.Error(err)
  }
  err = bc.Query("c3"+fmt.Sprintf("%d", 66), `{"Name":"get_call_info", "Args":["1"]}`, "", "null")
  if err != nil {
    t.Error(err)
  }


  // Circle: contract 1 calls contract 2, contract 2 calls contract 3, contract 3 calls contract 1...

  // deploy 4 identical contracts using definition2
  for i := 1; i <= 4; i++ {
    err = bc.ConnectBlock(
      NewLuaTxDef("user", fmt.Sprintf("c4%d", i), 0, definition2),
    )
    if err != nil {
      t.Error(err)
    }
  }
  // build a list of contract IDs, used to call the first contract
  contracts = make([]string, 4)
  for i := 1; i <= 4; i++ {
    contracts[i-1] = types.EncodeAddress(strHash(fmt.Sprintf("c4%d", i)))
  }
  contracts_str, err = json.Marshal(contracts)
	if err != nil {
		t.Error(err)
	}
  // call first contract - recursion depth 64
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c4"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 64]}`, string(contracts_str))),
  )
  if err != nil {
    t.Error(err)
  }
  // check state on all the 4 contracts
  // each contract should have (64 / 4) = 16 calls
  for i := 1; i <= 4; i++ {
    err = bc.Query(fmt.Sprintf("c4%d", i), `{"Name":"get_total_calls"}`, "", "16")
    if err != nil {
      t.Error(err)
    }
    for j := 1; j <= 16; j++ {
      err = bc.Query(fmt.Sprintf("c4%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, j), "", fmt.Sprintf("%d", i+4*(j-1)))
      if err != nil {
        t.Error(err)
      }
    }
  }

  // call first contract - recursion depth 66
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c4"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 66]}`, string(contracts_str))).Fail("exceeded the maximum call depth"),
  )
  if err != nil {
    t.Error(err)
  }
  // check state on all the 4 contracts
  // each contract should have (64 / 4) = 16 calls
  for i := 1; i <= 4; i++ {
    err = bc.Query(fmt.Sprintf("c4%d", i), `{"Name":"get_total_calls"}`, "", "16")
    if err != nil {
      t.Error(err)
    }
    for j := 1; j <= 16; j++ {
      err = bc.Query(fmt.Sprintf("c4%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, j), "", fmt.Sprintf("%d", i+4*(j-1)))
      if err != nil {
        t.Error(err)
      }
    }
  }


  // ZigZag: contract 1 calls contract 2, contract 2 calls contract 1...

  // deploy 2 identical contracts using definition2
  for i := 1; i <= 2; i++ {
    err = bc.ConnectBlock(
      NewLuaTxDef("user", fmt.Sprintf("c5%d", i), 0, definition2),
    )
    if err != nil {
      t.Error(err)
    }
  }
  // build a list of contract IDs, used to call the first contract
  contracts = make([]string, 2)
  for i := 1; i <= 2; i++ {
    contracts[i-1] = types.EncodeAddress(strHash(fmt.Sprintf("c5%d", i)))
  }
  contracts_str, err = json.Marshal(contracts)
	if err != nil {
		t.Error(err)
	}
  // call first contract - recursion depth 64
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c5"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 64]}`, string(contracts_str))),
  )
  if err != nil {
    t.Error(err)
  }
  // check state on all the 2 contracts
  // each contract should have (64 / 2) = 32 calls
  for i := 1; i <= 2; i++ {
    err = bc.Query(fmt.Sprintf("c5%d", i), `{"Name":"get_total_calls"}`, "", "32")
    if err != nil {
      t.Error(err)
    }
    for j := 1; j <= 32; j++ {
      err = bc.Query(fmt.Sprintf("c5%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, j), "", fmt.Sprintf("%d", i+2*(j-1)))
      if err != nil {
        t.Error(err)
      }
    }
  }

  // call first contract - recursion depth 66
  err = bc.ConnectBlock(
    NewLuaTxCall("user", "c5"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 66]}`, string(contracts_str))).Fail("exceeded the maximum call depth"),
  )
  if err != nil {
    t.Error(err)
  }
  // check state on all the 2 contracts
  // each contract should have (64 / 2) = 32 calls
  for i := 1; i <= 2; i++ {
    err = bc.Query(fmt.Sprintf("c5%d", i), `{"Name":"get_total_calls"}`, "", "32")
    if err != nil {
      t.Error(err)
    }
    for j := 1; j <= 32; j++ {
      err = bc.Query(fmt.Sprintf("c5%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, j), "", fmt.Sprintf("%d", i+2*(j-1)))
      if err != nil {
        t.Error(err)
      }
    }
  }

}
