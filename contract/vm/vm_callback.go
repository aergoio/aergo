package contract

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.1
#cgo LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <string.h>
#include "vm.h"
#include "bignum_module.h"

struct proof {
	void *data;
	size_t len;
};

#define RLP_TSTRING 0
#define RLP_TLIST 1

struct rlp_obj {
	int rlp_obj_type;
	void *data;
	size_t size;
};
*/
import "C"
import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)



//export luaSetVariable
func luaSetVariable(L *LState, key unsafe.Pointer, keyLen C.int, value *C.char) *C.char {
	args := []string{C.GoBytes(key, keyLen), C.GoString(value)}
	result, err := sendRequest("set", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	if len(result) > 0 {
		return C.CString(result)
	}
	return nil
}

//export luaGetVariable
func luaGetVariable(L *LState, key unsafe.Pointer, keyLen C.int, blkno *C.char) (*C.char, *C.char) {
	args := []string{C.GoBytes(key, keyLen), C.GoString(blkno)}
	result, err := sendRequest("get", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return nil, C.CString(err.Error())
	}
	if len(result) > 0 {
		return C.CString(result), nil
	}
	return nil, nil
}

//export luaDelVariable
func luaDelVariable(L *LState, key unsafe.Pointer, keyLen C.int) *C.char {
	args := []string{C.GoBytes(key, keyLen)}
	result, err := sendRequest("del", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	if len(result) > 0 {
		return C.CString(result)
	}
	return nil
}

//export luaCallContract
func luaCallContract(L *LState,
	contractId *C.char, fname *C.char, arguments *C.char,
	amount *C.char, gas uint64,
) (C.int, *C.char) {

	contractAddress := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(arguments)
	amountStr := C.GoString(amount)
	gasStr := string((*[8]byte)(unsafe.Pointer(&gas))[:])

	args := []string{contractAddress, fnameStr, argsStr, amountStr, gasStr}
	result, err := sendRequest("call", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return -1, C.CString(err.Error())
	}
	return C.int(result), nil
}

//export luaDelegateCallContract
func luaDelegateCallContract(L *LState
	contractId *C.char, fname *C.char, arguments *C.char, gas uint64
) (C.int, *C.char) {

	contractAddress := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(arguments)
	gasStr := string((*[8]byte)(unsafe.Pointer(&gas))[:])

	args := []string{contractAddress, fnameStr, argsStr, gasStr}
	result, err := sendRequest("delegate-call", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return -1, C.CString(err.Error())
	}
	return C.int(result), nil
}

//export luaSendAmount
func luaSendAmount(L *LState, contractId *C.char, amount *C.char) *C.char {
	args := []string{C.GoString(contractId), C.GoString(amount)}
	result, err := sendRequest("send", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaPrint
func luaPrint(L *LState, arguments *C.char) {
	args := []string{C.GoString(arguments)}
	result, err := sendRequest("print", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaSetRecoveryPoint
func luaSetRecoveryPoint(L *LState) (C.int, *C.char) {
	args := []string{}
	result, err := sendRequest("setRecoveryPoint", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return -1, C.CString(err.Error())
	}
	return C.int(result), nil
}

//export luaClearRecovery
func luaClearRecovery(L *LState, start int, isError bool) *C.char {
	args := []string{fmt.Sprintf("%d", start), fmt.Sprintf("%d", int(isError))}
	result, err := sendRequest("clearRecovery", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaGetBalance
func luaGetBalance(L *LState, contractId *C.char) (*C.char, *C.char) {
	args := []string{C.GoString(contractId)}
	result, err := sendRequest("balance", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return nil, C.CString(err.Error())
	}
	return C.CString(result), nil
}

//export luaGetSender
func luaGetSender(L *LState) *C.char {
	ctx := contexts[service]
	setInstMinusCount(ctx, L, 1000)
	return C.CString(types.EncodeAddress(ctx.curContract.sender))
}

//export luaGetTxHash
func luaGetTxHash(L *LState) *C.char {
	args := []string{}
	result, err := sendRequest("getTxHash", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaGetBlockNo
func luaGetBlockNo(L *LState) C.lua_Integer {
	args := []string{}
	result, err := sendRequest("getBlockNo", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.lua_Integer(0)
	}
	return C.lua_Integer(result)
}

//export luaGetTimeStamp
func luaGetTimeStamp(L *LState) C.lua_Integer {
	args := []string{}
	result, err := sendRequest("getTimeStamp", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.lua_Integer(0)
	}
	return C.lua_Integer(result)
}

//export luaGetContractId
func luaGetContractId(L *LState) *C.char {
	args := []string{}
	result, err := sendRequest("getContractId", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaGetAmount
func luaGetAmount(L *LState) *C.char {
	args := []string{}
	result, err := sendRequest("getAmount", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaGetOrigin
func luaGetOrigin(L *LState) *C.char {
	args := []string{}
	result, err := sendRequest("getOrigin", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaGetPrevBlockHash
func luaGetPrevBlockHash(L *LState) *C.char {
	args := []string{}
	result, err := sendRequest("getPrevBlockHash", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	return C.CString(result)
}

func checkHexString(data string) bool {
	if len(data) >= 2 && data[0] == '0' && (data[1] == 'x' || data[1] == 'X') {
		return true
	}
	return false
}

//export luaCryptoSha256
func luaCryptoSha256(L *LState, arg unsafe.Pointer, argLen C.int) (*C.char, *C.char) {
	data := C.GoBytes(arg, argLen)
	args := []string{string(data)}
	result, err := sendRequest("sha256", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return nil, C.CString(err.Error())
	}
	return C.CString(result), nil
}

func decodeHex(hexStr string) ([]byte, error) {
	if checkHexString(hexStr) {
		hexStr = hexStr[2:]
	}
	return hex.Decode(hexStr)
}

//export luaECVerify
func luaECVerify(L *LState, msg *C.char, sig *C.char, addr *C.char) (C.int, *C.char) {
	args := []string{C.GoString(msg), C.GoString(sig), C.GoString(addr)}
	result, err := sendRequest("ecVerify", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.int(-1), C.CString(err.Error())
	}
	return C.int(result), nil
}

func luaCryptoToBytes(data unsafe.Pointer, dataLen C.int) ([]byte, bool) {
	var d []byte
	b := C.GoBytes(data, dataLen)
	isHex := checkHexString(string(b))
	if isHex {
		var err error
		d, err = hex.Decode(string(b[2:]))
		if err != nil {
			isHex = false
		}
	}
	if !isHex {
		d = b
	}
	return d, isHex
}

func luaCryptoRlpToBytes(data unsafe.Pointer) []byte {
	x := (*C.struct_rlp_obj)(data)
	if x.rlp_obj_type == C.RLP_TSTRING {
		b, _ := luaCryptoToBytes(x.data, C.int(x.size))
		// add a first byte to the byte array to indicate the type of the RLP object
		b = append([]byte{byte(C.RLP_TSTRING)}, b...)
		return b
	}
	elems := (*[1 << 30]C.struct_rlp_obj)(unsafe.Pointer(x.data))[:C.int(x.size):C.int(x.size)]
	list := make([][]byte, len(elems))
	for i, elem := range elems {
		b, _ := luaCryptoToBytes(elem.data, C.int(elem.size))
		list[i] = b
	}
	// serialize the list as a single byte array, including the type byte
	ret := SerializeMessageBytes([]byte{byte(C.RLP_TLIST)}, list...)
	return ret
}

//export luaCryptoVerifyProof
func luaCryptoVerifyProof(
	L *LState,
	key unsafe.Pointer, keyLen C.int,
	value unsafe.Pointer,
	hash unsafe.Pointer, hashLen C.int,
	proof unsafe.Pointer, nProof C.int,
) C.int {
	// convert to bytes
	k, _ := luaCryptoToBytes(key, keyLen)
	v := luaCryptoRlpToBytes(value)
	h, _ := luaCryptoToBytes(hash, hashLen)
	// read each proof element into a string array
	cProof := (*[1 << 30]C.struct_proof)(proof)[:nProof:nProof]
	proofElems := make([]string, int(nProof))
	for i, p := range cProof {
		data, _ := luaCryptoToBytes(p.data, C.int(p.len))
		proofElems[i] = string(data)
	}
	// convert the proof elements into a single byte array
	proofBytes := SerializeMessage(proofElems...)

	// send request
	args := []string{string(k), string(v), string(h), string(proofBytes)}
	result, err := sendRequest("verifyEthStorageProof", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.int(0)
	}
	return C.int(result)
}

//export luaCryptoKeccak256
func luaCryptoKeccak256(data unsafe.Pointer, dataLen C.int) (unsafe.Pointer, int) {
	args := []string{string(C.GoBytes(data, dataLen))}
	result, err := sendRequest("keccak256", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return nil, 0
	}
	return C.CBytes(result), len(result)
}

//export luaDeployContract
func luaDeployContract(
	L *LState,
	contract *C.char,
	arguments *C.char,
	amount *C.char,
) (C.int, *C.char) {

	contractStr := C.GoString(contract)
	argsStr := C.GoString(arguments)
	amountStr := C.GoString(amount)

	args := []string{contractStr, argsStr, amountStr}
	result, err := sendRequest("deploy", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return -1, C.CString(err.Error())
	}
	return C.int(result), nil
}

//export isPublic
func isPublic() C.int {
	if PubNet {
		return C.int(1)
	} else {
		return C.int(0)
	}
}

//export luaRandomInt
func luaRandomInt(L *LState, min, max C.int) C.int {
	args := []string{C.GoString(min), C.GoString(max)}
	result, err := sendRequest("randomInt", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.int(0)
	}
	return C.int(result)
}

//export luaEvent
func luaEvent(L *LState, eventName *C.char, arguments *C.char) *C.char {
	args := []string{C.GoString(eventName), C.GoString(arguments)}
	result, err := sendRequest("event", args)
	if err != nil {
		if isUncatchable(err) {
			C.luaL_setuncatchablerror(L)
		}
		return C.CString(err.Error())
	}
	if len(result) > 0 {
		return C.CString(result)
	}
	return nil
}

//export luaToPubkey
func luaToPubkey(L *LState, address *C.char) *C.char {
	args := []string{C.GoString(address)}
	result, err := sendRequest("toPubkey", args)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaToAddress
func luaToAddress(L *LState, pubkey *C.char) *C.char {
	args := []string{C.GoString(pubkey)}
	result, err := sendRequest("toAddress", args)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaIsContract
func luaIsContract(L *LState, contractId *C.char) (C.int, *C.char) {
	args := []string{C.GoString(contractId)}
	result, err := sendRequest("isContract", args)
	if err != nil {
		return -1, C.CString(err.Error())
	}
	return C.int(result), nil
}

//export luaNameResolve
func luaNameResolve(L *LState, name_or_address *C.char) *C.char {
	args := []string{C.GoString(name_or_address)}
	result, err := sendRequest("nameResolve", args)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString(result)
}

//export luaGovernance
func luaGovernance(L *LState, gType C.char, arg *C.char) *C.char {
	args := []string{C.GoString(gType), C.GoString(arg)}
	result, err := sendRequest("governance", args)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString(result)
}



// checks whether the block creation timeout occurred
//
//export luaCheckTimeout
func luaCheckTimeout() C.int {
	if timedout {
		return 1
	}
	return 0
}

//export luaIsFeeDelegation
func luaIsFeeDelegation(L *LState) (C.int, *C.char) {
	if ctx.isFeeDelegation {
		return 1, nil
	}
	return 0, nil
}

//export LuaGetDbHandleSnap
func LuaGetDbHandleSnap(L *LState, snap *C.char) *C.char {

	stateSet := contexts[service]
	curContract := stateSet.curContract
	callState := curContract.callState

	if stateSet.isQuery != true {
		return C.CString("[Contract.luaGetDBSnap] not permitted in transaction")
	}

	if callState.tx != nil {
		return C.CString("[Contract.luaGetDBSnap] transaction already started")
	}

	rp, err := strconv.ParseUint(C.GoString(snap), 10, 64)
	if err != nil {
		return C.CString("[Contract.luaGetDBSnap] snapshot is not valid" + C.GoString(snap))
	}

	aid := types.ToAccountID(curContract.contractId)
	tx, err := beginReadOnly(aid.String(), rp)
	if err != nil {
		return C.CString("Error Begin SQL Transaction")
	}

	callState.tx = tx
	return nil
}

//export LuaGetDbSnapshot
func LuaGetDbSnapshot(service C.int) *C.char {
	stateSet := contexts[service]
	curContract := stateSet.curContract

	return C.CString(strconv.FormatUint(curContract.rp, 10))
}

//export luaGetStaking
func luaGetStaking(L *LState, addr *C.char) (*C.char, C.lua_Integer, *C.char) {
	args := []string{C.GoString(addr)}
	result, err := sendRequest("getStaking", args)
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}
	// extract amount and when from result
	amount := result[0]
	when := result[1]
	return C.CString(amount), C.lua_Integer(when), nil
}
