package vm  // or main

/*
 #cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.1 -I${SRCDIR}/../libtool/include
 #cgo !windows CFLAGS: -DLJ_TARGET_POSIX
 #cgo darwin LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a ${SRCDIR}/../libtool/lib/libgmp.dylib -lm
 #cgo windows LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a ${SRCDIR}/../libtool/bin/libgmp-10.dll -lm
 #cgo !darwin,!windows LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -L${SRCDIR}/../libtool/lib64 -L${SRCDIR}/../libtool/lib -lgmp -lm


 #include <stdlib.h>
 #include <string.h>
 #include "vm.h"
 #include "bignum_module.h"
*/
import "C"



////////////////////////////////////////////////////////////////////////////////
// Lua
////////////////////////////////////////////////////////////////////////////////

func (ce *executor) pushArgs() {
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

////////////////////////////////////////////////////////////////////////////////

func (ce *executor) call(instLimit C.int, target *LState) (ret C.int) {

	defer func() {
		if ret == 0 && target != nil {
			if C.luaL_hasuncatchablerror(ce.L) != C.int(0) {
				C.luaL_setuncatchablerror(target)
			}
			if C.luaL_hassyserror(ce.L) != C.int(0) {
				C.luaL_setsyserror(target)
			}
		}
	}()

	if ce.err != nil {
		return 0
	}

	defer ce.refreshRemainingGas()

	if ce.isView == true {
		ce.ctx.nestedView++
		defer func() {
			ce.ctx.nestedView--
		}()
	}

	ce.vmLoadCall()
	if ce.err != nil {
		return 0
	}
	if ce.preErr != nil {
		ce.err = ce.preErr
		return 0
	}

	if ce.isAutoload {
		if loaded := vmAutoload(ce.L, ce.fname); !loaded {
			if ce.fname != constructor {
				ce.err = errors.New(fmt.Sprintf("contract autoload failed %s : %s",
					types.EncodeAddress(ce.ctx.curContract.contractId), ce.fname))
			}
			return 0
		}
	} else {
		C.vm_remove_constructor(ce.L)
		resolvedName := C.CString(ce.fname)
		C.vm_get_abi_function(ce.L, resolvedName)
		C.free(unsafe.Pointer(resolvedName))
	}

	ce.pushArgs()
	if ce.err != nil {
		ctrLgr.Debug().Err(ce.err)
		              .Stringer("contract", types.LogAddr(ce.ctx.curContract.contractId))
		              .Msg("invalid argument")
		return 0
	}

	ce.setCountHook(instLimit)
	nRet := C.int(0)
	cErrMsg := C.vm_pcall(ce.L, ce.numArgs, &nRet)

	if cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		if C.luaL_hassyserror(ce.L) != C.int(0) {
			ce.err = newVmSystemError(errors.New(errMsg))
		} else {
			isUncatchable := C.luaL_hasuncatchablerror(ce.L) != C.int(0)
			if isUncatchable && (errMsg == C.ERR_BF_TIMEOUT || errMsg == vmTimeoutErrMsg) {
				ce.err = &VmTimeoutError{}
			} else {
				ce.err = errors.New(errMsg)
			}
		}
		ctrLgr.Debug().Err(ce.err).Stringer(
			"contract",
			types.LogAddr(ce.ctx.curContract.contractId),
		).Msg("contract is failed")
		return 0
	}

	if target == nil {
		var errRet C.int
		retMsg := C.GoString(C.vm_get_json_ret(ce.L, nRet, &errRet))
		if errRet == 1 {
			ce.err = errors.New(retMsg)
		} else {
			ce.jsonRet = retMsg
		}
	} else {
		if c2ErrMsg := C.vm_copy_result(ce.L, target, nRet); c2ErrMsg != nil {
			errMsg := C.GoString(c2ErrMsg)
			ce.err = errors.New(errMsg)
			ctrLgr.Debug().Err(ce.err).Stringer(
				"contract",
				types.LogAddr(ce.ctx.curContract.contractId),
			).Msg("failed to move results")
		}
	}

	if ce.ctx.traceFile != nil {
		address := types.EncodeAddress(ce.ctx.curContract.contractId)
		codeFile := fmt.Sprintf("%s%s%s.code", os.TempDir(), string(os.PathSeparator), address)
		if _, err := os.Stat(codeFile); os.IsNotExist(err) {
			f, err := os.OpenFile(codeFile, os.O_WRONLY|os.O_CREATE, 0644)
			if err == nil {
				_, _ = f.Write(ce.code)
				_ = f.Close()
			}
		}
		_, _ = ce.ctx.traceFile.WriteString(fmt.Sprintf("contract %s used fee: %s\n",
			address, ce.ctx.usedFee().String()))
	}

	return nRet
}

