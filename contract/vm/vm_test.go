package main

import (
	"testing"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/aergoio/aergo/v2/cmd/aergoluac/util"
)

func TestCompile(t *testing.T) {

	invalidCode := `
		function add(a, b)
			return a + b
		end
	`

	validCode := `
		function add(a, b)
			return a + b
		end
		abi.register(add)
	`

	// Test case: Valid Lua code
	byteCode, err := Compile(validCode, false)
	assert.NoError(t, err, "Expected no error for valid Lua code")
	assert.NotNil(t, byteCode, "Expected bytecode for valid Lua code")

	// Test case: Invalid Lua code
	byteCode, err = Compile(invalidCode, false)
	assert.Error(t, err, "Expected an error for invalid Lua code")
	assert.Nil(t, byteCode, "Expected no bytecode for invalid Lua code")

	// Test case: Valid Lua code with parent
	byteCode, err = Compile(validCode, true)
	assert.NoError(t, err, "Expected no error for valid Lua code with parent")
	assert.NotNil(t, byteCode, "Expected bytecode for valid Lua code with parent")

	// Test case: Invalid Lua code with parent
	byteCode, err = Compile(invalidCode, true)
	assert.Error(t, err, "Expected an error for invalid Lua code with parent")
	assert.Nil(t, byteCode, "Expected no bytecode for invalid Lua code with parent")
}

func TestExecuteBasic(t *testing.T) {

	contractCode := `
		function add(a, b)
			return a + b
		end
		function hello(name)
			return "Hello, " .. name
		end
		function many()
			return 123, bignum.number(456), "abc", true, nil
		end
		function echo(...)
			return ...
		end
		abi.register(add, hello, many, echo)
	`

	// set global variables
	hardforkVersion = 3
	isPubNet = true

	// initialize the Lua VM
	InitializeVM()

	// compile contract
	byteCodeAbi, err := Compile(contractCode, false)
	assert.NoError(t, err)
	assert.NotNil(t, byteCodeAbi)

	bytecode := util.LuaCode(byteCodeAbi).ByteCode()

	// execute contract - add
	result, err, usedGas := Execute("testAddress", string(bytecode), "add", `[1,2]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `3`, result)

	// execute contract - hello
	result, err, usedGas = Execute("testAddress", string(bytecode), "hello", `["World"]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `"Hello, World"`, result)

	// execute contract - many
	result, err, usedGas = Execute("testAddress", string(bytecode), "many", `[]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `[123,{"_bignum":"456"},"abc",true,null]`, result)

	// execute contract - echo
	result, err, usedGas = Execute("testAddress", string(bytecode), "echo", `[123,4.56,{"_bignum":"789"},"abc",true,null]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `[123,4.56,{"_bignum":"789"},"abc",true,null]`, result)

}

func TestExecuteQueryBasic(t *testing.T) {

	contractCode := `
		function add(a, b)
			return a + b
		end
		function hello(name)
			return "Hello, " .. name
		end
		function many()
			return 123, bignum.number(456), "abc", true, nil
		end
		function echo(...)
			return ...
		end
		abi.register(add, hello, many, echo)
	`

	// set global variables
	hardforkVersion = 3
	isPubNet = true

	// initialize the Lua VM
	InitializeVM()

	// compile contract
	byteCodeAbi, err := Compile(contractCode, false)
	assert.NoError(t, err)
	assert.NotNil(t, byteCodeAbi)

	bytecode := util.LuaCode(byteCodeAbi).ByteCode()

	// execute contract - add
	result, err, usedGas := Execute("testAddress", string(bytecode), "add", `[1,2]`, 0, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, usedGas, uint64(0), "Expected no gas to be used")
	assert.Equal(t, `3`, result)

	// execute contract - hello
	result, err, usedGas = Execute("testAddress", string(bytecode), "hello", `["World"]`, 0, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, usedGas, uint64(0), "Expected no gas to be used")
	assert.Equal(t, `"Hello, World"`, result)

	// execute contract - many
	result, err, usedGas = Execute("testAddress", string(bytecode), "many", `[]`, 0, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, usedGas, uint64(0), "Expected no gas to be used")
	assert.Equal(t, `[123,{"_bignum":"456"},"abc",true,null]`, result)

	// execute contract - echo
	result, err, usedGas = Execute("testAddress", string(bytecode), "echo", `[123,4.56,{"_bignum":"789"},"abc",true,null]`, 0, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, usedGas, uint64(0), "Expected no gas to be used")
	assert.Equal(t, `[123,4.56,{"_bignum":"789"},"abc",true,null]`, result)

}


type vmCallback struct {
	method string
	args []string
	result string
	err error
}

var callbacks []vmCallback

