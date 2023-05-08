package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/aergoio/aergo/types"
)

func TestContractHello(t *testing.T) {
	code := `function hello(say) return "Hello " .. say end abi.register(hello)`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
	)
	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "hello", 0, code),
	)
	tx := NewLuaTxCall("ktlee", "hello", 0, `{"Name":"hello", "Args":["World"]}`)
	_ = bc.ConnectBlock(tx)
	receipt := bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `"Hello World"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestContractSystem(t *testing.T) {
	code := `function testState()
	system.setItem("key1", 999)
	return system.getSender(), system.getTxhash(),system.getContractID(), system.getTimestamp(), system.getBlockheight(), system.getItem("key1")
  end 
abi.register(testState)`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
	)
	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "system", 0, code),
	)
	tx := NewLuaTxCall("ktlee", "system", 0, `{"Name":"testState", "Args":[]}`)
	_ = bc.ConnectBlock(tx)
	receipt := bc.GetReceipt(tx.Hash())
	exRv := fmt.Sprintf(`["Amg6nZWXKB6YpNgBPv9atcjdm6hnFvs5wMdRgb2e9DmaF5g9muF2","GcVQWfSWUKoPuxbhoqx18hD4JKs2L1cvVvJZFzXKWaQ2","AmhNNBNY7XFk4p5ym4CJf8nTcRTEHjWzAeXJfhP71244CjBCAQU3",%d,3,999]`, bc.cBlock.Header.Timestamp/1e9)
	if receipt.GetRet() != exRv {
		t.Errorf("expected: %s, but got: %s", exRv, receipt.GetRet())
	}
}

func TestContractSend(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
	function constructor()
	end
    function send(addr)
        contract.send(addr,1)
		contract.call.value(1)(addr)
    end
    abi.register(send, constructor)
	abi.payable(constructor)
`
	definition2 := `
    function default()
		system.print("default called")
    end
    abi.register(default)
	abi.payable(default)
`
	definition3 := `
    function test()
    end
    abi.register(test)
`
	definition4 := `
    function default()
    end
    abi.register(default)
`
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "test1", 50, definition),
		NewLuaTxDef("ktlee", "test2", 0, definition2),
		NewLuaTxDef("ktlee", "test3", 0, definition3),
		NewLuaTxDef("ktlee", "test4", 0, definition4),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, types.EncodeAddress(strHash("test2")))),
	)
	if err != nil {
		t.Error(err)
	}
	state, err := bc.GetAccountState("test2")
	if state.GetBalanceBigInt().Uint64() != 2 {
		t.Error("balance error", state.GetBalanceBigInt())
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, types.EncodeAddress(strHash("test3")))).Fail(`[Contract.LuaSendAmount] call err: not found function: default`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, types.EncodeAddress(strHash("test4")))).Fail(`[Contract.LuaSendAmount] call err: 'default' is not payable`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, types.EncodeAddress(strHash("ktlee")))),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestContractSendF(t *testing.T) {
	bc, err := LoadDummyChain(OnPubNet)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
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
	definition2 := `
    function default()
		system.print("default called")
    end
    abi.register(default)
	abi.payable(default)
`
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "test1", 50000000000000000, definition),
		NewLuaTxDef("ktlee", "test2", 0, definition2),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "test1", 0,
		fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`,
			types.EncodeAddress(strHash("test2"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	r := bc.GetReceipt(tx.Hash())
	expectedFee := uint64(105087)
	if r.GetGasUsed() != expectedFee {
		t.Errorf("expected: %d, but got: %d", expectedFee, r.GetGasUsed())
	}
	state, err := bc.GetAccountState("test2")
	if state.GetBalanceBigInt().Uint64() != 2 {
		t.Error("balance error", state.GetBalanceBigInt())
	}
	tx = NewLuaTxCall("ktlee", "test1", 0,
		fmt.Sprintf(`{"Name":"send2", "Args":["%s"]}`,
			types.EncodeAddress(strHash("test2"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	r = bc.GetReceipt(tx.Hash())
	expectedFee = uint64(105179)
	if r.GetGasUsed() != expectedFee {
		t.Errorf("expected: %d, but got: %d", expectedFee, r.GetGasUsed())
	}
	state, err = bc.GetAccountState("test2")
	if state.GetBalanceBigInt().Uint64() != 6 {
		t.Error("balance error", state.GetBalanceBigInt())
	}
}

func TestContractQuery(t *testing.T) {
	code := `function inc()
a = system.getItem("key1")
if (a == nil) then
	system.setItem("key1", 1)
	return
end
system.setItem("key1", a + 1)
end
function query(a)
	return system.getItem(a)
end
abi.register(query)
abi.payable(inc)`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
	)
	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "query", 0, code),
		NewLuaTxCall("ktlee", "query", 2, `{"Name":"inc", "Args":[]}`),
	)

	query, err := bc.GetAccountState("query")
	if err != nil {
		t.Error(err)
	}
	if query.GetBalanceBigInt().Uint64() != uint64(2) {
		t.Error(query.Balance)
	}

	err = bc.Query("query", `{"Name":"inc", "Args":[]}`, "set not permitted in query", "")
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
}

func TestContractCall(t *testing.T) {
	definition1 := `
	function constructor(init)
		system.setItem("count", init)
	end
	function inc()
		count = system.getItem("count")
		system.setItem("count", count + 1)
		return count
	end

	function get()
		return system.getItem("count")
	end

	function set(val)
		system.setItem("count", val)
	end
	abi.register(inc,get,set)
	`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "counter", 0, definition1).Constructor("[1]"),
		NewLuaTxCall("ktlee", "counter", 0, `{"Name":"inc", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "2")
	if err != nil {
		t.Error(err)
	}

	definition2 := `
	function constructor(addr)
		system.setItem("count", 99)
		system.setItem("addr", addr)
	end
	function add(amount)
		return contract.call.value(amount)(system.getItem("addr"), "inc")
	end
	function dadd()
		return contract.delegatecall(system.getItem("addr"), "inc")
	end
	function get()
		addr = system.getItem("addr")
		a = contract.call(addr, "get")
		return a
	end
	function dget()
		addr = system.getItem("addr")
		a = contract.delegatecall(addr, "get")
		return a
	end
	function set(val)
		contract.call(system.getItem("addr"), "set", val)
	end
	function dset(val)
		contract.delegatecall(system.getItem("addr"), "set", val)
	end
	abi.register(add,dadd, get, dget, set, dset)
	`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "caller", 0, definition2).
			Constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(strHash("counter")))),
		NewLuaTxCall("ktlee", "caller", 0, `{"Name":"add", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "3")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("caller", `{"Name":"dget", "Args":[]}`, "", "99")
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "caller", 0, `{"Name":"dadd", "Args":[]}`)
	_ = bc.ConnectBlock(tx)
	receipt := bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `99` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	tx = NewLuaTxCall("ktlee", "caller", 0, `{"Name":"dadd", "Args":[]}`)
	_ = bc.ConnectBlock(tx)
	receipt = bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `100` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "3")
	if err != nil {
		t.Error(err)
	}
}

func TestPingpongCall(t *testing.T) {
	definition1 := `
	function constructor()
		system.setItem("key",  "empty")
	end
	function start(addr)
		system.setItem("key",  "start")
		contract.call(addr, "called")
	end

	function callback()
		system.setItem("key",  "callback")
	end

	function get()
		return system.getItem("key")
	end

	abi.register(start, callback, get)
	`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "a", 0, definition1),
	)

	definition2 := `
	function constructor(addr)
		system.setItem("key",  "empty")
		system.setItem("addr",  addr)
	end

	function called()
		system.setItem("key",  "called")
		contract.call(system.getItem("addr"), "callback")
	end

	function get()
		return system.getItem("key")
	end

	abi.register(called, get)
	`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "b", 0, definition2).
			Constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(strHash("a")))),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "a", 0,
		fmt.Sprintf(`{"Name":"start", "Args":["%s"]}`, types.EncodeAddress(strHash("b"))))
	_ = bc.ConnectBlock(tx)
	err = bc.Query("a", `{"Name":"get", "Args":[]}`, "", `"callback"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("b", `{"Name":"get", "Args":[]}`, "", `"called"`)
	if err != nil {
		t.Error(err)
	}
}

func TestRollback(t *testing.T) {
	code := `function inc()
a = system.getItem("key1")
if (a == nil) then
	system.setItem("key1", 1)
	return
end
system.setItem("key1", a + 1)
end
function query(a)
	return system.getItem(a)
end
abi.register(query)
abi.payable(inc)`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
	)
	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "query", 0, code),
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
	)
	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
	)
	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
	)

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "5")
	if err != nil {
		t.Error(err)
	}

	err = bc.DisConnectBlock()
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "3")
	if err != nil {
		t.Error(err)
	}

	err = bc.DisConnectBlock()
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
	)

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
}

func TestAbi(t *testing.T) {
	errMsg := "no exported functions"

	noAbi := `
	function dummy()
		system.print("dummy")
	end`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "a", 0, noAbi),
	)
	if err == nil {
		t.Errorf("expected: %s, but got: nil", errMsg)
	} else if !strings.Contains(err.Error(), "no exported functions") {
		t.Errorf("expected: %s, but got: %s", errMsg, err.Error())
	}

	empty := `
	function dummy()
		system.print("dummy")
	end
	abi.register()`

	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "a", 0, empty),
	)
	if err == nil {
		t.Errorf("expected: %s, but got: nil", errMsg)
	} else if !strings.Contains(err.Error(), "no exported functions.") {
		t.Errorf("expected: %s, but got: %s", errMsg, err.Error())
	}

	localFunc := `
	function dummy()
		system.print("dummy")
	end
	local function helper()
		system.print("helper")
	end
	abi.register(helper)`

	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "a", 0, localFunc),
	)
	if err == nil {
		t.Errorf("expected: %s, but got: nil", errMsg)
	} else if !strings.Contains(err.Error(), "global function expected") {
		t.Errorf("expected: %s, but got: %s", errMsg, err.Error())
	}
}

func TestGetABI(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "hello", 0,
			`state.var {
	Say = state.value()
}

function hello(say) 
  return "Hello " .. say 
end 

abi.register(hello)`),
	)
	abi, err := bc.GetABI("hello")
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(abi)
	if err != nil {
		t.Error(err)
	}
	if string(b) != `{"version":"0.2","language":"lua","functions":[{"name":"hello","arguments":[{"name":"say"}]}],"state_variables":[{"name":"Say","type":"value"}]}` {
		t.Error(string(b))
	}
}

