package main

/*
#cgo CFLAGS: -I${SRCDIR}/../../libtool/include/luajit-2.1
#cgo LDFLAGS: ${SRCDIR}/../../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <string.h>
#include "vm.h"
#include "util.h"
#include "db_module.h"
#include "../db_msg.h"
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
	"fmt"
	"strconv"
	"strings"
	"unsafe"
	"errors"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/contract/msg"
)

var nestedView int

//export luaViewStart
func luaViewStart() {
	nestedView++
}

//export luaViewEnd
func luaViewEnd() {
	nestedView--
}

//export luaSetVariable
func luaSetVariable(L *LState, key *C.char, keyLen C.int, value *C.char) *C.char {
	args := []string{C.GoStringN(key, keyLen), C.GoString(value)}
	_, err := sendRequest("set", args)
	if err != nil {
		return handleError(L, err)
	}
	return nil
}

//export luaGetVariable
func luaGetVariable(L *LState, key *C.char, keyLen C.int, blkno *C.char) (*C.char, *C.char) {
	args := []string{C.GoStringN(key, keyLen), C.GoString(blkno)}
	result, err := sendRequest("get", args)
	if err != nil {
		return nil, handleError(L, err)
	}
	if len(result) > 0 {
		return C.CString(result), nil
	}
	return nil, nil
}

//export luaDelVariable
func luaDelVariable(L *LState, key *C.char, keyLen C.int) *C.char {
	args := []string{C.GoStringN(key, keyLen)}
	_, err := sendRequest("del", args)
	if err != nil {
		return handleError(L, err)
	}
	return nil
}

//export luaCallContract
func luaCallContract(L *LState,
	address *C.char, fname *C.char, arguments *C.char,
	amount *C.char, gas uint64,
) (*C.char, *C.char) {

	contractAddress := C.GoString(address)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(arguments)
	amountStr := C.GoString(amount)
	resourceLimit := getResourceLimit(gas)
	limitStr := string((*[8]byte)(unsafe.Pointer(&resourceLimit))[:])

	args := []string{contractAddress, fnameStr, argsStr, amountStr, limitStr}
	result, err := sendRequest("call", args)

	// extract the used gas or used instructions from the result
	usedResources, result := extractUsedResources(result)
	// deduct the used resources from the remaining
	err = updateRemainingResources(usedResources, err)

	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export luaDelegateCallContract
func luaDelegateCallContract(L *LState,
	address *C.char, fname *C.char, arguments *C.char, gas uint64,
) (*C.char, *C.char) {

	contractAddress := C.GoString(address)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(arguments)
	resourceLimit := getResourceLimit(gas)
	limitStr := string((*[8]byte)(unsafe.Pointer(&resourceLimit))[:])

	args := []string{contractAddress, fnameStr, argsStr, limitStr}
	result, err := sendRequest("delegate-call", args)

	// extract the used gas or used instructions from the result
	usedResources, result := extractUsedResources(result)
	// deduct the used resources from the remaining
	err = updateRemainingResources(usedResources, err)

	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export luaSendAmount
func luaSendAmount(L *LState, address *C.char, amount *C.char) *C.char {
	resourceLimit := getResourceLimit(0)
	limitStr := string((*[8]byte)(unsafe.Pointer(&resourceLimit))[:])
	args := []string{C.GoString(address), C.GoString(amount), limitStr}
	result, err := sendRequest("send", args)

	// extract the used gas or used instructions from the result
	usedResources, result := extractUsedResources(result)
	// deduct the used resources from the remaining
	err = updateRemainingResources(usedResources, err)

	if err != nil {
		return handleError(L, err)
	}
	// it does not return the result
	return nil
}

//export luaPrint
func luaPrint(L *LState, arguments *C.char) *C.char {
	args := []string{C.GoString(arguments)}
	_, err := sendRequest("print", args)
	if err != nil {
		return handleError(L, err)
	}
	return nil
}

//export luaSetRecoveryPoint
func luaSetRecoveryPoint(L *LState) (C.int, *C.char) {
	args := []string{}
	result, err := sendRequest("setRecoveryPoint", args)
	if err != nil {
		return -1, handleError(L, err)
	}
	// if on a query or inside a view function
	if result == "" {
		return 0, nil
	}
	resultInt, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return -1, handleError(L, fmt.Errorf("uncatchable: luaSetRecoveryPoint: failed to parse result: %v", err))
	}
	return C.int(resultInt), nil
}

//export luaClearRecovery
func luaClearRecovery(L *LState, start int, isError bool) *C.char {
	args := []string{fmt.Sprintf("%d", start), fmt.Sprintf("%t", isError)}
	_, err := sendRequest("clearRecovery", args)
	if err != nil {
		return handleError(L, err)
	}
	return nil
}

//export luaGetBalance
func luaGetBalance(L *LState, address *C.char) (*C.char, *C.char) {
	args := []string{C.GoString(address)}
	result, err := sendRequest("balance", args)
	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export luaGetSender
func luaGetSender(L *LState) *C.char {
	return C.CString(contractCaller)
}

//export luaGetTxHash
func luaGetTxHash(L *LState) (*C.char, *C.char) {
	args := []string{}
	result, err := sendRequest("getTxHash", args)
	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export luaGetBlockNo
func luaGetBlockNo(L *LState) (C.lua_Integer, *C.char) {
	args := []string{}
	result, err := sendRequest("getBlockNo", args)
	if err != nil {
		return C.lua_Integer(0), handleError(L, err)
	}
	blockNo, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return C.lua_Integer(0), handleError(L, fmt.Errorf("uncatchable: luaGetBlockNo: failed to parse result: %v", err))
	}
	return C.lua_Integer(blockNo), nil
}

//export luaGetTimeStamp
func luaGetTimeStamp(L *LState) (C.lua_Integer, *C.char) {
	args := []string{}
	result, err := sendRequest("getTimeStamp", args)
	if err != nil {
		return C.lua_Integer(0), handleError(L, err)
	}
	timestamp, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return C.lua_Integer(0), handleError(L, fmt.Errorf("uncatchable: failed to parse timestamp: %v", err))
	}
	return C.lua_Integer(timestamp), nil
}

//export luaGetContractId
func luaGetContractId(L *LState) *C.char {
	return C.CString(contractAddress)
}

//export luaGetAmount
func luaGetAmount(L *LState) (*C.char, *C.char) {
	args := []string{}
	result, err := sendRequest("getAmount", args)
	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export luaGetOrigin
func luaGetOrigin(L *LState) (*C.char, *C.char) {
	args := []string{}
	result, err := sendRequest("getOrigin", args)
	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export luaGetPrevBlockHash
func luaGetPrevBlockHash(L *LState) (*C.char, *C.char) {
	args := []string{}
	result, err := sendRequest("getPrevBlockHash", args)
	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
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
		return nil, handleError(L, err)
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
		return C.int(-1), handleError(L, err)
	}
	resultInt, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return C.int(-1), handleError(L, fmt.Errorf("uncatchable: luaECVerify: failed to parse result: %v", err))
	}
	return C.int(resultInt), nil
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
	ret := msg.SerializeMessageBytes(append([][]byte{[]byte{byte(C.RLP_TLIST)}}, list...)...)
	return ret
}

//export luaCryptoVerifyProof
func luaCryptoVerifyProof(
	L *LState,
	key unsafe.Pointer, keyLen C.int,
	value unsafe.Pointer,
	hash unsafe.Pointer, hashLen C.int,
	proof unsafe.Pointer, nProof C.int,
) (C.int, *C.char) {
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
	proofBytes := msg.SerializeMessage(proofElems...)

	// send request
	args := []string{string(k), string(v), string(h), string(proofBytes)}
	result, err := sendRequest("verifyEthStorageProof", args)
	if err != nil {
		return C.int(0), handleError(L, err)
	}
	resultInt, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return C.int(0), handleError(L, fmt.Errorf("uncatchable: luaCryptoVerifyProof: failed to parse result: %v", err))
	}
	return C.int(resultInt), nil
}

//export luaCryptoKeccak256
func luaCryptoKeccak256(L *LState, data *C.char, dataLen C.int) (unsafe.Pointer, C.int, *C.char) {
	args := []string{C.GoStringN(data, dataLen)}
	result, err := sendRequest("keccak256", args)
	if err != nil {
		return nil, 0, handleError(L, err)
	}
	return C.CBytes([]byte(result)), C.int(len(result)), nil
}

//export luaDeployContract
func luaDeployContract(
	L *LState,
	contract *C.char,
	arguments *C.char,
	amount *C.char,
) (*C.char, *C.char) {

	contractStr := C.GoString(contract)
	argsStr := C.GoString(arguments)
	amountStr := C.GoString(amount)
	resourceLimit := getResourceLimit(0)
	limitStr := string((*[8]byte)(unsafe.Pointer(&resourceLimit))[:])

	args := []string{contractStr, argsStr, amountStr, limitStr}
	result, err := sendRequest("deploy", args)

	// extract the used gas or used instructions from the result
	usedResources, result := extractUsedResources(result)
	// deduct the used resources from the remaining
	err = updateRemainingResources(usedResources, err)

	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export isPublic
func isPublic() C.int {
	if isPubNet {
		return C.int(1)
	} else {
		return C.int(0)
	}
}

//export luaRandomInt
func luaRandomInt(L *LState, min, max C.int) (C.int, *C.char) {
	args := []string{fmt.Sprintf("%d", min), fmt.Sprintf("%d", max)}
	result, err := sendRequest("randomInt", args)
	if err != nil {
		return C.int(0), handleError(L, err)
	}
	value, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return C.int(0), handleError(L, fmt.Errorf("uncatchable: luaRandomInt: failed to parse result: %v", err))
	}
	return C.int(value), nil
}

//export luaEvent
func luaEvent(L *LState, eventName *C.char, arguments *C.char) *C.char {
	args := []string{C.GoString(eventName), C.GoString(arguments)}
	_, err := sendRequest("event", args)
	if err != nil {
		return handleError(L, err)
	}
	return nil
}

//export luaToPubkey
func luaToPubkey(L *LState, address *C.char) (*C.char, *C.char) {
	args := []string{C.GoString(address)}
	result, err := sendRequest("toPubkey", args)
	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export luaToAddress
func luaToAddress(L *LState, pubkey *C.char) (*C.char, *C.char) {
	args := []string{C.GoString(pubkey)}
	result, err := sendRequest("toAddress", args)
	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export luaIsContract
func luaIsContract(L *LState, address *C.char) (C.int, *C.char) {
	args := []string{C.GoString(address)}
	result, err := sendRequest("isContract", args)
	if err != nil {
		return -1, handleError(L, err)
	}
	resultInt, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return -1, handleError(L, fmt.Errorf("uncatchable: luaIsContract: failed to parse result: %v", err))
	}
	return C.int(resultInt), nil
}

//export luaNameResolve
func luaNameResolve(L *LState, name_or_address *C.char) (*C.char, *C.char) {
	args := []string{C.GoString(name_or_address)}
	result, err := sendRequest("nameResolve", args)
	if err != nil {
		return nil, handleError(L, err)
	}
	return C.CString(result), nil
}

//export luaGovernance
func luaGovernance(L *LState, gType C.char, arg *C.char) *C.char {
	args := []string{fmt.Sprintf("%c", gType), C.GoString(arg)}
	_, err := sendRequest("governance", args)
	if err != nil {
		return handleError(L, err)
	}
	return nil
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
func luaIsFeeDelegation() C.bool {
	return C.bool(contractIsFeeDelegation)
}

//export luaGetStaking
func luaGetStaking(L *LState, addr *C.char) (*C.char, C.lua_Integer, *C.char) {
	args := []string{C.GoString(addr)}
	result, err := sendRequest("getStaking", args)
	if err != nil {
		return nil, 0, handleError(L, err)
	}
	// extract amount and when from result - result = staking.GetAmountBigInt().String() + "," + staking.When.String()
	sep := strings.Index(result, ",")
	amount := result[:sep]
	when, err := strconv.ParseInt(result[sep+1:], 10, 64)
	if err != nil {
		return nil, 0, handleError(L, fmt.Errorf("uncatchable: luaGetStaking: failed to parse 'when': %v", err))
	}
	return C.CString(amount), C.lua_Integer(when), nil
}

func getResourceLimit(gas uint64) uint64 {
	if IsGasSystem() {
		return getGasLimit(gas)
	} else {
		return getRemainingInstructions()
	}
}

func getGasLimit(definedGasLimit uint64) uint64 {

	remainingGas := getRemainingGas()

	if definedGasLimit > 0 && definedGasLimit < remainingGas {
		// if specified via contract.call.gas(limit)(...)
		return definedGasLimit
	} else {
		// if not specified, use the remaining gas from the transaction
		return remainingGas
	}
}

//export luaSendRequest
func luaSendRequest(L *LState, method *C.char, arguments *C.buffer, response *C.rresponse) {
	var args []string
	if arguments != nil {
		args = []string{C.GoStringN(arguments.ptr, arguments.len)}
	} else {
		args = []string{}
	}
	result, err := sendRequest(C.GoString(method), args)
	if err != nil {
		response.error = handleError(L, err)
	} else {
		response.result.ptr = C.CString(result)
		response.result.len = C.int(len(result))
	}
}

var sendRequest = sendRequestFunc

func sendRequestFunc(method string, args []string) (string, error) {

	// send the execution request to the VM instance
	err := sendApiMessage(method, args)
	if err != nil {
		return "", err
	}

	// wait for the response
	response, err := msg.WaitForMessage(conn, time.Time{})
	if err != nil {
		return "", err  //FIXME: this is a system error
	}

	/*/ decrypt the message
	response, err = msg.Decrypt(response, secretKey)
	if err != nil {
		return "", err
	}
	*/

	list, err := msg.DeserializeMessage(response)
	if err != nil {
		return "", err
	}

	result := list[0]
	errstr := list[1]
	if errstr != "" {
		err = errors.New(errstr)
	}

	// return the result
	return result, err
}

func sendApiMessage(method string, args []string) error {

	inViewStr := "0"
	if nestedView > 0 {
		inViewStr = "1"
	}

	// create new slice with the method, args and whether it is within a view function
	list := []string{method}
	list = append(list, args...)
	list = append(list, inViewStr)

	return sendMessage(list)
}

func sendMessage(list []string) error {

	// build the message
	message := msg.SerializeMessage(list...)

	/*/ encrypt the message
	message, err = msg.Encrypt(message, secretKey)
	if err != nil {
		fmt.Printf("Error: failed to encrypt message: %v\n", err)
		closeApp(1)
	} */

	// send the message to the VM API
	return msg.SendMessage(conn, message)
}

func handleError(L *LState, err error) *C.char {
	errstr := err.Error()
	if strings.HasPrefix(errstr, "uncatchable: ") {
		errstr = errstr[len("uncatchable: "):]
		C.luaL_setuncatchablerror(L)
	}
	if strings.HasPrefix(errstr, "syserror: ") {
		errstr = errstr[len("syserror: "):]
		C.luaL_setsyserror(L)
	}
	return C.CString(errstr)
}
