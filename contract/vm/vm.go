package main

/*
 #cgo CFLAGS: -I${SRCDIR}/../../libtool/include/luajit-2.1 -I${SRCDIR}/../../libtool/include
 #cgo !windows CFLAGS: -DLJ_TARGET_POSIX
 #cgo darwin LDFLAGS: ${SRCDIR}/../../libtool/lib/libluajit-5.1.a ${SRCDIR}/../../libtool/lib/libgmp.dylib -lm
 #cgo windows LDFLAGS: ${SRCDIR}/../../libtool/lib/libluajit-5.1.a ${SRCDIR}/../../libtool/bin/libgmp-10.dll -lm
 #cgo !darwin,!windows LDFLAGS: ${SRCDIR}/../../libtool/lib/libluajit-5.1.a -L${SRCDIR}/../../libtool/lib64 -L${SRCDIR}/../../libtool/lib -lgmp -lm


 #include <stdlib.h>
 #include <string.h>
 #include "vm.h"
 #include "util.h"
 #include "bignum_module.h"
*/
import "C"
import (
	"errors"
	"fmt"
	//"reflect"
	//"sort"
	"strings"
	"unsafe"
	"encoding/binary"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/cmd/aergoluac/luac"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/types"
)

const vmTimeoutErrMsg = "contract timeout during vm execution"

var logger *log.Logger

type LState = C.lua_State   // C.struct_lua_State

var lstate *LState          // *C.lua_State

var contractAddress string
var contractCaller string
var contractGasLimit uint64
var contractIsFeeDelegation bool

////////////////////////////////////////////////////////////////////////////////

func InitializeVM() {
	if lstate == nil {
		// these are called only once on tests
		logger = log.NewLogger("vm")
		C.init_bignum()
		C.initViewFunction()
	} else {
		C.lua_close(lstate)
	}
	lstate = C.vm_newstate(C.int(hardforkVersion))
}

////////////////////////////////////////////////////////////////////////////////

type executor struct {
	L          *LState
	code       []byte
	fname      string
	args       string
	numArgs    C.int
	isAutoload bool
	jsonRet    string
	err        error
	abiErr     error
}

func newExecutor(bytecode []byte, fname string, args string, abiError string) *executor {

	ce := &executor{
		L: lstate,
		code: bytecode,
	}

	if abiError != "" {
		ce.abiErr = errors.New(abiError)
	}

	// set the gas limit on the Lua state
	setGas()

	// load the contract code into the Lua state
	ce.vmLoadCode()
	if ce.err != nil {
		return ce
	}

	// if fname starts with "autoload:" then it is an autoload function
	if strings.HasPrefix(fname, "autoload:") {
		ce.isAutoload = true
		fname = fname[9:]
	}
	ce.fname = fname
	ce.args = args

	return ce
}

////////////////////////////////////////////////////////////////////////////////
// Lua
////////////////////////////////////////////////////////////////////////////////

// push the arguments to the stack
func (ce *executor) pushArguments() {
	args := C.CString(ce.args)
	ce.numArgs = C.lua_util_json_array_to_lua(ce.L, args, C.bool(true));
	C.free(unsafe.Pointer(args))
	if ce.numArgs == -1 {
		ce.err = errors.New("invalid arguments. must be valid JSON array")
	}
}

/*
func (ce *executor) processArgs() {
	for _, v := range ce.ci.Args {
		if err := pushValue(ce.L, v); err != nil {
			ce.err = err
			return
		}
	}
}

func pushValue(L *LState, v interface{}) error {
	switch arg := v.(type) {
	case string:
		argC := C.CBytes([]byte(arg))
		C.lua_pushlstring(L, (*C.char)(argC), C.size_t(len(arg)))
		C.free(argC)
	case float64:
		if arg == float64(int64(arg)) {
			C.lua_pushinteger(L, C.lua_Integer(arg))
		} else {
			C.lua_pushnumber(L, C.double(arg))
		}
	case bool:
		var b int
		if arg {
			b = 1
		}
		C.lua_pushboolean(L, C.int(b))
	case json.Number:
		str := arg.String()
		intVal, err := arg.Int64()
		if err == nil {
			C.lua_pushinteger(L, C.lua_Integer(intVal))
		} else {
			ftVal, err := arg.Float64()
			if err != nil {
				return errors.New("unsupported number type:" + str)
			}
			C.lua_pushnumber(L, C.double(ftVal))
		}
	case nil:
		C.lua_pushnil(L)
	case []interface{}:
		err := toLuaArray(L, arg)
		if err != nil {
			return err
		}
	case map[string]interface{}:
		err := toLuaTable(L, arg)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported type:" + reflect.TypeOf(v).Name())
	}
	return nil
}

func toLuaArray(L *LState, arr []interface{}) error {
	C.lua_createtable(L, C.int(len(arr)), C.int(0))
	n := C.lua_gettop(L)
	for i, v := range arr {
		if err := pushValue(L, v); err != nil {
			return err
		}
		C.lua_rawseti(L, n, C.int(i+1))
	}
	return nil
}

func toLuaTable(L *LState, tab map[string]interface{}) error {
	C.lua_createtable(L, C.int(0), C.int(len(tab)))
	n := C.lua_gettop(L)
	// get the keys and sort them
	keys := make([]string, 0, len(tab))
	for k := range tab {
		keys = append(keys, k)
	}
	if C.vm_is_hardfork(L, 3) {
		sort.Strings(keys)
	}
	for _, k := range keys {
		v := tab[k]
		if len(tab) == 1 && strings.EqualFold(k, "_bignum") {
			if arg, ok := v.(string); ok {
				C.lua_settop(L, -2)
				argC := C.CString(arg)
				msg := C.lua_set_bignum(L, argC)
				C.free(unsafe.Pointer(argC))
				if msg != nil {
					return errors.New(C.GoString(msg))
				}
				return nil
			}
		}
		// push a key
		key := C.CString(k)
		C.lua_pushstring(L, key)
		C.free(unsafe.Pointer(key))

		if err := pushValue(L, v); err != nil {
			return err
		}
		C.lua_rawset(L, n)
	}
	return nil
}
*/