func TestPayable(t *testing.T) {
	src := `
state.var {
	Data = state.value()
}
function save(data)
	Data:set(data)
end
function load()
	return Data:get()
end

abi.register(load)
abi.payable(save)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
	)

	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "payable", 1, src),
	)
	if err == nil {
		t.Error("expected: 'constructor' is not payable")
	} else {
		if !strings.Contains(err.Error(), "'constructor' is not payable") {
			t.Error(err)
		}
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`).Fail("not found contract"),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "payable", 0, src),
		NewLuaTxCall("ktlee", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("payable", `{"Name":"load"}`, "", `"blahblah"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "payable", 1, `{"Name":"save", "Args": ["payed"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("payable", `{"Name":"load"}`, "", `"payed"`)
	if err != nil {
		t.Error(err)
	}
}

func TestDefault(t *testing.T) {
	src := `
function default()
	return "default"
end
abi.register(default)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "default", 0, src),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "default", 0, "")
	err = bc.ConnectBlock(tx)
	receipt := bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `"default"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "default", 1, "").Fail(`'default' is not payable`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("default", `{"Name":"a"}`, "not found function: a", "")
	if err != nil {
		t.Error(err)
	}
}

func TestReturn(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "return_num", 0, "function return_num() return 10 end abi.register(return_num)"),
		NewLuaTxCall("ktlee", "return_num", 0, `{"Name":"return_num", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("return_num", `{"Name":"return_num", "Args":[]}`, "", "10")
	if err != nil {
		t.Error(err)
	}

	foo := `function foo()
	return {1,2,3}
end
function foo2(bar)
	return bar
	end
abi.register(foo,foo2)`

	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "foo", 0, foo),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("foo", `{"Name":"foo", "Args":[]}`, "", "[1,2,3]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("foo", `{"Name":"foo2", "Args":["foo314"]}`, "", `"foo314"`)
	if err != nil {
		t.Error(err)
	}
}
func TestReturnUData(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
	function test_die()
	return contract.call(system.getContractID(), "return_object")
	end
	function return_object()
	return db.query("select 1")
	end
	abi.register(test_die, return_object)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "rs-return", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "rs-return", 0, `{"Name": "test_die", "Args":[]}`).Fail(`unsupport type: userdata`),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestEvent(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
    function test_ev()
        contract.event("ev1", 1,"local", 2, "form")
        contract.event("ev1", 3,"local", 4, "form")
    end
    abi.register(test_ev)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "event", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "event", 0, `{"Name": "test_ev", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestView(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
    function test_view(a)
      contract.event("ev1", 1,"local", 2, "form")
      contract.event("ev1", 3,"local", 4, "form")
    end

	function k(a)
		return a
	end

	function tx_in_view_function()
		k2()
	end

	function k2()
		test_view()
	end

	function k3()
		ret = contract.pcall(test_view)
		assert (ret == false)
		contract.event("ev2", 4, "global")
	end
	function tx_after_view_function()
		assert(k(1) == 1)
        contract.event("ev1", 1,"local", 2, "form")
	end
	function sqltest()
  		db.exec([[create table if not exists book (
			page number,
			contents text
		)]])
	end
    abi.register(test_view, tx_after_view_function, k2, k3)
    abi.register_view(test_view, k, tx_in_view_function, sqltest)
`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "view", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "view", 0, `{"Name": "test_view", "Args":[]}`).Fail("[Contract.Event] event not permitted in query"),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("view", `{"Name":"k", "Args":[10]}`, "", "10")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "view", 0, `{"Name": "tx_in_view_function", "Args":[]}`).Fail("[Contract.Event] event not permitted in query"),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "view", 0, `{"Name": "tx_after_view_function", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "view", 0, `{"Name": "k2", "Args":[]}`).Fail("[Contract.Event] event not permitted in query"),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "view", 0, `{"Name": "k3", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "view", 0, `{"Name": "sqltest", "Args":[]}`).Fail("not permitted in view function"),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestDeploy(t *testing.T) {
	deploy := `
function hello()
	hello = [[
function hello(say)
	return "Hello " .. say 
end

local type_check = {}
function type_check.isValidAddress(address)
    -- check existence of invalid alphabets
    if nil ~= string.match(address, '[^123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]') then
        return false
    end
    -- check lenght is in range
    if 52 ~= string.len(address) then
        return false
    end
    -- TODO add checksum verification?
    return true
end
function type_check.isValidNumber(value)
    if nil ~= string.match(value, '[^0123456789]') then
        return false
    end
    return true
end

-- The a bridge token is a mintable and burnable token controlled by
-- the bridge contract. It represents all tokens locked on the other side of the 
-- bridge with a 1:1 ratio.
-- This contract is depoyed by the merkle bridge when a new type of token 
-- is transfered
state.var {
    Symbol = state.value(),
    Name = state.value(),
    Decimals = state.value(),
    TotalSupply = state.value(),
    Balances = state.map(),
    Nonces = state.map(),
    -- Contract ID is a unique id that cannot be shared by another contract, even one on a sidechain
    -- This is neeeded for replay protection of signed transfer, because users might have the same private key
    -- on different sidechains
    ContractID = state.value(),
    Owner = state.value(),
}

function constructor() 
    Symbol:set("TOKEN")
    Name:set("Standard Token on Aergo")
    Decimals:set(18)
    TotalSupply:set(bignum.number(0))
    Owner:set(system.getSender())
    -- contractID is the hash of system.getContractID (prevent replay between contracts on the same chain) and system.getPrevBlockHash (prevent replay between sidechains).
    -- take the first 16 bytes to save size of signed message
    local id = crypto.sha256(system.getContractID()..system.getPrevBlockHash())
    id = string.sub(id, 3, 32)
    ContractID:set(id)
    return true
end

---------------------------------------
-- Transfer sender's token to target 'to'
-- @type        call
-- @param to    a target address
-- @param value string amount of tokens to send
-- @return      success
---------------------------------------
function transfer(to, value) 
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    assert(type_check.isValidAddress(to), "invalid address format: " .. to)
    local from = system.getSender()
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    assert(bvalue > b0, "invalid value")
    assert(to ~= from, "same sender and receiver")
    assert(Balances[from] and bvalue <= Balances[from], "not enough balance")
    Balances[from] = Balances[from] - bvalue
    Nonces[from] = (Nonces[from] or 0) + 1
    Balances[to] = (Balances[to] or b0) + bvalue
    -- TODO event notification
    return true
end

---------------------------------------
-- Transfer tokens according to signed data from the owner
-- @type  call
-- @param from      sender's address
-- @param to        receiver's address
-- @param value     string amount of token to send in aer
-- @param nonce     nonce of the sender to prevent replay
-- @param fee       string fee given to the tx broadcaster
-- @param deadline  block number before which the tx can be executed
-- @param signature signature proving sender's consent
-- @return          success
---------------------------------------
function signed_transfer(from, to, value, nonce, signature, fee, deadline)
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    assert(type_check.isValidNumber(fee), "invalid fee format (must be string)")
    local bfee = bignum.number(fee)
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    -- check addresses
    assert(type_check.isValidAddress(to), "invalid address format: " .. to)
    assert(type_check.isValidAddress(from), "invalid address format: " .. from)
    assert(to ~= from, "same sender and receiver")
    -- check amounts, fee
    assert(bfee >= b0, "fee must be positive")
    assert(bvalue >= b0, "value must be positive")
    assert(Balances[from] and (bvalue+bfee) <= Balances[from], "not enough balance")
    -- check deadline
    assert(deadline == 0 or system.getBlockheight() < deadline, "deadline has passed")
    -- check nonce
    if Nonces[from] == nil then Nonces[from] = 0 end
    assert(Nonces[from] == nonce, "nonce is invalid or already spent")
    -- construct signed transfer and verifiy signature
    data = crypto.sha256(to..bignum.tostring(bvalue)..tostring(nonce)..bignum.tostring(bfee)..tostring(deadline)..ContractID:get())
    assert(crypto.ecverify(data, signature, from), "signature of signed transfer is invalid")
    -- execute transfer
    Balances[from] = Balances[from] - bvalue - bfee
    Balances[to] = (Balances[to] or b0) + bvalue
    Balances[system.getOrigin()] = (Balances[system.getOrigin()] or b0) + bfee
    Nonces[from] = Nonces[from] + 1
    -- TODO event notification
    return true
end


---------------------------------------
-- mint, burn and signed_burn are specific to the token contract controlled by
-- the merkle bridge contract and representing transfered assets.
---------------------------------------

---------------------------------------
-- Mint tokens to 'to'
-- @type        call
-- @param to    a target address
-- @param value string amount of token to mint
-- @return      success
---------------------------------------
function mint(to, value)
    assert(system.getSender() == Owner:get(), "Only bridge contract can mint")
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    assert(type_check.isValidAddress(to), "invalid address format: " .. to)
    local new_total = TotalSupply:get() + bvalue
    TotalSupply:set(new_total)
    Balances[to] = (Balances[to] or b0) + bvalue;
    -- TODO event notification
    return true
end

---------------------------------------
-- burn the tokens of 'from'
-- @type        call
-- @param from  a target address
-- @param value an amount of token to send
-- @return      success
---------------------------------------
function burn(from, value)
    assert(system.getSender() == Owner:get(), "Only bridge contract can burn")
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    assert(type_check.isValidAddress(from), "invalid address format: " ..from)
    assert(Balances[from] and bvalue <= Balances[from], "Not enough funds to burn")
    new_total = TotalSupply:get() - bvalue
    TotalSupply:set(new_total)
    Balances[from] = Balances[from] - bvalue
    -- TODO event notification
    return true
end

---------------------------------------
-- signed_burn the tokens of 'from' according to signed data from the owner
-- @type            call
-- @param from      a target address
-- @param value     an amount of token to send
-- @param nonce     nonce of the sender to prevent replay
-- @param fee       string fee given to the tx broadcaster
-- @param deadline  block number before which the tx can be executed
-- @param signature signature proving sender's consent
-- @return          success
---------------------------------------
function signed_burn(from, value, nonce, signature, fee, deadline)
    assert(system.getSender() == Owner:get(), "Only bridge contract can burn")
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    assert(type_check.isValidNumber(fee), "invalid fee format (must be string)")
    local bfee = bignum.number(fee)
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    -- check addresses
    assert(type_check.isValidAddress(from), "invalid address format: " .. from)
    -- check amounts, fee
    assert(bfee >= b0, "fee must be positive")
    assert(bvalue >= b0, "value must be positive")
    assert(Balances[from] and (bvalue+bfee) <= Balances[from], "not enough balance")
    -- check deadline
    assert(deadline == 0 or system.getBlockheight() < deadline, "deadline has passed")
    -- check nonce
    if Nonces[from] == nil then Nonces[from] = 0 end
    assert(Nonces[from] == nonce, "nonce is invalid or already spent")
    -- construct signed transfer and verifiy signature
    data = crypto.sha256(system.getSender()..bignum.tostring(bvalue)..tostring(nonce)..bignum.tostring(bfee)..tostring(deadline)..ContractID:get())
    assert(crypto.ecverify(data, signature, from), "signature of signed transfer is invalid")
    -- execute burn
    new_total = TotalSupply:get() - bvalue
    TotalSupply:set(new_total)
    Balances[from] = Balances[from] - bvalue - bfee
    Balances[system.getOrigin()] = (Balances[system.getOrigin()] or b0) + bfee
    Nonces[from] = Nonces[from] + 1
    -- TODO event notification
    return true
end


-- register functions to abi
abi.register(transfer, signed_transfer, mint, burn, signed_burn, hello)
	]]
	addr = contract.deploy(hello)
	ret = contract.call(addr, "hello", "world")
	return addr, ret
end

function helloQuery(addr)
	return contract.call(addr, "hello", "world")
end

function testConst()
	src = [[
		function hello(say, key) 
			return "Hello " .. say .. system.getItem(key) 
		end 
		function constructor(key, item) 
			system.setItem(key, item)
			return key, item
		end 
		abi.register(hello) 
		abi.payable(constructor)
	]]
	addr, key, item = contract.deploy.value(100)(src, "key", 2)
	ret = contract.call(addr, "hello", "world", "key")
	return addr, ret
end

function testFail()
	src = [[
		function hello(say, key) 
			return "Hello " .. say .. system.getItem(key) 
		end 
		function constructor()
		end 
		abi.register(hello) 
	]]
	addr = contract.deploy.value(100)(src)
	return addr
end
 
paddr = nil
function deploy()
	src = [[
		function hello(say, key) 
			return "Hello " .. say .. system.getItem(key) 
		end 
		function getcre()
			return system.getCreator()
		end
		function constructor()
		end 
		abi.register(hello, getcre) 
	]]
	paddr = contract.deploy(src)
	system.print("addr :", paddr)
	ret = contract.call(paddr, "hello", "world", "key")
end

function testPcall()
	ret = contract.pcall(deploy)
	return contract.call(paddr, "getcre")
end
function constructor()
end

abi.register(hello, helloQuery, testConst, testFail, testPcall)
abi.payable(constructor)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "deploy", 50000000000, deploy),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"hello"}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `["AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf","Hello world"]` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.Query("deploy", `{"Name":"helloQuery", "Args":["AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf"]}`, "", `"Hello world"`)
	if err != nil {
		t.Error(err)
	}
	tx = NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"testConst"}`)
	err = bc.ConnectBlock(tx)
	receipt = bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `["Amhmj6kKZz7mPstBAPJWRe1e8RHP7bZ5pV35XatqTHMWeAVSyMkc","Hello world2"]` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	deployAcc, err := bc.GetAccountState("deploy")
	if err != nil {
		t.Error(err)
	}
	if deployAcc.GetBalanceBigInt().Uint64() != uint64(49999999900) {
		t.Error(deployAcc.GetBalanceBigInt().Uint64())
	}
	deployAcc, _ = bc.GetAccountState("deploy")
	tx = NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"testFail"}`)
	err = bc.ConnectBlock(tx)
	deployAcc, _ = bc.GetAccountState("deploy")
	if deployAcc.Nonce != 2 {
		t.Error("nonce rollback failed", deployAcc.Nonce)
	}
	tx = NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"testPcall"}`)
	err = bc.ConnectBlock(tx)
	deployAcc, _ = bc.GetAccountState("deploy")
	if deployAcc.Nonce != 2 {
		t.Error("nonce rollback failed", deployAcc.Nonce)
	}
	receipt = bc.GetReceipt(tx.Hash())
	if !strings.Contains(receipt.GetRet(), "cannot find contract") {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestDeploy2(t *testing.T) {
	deploy := `
function hello()
	src = [[
state.var{
  counts = state.array(10)
}

counts[1] = 10
function inc(key)
  if counts[key] == nil then
    counts[key] = 0
  end
  counts[key] = counts[key] + 1
end

function get(key)
  return counts[key]
end

function set(key,val)
  counts[key] = val
end

function len()
  return counts:length()
end

function iter()
  local rv = {}
  for i, v in counts:ipairs() do
    if v == nil then
      rv[i] = "nil"
    else
      rv[i] = v
    end
  end
  return rv
end

abi.register(inc,get,set,len,iter)
	]]
	paddr = contract.deploy(src)
	system.print("addr :", paddr)
	ret = contract.call(paddr, "hello", "world", "key")
end

function constructor()
end

abi.register(hello)
abi.payable(constructor)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "deploy", 50000000000, deploy),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"hello"}`).Fail(`not permitted state referencing at global scope`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
}

func TestNDeploy(t *testing.T) {

	bc, err := LoadDummyChain()

	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function constructor()
  testall()
end

function testall()
  deploytest()
  sendtest()
end

function deploytest()
  src = [[
  function default()
    contract.send(system.getSender(), system.getAmount())
  end

  function getargs(...)
    tb = {...}
  end

  abi.payable(default)
  abi.register(getargs)
  ]]

  addr = contract.deploy(src)
  id = 'deploy_src'; system.setItem(id, addr)
  system.print(id, system.getItem(id))

  korean_char_src = [[
  function 함수()
    변수 = 1
    결과 = 변수 + 3
    system.print('결과', 결과)
  end

  abi.register(함수)
  ]]

  
  korean_char_src222 = [[
    function default()
      contract.send(system.getSender(), system.getAmount())
    end
  
    function getargs(...)
      tb = {...}
    end
  
    function x()
    end

    abi.payable(default)
    abi.register(getargs)
  ]]

  korean_addr =  contract.deploy(korean_char_src)
  id = 'korean_char_src'; system.setItem(id, korean_addr)
  system.print(id, system.getItem(id))
end

function sendtest()
  addr = system.getItem("deploy_src")
  system.print('ADDRESS', addr, system.getAmount())
  
  id = 's01'; system.setItem(id,{pcall(function() contract.send(addr, system.getAmount()) end)})
  system.print(id, system.getItem(id))
end

function default()
  -- do nothing
end

abi.payable(constructor, default)
abi.register(testall)
`
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "n-deploy", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestDeployFee(t *testing.T) {
	src := `
	paddr = nil
	function deploy()
	src = [[
		function hello(say, key)
	return "Hello " .. say .. key
	end
	function getcre()
	return system.getCreator()
	end
	function constructor()
	end
	abi.register(hello, getcre)
]]
	paddr = contract.deploy(src)
	system.print("addr :", paddr)
	ret = contract.call(paddr, "hello", "world", "key")
	end
	function testPcall()
		ret = contract.pcall(deploy)
		return contract.call(paddr, "getcre")
	end

	abi.register(testPcall)`

	bc, err := LoadDummyChain(
		func(d *DummyChain) {
			d.timeout = 50
		},
		OnPubNet,
	)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "deploy", 0, src),
	)
	if err != nil {
		t.Error(err)
	}
	state, err := bc.GetAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	bal := state.GetBalanceBigInt().Uint64()
	tx := NewLuaTxCall("ktlee", "deploy", 0, `{"Name": "testPcall"}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	r := bc.GetReceipt(tx.Hash())
	expectedFee := uint64(117861)
	if r.GetGasUsed() != expectedFee {
		t.Errorf("expected: %d, but got: %d", expectedFee, r.GetGasUsed())
	}
	state, err = bc.GetAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	if bal-expectedFee != state.GetBalanceBigInt().Uint64() {
		t.Errorf(
			"expected: %d, but got: %d",
			bal-expectedFee,
			state.GetBalanceBigInt().Uint64(),
		)
	}
}

func xestInfiniteLoop(t *testing.T) {
	bc, err := LoadDummyChain(
		func(d *DummyChain) {
			d.timeout = 50
		},
	)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function infiniteLoop()
    local t = 0
	while true do
	    t = t + 1
	end
	return t
end
function infiniteCall()
	infiniteCall()
end
function catch()
	return pcall(infiniteLoop)
end
function contract_catch()
	return contract.pcall(infiniteLoop)
end
abi.register(infiniteLoop, infiniteCall, catch, contract_catch)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "loop", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"infiniteLoop"}`,
		),
	)
	errTimeout := "exceeded the maximum instruction count"
	if err == nil {
		t.Errorf("expected: %s", errTimeout)
	}
	if err != nil && !strings.Contains(err.Error(), errTimeout) {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"catch"}`,
		),
	)
	if err == nil {
		t.Errorf("expected: %s", errTimeout)
	}
	if err != nil && !strings.Contains(err.Error(), errTimeout) {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"contract_catch"}`,
		),
	)
	if err == nil {
		t.Errorf("expected: %s", errTimeout)
	}
	if err != nil && !strings.Contains(err.Error(), errTimeout) {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"infiniteCall"}`,
		).Fail("stack overflow"),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestInfiniteLoopOnPubNet(t *testing.T) {
	bc, err := LoadDummyChain(
		func(d *DummyChain) {
			d.timeout = 50
		},
		OnPubNet,
	)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function infiniteLoop()
    local t = 0
	while true do
	    t = t + 1
	end
	return t
end
function infiniteCall()
	infiniteCall()
end
function catch()
	return pcall(infiniteLoop)
end
function contract_catch()
	return contract.pcall(infiniteLoop)
end
abi.register(infiniteLoop, infiniteCall, catch, contract_catch)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "loop", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"infiniteLoop"}`,
		),
	)
	errTimeout := VmTimeoutError{}
	if err == nil {
		t.Errorf("expected: %v", errTimeout)
	}
	if err != nil && !strings.Contains(err.Error(), errTimeout.Error()) {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"catch"}`,
		),
	)
	if err == nil {
		t.Errorf("expected: %v", errTimeout)
	}
	if err != nil && !strings.Contains(err.Error(), errTimeout.Error()) {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"contract_catch"}`,
		),
	)
	if err == nil {
		t.Errorf("expected: %v", errTimeout)
	}
	if err != nil && !strings.Contains(err.Error(), errTimeout.Error()) {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"infiniteCall"}`,
		).Fail("stack overflow"),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateSize(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function infiniteLoop()
	for i = 1, 100000000000000 do
		system.setItem("key_"..i, "value_"..i)
	end
end
abi.register(infiniteLoop)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "loop", 0, definition),
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"infiniteLoop"}`,
		),
	)
	errMsg := "exceeded size of updates in the state database"
	if err == nil {
		t.Errorf("expected: %s", errMsg)
	}
	if err != nil && !strings.Contains(err.Error(), errMsg) {
		t.Error(err)
	}
}
func TestTimeoutCnt(t *testing.T) {
	timeout := 500
	src := `
function infinite_loop(n)
	while true do
	end
	return 0
end

abi.register(infinite_loop)
`
	bc, err := LoadDummyChain(
		func(d *DummyChain) {
			d.timeout = timeout // milliseconds
		},
		OnPubNet,
	)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "timeout-cnt", 0, src),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "timeout-cnt", 0, `{"Name": "infinite_loop"}`).Fail("contract timeout"),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("timeout-cnt", `{"Name": "infinite_loop"}`, "exceeded the maximum instruction count")
	if err != nil {
		t.Error(err)
	}

	src2 := `
function a()
    src = [[
while true do
end
    function b()
    end
    abi.register(b)
    ]]
    contract.deploy(src)
end

abi.register(a)
`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "timeout-cnt2", 0, src2),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "timeout-cnt2", 0, `{"Name": "a"}`).Fail("contract timeout"),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSnapshot(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
	state.var{
		counts = state.map(),
		data = state.value(),
		array = state.array(10)
	}

	function inc()
		a = system.getItem("key1")
		if (a == nil) then
			system.setItem("key1", 1)
			return
		end
		system.setItem("key1", a + 1)
		counts["key1"] = a + 1
		data:set(a+1)
		array[1] = a + 1
	end
	function query(a)
			return system.getItem("key1", a), state.getsnap(counts, "key1", a), state.getsnap(data,a), state.getsnap(array, 1, a)
	end
	function query2()
			return state.getsnap(array, 1)
	end
	abi.register(inc, query, query2)
	abi.payable(inc)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "snap", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "snap", 0, `{"Name": "inc", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "snap", 0, `{"Name": "inc", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "snap", 0, `{"Name": "inc", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("snap", `{"Name":"query"}`, "", "[3,3,3,3]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("snap", `{"Name":"query", "Args":[2]}`, "", "[1,null,null,null]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("snap", `{"Name":"query", "Args":[3]}`, "", "[2,2,2,2]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("snap", `{"Name":"query2", "Args":[]}`,
		"invalid argument at getsnap, need (state.array, index, blockheight)", "")
	if err != nil {
		t.Error(err)
	}
}

func TestKvstore(t *testing.T) {
	definition := `
	state.var{
		counts = state.map(),
		name = state.value()
	}

	function inc(key)
		if counts[key] == nil then
			counts[key] = 0
		end
		counts[key] = counts[key] + 1
	end

	function get(key)
		return counts[key]
	end

	function set(key,val)
		counts[key] = val
	end

	function setname(n)
		name:set(n)
	end

	function getname()
		return name:get()
	end

	abi.register(inc,get,set,setname,getname)`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "map", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"inc", "Args":["ktlee"]}`),
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"setname", "Args":["eve2adam"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock()
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("map", `{"Name":"get", "Args":["ktlee"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("map", `{"Name":"get", "Args":["htwo"]}`, "", "null")
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"inc", "Args":["ktlee"]}`),
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"inc", "Args":["htwo"]}`),
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"set", "Args":["wook", 100]}`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("map", `{"Name":"get", "Args":["ktlee"]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("map", `{"Name":"get", "Args":["htwo"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("map", `{"Name":"get", "Args":["wook"]}`, "", "100")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("map", `{"Name":"getname"}`, "", `"eve2adam"`)
	if err != nil {
		t.Error(err)
	}
}

// sql tests
func TestSqlConstrains(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function init()
    db.exec([[create table if not exists r (
  id integer primary key
, n integer check(n >= 10)
, nonull text not null
, only integer unique)
]])
    db.exec("insert into r values (1, 11, 'text', 1)")
	db.exec("create table if not exists s (rid integer references r(id))")
end

function pkFail()
	db.exec("insert into r values (1, 12, 'text', 2)")
end

function checkFail()
	db.exec("insert into r values (2, 9, 'text', 3)")
end

function fkFail()
	db.exec("insert into s values (2)")
end

function notNullFail()
	db.exec("insert into r values (2, 13, null, 2)")
end

function uniqueFail()
	db.exec("insert into r values (2, 13, 'text', 1)")
end

abi.register(init, pkFail, checkFail, fkFail, notNullFail, uniqueFail)`

	err = bc.ConnectBlock(
		NewLuaTxAccount(
			"ktlee",
			100000000000000000,
		),
		NewLuaTxDef(
			"ktlee",
			"constraint",
			0,
			definition,
		),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"init"}`,
		),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"pkFail"}`,
		).Fail("UNIQUE constraint failed: r.id"),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"checkFail"}`,
		).Fail("CHECK constraint failed: r"),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"fkFail"}`,
		).Fail("FOREIGN KEY constraint failed"),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"notNullFail"}`,
		).Fail("NOT NULL constraint failed: r.nonull"),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"uniqueFail"}`,
		).Fail("UNIQUE constraint failed: r.only"),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlAutoincrement(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function init()
    db.exec("create table if not exists auto_test (a integer primary key autoincrement, b text)")
	local n = db.exec("insert into auto_test(b) values (?),(?)", 10000, 1)
	assert(n == 2, "change count mismatch");
end

function get()
	db.exec("insert into auto_test(b) values ('ss')")
	assert(db.last_insert_rowid() == 3, "id is not valid")
end

abi.register(init, get)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "auto", 0, definition),
		NewLuaTxCall("ktlee", "auto", 0, `{"Name":"init"}`),
	)

	tx := NewLuaTxCall("ktlee", "auto", 0, `{"Name":"get"}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlOnConflict(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function constructor()
    db.exec("create table if not exists t (col integer primary key)")
	db.exec("insert into t values (1)")
end

function stmt_exec(stmt)
	db.exec(stmt)
end

function stmt_exec_pcall(stmt)
	pcall(db.exec, stmt)
end

function get()
	local rs = db.query("select col from t order by col")
	local t = {}
	while rs:next() do
		local col = rs:get()
		table.insert(t, col)
	end
	return t
end

abi.register(stmt_exec, stmt_exec_pcall, get)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "on_conflict", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"on_conflict",
			0,
			`{"name":"stmt_exec", "args": ["insert into t values (2)"]}`,
		),
		NewLuaTxCall(
			"ktlee",
			"on_conflict",
			0,
			`{"name":"stmt_exec", "args": ["insert into t values (3),(2),(4)"]}`,
		).Fail(`UNIQUE constraint failed: t.col`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query(
		"on_conflict",
		`{"name":"get"}`,
		"",
		`[1,2]`,
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"on_conflict",
			0,
			`{"name":"stmt_exec", "args": ["replace into t values (2)"]}`,
		),
		NewLuaTxCall(
			"ktlee",
			"on_conflict",
			0,
			`{"name":"stmt_exec", "args": ["insert or ignore into t values (3),(2),(4)"]}`,
		),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query(
		"on_conflict",
		`{"name":"get"}`,
		"",
		`[1,2,3,4]`,
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"on_conflict",
			0,
			`{"name":"stmt_exec", "args": ["insert into t values (5)"]}`,
		),
		NewLuaTxCall(
			"ktlee",
			"on_conflict",
			0,
			`{"name":"stmt_exec", "args": ["insert or rollback into t values (5)"]}`,
		).Fail("syntax error"),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query(
		"on_conflict",
		`{"name":"get"}`,
		"",
		`[1,2,3,4,5]`,
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"on_conflict",
			0,
			`{"name":"stmt_exec_pcall", "args": ["insert or fail into t values (6),(7),(5),(8),(9)"]}`,
		),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query(
		"on_conflict",
		`{"name":"get"}`,
		"",
		`[1,2,3,4,5,6,7]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlDupCol(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function get()
	db.query("select * from (select 1+1, 1+1, 1+1, 1+1, 1+1, 1+1)")
	return "success"
end

abi.register(get)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "dup_col", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"dup_col",
		`{"name":"get"}`,
		`too many duplicate column name "1+1", max: 5`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmSimple(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function createAndInsert()
    db.exec("create table if not exists dual(dummy char(1))")
	db.exec("insert into dual values ('X')")
    local insertYZ = db.prepare("insert into dual values (?),(?)")
    insertYZ:exec("Y", "Z")
end

function insertRollbackData()
	db.exec("insert into dual values ('A'),('B'),('C')")
end

function query()
    local rt = {}
    local stmt = db.prepare("select ?+1, round(?, 1), dummy || ? as col3 from dual order by col3")
    local rs = stmt:query(1, 3.14, " Hello Blockchain")
    while rs:next() do
        local col1, col2, col3 = rs:get()
        table.insert(rt, col1)
        table.insert(rt, col2)
        table.insert(rt, col3)
    end
    return rt
end

function count()
	local rs = db.query("select count(*) from dual")
	if rs:next() then
		local n = rs:get()
		--rs:next()
		return n
	else
		return "error in count()"
	end
end

function all()
    local rt = {}
    local rs = db.query("select dummy from dual order by 1")
    while rs:next() do
        local col = rs:get()
        table.insert(rt, col)
    end
    return rt
end

abi.register(createAndInsert, insertRollbackData, query, count, all)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "simple-query", 0, definition),
	)
	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`),
	)
	err = bc.Query(
		"simple-query",
		`{"Name": "query", "Args":[]}`,
		"",
		`[2,3.1,"X Hello Blockchain",2,3.1,"Y Hello Blockchain",2,3.1,"Z Hello Blockchain"]`,
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"",
		`3`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`),
	)
	err = bc.Query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"",
		`6`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.DisConnectBlock()

	err = bc.Query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"",
		`3`,
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.DisConnectBlock()
	if err != nil {
		t.Error(err)
	}
	err = bc.DisConnectBlock()
	if err != nil {
		t.Error(err)
	}

	// there is only a genesis block
	err = bc.Query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"not found contract",
		"",
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmFail(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function init()
    db.exec("create table if not exists total(n int)")
	db.exec("insert into total values (0)")
end

function add(n)
	local stmt = db.prepare("update total set n = n + ?")
	stmt:exec(n)
end

function addFail(n)
	local stmt = db.prepare("update set n = n + ?")
	stmt:exec(n)
end

function get()
	local rs = db.query("select n from total")
	rs:next()
	n = rs:get()
	return n
end
abi.register(init, add, addFail, get)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "fail", 0, definition),
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"init"}`),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[1]}`),
	)

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[2]}`),
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"addFail", "Args":[3]}`).
			Fail(`near "set": syntax error`),
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[4]}`),
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[5]}`),
	)

	err = bc.Query("fail", `{"Name":"get"}`, "", "12")
	if err != nil {
		t.Error(err)
	}

	_ = bc.DisConnectBlock()

	err = bc.Query("fail", `{"Name":"get"}`, "", "7")
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmPubNet(t *testing.T) {
	bc, err := LoadDummyChain(OnPubNet)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function createAndInsert()
    db.exec("create table if not exists dual(dummy char(1))")
	db.exec("insert into dual values ('X')")
    local insertYZ = db.prepare("insert into dual values (?),(?)")
    insertYZ:exec("Y", "Z")
end
abi.register(createAndInsert)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "simple-query", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`).Fail(`attempt to index global 'db'`),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmDateTime(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function init()
    db.exec("create table if not exists dt_test (n datetime, b bool)")
	local n = db.exec("insert into dt_test values (?, ?),(date('2004-10-24', '+1 month', '-1 day'), 0)", 10000, 1)
	assert(n == 2, "change count mismatch");
end

function nowNull()
	db.exec("insert into dt_test values (date('now'), 0)")
end

function localtimeNull()
	db.exec("insert into dt_test values (datetime('2018-05-25', ?), 1)", 'localtime')
end

function get()
	local rs = db.query("select n, b from dt_test order by 1, 2")
	local r = {}
	while rs:next() do
		local d, b = rs:get()
		table.insert(r, { date= d, bool= b })
	end
	return r
end
abi.register(init, nowNull, localtimeNull, get)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "datetime", 0, definition),
		NewLuaTxCall("ktlee", "datetime", 0, `{"Name":"init"}`),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "datetime", 0, `{"Name":"nowNull"}`),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "datetime", 0, `{"Name":"localtimeNull"}`),
	)

	err = bc.Query(
		"datetime",
		`{"Name":"get"}`,
		"",
		`[{"bool":0},{"bool":1},{"bool":1,"date":"1970-01-01 02:46:40"},{"bool":0,"date":"2004-11-23"}]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmCustomer(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function createTable()
  db.exec([[create table if not exists customer(
        id varchar(10),
        passwd varchar(20),
        name varchar(30),
        birth char(8),
        mobile varchar(20)
    )]])
end

function query(id)
    local rt = {}
    local rs = db.query("select * from customer where id like '%' || ? || '%'", id)
    while rs:next() do
        local col1, col2, col3, col4, col5 = rs:get()
        local item = {
                    id = col1,
                    passwd = col2,
                    name = col3,
                    birth = col4,
                    mobile = col5
            }
        table.insert(rt, item)
    end
    return rt
end

function insert(id , passwd, name, birth, mobile)
    local n = db.exec("insert into customer values (?,?,?,?,?)", id, passwd, name, birth, mobile)
	assert(n == 1, "insert count mismatch")
end

function update(id , passwd)
    local n = db.exec("update customer set passwd =? where id =?", passwd, id)
	assert(n == 1, "update count mismatch")
end

function delete(id)
    local n = db.exec("delete from customer where id =?", id)
	assert(n == 1, "delete count mismatch")
end

function count()
	local rs = db.query("select count(*) from customer")
	if rs:next() then
		local n = rs:get()
		return n
	else
		return "error in count()"
	end
end

abi.register(createTable, query, insert, update, delete, count)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "customer", 0, definition),
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"createTable"}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"insert", "Args":["id1","passwd1","name1","20180524","010-1234-5678"]}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"insert", "Args":["id2","passwd2","name2","20180524","010-1234-5678"]}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"update", "Args":["id2","passwd3"]}`,
		),
	)

	err = bc.Query("customer", `{"Name":"count"}`, "", "2")
	if err != nil {
		t.Error(err)
	}

	_ = bc.DisConnectBlock()

	err = bc.Query(
		"customer",
		`{"Name":"query", "Args":["id2"]}`,
		"",
		`[{"birth":"20180524","id":"id2","mobile":"010-1234-5678","name":"name2","passwd":"passwd2"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"delete", "Args":["id2"]}`,
		),
	)

	err = bc.Query(
		"customer",
		`{"Name":"query", "Args":["id2"]}`,
		"",
		`{}`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmDataType(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function createDataTypeTable()
  db.exec([[create table if not exists datatype_table(
        var1 varchar(10),
        char1 char(10),
        int1 int(5),
        float1 float(6),
        blockheight1 long
    )]])
end

function dropDataTypeTable()
   db.exec("drop table datatype_table")
end

function insertDataTypeTable()
    local stmt = db.prepare("insert into datatype_table values ('ABCD','fgh',1,3.14,?)")
    stmt:exec(system.getBlockheight())
end
function queryOrderByDesc()
    local rt = {}
    local rs = db.query("select * from datatype_table order by blockheight1 desc")
    while rs:next() do
        local col1, col2, col3, col4, col5 = rs:get()
        item = {
                    var1 = col1,
                    char1 = col2,
                    int1 = col3,
                    float1 = col4,
                    blockheight1 = col5
            }
        table.insert(rt, item)
    end
    return rt
end

function queryGroupByBlockheight1()
    local rt = {}
    local rs = db.query("select blockheight1, count(*), sum(int1), avg(float1) from datatype_table group by blockheight1")
    while rs:next() do
        local col1, col2, col3, col4 = rs:get()
        item = {
                    blockheight1 = col1,
                    count1 = col2,
                    sum_int1 = col3,
                    avg_float1 =col4
            }
        table.insert(rt, item)
    end
    return rt
end

abi.register(createDataTypeTable, dropDataTypeTable, insertDataTypeTable, queryOrderByDesc, queryGroupByBlockheight1)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "datatype", 0, definition),
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"createDataTypeTable"}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"insertDataTypeTable"}`,
		),
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"insertDataTypeTable"}`,
		),
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"insertDataTypeTable"}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"insertDataTypeTable"}`,
		),
	)

	err = bc.Query(
		"datatype",
		`{"Name":"queryOrderByDesc"}`,
		"",
		`[{"blockheight1":3,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"datatype",
		`{"Name":"queryGroupByBlockheight1"}`,
		"",
		`[{"avg_float1":3.14,"blockheight1":2,"count1":3,"sum_int1":3},{"avg_float1":3.14,"blockheight1":3,"count1":1,"sum_int1":1}]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmFunction(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function sql_func()
    local rt = {}
    local rs = db.query("select round(3.14),min(1,2,3), max(4,5,6)")
	if rs:next() then
	    local col1, col2, col3 = rs:get()
        table.insert(rt, col1)
        table.insert(rt, col2)
        table.insert(rt, col3)
        return rt
	else
		return "error in func()"
	end
end

function abs_func()
    local rt = {}
    local rs = db.query("select abs(-1),abs(0), abs(1)")
	if rs:next() then
	    local col1, col2, col3 = rs:get()
        table.insert(rt, col1)
        table.insert(rt, col2)
        table.insert(rt, col3)
        return rt
	else
		return "error in abs()"
	end
end

function typeof_func()
    local rt = {}
    local rs = db.query("select typeof(-1), typeof('abc'), typeof(3.14), typeof(null)")
	if rs:next() then
	    local col1, col2, col3, col4 = rs:get()
        table.insert(rt, col1)
        table.insert(rt, col2)
        table.insert(rt, col3)
        table.insert(rt, col4)
        return rt
	else
		return "error in typeof()"
	end
end

abi.register(sql_func, abs_func, typeof_func)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "fns", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("fns", `{"Name":"sql_func"}`, "", `[3,1,6]`)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("fns", `{"Name":"abs_func"}`, "", `[1,0,1]`)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("fns", `{"Name":"typeof_func"}`,
		"", `["integer","text","real","null"]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmBook(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function createTable()
  db.exec([[create table if not exists book (
        page number,
        contents text
    )]])

  db.exec([[create table if not exists copy_book (
        page number,
        contents text
    )]])
end

function makeBook()
   	local stmt = db.prepare("insert into book values (?,?)")
	for i = 1, 100 do    
   		stmt:exec(i, "value=" .. i*i)
    end
end

function copyBook()
    local rs = db.query("select page, contents from book order by page asc")
    while rs:next() do
        local col1, col2 = rs:get()
        local stmt_t = db.prepare("insert into copy_book values (?,?)")
        stmt_t:exec(col1, col2)
    end
end


function viewCopyBook()
    local rt = {}
    local rs = db.query("select max(page), min(contents) from copy_book")
    while rs:next() do
        local col1, col2 = rs:get()
        table.insert(rt, col1)
		table.insert(rt, col2)
    end
    return rt
end

function viewJoinBook()
    local rt = {}
    local rs = db.query([[select c.page, b.page, c.contents  
							from copy_book c, book b 
							where c.page = b.page and c.page = 10 ]])
    while rs:next() do
        local col1, col2, col3 = rs:get()
        table.insert(rt, col1)
		table.insert(rt, col2)
		table.insert(rt, col3)
    end
    return rt
end

abi.register(createTable, makeBook, copyBook, viewCopyBook, viewJoinBook)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "book", 0, definition),
		NewLuaTxCall(
			"ktlee",
			"book",
			0,
			`{"Name":"createTable"}`,
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"book",
			0,
			`{"Name":"makeBook"}`,
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"book",
			0,
			`{"Name":"copyBook"}`,
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"book",
		`{"Name":"viewCopyBook"}`,
		"",
		`[100,"value=1"]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmDateformat(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function init()
	db.exec("drop table if exists dateformat_test")
	db.exec([[create table if not exists dateformat_test
	(
		col1 date ,
		col2 datetime ,
		col3 text
	)]])
	db.exec("insert into dateformat_test values (date('2004-10-24 11:11:11'), datetime('2004-10-24 11:11:11'),strftime('%Y%m%d%H%M%S','2004-10-24 11:11:11'))")
	db.exec("insert into dateformat_test values (date(1527504338,'unixepoch'), datetime(1527504338,'unixepoch'), strftime('%Y%m%d%H%M%S',1527504338,'unixepoch') )")
end

function get()
    local rt = {}
    local rs = db.query([[select col1, col2, col3
                            from dateformat_test ]])
    while rs:next() do
        local col1, col2, col3 = rs:get()
        table.insert(rt, {col1,col2,col3} )
    end
    return rt
end

abi.register(init, get)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef(
			"ktlee",
			"data_format",
			0,
			definition,
		),
		NewLuaTxCall("ktlee", "data_format", 0, `{"Name":"init"}`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"data_format",
		`{"Name":"get"}`,
		"",
		`[["2004-10-24","2004-10-24 11:11:11","20041024111111"],["2018-05-28","2018-05-28 10:45:38","20180528104538"]]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmRecursiveData(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function r()
	local t = {}
	t["name"] = "ktlee"
	t["self"] = t
	return t
end
abi.register(r)`

	tx := NewLuaTxCall("ktlee", "r", 0, `{"Name":"r"}`)
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "r", 0, definition),
		tx,
	)
	if err == nil {
		t.Error(err)
	}
	if err.Error() != `nested table error` {
		t.Errorf("contract Call ret error :%s", err.Error())
	}
}

func TestSqlJdbc(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function init()
    db.exec("create table if not exists total(a int, b int, c text)")
    db.exec("insert into total(a,c) values (1,2)")
    db.exec("insert into total values (2,2,3)")
    db.exec("insert into total values (3,2,3)")
    db.exec("insert into total values (4,2,3)")
    db.exec("insert into total values (5,2,3)")
    db.exec("insert into total values (6,2,3)")
    db.exec("insert into total values (7,2,3)")
end

function exec(sql, ...)
    local stmt = db.prepare(sql)
    stmt:exec(...)
end

function query(sql, ...)
    local stmt = db.prepare(sql)
	local rs = stmt:query(...)
    local r = {}
    local colcnt = rs:colcnt()
	local colmetas
    while rs:next() do
		if colmetas == nil then
			colmetas = stmt:column_info()
		end

		local k = {rs:get()}
		for i = 1, colcnt do
			if k[i] == nil then
				k[i] = {}
			end
        end
        table.insert(r, k)
    end
--  if (#r == 0) then
--      return {"colcnt":0, "rowcnt":0}
--  end

    return {snap=db.getsnap(), colcnt=colcnt, rowcnt=#r, data=r, colmetas=colmetas}
end

function queryS(snap, sql, ...)
	db.open_with_snapshot(snap)

    local stmt = db.prepare(sql)
	local rs = stmt:query(...)
    local r = {}
    local colcnt = rs:colcnt()
	local colmetas
    while rs:next() do
		if colmetas == nil then
			colmetas = stmt:column_info()
		end

		local k = {rs:get()}
		for i = 1, colcnt do
			if k[i] == nil then
				k[i] = {}
			end
        end
        table.insert(r, k)
    end
--  if (#r == 0) then
--      return {"colcnt":0, "rowcnt":0}
--  end

    return {snap=db.getsnap(), colcnt=colcnt, rowcnt=#r, data=r, colmetas=colmetas}
end
function getmeta(sql)
    local stmt = db.prepare(sql)

	return stmt:column_info(), stmt:bind_param_cnt()
end
abi.register(init, exec, query, getmeta, queryS)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "jdbc", 0, definition),
		NewLuaTxCall("ktlee", "jdbc", 0, `{"Name":"init"}`),
	)

	err = bc.Query("jdbc", `{"Name":"query", "Args":["select a,b,c from total"]}`, "",
		`{"colcnt":3,"colmetas":{"colcnt":3,"decltypes":["int","int","text"],"names":["a","b","c"]},"data":[[1,{},"2"],[2,2,"3"],[3,2,"3"],[4,2,"3"],[5,2,"3"],[6,2,"3"],[7,2,"3"]],"rowcnt":7,"snap":"2"}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("jdbc", `{"Name":"getmeta", "Args":["select a,b,?+1 from total"]}`, "",
		`[{"colcnt":3,"decltypes":["int","int",""],"names":["a","b","?+1"]},1]`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "jdbc", 0, `{"Name": "exec", "Args":["insert into total values (3,4,5)"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("jdbc", `{"Name":"query", "Args":["select a,b,c from total"]}`, "",
		`{"colcnt":3,"colmetas":{"colcnt":3,"decltypes":["int","int","text"],"names":["a","b","c"]},"data":[[1,{},"2"],[2,2,"3"],[3,2,"3"],[4,2,"3"],[5,2,"3"],[6,2,"3"],[7,2,"3"],[3,4,"5"]],"rowcnt":8,"snap":"3"}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("jdbc", `{"Name":"queryS", "Args":["2", "select a,b,c from total"]}`, "",
		`{"colcnt":3,"colmetas":{"colcnt":3,"decltypes":["int","int","text"],"names":["a","b","c"]},"data":[[1,{},"2"],[2,2,"3"],[3,2,"3"],[4,2,"3"],[5,2,"3"],[6,2,"3"],[7,2,"3"]],"rowcnt":7,"snap":"3"}`)
	if err != nil {
		t.Error(err)
	}
}

// type Tests
func TestTypeOP(t *testing.T) {
	src, err := ioutil.ReadFile("op.lua")
	if err != nil {
		t.Error(err)
	}
	bc, err := LoadDummyChain(OnPubNet)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()
	balance, _ := new(big.Int).SetString("10000000000000000", 10)
	err = bc.ConnectBlock(
		NewLuaTxAccountBig("ktlee", balance),
		NewLuaTxDef("ktlee", "op", 0, string(src)),
	)
	if err != nil {
		t.Error(err)
	}
	state, err := bc.GetAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	bal := state.GetBalanceBigInt().Uint64()
	tx := NewLuaTxCall("ktlee", "op", 0, `{"Name": "main"}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	r := bc.GetReceipt(tx.Hash())
	expectedFee := uint64(117610)
	if r.GetGasUsed() != expectedFee {
		t.Errorf("expected: %d, but got: %d", expectedFee, r.GetGasUsed())
	}
	state, err = bc.GetAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	if bal-expectedFee != state.GetBalanceBigInt().Uint64() {
		t.Errorf(
			"expected: %d, but got: %d",
			bal-expectedFee,
			state.GetBalanceBigInt().Uint64(),
		)
	}
}

func TestTypeBF(t *testing.T) {
	src, err := ioutil.ReadFile("bf.lua")
	if err != nil {
		t.Error(err)
	}
	bc, err := LoadDummyChain(OnPubNet)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()
	balance, _ := new(big.Int).SetString("10000000000000000", 10)
	err = bc.ConnectBlock(
		NewLuaTxAccountBig("ktlee", balance),
		NewLuaTxDef("ktlee", "op", 0, string(src)),
	)
	if err != nil {
		t.Error(err)
	}
	state, err := bc.GetAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}

	feeTest := func(expectedFee uint64) {
		bal := state.GetBalanceBigInt().Uint64()
		tx := NewLuaTxCall("ktlee", "op", 0, `{"Name": "main"}`)
		err = bc.ConnectBlock(tx)
		if err != nil {
			t.Error(err)
		}
		r := bc.GetReceipt(tx.Hash())
		if r.GetGasUsed() != expectedFee {
			t.Errorf("expected: %d, but got: %d", expectedFee, r.GetGasUsed())
		}
		state, err = bc.GetAccountState("ktlee")
		if err != nil {
			t.Error(err)
		}
		if bal-expectedFee != state.GetBalanceBigInt().Uint64() {
			t.Errorf(
				"expected: %d, but got: %d",
				bal-expectedFee,
				state.GetBalanceBigInt().Uint64(),
			)
		}
	}

	feeTest(47456244)

	OldV3 := HardforkConfig.V3
	HardforkConfig.V3 = types.BlockNo(0)
	feeTest(47513803)
	HardforkConfig.V3 = OldV3
}

func TestTypeMaxString(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function oom()
	 local s = "hello"

	 while 1 do
		 s = s .. s
	 end
end

function p()
	pcall(oom)
end

function cp()
	contract.pcall(oom)
end
abi.register(oom, p, cp)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "oom", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	errMsg := "not enough memory"
	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"oom",
			0,
			`{"Name":"oom"}`,
		).Fail(errMsg),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"oom",
			0,
			`{"Name":"p"}`,
		).Fail(errMsg),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"oom",
			0,
			`{"Name":"cp"}`,
		).Fail(errMsg),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestTypeMaxStringOnPubNet(t *testing.T) {
	bc, err := LoadDummyChain(OnPubNet)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function oom()
	 local s = "hello"

	 while 1 do
		 s = s .. s
	 end
end

function p()
	pcall(oom)
end

function cp()
	contract.pcall(oom)
end
abi.register(oom, p, cp)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "oom", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	errMsg := "string length overflow"
	errMsg1 := "not enough memory"
	var travis bool
	if os.Getenv("TRAVIS") == "true" {
		travis = true
	}
	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"oom",
			0,
			`{"Name":"oom"}`,
		),
	)
	if err == nil {
		t.Errorf("expected: %s", errMsg)
	} else if !strings.Contains(err.Error(), errMsg) && !strings.Contains(err.Error(), errMsg1) {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"oom",
			0,
			`{"Name":"p"}`,
		),
	)
	if err != nil && (!travis || !strings.Contains(err.Error(), errMsg1)) {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"oom",
			0,
			`{"Name":"cp"}`,
		),
	)
	if err != nil && (!travis || !strings.Contains(err.Error(), errMsg1)) {
		t.Error(err)
	}
}

func TestTypeNsec(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
	function test_nsec()
		system.print(nsec())
	end
	abi.register(test_nsec)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "nsec", 0, definition),
	)
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "nsec", 0, `{"Name": "test_nsec"}`).Fail(`attempt to call global 'nsec' (a nil value)`),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestTypeUtf(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
	function string.tohex(str)
    return (str:gsub('.', function (c)
        return string.format('%02X', string.byte(c))
    end))
	end

	function query()
		assert (utf8.char(256) == json.decode('"\\u0100"'), "test1")
		a = utf8.char(256,128)
		b = utf8.char(256,10000,45)
		assert(string.len(a) == 4 and utf8.len(a) == 2, "test2")

		for p,c in utf8.codes(a) do
			if p == 1 then
				assert(c == 256, "test11")
			else
				assert(c == 128, "test12")
			end
		end
		assert(utf8.offset(b,1)==1, "test3")
		assert(utf8.offset(b,2)==3, "test4")
		assert(utf8.offset(b,3)==6, "test5")

		assert(utf8.codepoint(b,1)==256, "test6")

		k1, k2, k3 = utf8.codepoint(b,1,3)
		assert(k1 == 256 and k2 == 10000 and k3 == nil, "test7" .. k1 .. k2)
		
		k1, k2, k3 = utf8.codepoint(b,1,6)
		assert(k1 == 256 and k2 == 10000 and k3 == 45, "test7" .. k1 .. k2 .. k3)
	end

	function query2()
		a = bignum.number(1000000000000)
		b = bignum.number(0)
		return (bignum.tobyte(a)):tohex(), (bignum.tobyte(b)):tohex()
	end

	function query3()
		a = bignum.number(-1)
		return (bignum.tobyte(a)):tohex()
	end
	abi.register(query, query2, query3)
	`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "utf", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("utf", `{"Name":"query"}`, "", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("utf", `{"Name":"query2"}`, "", `["E8D4A51000","00"]`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("utf", `{"Name":"query3"}`, "bignum not allowed negative value", "")
	if err != nil {
		t.Error(err)
	}
}

func TestTypeDupVar(t *testing.T) {
	dupVar := `
state.var{
	Var1 = state.value(),
}
function GetVar1()
	return Var1:get()
end
state.var{
	Var1 = state.value(),
}
abi.register(GetVar1)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 1000000000000000000),
	)
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "dupVar", 0, dupVar),
	)
	if err == nil {
		t.Error("duplicated variable: 'Var1'")
	}
	if !strings.Contains(err.Error(), "duplicated variable: 'Var1'") {
		t.Error(err)
	}
	dupVar = `
state.var{
	Var1 = state.value(),
}
function GetVar1()
	return Var1:get()
end
function Work()
	state.var{
		Var1 = state.value(),
	}
end
abi.register(GetVar1, Work)
`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "dupVar1", 0, dupVar),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "dupVar1", 0, `{"Name": "Work"}`).Fail("duplicated variable: 'Var1'"),
	)

	if err != nil {
		t.Error(err)
	}
}

func TestTypeInvalidKey(t *testing.T) {
	src := `
state.var {
	h = state.map(),
	arr = state.array(10),
	v = state.value()
}

t = {}

function key_table()
	local k = {}
	t[k] = "table"
end

function key_func()
	t[key_table] = "function"
end

function key_statemap(key)
	t[h] = "state.map"
end

function key_statearray(key)
	t[arr] = "state.array"
end

function key_statevalue(key)
	t[v] = "state.value"
end

function key_upval(key)
	local k = {}
	local f = function()
		t[k] = "upval"
	end
	f()
end

function key_nil(key)
	h[nil] = "nil"
end

abi.register(key_table, key_func, key_statemap, key_statearray, key_statevalue, key_upval, key_nil)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "invalidkey", 0, src),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_table"}`).Fail(
			"cannot use 'table' as a key",
		),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_func"}`).Fail(
			"cannot use 'function' as a key",
		),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_statemap"}`).Fail(
			"cannot use 'userdata' as a key",
		),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_statearray"}`).Fail(
			"cannot use 'userdata' as a key",
		),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_statevalue"}`).Fail(
			"cannot use 'userdata' as a key",
		),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_upval"}`).Fail(
			"cannot use 'table' as a key",
		),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "invalidkey", 0, `{"Name":"key_nil"}`).Fail(
			"invalid key type: 'nil', state.map: 'h'",
		),
	)
	if err != nil {
		t.Error(err)
	}
}
func TestTypeByteKey(t *testing.T) {
	bk := `
state.var {
    c = state.map(),
}

function constructor()
    c[fromhex('00')] = "kk"
    c[fromhex('61')] = "kk"
    system.setItem(fromhex('00'), "kk")
end

function fromhex(str)
    return (str:gsub('..', function (cc)
        return string.char(tonumber(cc, 16))
    end))
end
function get()
	return c[fromhex('00')], system.getItem(fromhex('00')), system.getItem(fromhex('0000'))
end
function getcre()
	return system.getCreator()
end
abi.register(get, getcre)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "bk", 0, bk),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("bk", `{"Name":"get"}`, "", `["kk","kk"]`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("bk", `{"Name":"getcre"}`, "", `"Amg6nZWXKB6YpNgBPv9atcjdm6hnFvs5wMdRgb2e9DmaF5g9muF2"`)
	if err != nil {
		t.Error(err)
	}
}

func TestTypeArray(t *testing.T) {
	definition := `
	state.var{
		counts = state.array(10)
	}

	function inc(key)
		if counts[key] == nil then
			counts[key] = 0
		end
		counts[key] = counts[key] + 1
	end

	function get(key)
		return counts[key]
	end

	function set(key,val)
		counts[key] = val
	end

	function len()
		return counts:length()
	end

	function iter()
		local rv = {}
		for i, v in counts:ipairs() do 
			if v == nil then
				rv[i] = "nil"
			else
				rv[i] = v
			end
		end
		return rv
	end

	abi.register(inc,get,set,len,iter)`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "array", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[1]}`),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[0]}`).Fail("index out of range"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[1]}`),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[1.00000001]}`).Fail("integer expected, got number"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":["1"]}`).Fail("integer expected, got string)"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[true]}`).Fail("integer expected, got boolean"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[[1, 2]]}`).Fail("integer expected, got table"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[null]}`).Fail("integer expected, got nil)"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[{}]}`).Fail("integer expected, got table)"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[""]}`).Fail("integer expected, got string)"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"set", "Args":[2,"ktlee"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"get", "Args":[11]}`, "index out of range", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"get", "Args":[1]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"get", "Args":[2]}`, "", `"ktlee"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"len"}`, "", `10`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"iter"}`, "", `[2,"ktlee","nil","nil","nil","nil","nil","nil","nil","nil"]`)
	if err != nil {
		t.Error(err)
	}
	overflow := `
	state.var{
		counts = state.array(1000000000000)
	}

	function get()
		return "hello"
	end
	
	abi.register(get)
	`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "overflow", 0, overflow),
	)
	errMsg := "integer expected, got number"
	if err == nil {
		t.Errorf("expected: '%s', but got: nil", errMsg)
	} else if !strings.Contains(err.Error(), errMsg) {
		t.Errorf("expected: %s, but got: %s", errMsg, err.Error())
	}
}