func vmAutoload(L *LState, funcName string) bool {
	s := C.CString(funcName)
	loaded := C.vm_autoload(L, s)
	C.free(unsafe.Pointer(s))
	return loaded != C.int(0)
}

func (ce *executor) vmLoadCode(id []byte) {
	var chunkId *C.char
	if ce.ctx.blockInfo.ForkVersion >= 3 {
		chunkId = C.CString("@" + types.EncodeAddress(id))
	} else {
		chunkId = C.CString(hex.Encode(id))
	}
	defer C.free(unsafe.Pointer(chunkId))

	if cErrMsg := C.vm_loadbuff(
		ce.L,
		(*C.char)(unsafe.Pointer(&ce.code[0])),
		C.size_t(len(ce.code)),
		chunkId,
		ce.ctx.service - C.int(maxContext),
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		ce.err = errors.New(errMsg)
		ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(id)).Msg("failed to load code")
	}
}

func (ce *executor) vmLoadCall() {
	if cErrMsg := C.vm_loadcall(ce.L); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		isUncatchable := C.luaL_hasuncatchablerror(ce.L) != C.int(0)
		if isUncatchable && (errMsg == C.ERR_BF_TIMEOUT || errMsg == vmTimeoutErrMsg) {
			ce.err = &VmTimeoutError{}
		} else {
			ce.err = errors.New(errMsg)
		}
	}
	C.luaL_set_service(ce.L, ce.ctx.service)
}


////////////////////////////////////////////////////////////////////////////////
// GAS
////////////////////////////////////////////////////////////////////////////////

// set the remaining gas on the given LState
func (ctx *vmContext) setRemainingGas(L *LState) {
	if ctx.IsGasSystem() {
		C.lua_gasset(L, C.ulonglong(ctx.remainingGas))
		//defer func() {
			if ctrLgr.IsDebugEnabled() {
				ctrLgr.Debug().Uint64("gas used", ce.ctx.usedGas()).Str("lua vm", "loaded").Msg("gas information")
			}
		//}()
	}
}

func (ce *executor) setGas() {
	if ce == nil || ce.L == nil || ce.err != nil {
		return
	}
	C.lua_gasset(ce.L, C.ulonglong(ce.ctx.remainingGas))
}

func (ce *executor) gas() uint64 {
	return uint64(C.lua_gasget(ce.L))
}

////////////////////////////////////////////////////////////////////////////////

func setInstCount(ctx *vmContext, parent *LState, child *LState) {
	if !ctx.IsGasSystem() {
		C.vm_setinstcount(parent, C.vm_instcount(child))
	}
}

func setInstMinusCount(ctx *vmContext, L *LState, deduc C.int) {
	if !ctx.IsGasSystem() {
		C.vm_setinstcount(L, minusCallCount(ctx, C.vm_instcount(L), deduc))
	}
}

func minusCallCount(ctx *vmContext, curCount, deduc C.int) C.int {
	if ctx.IsGasSystem() {
		return 0
	}
	remain := curCount - deduc
	if remain <= 0 {
		remain = 1
	}
	return remain
}

////////////////////////////////////////////////////////////////////////////////


	if ctx.blockInfo.ForkVersion >= 2 {
		C.luaL_set_hardforkversion(ce.L, C.int(ctx.blockInfo.ForkVersion))
	}