////////////////////////////////////////////////////////////////////////////////

func (ce *executor) call(hasParent bool) {

	defer func() {
		if ce.err != nil && hasParent {
			if bool(C.luaL_hasuncatchablerror(ce.L)) || bool(C.luaL_hassyserror(ce.L)) {
				ce.err = errors.New("uncatchable: " + ce.err.Error())
			}
		}
	}()

	if ce.err != nil {
		return
	}

	// execute code from the global scope, like declaring state variables and functions
	// as well as abi.register, abi.register_view, abi.payable, etc.
	ce.vmPreRun()
	if ce.err != nil {
		return
	}
	// if there is no error in the code pre-execution but failed to process the ABI, return the error now
	if ce.abiErr != nil {
		ce.err = ce.abiErr
		return
	}

	// push the function to be called to the stack
	if ce.isAutoload {
		// used for constructor and check_delegation functions
		loaded := vmPushGlobalFunction(ce.L, ce.fname)
		if !loaded {
			if ce.fname == "constructor" {
				// the constructor function was not found
				if hasParent {
					ce.jsonRet = "[]"
				}
			} else {
				ce.err = errors.New(fmt.Sprintf("contract autoload failed %s function: %s",
					contractAddress, ce.fname))
			}
			return
		}
	} else {
		// used for normal functions
		vmPushAbiFunction(ce.L, ce.fname)
	}

	// push the arguments to the stack
	ce.pushArguments()
	if ce.err != nil {
		logger.Debug().Err(ce.err).Str("contract", contractAddress).Msg("invalid argument")
		return
	}
	if !ce.isAutoload {
		ce.numArgs = ce.numArgs + 1
	}

	// set the instruction limit and timeout hook
	ce.setCountHook(C.int(5000000))

	// call the function
	nRet := C.int(0)
	cErrMsg := C.vm_call(ce.L, ce.numArgs, &nRet)

	// check for errors
	if cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		if (errMsg == C.ERR_BF_TIMEOUT || errMsg == vmTimeoutErrMsg) {
			ce.err = errors.New(vmTimeoutErrMsg)  // &VmTimeoutError{}
		} else {
			ce.err = errors.New(errMsg)
		}
		logger.Debug().Err(ce.err).Str("contract", contractAddress).Msg("contract execution failed")
		return
	}

	// convert the result to json
	var isError C.int
	retMsg := C.GoString(C.vm_get_json_ret(ce.L, nRet, C.bool(hasParent), &isError))
	if isError == 1 {
		ce.err = errors.New(retMsg)
	} else {
		ce.jsonRet = retMsg
	}

/*/ this can be moved to server side
	if ce.ctx.traceFile != nil {
		// write the contract code to a file in the temp directory
		address := types.EncodeAddress(ce.contractId)
		codeFile := fmt.Sprintf("%s%s%s.code", os.TempDir(), string(os.PathSeparator), address)
		if _, err := os.Stat(codeFile); os.IsNotExist(err) {
			f, err := os.OpenFile(codeFile, os.O_WRONLY|os.O_CREATE, 0644)
			if err == nil {
				_, _ = f.Write(ce.code)
				_ = f.Close()
			}
		}
		// write the used fee to the trace file
		str := fmt.Sprintf("contract %s used fee: %s\n", address, ce.ctx.usedFee().String())
		_, _ = ce.ctx.traceFile.WriteString(str)
	}
*/

}

// push the function to be called to the stack
func vmPushGlobalFunction(L *LState, funcName string) bool {
	fname := C.CString(funcName)
	loaded := C.vm_push_global_function(L, fname)
	C.free(unsafe.Pointer(fname))
	return loaded != C.int(0)
}