func TestTypeMultiArray(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
	state.var{
		mcounts = state.map(2),
		array = state.array(10, 11),
		tcounts = state.map(3)
	}

	function inc()
		a = system.getItem("key1")
		if (a == nil) then
			system.setItem("key1", 1)
			return
		end
		system.setItem("key1", a + 1)
		mcounts[system.getSender()]["key1"] = a + 1
		array[1][10] = "k"
		array[10][5] = "l"
		tcounts[0][0][0] = 2
	end
	function query(a)
		return system.getItem("key1"), mcounts[a]["key1"], tcounts[0][0][0], tcounts[1][2][3], array:length(), array[1]:length()
	end
	function del()
		tcounts[0][0]:delete(0)
		tcounts[1][2]:delete(3)
	end
	function iter(a)
		local rv = {}
		for i, x in array:ipairs() do 
			for j, y in x:ipairs() do 
				if y ~= nil then
					rv[i..","..j] =  y
				end
			end
		end
		return rv
	end

	function seterror()
		rv, err = pcall(function () mcounts[1]["k2y1"] = 4 end)
		assert(rv == false and string.find(err, "string expected, got number"))
		rv, err = pcall(function () mcounts["middle"] = 4 end)
		assert(rv == false and string.find(err, "not permitted to set intermediate dimension of map"))
		rv, err = pcall(function () array[1] = 4 end)
		assert(rv == false and string.find(err, "not permitted to set intermediate dimension of array"))
		rv, err = pcall(function () tcounts[0]:delete(0) end)
		assert(rv == false and string.find(err, "not permitted to set intermediate dimension of map"))
		rv, err = pcall(function () tcounts[0][1]:delete() end)
		assert(rv == false and string.find(err, "invalid key type: 'no value', state.map: 'tcounts'"))
		rv, err = pcall(function () array[0]:append(2) end)
		assert(rv == false and string.find(err, "the fixed array cannot use 'append' method"))
		rv, err = pcall(function () state.var {k = state.map(6)} end)
		assert(rv == false and string.find(err, "dimension over max limit"), err)
		rv, err = pcall(function () state.var {k = state.array(1,2,3,4,5,6)} end)
		assert(rv == false and string.find(err, "dimension over max limit"), err)
	end

	abi.register(inc, query, iter, seterror, del)
	abi.payable(inc)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "ma", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "ma", 0, `{"Name": "inc", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "ma", 0, `{"Name": "inc", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("ma", fmt.Sprintf(`{"Name":"query", "Args":["%s"]}`,
		types.EncodeAddress(strHash("ktlee"))), "", "[2,2,2,null,10,11]")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "ma", 0, `{"Name": "del", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("ma", fmt.Sprintf(`{"Name":"query", "Args":["%s"]}`,
		types.EncodeAddress(strHash("ktlee"))), "", "[2,2,null,null,10,11]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("ma", `{"Name":"iter"}`, "", `{"1,10":"k","10,5":"l"}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("ma", `{"Name":"seterror"}`, "", ``)
	if err != nil {
		t.Error(err)
	}
	definition2 := `
state.var {
  -- global map
  alpha = state.map(2),
  beta = state.map(2),
}

function constructor()

  local d = alpha["dict"]
  d["a"] = "A"
  alpha["dict"]["b"] = "B"

  assert(alpha["dict"]["a"]=="A" and alpha["dict"]["b"]=="B")

  -- with local variable
  local d2 = beta["dict"]
  d2["a"] = "A"
  d2["value"] = "v0"
  beta["dict"]["b"] = "B"
  beta["dict"]["value"] = "v1"
  assert(beta["dict"]["a"]=="A" and beta["dict"]["b"]=="B" and beta["dict"]["value"]=="v1")
end

function abc()
  local d = alpha["dict"]
  d["c"] = "C"
  alpha["dict"]["d"] = "D"

  local d = beta["dict"]
  d["a"] = "A"
  d["value"] = "v2"
  beta["dict"]["b"] = "B"
  beta["dict"]["value"] = "v3"
  return alpha["dict"]["c"], alpha["dict"]["d"], beta["dict"]["a"], beta["dict"]["b"], beta["dict"]["value"]
end

function query()
  return alpha["dict"]["a"], alpha["dict"]["b"], alpha["dict"]["c"], alpha["dict"]["d"], 
beta["dict"]["a"], beta["dict"]["b"], beta["dict"]["value"] 
end

abi.register(abc, query)
`
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "ma", 0, definition2),
	)
	err = bc.Query("ma", `{"Name":"query", "Args":[]}`,
		"", `["A","B",null,null,"A","B","v1"]`)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "ma", 0, `{"Name": "abc", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `["C","D","A","B","v3"]` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.Query("ma", `{"Name":"query", "Args":[]}`,
		"", `["A","B","C","D","A","B","v3"]`)
	if err != nil {
		t.Error(err)
	}
}

func TestTypeArrayArg(t *testing.T) {
	definition1 := `
	function copy(arr)
		assert(type(arr) == "table", "table expected")
		local rv = {}
		for i, v in ipairs(arr) do
			table.insert(rv, i, v)
        end
		return rv
	end
	function two_arr(arr1, arr2)
		assert(type(arr1) == "table", "table expected")
		assert(type(arr2) == "table", "table expected")
		local rv = {}
		table.insert(rv, 1, #arr1)
		table.insert(rv, 2, #arr2)
		return rv
	end
	function mixed_args(arr1, map1, n)
		assert(type(arr1) == "table", "table expected")
		assert(type(map1) == "table", "table expected")
		local rv = {}
		table.insert(rv, 1, arr1)
		table.insert(rv, 2, map1)
		table.insert(rv, 3, n)
		return rv
	end

	abi.register(copy, two_arr, mixed_args)
	`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "a", 0, definition1),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "copy", "Args":[1, 2, 3]}`, "table expected", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "copy", "Args":[[1, 2, 3]]}`, "", "[1,2,3]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "two_arr", "Args":[[1, 2, 3],[4, 5]]}`, "", "[3,2]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "mixed_args", "Args":[[1, 2, 3], {"name": "kslee", "age": 39}, 7]}`,
		"",
		`[[1,2,3],{"age":39,"name":"kslee"},7]`,
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "mixed_args", "Args":[
[[1, 2, 3],["first", "second"]],
{"name": "kslee", "age": 39, "address": {"state": "XXX-do", "city": "YYY-si"}},
"end"
]}`,
		"",
		`[[[1,2,3],["first","second"]],{"address":{"city":"YYY-si","state":"XXX-do"},"age":39,"name":"kslee"},"end"]`,
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "mixed_args", "Args":[
[{"name": "wook", "age": 50}, {"name": "hook", "age": 42}],
{"name": "kslee", "age": 39, "scores": [10, 20, 30, 40, 50]},
"hmm..."
]}`,
		"",
		`[[{"age":50,"name":"wook"},{"age":42,"name":"hook"}],{"age":39,"name":"kslee","scores":[10,20,30,40,50]},"hmm..."]`,
	)
	if err != nil {
		t.Error(err)
	}
}