func TestExecuteWithCallback(t *testing.T) {

	sendRequest = func(method string, args []string) (string, error) {
		//fmt.Println("method: ", method, "args: ", args)
		// get the next callback
		callback := callbacks[0]
		callbacks = callbacks[1:]
		// check that the method and args are correct
		assert.Equal(t, callback.method, method)
		assert.Equal(t, callback.args, args)
		return callback.result, callback.err
	}

	contractCode := `
		state.var {
			kv = state.map()
		}
		function set(key, value)
			kv[key] = value
		end
		function get(key)
			return kv[key]
		end
		function send(to, amount)
			return contract.send(to, amount)
		end
		function call(...)
			return contract.call(...)
		end
		function call_with_send(amount, ...)
			return contract.call.value(amount)(...)
		end
		function delegatecall(...)
			return contract.delegatecall(...)
		end
		function deploy(...)
			return contract.deploy(...)
		end
		function deploy_with_send(amount, ...)
			return contract.deploy.value(amount)(...)
		end
		function get_info()
			return system.getContractID(), contract.balance(), system.getAmount(), system.getSender(), system.getOrigin(), system.isFeeDelegation()
		end
		function events()
			contract.event('first', 123, 'abc')
			contract.event('second', '456', 7.89)
		end
		abi.register(set, get, send, call, call_with_send, delegatecall, deploy, deploy_with_send, get_info, events)
	`

	contract2 := `
		state.var {
			_owner = state.value(),
			_name = state.value()
		}
		function default()
			-- do nothing, only receive aergo
		end
		function constructor(first_name)
			_name.set(first_name)
			_owner.set(contract.getSender())
		end
		abi.payable(constructor, default)
	`

	contract3 := `
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
		abi.register(sql_func)
	`

	// set global variables
	hardforkVersion = 3
	isPubNet = true

	// initialize the Lua VM
	InitializeVM()

	// compile contract
	byteCodeAbi, err := Compile(contractCode, false)
	assert.NoError(t, err)
	assert.NotNil(t, byteCodeAbi)

	bytecode := util.LuaCode(byteCodeAbi).ByteCode()

	InitializeVM()

	// execute contract - set
	callbacks = []vmCallback{
		{"get", []string{"_sv_meta-type_kv", ""}, "null", nil},
		{"set", []string{"_sv_meta-type_kv", "4"}, "", nil},
		{"set", []string{"_sv_kv-key", "12345"}, "", nil},
	}
	result, err, usedGas := Execute("testAddress", string(bytecode), "set", `["key",12345]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, ``, result)

	InitializeVM()

	// execute contract - get
	callbacks = []vmCallback{
		{"get", []string{"_sv_meta-type_kv", ""}, "4", nil},
		{"get", []string{"_sv_kv-key", ""}, `12345`, nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "get", `["key"]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `12345`, result)

	InitializeVM()

	// execute contract - send - simple
	callbacks = []vmCallback{
		// the last argument is the gas in bytes, the first 8 bytes of the result is the used gas
		{"send", []string{"0x12345", "1000000000000000000", "\xa8*\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "send", `["0x12345","1000000000000000000"]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, ``, result)

	InitializeVM()

	// execute contract - send - with successful call
	callbacks = []vmCallback{
		// the last argument is the gas in bytes, the first 8 bytes of the result is the used gas
		{"send", []string{"0x12345", "1000000000000000000", "\xa8*\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00[]", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "send", `["0x12345","1000000000000000000"]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, ``, result)

	InitializeVM()

	// execute contract - send - with failed call
	callbacks = []vmCallback{
		// the last argument is the gas in bytes, the first 8 bytes of the result is the used gas
		{"send", []string{"0x12345", "1000000000000000000", "\xa8*\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00", errors.New("failed call")},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "send", `["0x12345","1000000000000000000"]`, 1000000, "testCaller", false, false, "")
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)

	InitializeVM()

	// execute contract - send - with invalid address
	callbacks = []vmCallback{
		// the last argument is the gas in bytes, the first 8 bytes of the result is the used gas
		{"send", []string{"chucku-chucku", "1000000000000000000", "\xa8*\x0f\x00\x00\x00\x00\x00"}, "", errors.New("invalid address")},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "send", `["chucku-chucku","1000000000000000000"]`, 1000000, "testCaller", false, false, "")
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)

	InitializeVM()

	// execute contract - send - with invalid amount
	callbacks = []vmCallback{
		// the last argument is the gas in bytes, the first 8 bytes of the result is the used gas
		{"send", []string{"0x12345", "abc", "\xa8*\x0f\x00\x00\x00\x00\x00"}, "", errors.New("invalid amount")},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "send", `["0x12345","abc"]`, 1000000, "testCaller", false, false, "")
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)

	InitializeVM()

	// execute contract - call
	callbacks = []vmCallback{
		// the last argument is the gas in bytes, the first 8 bytes of the result is the used gas
		{"call", []string{"0x12345", "add", "[1,2]", "", ".#\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00[3]", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "call", `["0x12345","add",1,2]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `3`, result)

	InitializeVM()

	// execute contract - call with send
	callbacks = []vmCallback{
		// the last argument is the gas in bytes, the first 8 bytes of the result is the used gas
		{"call", []string{"0x12345", "buy", `[1,"NFT"]`, "9876543210", "\xf8\x1d\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00[\"purchased\",1,\"NFT\"]", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "call_with_send", `["9876543210","0x12345","buy",1,"NFT"]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `["purchased",1,"NFT"]`, result)

	InitializeVM()

	// execute contract - call with send to default function
	callbacks = []vmCallback{
		// the last argument is the gas in bytes, the first 8 bytes of the result is the used gas
		{"send", []string{"0x12345", "9876543210", "\xca)\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00[]", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "call_with_send", `["9876543210","0x12345"]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)

	InitializeVM()

	// execute contract - delegated call
	callbacks = []vmCallback{
		{"delegate-call", []string{"0x12345", "add", "[1,2]", "\x9d#\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00[3]", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "delegatecall", `["0x12345","add",1,2]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `3`, result)

	InitializeVM()

	// execute contract - deploy
	callbacks = []vmCallback{
		{"deploy", []string{contractCode, "[]", "", "\xd5\x17\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00[\"Amhs9v8EeAAWrrvEFrvMng4UksHRsR7wN1iLqKkXw5bqMV18JP3h\"]", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "deploy", `["`+contractCode+`"]`, 1000005, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.Equal(t, `"Amhs9v8EeAAWrrvEFrvMng4UksHRsR7wN1iLqKkXw5bqMV18JP3h"`, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)

	InitializeVM()

	// execute contract - deploy with invalid return
	callbacks = []vmCallback{
		{"deploy", []string{contractCode, "[]", "", "\xd5\x17\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00...", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "deploy", `["`+contractCode+`"]`, 1000005, "testCaller", false, false, "")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "uncatchable: internal error: result from call is not a valid JSON array")
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)

	InitializeVM()

	// execute contract - deploy with incomplete return
	callbacks = []vmCallback{
		{"deploy", []string{contractCode, "[]", "", "\xd5\x17\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00[]", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "deploy", `["`+contractCode+`"]`, 1000005, "testCaller", false, false, "")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "uncatchable: internal error: result from call is not a valid JSON array")
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)

	InitializeVM()

	// execute contract - deploy with send
	callbacks = []vmCallback{
		{"deploy", []string{contract2, "[250]", "9876543210", "|\x11\x0f\x00\x00\x00\x00\x00"}, "\x09\x00\x01\x00\x00\x00\x00\x00[\"Amhs9v8EeAAWrrvEFrvMng4UksHRsR7wN1iLqKkXw5bqMV18JP3h\"]", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "deploy_with_send", `["9876543210","`+contract2+`",250]`, 1000005, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.Equal(t, `"Amhs9v8EeAAWrrvEFrvMng4UksHRsR7wN1iLqKkXw5bqMV18JP3h"`, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)

	InitializeVM()

	// execute contract - get_info
	callbacks = []vmCallback{
		{"balance", []string{""}, "123000", nil},
		{"getAmount", []string{}, "1000000", nil},
		{"getOrigin", []string{}, "anotherAddress", nil},
		{"isFeeDelegation", []string{}, "false", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "get_info", `[]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `["testAddress","123000","1000000","testCaller","anotherAddress",false]`, result)

	InitializeVM()

	// execute contract - events
	callbacks = []vmCallback{
		{"event", []string{"first", `[123,"abc"]`}, "", nil},
		{"event", []string{"second", `["456",7.89]`}, "", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "events", `[]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)



	isPubNet = false

	InitializeVM()

	// compile contract
	byteCodeAbi, err = Compile(contract3, false)
	assert.NoError(t, err)
	assert.NotNil(t, byteCodeAbi)

	bytecode = util.LuaCode(byteCodeAbi).ByteCode()

	InitializeVM()

	// execute contract - sql_func
	callbacks = []vmCallback{
		{"dbQuery", []string{"+\x00\x00\x00sselect round(3.14),min(1,2,3), max(4,5,6)\x00\x01\x00\x00\x00y"}, "\x05\x00\x00\x00i\x01\x00\x00\x00", nil},
		{"rsNext", []string{"\x05\x00\x00\x00i\x01\x00\x00\x00"}, "\x02\x00\x00\x00b\x01", nil},
		{"rsGet", []string{"\x05\x00\x00\x00i\x01\x00\x00\x00"}, "\x05\x00\x00\x00i\x03\x00\x00\x00\x05\x00\x00\x00i\x01\x00\x00\x00\x05\x00\x00\x00i\x06\x00\x00\x00", nil},
		//{"rsNext", []string{"\x05\x00\x00\x00i\x01\x00\x00\x00"}, "\x02\x00\x00\x00b\x00", nil},
	}
	result, err, usedGas = Execute("testAddress", string(bytecode), "sql_func", `[]`, 1000000, "testCaller", false, false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Greater(t, usedGas, uint64(0), "Expected some gas to be used")
	fmt.Println("used gas: ", usedGas)
	assert.Equal(t, `[3,1,6]`, result)

}