// push the function to be called to the stack
func vmPushAbiFunction(L *LState, funcName string) {
	C.vm_remove_constructor(L)
	fname := C.CString(funcName)
	C.vm_push_abi_function(L, fname)
	C.free(unsafe.Pointer(fname))
}

// load the contract code
func (ce *executor) vmLoadCode() {
	var chunkId *C.char
	if hardforkVersion >= 3 {
		chunkId = C.CString("@" + contractAddress)
	} else {
		contractId, err := types.DecodeAddress(contractAddress)
		if err != nil {
			ce.err = err
			return
		}
		chunkId = C.CString(hex.Encode(contractId))
	}
	defer C.free(unsafe.Pointer(chunkId))

	cErrMsg := C.vm_load_code(
		ce.L,
		(*C.char)(unsafe.Pointer(&ce.code[0])),
		C.size_t(len(ce.code)),
		chunkId,
	)

	if cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		ce.err = errors.New(errMsg)
		logger.Debug().Err(ce.err).Str("contract", contractAddress).Msg("failed to load code")
	}
}

// execute code from the global scope, like declaring state variables and functions
// as well as abi.register, abi.register_view, abi.payable, etc.
func (ce *executor) vmPreRun() {
	cErrMsg := C.vm_pre_run(ce.L)
	if cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		isUncatchable := bool(C.luaL_hasuncatchablerror(ce.L))
		if isUncatchable && (errMsg == C.ERR_BF_TIMEOUT || errMsg == vmTimeoutErrMsg) {
			ce.err = errors.New(vmTimeoutErrMsg) // &VmTimeoutError{}
		} else {
			ce.err = errors.New(errMsg)
		}
	}
}


////////////////////////////////////////////////////////////////////////////////
// GAS
////////////////////////////////////////////////////////////////////////////////


func IsGasSystem() bool {
	return contractGasLimit > 0
}

func setGas() {
	if IsGasSystem() {
		C.lua_gasset(lstate, C.ulonglong(contractGasLimit))
	}
}

func getRemainingGas() uint64 {
	return uint64(C.lua_gasget(lstate))
}

func getUsedGas() uint64 {
	return contractGasLimit - getRemainingGas()
}

func addConsumedGas(gas uint64, err error) error {
	if !IsGasSystem() {
		return err
	}
	remainingGas := getRemainingGas()
	if gas > remainingGas {
		if err == nil {
			err = errors.New("uncatchable: gas limit exceeded")
		}
		return err
	}
	remainingGas -= gas
	C.lua_gasset(lstate, C.ulonglong(remainingGas))
	return err
}

// extract the used gas from the result
func extractUsedGas(result string) (uint64, string) {
	if len(result) < 8 {
		return 0, result
	}
	usedGas := binary.LittleEndian.Uint64([]byte(result[:8]))
	result = result[8:]
	return usedGas, result
}


////////////////////////////////////////////////////////////////////////////////

/*
func setInstCount(ctx *vmContext, parent *LState, child *LState) {
	if !IsGasSystem() {
		C.vm_setinstcount(parent, C.vm_instcount(child))
	}
}

func setInstMinusCount(ctx *vmContext, L *LState, deduc C.int) {
	if !IsGasSystem() {
		C.vm_setinstcount(L, minusCallCount(ctx, C.vm_instcount(L), deduc))
	}
}

func minusCallCount(ctx *vmContext, curCount, deduc C.int) C.int {
	if !IsGasSystem() {
		return 0
	}
	remain := curCount - deduc
	if remain <= 0 {
		remain = 1
	}
	return remain
}
*/

////////////////////////////////////////////////////////////////////////////////

func Execute(
	address string,
	code string,
	fname string,
	args string,
	gas uint64,
	caller string,
	hasParent bool,
	isFeeDelegation bool,
	abiError string,
) (string, error, uint64) {

	contractAddress = address
	contractCaller = caller
	contractGasLimit = gas
	contractIsFeeDelegation = isFeeDelegation

	ex := newExecutor([]byte(code), fname, args, abiError)

	ex.call(hasParent)

	totalUsedGas := getUsedGas()

	return ex.jsonRet, ex.err, totalUsedGas
}

////////////////////////////////////////////////////////////////////////////////

func Compile(code string, hasParent bool) ([]byte, error) {
	L := luac.NewLState()
	if L == nil {
		return nil, errors.New("syserror: failed to create LState")
	}
	defer luac.CloseLState(L)
	var lState = (*LState)(L)

	if hasParent {
		// mark as running a call
		//C.luaL_set_loading(lState, C.bool(false))
		// set the hardfork version
		//C.luaL_set_hardforkversion(lState, 5)
		// set the timeout hook
		C.vm_set_timeout_hook(lState)
	}

	byteCodeAbi, err := luac.Compile(L, code)
	if err != nil {
		// if there is an uncatchable error, return it to the parent
		if hasParent && bool(C.luaL_hasuncatchablerror(lState)) {
			err = errors.New("uncatchable: " + err.Error())
		}
		return nil, err
	}

	return byteCodeAbi.Bytes(), nil
}