// end of test-cases

func TestTypeMapKey(t *testing.T) {
	definition := `
	state.var{
		counts = state.map()
	}
	function setCount(key, value)
		counts[key] = value
	end
	function getCount(key)
		return counts[key]
	end
	function delCount(key)
		counts:delete(key)
	end
	abi.register(setCount, getCount, delCount)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "a", 0, definition),
	)

	err = bc.Query("a", `{"Name":"getCount", "Args":[1]}`, "", "null")
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":[1, 10]}`),
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":["1", 20]}`).Fail("(number expected, got string)"),
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":[1.1, 30]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":["1"]}`, "(number expected, got string)", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":[1]}`, "", "10")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":[1.1]}`, "", "30")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "a", 0,
			`{"Name":"setCount", "Args":[true, 40]}`,
		).Fail(`invalid key type: 'boolean', state.map: 'counts'`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"delCount", "Args":[1.1]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":[1.1]}`, "", "null")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":[2]}`, "", "null")
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "x", 0, definition),
	)
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":["1", 10]}`),
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":[1, 20]}`).Fail("string expected, got number)"),
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":["third", 30]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("x", `{"Name":"getCount", "Args":["1"]}`, "", "10")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("x", `{"Name":"getCount", "Args":["third"]}`, "", "30")
	if err != nil {
		t.Error(err)
	}
}

func TestTypeStateVarFieldUpdate(t *testing.T) {
	src := `
state.var{
   Person = state.value()
}

function constructor()
  Person:set({ name = "kslee", age = 38, address = "blahblah..." })
end

function InvalidUpdateAge(age)
  Person:get().age = age
end

function ValidUpdateAge(age)
  local p = Person:get()
  p.age = age
  Person:set(p)
end

function GetPerson()
  return Person:get()
end

abi.register(InvalidUpdateAge, ValidUpdateAge, GetPerson)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "c", 0, src),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "c", 0, `{"Name":"InvalidUpdateAge", "Args":[10]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("c", `{"Name":"GetPerson"}`, "",
		`{"address":"blahblah...","age":38,"name":"kslee"}`,
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "c", 0, `{"Name":"ValidUpdateAge", "Args":[10]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("c", `{"Name":"GetPerson"}`, "",
		`{"address":"blahblah...","age":10,"name":"kslee"}`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestTypeDatetime(t *testing.T) {
	src := `
state.var {
    cdate = state.value()
}

function constructor()
	cdate:set(906000490)
end

function CreateDate()
	return system.date("%c", cdate:get())
end

function Extract(fmt)
	return system.date(fmt, cdate:get())
end

function Difftime()
	system.print(system.date("%c", cdate:get()))
	s = system.date("*t", cdate:get())
	system.print(s)
	s.hour = 2 
	s.min = 0
	s.sec = 0
	system.print(system.date("*t", system.time(s)))
	return system.difftime(cdate:get(), system.time(s))
end

abi.register(CreateDate, Extract, Difftime)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "datetime", 0, src),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "CreateDate"}`, "", `"1998-09-17 02:48:10"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%x"]}`, "", `"09/17/98"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%X"]}`, "", `"02:48:10"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%A"]}`, "", `"Thursday"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%I:%M:%S %p"]}`, "", `"02:48:10 AM"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Difftime"}`, "", `2890`)
	if err != nil {
		t.Error(err)
	}
}

func TestTypeDynamicArray(t *testing.T) {
	zeroLen := `
state.var {
    fixedArray = state.array(0)
}

function Length()
	return fixedArray:length()
end

abi.register(Length)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
	)
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "zeroLen", 0, zeroLen),
	)
	if err == nil {
		t.Error("expected: the array length must be greater than zero")
	}
	if !strings.Contains(err.Error(), "the array length must be greater than zero") {
		t.Errorf(err.Error())
	}

	dArr := `
state.var {
    dArr = state.array()
}

function Append(val)
	dArr:append(val)
end

function Get(idx)
	return dArr[idx]
end

function Set(idx, val)
	dArr[idx] = val
end

function Length()
	return dArr:length()
end

abi.register(Append, Get, Set, Length)
`
	tx := NewLuaTxDef("ktlee", "dArr", 0, dArr)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Length"}`, "", "0")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [10]}`),
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [20]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Get", "Args": [1]}`, "", "10")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Get", "Args": [2]}`, "", "20")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Get", "Args": [3]}`, "index out of range", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Length"}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [30]}`),
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [40]}`),
	)
	err = bc.Query("dArr", `{"Name": "Length"}`, "", "4")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Set", "Args": [3, 50]}`),
	)
	err = bc.Query("dArr", `{"Name": "Get", "Args": [3]}`, "", "50")
	if err != nil {
		t.Error(err)
	}
}

func TestTypeCrypto(t *testing.T) {
	src := `
function get(a)
	return crypto.sha256(a)
end

function checkEther()
	return crypto.ecverify("0xce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008",
"0x90f27b8b488db00b00606796d2987f6a5f59ae62ea05effe84fef5b8b0e549984a691139ad57a3f0b906637673aa2f63d1f55cb1a69199d4009eea23ceaddc9301",
"0xbcf9061f21320aa7e824b00d0152398b2d7a6e44")
end

function checkAergo()
	return crypto.ecverify("11e96f2b58622a0ce815b81f94da04ae7a17ba17602feb1fd5afa4b9f2467960",
"304402202e6d5664a87c2e29856bf8ff8b47caf44169a2a4a135edd459640be5b1b6ef8102200d8ea1f6f9ecdb7b520cdb3cc6816d773df47a1820d43adb4b74fb879fb27402",
"AmPbWrQbtQrCaJqLWdMtfk2KiN83m2HFpBbQQSTxqqchVv58o82i")
end

function keccak256(s)
	return crypto.keccak256(s)
end

abi.register(get, checkEther, checkAergo, keccak256)
`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "crypto", 0, src),
	)
	err = bc.Query("crypto", `{"Name": "get", "Args" : ["ab\u0000\u442a"]}`, "", `"0xc58f6dca13e4bba90a326d8605042862fe87c63a64a9dd0e95608a2ee68dc6f0"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("crypto", `{"Name": "get", "Args" : ["0x616200e490aa"]}`, "", `"0xc58f6dca13e4bba90a326d8605042862fe87c63a64a9dd0e95608a2ee68dc6f0"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("crypto", `{"Name": "checkEther", "Args" : []}`, "", `true`)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("crypto", `{"Name": "checkAergo", "Args" : []}`, "", `true`)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"crypto",
		`{"Name": "keccak256", "Args" : ["0x616263"]}`,
		"",
		`"0x4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45"`)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"crypto",
		`{"Name": "keccak256", "Args" : ["0x616572676F"]}`,
		"",
		`"0xe98bb03ab37161f8bbfe131f711dcccf3002a9cd9ec31bbd52edf181f7ab09a0"`)
	if err != nil {
		t.Error(err)
	}
}

func TestTypeBignum(t *testing.T) {
	bigNum := `
function test(addr)
	bal = contract.balance()
	contract.send(addr, bal / 2)
	return contract.balance()
end

function sendS(addr)
	contract.send(addr, "1 gaer 99999")
	return contract.balance()
end

function testBignum()
	bg = bignum.number("999999999999999999999999999999")
	system.setItem("big", bg)
	bi = system.getItem("big")
	return tostring(bi)
end

function argBignum(a)
	b = a + 1
	return tostring(b)
end

function calladdBignum(addr, a)
	return tostring(contract.call(addr, "add", a, 2) + 3)
end

function checkBignum()
	a = 1
	b = bignum.number(1)
	
	return bignum.isbignum(a), bignum.isbignum(b), bignum.isbignum("2333")
end
function calcBignum()
	bg1 = bignum.number("999999999999999999999999999999")
	bg2 = bignum.number("999999999999999999999999999999")
	bg3 = bg1 + bg2
	bg4 = bg1 * 2
	bg5 = 2 * bg1
	n1 = 999999999999999
	system.print(n1)
	bg6 = bignum.number(n1)
	assert (bg3 == bg4 and bg4 == bg5)
	bg5 = bg1 - bg3
	-- ispositive() and isnegative()
	assert (bignum.isnegative(bg5) and bg5 == bignum.neg(bg1))
	assert (bignum.ispositive(bg1) and bignum.ispositive(bg3))
	assert (not bignum.ispositive(bg5) and not bignum.iszero(bg5))
	assert (not bignum.isnegative(bg1) and not bignum.iszero(bg1))
	-- ispositive() and isneg()
	assert (bignum.isneg(bg5) and bg5 == bignum.neg(bg1))
	assert (bignum.ispositive(bg1) and bignum.ispositive(bg3))
	assert (not bignum.ispositive(bg5) and not bignum.iszero(bg5))
	assert (not bignum.isneg(bg1) and not bignum.iszero(bg1))
	system.print(bg3, bg5, bg6)
	bg6 = bignum.number(1)
	assert (bg6 > bg5)
	a = bignum.number(2)
	b = bignum.number(8)
	pow = a ^ b
	system.print(pow, a, b)
	assert(pow == bignum.number(256) and a == bignum.number(2) and b == bignum.number(8))
	assert(bignum.compare(bg6, 1) == 0)
	system.print((bg6 == 1), bignum.isbignum(pow))
	div1 = bignum.number(3)/2
	assert(bignum.compare(div1, 1) == 0)
	div = bg6 / 0
end

function negativeBignum()
	bg1 = bignum.number("-2")
	bg2 = bignum.sqrt(bg1)
end

function byteBignum()
	 state.var {
        value = state.value()
    }
	value = bignum.tobyte(bignum.number("177"))
	return bignum.frombyte(value)
end

function constructor()
end

abi.register(test, sendS, testBignum, argBignum, calladdBignum, checkBignum, calcBignum, negativeBignum, byteBignum)
abi.payable(constructor)
`
	callee := `
	function add(a, b)
		return a + b
	end
	abi.register(add)
	`
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "bigNum", 50000000000, bigNum),
		NewLuaTxDef("ktlee", "add", 0, callee),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "bigNum", 0, fmt.Sprintf(`{"Name":"test", "Args":["%s"]}`, types.EncodeAddress(strHash("ktlee"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `"25000000000"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	tx = NewLuaTxCall("ktlee", "bigNum", 0, fmt.Sprintf(`{"Name":"sendS", "Args":["%s"]}`, types.EncodeAddress(strHash("ktlee"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt = bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `"23999900001"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	tx = NewLuaTxCall("ktlee", "bigNum", 0, `{"Name":"testBignum", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt = bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `"999999999999999999999999999999"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.Query("bigNum", `{"Name":"argBignum", "Args":[{"_bignum":"99999999999999999999999999"}]}`, "", `"100000000000000000000000000"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("bigNum", fmt.Sprintf(`{"Name":"calladdBignum", "Args":["%s", {"_bignum":"999999999999999999"}]}`, types.EncodeAddress(strHash("add"))), "", `"1000000000000000004"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("bigNum", `{"Name":"checkBignum"}`, "", `[false,true,false]`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("bigNum", `{"Name":"calcBignum"}`, "bignum divide by zero", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("bigNum", `{"Name":"negativeBignum"}`, "bignum not allowed negative value", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("bigNum", `{"Name":"byteBignum"}`, "", `{"_bignum":"177"}`)
	if err != nil {
		t.Error(err)
	}
}

func checkRandomIntValue(v string, min, max int) error {
	n, _ := strconv.Atoi(v)
	if n < min || n > max {
		return errors.New("out of range")
	}
	return nil
}

func TestTypeRandom(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	random := `
function random(...)
	return system.random(...)
end
abi.register(random)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "random", 0, random),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[]}`).Fail(
			"1 or 2 arguments required",
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"random",
			0,
			`{"Name": "random", "Args":[0]}`).Fail("the maximum value must be greater than zero"),
	)
	if err != nil {
		t.Error(err)
	}

	tx := NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[3]}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.GetReceipt(tx.Hash())
	err = checkRandomIntValue(receipt.GetRet(), 1, 3)
	if err != nil {
		t.Errorf("error: %s, return value: %s", err.Error(), receipt.GetRet())
	}

	tx = NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[3, 10]}`)
	err = bc.ConnectBlock(tx)
	receipt = bc.GetReceipt(tx.Hash())
	err = checkRandomIntValue(receipt.GetRet(), 3, 10)
	if err != nil {
		t.Errorf("error: %s, return value: %s", err.Error(), receipt.GetRet())
	}

	err = bc.Query("random", `{"Name": "random", "Args":[1]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("random", `{"Name": "random", "Args":[4,4]}`, "", "4")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("random", `{"Name": "random", "Args":[0,4]}`, "system.random: the minimum value must be greater than zero", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("random", `{"Name": "random", "Args":[3,1]}`, "system.random: the maximum value must be greater than the minimum value", "")
	if err != nil {
		t.Error(err)
	}
}

func TestTypeSparseTable(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function is_table_equal(t1,t2,ignore_mt)
   local ty1 = type(t1)
   local ty2 = type(t2)
   if ty1 ~= ty2 then return false end
   -- non-table types can be directly compared
   if ty1 ~= 'table' and ty2 ~= 'table' then return t1 == t2 end
   -- as well as tables which have the metamethod __eq
   local mt = getmetatable(t1)
   if not ignore_mt and mt and mt.__eq then return t1 == t2 end
   for k1,v1 in pairs(t1) do
      local v2 = t2[k1]
      if v2 == nil or not is_table_equal(v1,v2) then return false end
   end
   for k2,v2 in pairs(t2) do
      local v1 = t1[k2]
      if v1 == nil or not is_table_equal(v1,v2) then return false end
   end
   return true
end

function r()
	local t = {}
	t[10000] = "1234"
	system.setItem("k", t)
	k = system.getItem("k")
	if is_table_equal(t, k, false) then
		return 1
    end
	return 0
end
abi.register(r)`

	tx := NewLuaTxCall("ktlee", "r", 0, `{"Name":"r"}`)
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "r", 0, definition),
		tx,
	)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `1` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}
func TestTypeBigTable(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	bigSrc := `
function constructor()
    db.exec("create table if not exists table1 (cid integer PRIMARY KEY, rgtime datetime)")
    db.exec("insert into table1 (rgtime) values (datetime('2018-10-30 16:00:00'))")
end

function inserts(n)
    for i = 1, n do
        db.exec("insert into table1 (rgtime) select rgtime from table1")
    end
end

abi.register(inserts)
`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "big", 0, bigSrc),
	)
	if err != nil {
		t.Error(err)
	}

	// About 900MB
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "big", 0, `{"Name": "inserts", "Args":[25]}`),
	)
	if err != nil {
		t.Error(err)
	}

	SetStateSQLMaxDBSize(20)

	bigSrc = `
function constructor()
    db.exec("create table if not exists aergojdbc001 (name text, yyyymmdd text)")
    db.exec("insert into aergojdbc001 values ('홍길동', '20191007')")
    db.exec("insert into aergojdbc001 values ('홍길동', '20191007')")
    db.exec("insert into aergojdbc001 values ('홍길동', '20191007')")
end

function inserts()
	db.exec("insert into aergojdbc001 select * from aergojdbc001")
end

abi.register(inserts)
`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "big20", 0, bigSrc),
	)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 17; i++ {
		err = bc.ConnectBlock(
			NewLuaTxCall("ktlee", "big20", 0, `{"Name": "inserts"}`),
		)
		if err != nil {
			t.Error(err)
		}
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "big20", 0, `{"Name": "inserts"}`).Fail("database or disk is full"),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestTypeJson(t *testing.T) {
	definition := `
	state.var{
		table = state.value()
	}

	function set(val)
		table:set(json.decode(val))
	end

	function get()
		return table:get()
	end

	function getenc()
		return json.encode(table:get())
	end
	
	function getlen()
		a = table:get()
		return a[1], string.len(a[1])
	end

	function getAmount()
		return system.getAmount()
	end

	abi.register(set, get, getenc, getlen)
	abi.payable(getAmount)`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "json", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["[1,2,3]"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", "[1,2,3]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"[1,2,3]"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0,
			`{"Name":"set", "Args":["{\"key1\":[1,2,3], \"run\", \"key2\":5, [4,5,6]}"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"1":"run","2":[4,5,6],"key1":[1,2,3],"key2":5}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"{\"1\":\"run\",\"2\":[4,5,6],\"key1\":[1,2,3],\"key2\":5}"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0,
			`{"Name":"set", "Args":["{\"key1\":{\"arg1\": 1,\"arg2\":null, \"arg3\":[]}, \"key2\":[5,4,3]}"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"key1":{"arg1":1,"arg3":{}},"key2":[5,4,3]}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"{\"key1\":{\"arg1\":1,\"arg3\":{}},\"key2\":[5,4,3]}"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0,
			`{"Name":"set", "Args":["{\"key1\":[1,2,3], \"key1\":5}"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"key1":5}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["[\"\\\"hh\\t\",\"2\",3]"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `["\"hh\u0009","2",3]`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getlen", "Args":[]}`, "", `["\"hh\u0009",4]`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"[\"\\\"hh\\u0009\",\"2\",3]"`)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "json", 100, `{"Name":"getAmount"}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != `"100"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0,
			`{"Name":"set", "Args":["{\"key1\":[1,2,3], \"key1\":5}}"]}`).Fail("not proper json format"),
	)
}

// feature tests
func TestFeatureVote(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
function constructor()
	system.setItem("owner", system.getSender())
end

function addCandidate(name)
	if system.getSender() ~= system.getItem("owner") then
		return
	end

	if (system.getItem(name) ~= nil) then
		return
	end
	
	local numCandidates;
	if (system.getItem("numCandidates") == nil) then
		numCandidates = 0;
	else
		numCandidates = tonumber(system.getItem("numCandidates"))
	end

	system.setItem("candidate_list_" .. numCandidates, name)

	numCandidates = numCandidates + 1;
	system.setItem("numCandidates", tostring(numCandidates));
	system.setItem(name, tostring(0));
end

function getCandidates()
	local numCandidates;
	if (system.getItem("numCandidates") == nil) then
		return {};
	else
		numCandidates = tonumber(system.getItem("numCandidates"))
	end

	local candidates = {};
	local i = 0;

	while true do
		if (numCandidates == i) then
			break;
		end
		local candidate = system.getItem("candidate_list_" .. i)
		local count = system.getItem(candidate)
		if count == nil then
			count = 0
		end
		table.insert(candidates, {id = i, name = candidate, count = count});
		i = i + 1;
	end
	return candidates;
end

function registerVoter(address)
	if system.getSender() ~= system.getItem("owner") then
		return
	end
	
	system.setItem("voter_" .. address, "0");
end

function vote(candidateID)
	local totalVoted
	local voter = system.getItem("voter_" .. system.getSender())
	if voter == nil then
		return
	end
	totalVoted = tonumber(system.getItem("voter_" .. system.getSender()))
	if totalVoted > 3 then
		return
	end
	if system.getItem(candidateID) == nil then
		return
	end
	local currentVotes;
	if (system.getItem(candidateID) == nil) then
		currentVotes = 0;
	else
		currentVotes = tonumber(system.getItem(candidateID))
	end
	currentVotes = currentVotes + 1

	system.setItem(candidateID, tostring(currentVotes))
	totalVoted = totalVoted + 1
	system.setItem("voter_" .. system.getSender(), tostring(totalVoted));
end

abi.register(addCandidate, getCandidates, registerVoter, vote)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("owner", 100000000000000000),
		NewLuaTxDef("owner", "vote", 0, definition),
		NewLuaTxAccount("user1", 100000000000000000),
		NewLuaTxAccount("user10", 100000000000000000),
		NewLuaTxAccount("user11", 100000000000000000),
	)

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			`{"Name":"addCandidate", "Args":["candidate1"]}`,
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			`{"Name":"addCandidate", "Args":["candidate2"]}`,
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			`{"Name":"addCandidate", "Args":["candidate3"]}`,
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"addCandidate", "Args":["candidate4"]}`,
		),
	)
	err = bc.Query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		// register voter
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user10"))),
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user10"))),
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user11"))),
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user1"))),
		),
		// vote
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user1"]}`,
		),
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user1"]}`,
		),
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user2"]}`,
		),
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user2"]}`,
		),
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user3"]}`,
		),
	)

	err = bc.Query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"user11",
			"vote",
			0,
			`{"Name":"vote", "Args":["candidate1"]}`,
		),
		NewLuaTxCall(
			"user10",
			"vote",
			0,
			`{"Name":"vote", "Args":["candidate1"]}`,
		),
	)

	err = bc.Query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"2","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestFeatureGovernance(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
    function test_gov()
		contract.stake("10000 aergo")
		contract.vote("16Uiu2HAm2gtByd6DQu95jXURJXnS59Dyb9zTe16rDrcwKQaxma4p")
    end

	function error_case()
		contract.stake("10000 aergo")
		assert(false)
	end
	
	function test_pcall()
		return contract.pcall(error_case)
	end
		
    abi.register(test_gov, test_pcall, error_case)
	abi.payable(test_gov, test_pcall)
`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "gov", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	amount, _ := new(big.Int).SetString("40000000000000000000000", 10)
	err = bc.ConnectBlock(
		NewLuaTxCallBig("ktlee", "gov", amount, `{"Name": "test_gov", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	oldstaking, err := bc.GetStaking("gov")
	if err != nil {
		t.Error(err)
	}
	oldgov, err := bc.GetAccountState("gov")
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "gov", 0, `{"Name": "test_pcall", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	staking, err := bc.GetStaking("gov")
	if err != nil {
		t.Error(err)
	}
	gov, err := bc.GetAccountState("gov")
	if err != nil {
		t.Error(err)
	}

	if bytes.Equal(oldstaking.Amount, staking.Amount) == false ||
		bytes.Equal(oldgov.GetBalance(), gov.GetBalance()) == false {
		t.Error("pcall error")
	}
	tx = NewLuaTxCall("ktlee", "gov", 0, `{"Name": "error_case", "Args":[]}`)
	_ = bc.ConnectBlock(tx)
	newstaking, err := bc.GetStaking("gov")
	if err != nil {
		t.Error(err)
	}
	newgov, err := bc.GetAccountState("gov")
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(oldstaking.Amount, newstaking.Amount) == false ||
		bytes.Equal(oldgov.GetBalance(), newgov.GetBalance()) == false {
		fmt.Println(new(big.Int).SetBytes(newstaking.Amount).String(), newgov.GetBalanceBigInt().String())
		t.Error("pcall error")
	}
}

func TestFeaturePcallRollback(t *testing.T) {
	definition1 := `
	function constructor(init)
		system.setItem("count", init)
	end

	function init()
		db.exec([[create table if not exists r (
	  id integer primary key
	, n integer check(n >= 10)
	, nonull text not null
	, only integer unique)
	]])
		db.exec("insert into r values (1, 11, 'text', 1)")
	end

	function pkins1()
		db.exec("insert into r values (3, 12, 'text', 2)")
		db.exec("insert into r values (1, 12, 'text', 2)")
	end

	function pkins2()
		db.exec("insert into r values (4, 12, 'text', 2)")
	end

	function pkget()
		local rs = db.query("select count(*) from r")
		if rs:next() then
			local n = rs:get()
			--rs:next()
			return n
		else
			return "error in count()"
		end
	end

	function inc()
		count = system.getItem("count")
		system.setItem("count", count + 1)
		return count
	end

	function get()
		return system.getItem("count")
	end

	function getOrigin()
		return system.getOrigin()
	end

	function set(val)
		system.setItem("count", val)
	end
	abi.register(inc,get,set, init, pkins1, pkins2, pkget, getOrigin)
	abi.payable(constructor, inc)
	`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "counter", 10, definition1).Constructor("[0]"),
		NewLuaTxCall("ktlee", "counter", 15, `{"Name":"inc", "Args":[]}`),
	)

	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "1")
	if err != nil {
		t.Error(err)
	}

	definition2 := `
	function constructor(addr)
		system.setItem("count", 99)
		system.setItem("addr", addr)
	end
	function add(amount)
		first = contract.call.value(amount)(system.getItem("addr"), "inc")
		status, res = pcall(contract.call.value(1000000), system.getItem("addr"), "inc")
		if status == false then
			return first
		end
		return res
	end
	function dadd()
		return contract.delegatecall(system.getItem("addr"), "inc")
	end
	function get()
		addr = system.getItem("addr")
		a = contract.call(addr, "get")
		return a
	end
	function dget()
		addr = system.getItem("addr")
		a = contract.delegatecall(addr, "get")
		return a
	end
	function send(addr, amount)
		contract.send(addr, amount)
		status, res = pcall(contract.call.value(1000000000)(system.getItem("addr"), "inc"))
		return status
	end
	function sql()
		contract.call(system.getItem("addr"), "init")
		pcall(contract.call, system.getItem("addr"), "pkins1")
		contract.call(system.getItem("addr"), "pkins2")
		return status
	end

	function sqlget()
		return contract.call(system.getItem("addr"), "pkget")
	end

	function getOrigin()
		return contract.call(system.getItem("addr"), "getOrigin")
	end
	abi.register(add, dadd, get, dget, send, sql, sqlget, getOrigin)
	abi.payable(constructor,add)
	`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "caller", 10, definition2).
			Constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(strHash("counter")))),
		NewLuaTxCall("ktlee", "caller", 15, `{"Name":"add", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "caller", 0, `{"Name":"sql", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("caller", `{"Name":"sqlget", "Args":[]}`, "", "2")
	if err != nil {
		t.Error(err)
	}

	tx := NewLuaTxCall("ktlee", "caller", 0, `{"Name":"getOrigin", "Args":[]}`)
	_ = bc.ConnectBlock(tx)
	receipt := bc.GetReceipt(tx.Hash())
	if receipt.GetRet() != "\""+types.EncodeAddress(strHash("ktlee"))+"\"" {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}

	definition3 := `
	function pass(addr)
		contract.send(addr, 1)
	end

	function add(addr, a, b)
		system.setItem("arg", a)
		contract.pcall(pass, addr)
		return a+b
	end

	function set(addr)
		contract.send(addr, 1)
		system.setItem("arg", 2)
		status, ret  = contract.pcall(add, addr, 1, 2)
	end

	function set2(addr)
		contract.send(addr, 1)
		system.setItem("arg", 2)
		status, ret  = contract.pcall(add, addar, 1)
	end

	function get()
		return system.getItem("arg")
	end
	
	function getBalance()
		return contract.balance()
	end

	abi.register(set, set2, get, getBalance)
	abi.payable(set, set2)
	`

	bc, err = LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxAccount("bong", 0),
		NewLuaTxDef("ktlee", "counter", 0, definition3),
	)
	if err != nil {
		t.Error(err)
	}
	tx = NewLuaTxCall("ktlee", "counter", 20,
		fmt.Sprintf(`{"Name":"set", "Args":["%s"]}`, types.EncodeAddress(strHash("bong"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("counter", `{"Name":"getBalance", "Args":[]}`, "", "\"18\"")
	if err != nil {
		t.Error(err)
	}
	state, err := bc.GetAccountState("bong")
	if state.GetBalanceBigInt().Uint64() != 2 {
		t.Error("balance error")
	}
	tx = NewLuaTxCall("ktlee", "counter", 10,
		fmt.Sprintf(`{"Name":"set2", "Args":["%s"]}`, types.EncodeAddress(strHash("bong"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	state, err = bc.GetAccountState("bong")
	if state.GetBalanceBigInt().Uint64() != 3 {
		t.Error("balance error")
	}
}

func TestFeaturePcallNested(t *testing.T) {
	definition1 := `
state.var {
    Map = state.map(),
}
function map(a)
  return Map[a]
end
function constructor()
end
function pcall3(to)
  contract.send(to, "1 aergo")
end
function pcall2(addr, to)
  status = pcall(contract.call, addr, "pcall3", to)
  system.print(status)
  assert(false)
end
function pcall1(addr, to)
  status = pcall(contract.call, addr, "pcall2", addr, to)
  system.print(status)
  Map[addr] = 2
  status = pcall(contract.call, addr, "pcall3", to)
  system.print(status)
  status = pcall(contract.call, addr, "pcall2", addr, to)
  system.print(status)
end
function default()
end
abi.register(map, pcall1, pcall2, pcall3, default)
abi.payable(pcall1, default, constructor)
	`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxAccount("bong", 0),
		NewLuaTxDef("ktlee", "pcall", 10000000000000000000, definition1),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "pcall", 0,
			fmt.Sprintf(`{"Name":"pcall1", "Args":["%s", "%s"]}`,
				types.EncodeAddress(strHash("pcall")), types.EncodeAddress(strHash("bong")))),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("pcall", fmt.Sprintf(`{"Name":"map", "Args":["%s"]}`,
		types.EncodeAddress(strHash("pcall"))), "", "2")
	if err != nil {
		t.Error(err)
	}
	state, err := bc.GetAccountState("bong")
	if state.GetBalanceBigInt().Uint64() != 1000000000000000000 {
		t.Error("balance error", state.GetBalanceBigInt().Uint64())
	}
}

func TestFeatureLuaCryptoVerifyProof(t *testing.T) {
	bc, err := LoadDummyChain(OnPubNet)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	definition := `
	function hextobytes(str)
		return (str:gsub('..', function (cc)
			return string.char(tonumber(cc, 16))
		end))
	end

	function verifyProofRaw(data)
		local k = "a6eef7e35abe7026729641147f7915573c7e97b47efa546f5f6e3230263bcb49"
		local v = "2710"
		local p0 = "f871a0379a71a6fb36a75e085aff02beec9f5934b9648d24e2901da307492219608b3780a006a684f73e33f5c18739fd1339977f6fe328eb5cbe64239244b0cec88744355180808080a023866491ea0336f72e659c2a7daf61285de093b04fa353c48069a807c2ba845f808080808080808080"
		local p1 = "e5a03eb5be412f275a18f6e4d622aee4ff40b21467c926224771b782d4c095d1444b83822710"
		local b = crypto.verifyProof(hextobytes(k), hextobytes(v), crypto.keccak256(hextobytes(p0)), hextobytes(p0), hextobytes(p1))
		return b
	end

	function verifyProofHex(data)
		local k = "0xa6eef7e35abe7026729641147f7915573c7e97b47efa546f5f6e3230263bcb49"
		local v = "0x2710"
		local p0 = "0xf871a0379a71a6fb36a75e085aff02beec9f5934b9648d24e2901da307492219608b3780a006a684f73e33f5c18739fd1339977f6fe328eb5cbe64239244b0cec88744355180808080a023866491ea0336f72e659c2a7daf61285de093b04fa353c48069a807c2ba845f808080808080808080"
		local p1 = "0xe5a03eb5be412f275a18f6e4d622aee4ff40b21467c926224771b782d4c095d1444b83822710"
		local b = crypto.verifyProof(k, v, crypto.keccak256(p0), p0, p1)
		return b
	end

	abi.register(verifyProofRaw, verifyProofHex)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100000000000000000),
		NewLuaTxDef("ktlee", "eth", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("eth", `{"Name":"verifyProofRaw"}`, "", `true`)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("eth", `{"Name":"verifyProofHex"}`, "", `true`)
	if err != nil {
		t.Error(err)
	}

	state, err := bc.GetAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	bal := state.GetBalanceBigInt().Uint64()
	tx := NewLuaTxCall("ktlee", "eth", 0, `{"Name": "verifyProofRaw"}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	r := bc.GetReceipt(tx.Hash())
	expectedFee := uint64(154137)
	if r.GetGasUsed() != expectedFee {
		t.Errorf("expected: %d, but got: %d", expectedFee, r.GetGasUsed())
	}
	state, err = bc.GetAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	if bal-expectedFee != state.GetBalanceBigInt().Uint64() {
		t.Errorf(
			"expected: %d, but got: %d",
			bal-expectedFee,
			state.GetBalanceBigInt().Uint64(),
		)
	}
	bal = state.GetBalanceBigInt().Uint64()
	tx = NewLuaTxCall("ktlee", "eth", 0, `{"Name": "verifyProofHex"}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	r = bc.GetReceipt(tx.Hash())
	expectedFee = uint64(108404)
	if r.GetGasUsed() != expectedFee {
		t.Errorf("expected: %d, but got: %d", expectedFee, r.GetGasUsed())
	}
	state, err = bc.GetAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	if bal-expectedFee != state.GetBalanceBigInt().Uint64() {
		t.Errorf(
			"expected: %d, but got: %d",
			bal-expectedFee,
			state.GetBalanceBigInt().Uint64(),
		)
	}
}

func TestFeatureFeeDelegation(t *testing.T) {
	definition := `
	state.var{
        whitelist = state.map(),
    }

    function reg(k)
		if (k == nil) then
        	whitelist[system.getSender()] = true
		else
        	whitelist[k] = true
		end
    end

    function query(a)
		if (system.isFeeDelegation() == true) then
        	whitelist[system.getSender()] = false
		end
        return 1,2,3,4,5
    end
    function check_delegation(fname,k)
		if (fname == "query") then
        	return whitelist[system.getSender()]
		end
		return false
    end
	function default()
	end
    abi.register(reg, query)
    abi.payable(default)
    abi.fee_delegation(query)
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
		NewLuaTxCallFeeDelegate("user1", "fd", 0, `{"Name": "check_delegation", "Args":[]}`).
			Fail("check_delegation function is not declared of fee delegation"),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("user1", "fd", 0, `{"Name": "query", "Args":[]}`).
			Fail("not enough balance"),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCallFeeDelegate("user1", "fd", 0, `{"Name": "query", "Args":[]}`).
			Fail("fee delegation is not allowed"),
	)
	if err != nil {
		t.Error(err)
	}
	contract1, err := bc.GetAccountState("fd")
	if err != nil {
		t.Error(err)
	}

	tx := NewLuaTxCallFeeDelegate("user1", "fd", 0, `{"Name": "query", "Args":["arg"]}`)
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "fd", 0, fmt.Sprintf(`{"Name":"reg", "Args":["%s"]}`,
			types.EncodeAddress(strHash("user1")))),
		tx,
	)
	if err != nil {
		t.Error(err)
	}

	contract2, err := bc.GetAccountState("fd")
	if err != nil {
		t.Error(err)
	}
	if contract1.GetBalanceBigInt().Uint64() == contract2.GetBalanceBigInt().Uint64() {
		t.Error("feedelegation error")
	}
	err = bc.ConnectBlock(
		tx.Fail("fee delegation is not allowed"),
	)

	definition2 := `
	state.var{
        whitelist = state.map(),
    }

    function reg(k)
		if (k == nil) then
        	whitelist[system.getSender()] = true
		else
        	whitelist[k] = true
		end
    end

    function query()
        whitelist[system.getSender()] = false
        return 1,2,3,4,5
    end
	function default()
	end
    abi.register(reg, query)
    abi.payable(default)
    abi.fee_delegation(query)
`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "fd2", 0, definition2),
	)
	if strings.Contains(err.Error(), "no 'check_delegation' function") == false {
		t.Error(err)
	}
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
