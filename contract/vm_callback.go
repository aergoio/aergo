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
	"math/big"
	"strconv"
	"strings"
	"unsafe"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
	"github.com/aergoio/aergo/v2/blacklist"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

var (
	mulAergo, mulGaer, zeroBig *big.Int
	vmLogger                   = log.NewLogger("contract.vm")
)

const (
	maxEventCntV2     = 50
	maxEventCntV4     = 128
	maxEventNameSize  = 64
	maxEventArgSize   = 4096
	luaCallCountDeduc = 1000
)

func init() {
	mulAergo = types.NewAmount(1, types.Aergo)
	mulGaer = types.NewAmount(1, types.Gaer)
	zeroBig = types.NewZeroAmount()
}

func maxEventCnt(ctx *vmContext) int32 {
	if ctx.blockInfo.ForkVersion >= 4 {
		return maxEventCntV4
	} else {
		return maxEventCntV2
	}
}

//export luaSetDB
func luaSetDB(L *LState, service C.int, key unsafe.Pointer, keyLen C.int, value *C.char) *C.char {
	ctx := contexts[service]
	if ctx == nil {
		return C.CString("[System.LuaSetDB] contract state not found")
	}
	if ctx.isQuery == true || ctx.nestedView > 0 {
		return C.CString("[System.LuaSetDB] set not permitted in query")
	}

	keyBytes := C.GoBytes(key, keyLen)
	keyStr := string(keyBytes)
	valueStr := C.GoString(value)
	valueBytes := []byte(valueStr)

	// log the operation with the modified key
	logOperation(ctx, "", "set_variable", convertKey(keyStr), valueStr)

	// set the state variable
	if err := ctx.curContract.callState.ctrState.SetData(keyBytes, valueBytes); err != nil {
		return C.CString(err.Error())
	}
	if err := ctx.addUpdateSize(int64(types.HashIDLength + len(valueBytes))); err != nil {
		C.luaL_setuncatchablerror(L)
		return C.CString(err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString("[Set]\n")
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("Key=%s Len=%v byte=%v\n",
			string(keyBytes), keyLen, keyBytes))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("Data=%s Len=%d byte=%v\n",
			string(valueBytes), len(valueBytes), valueBytes))
	}
	return nil
}

//export luaGetDB
func luaGetDB(L *LState, service C.int, key unsafe.Pointer, keyLen C.int, blkno *C.char) (*C.char, *C.char) {
	ctx := contexts[service]
	if ctx == nil {
		return nil, C.CString("[System.LuaGetDB] contract state not found")
	}
	if blkno != nil {
		bigNo, _ := new(big.Int).SetString(strings.TrimSpace(C.GoString(blkno)), 10)
		if bigNo == nil || bigNo.Sign() < 0 {
			return nil, C.CString("[System.LuaGetDB] invalid blockheight value :" + C.GoString(blkno))
		}
		blkNo := bigNo.Uint64()

		chainBlockHeight := ctx.blockInfo.No
		if chainBlockHeight == 0 {
			bestBlock, err := ctx.cdb.GetBestBlock()
			if err != nil {
				return nil, C.CString("[System.LuaGetDB] get best block error")
			}
			chainBlockHeight = bestBlock.GetHeader().GetBlockNo()
		}
		if blkNo < chainBlockHeight {
			blk, err := ctx.cdb.GetBlockByNo(blkNo)
			if err != nil {
				return nil, C.CString(err.Error())
			}
			accountId := types.ToAccountID(ctx.curContract.contractId)
			contractProof, err := ctx.bs.GetAccountAndProof(accountId[:], blk.GetHeader().GetBlocksRootHash(), false)
			if err != nil {
				return nil, C.CString("[System.LuaGetDB] failed to get snapshot state for account")
			} else if contractProof.Inclusion {
				trieKey := common.Hasher(C.GoBytes(key, keyLen))
				varProof, err := ctx.bs.GetVarAndProof(trieKey, contractProof.GetState().GetStorageRoot(), false)
				if err != nil {
					return nil, C.CString("[System.LuaGetDB] failed to get snapshot state variable in contract")
				}
				if varProof.Inclusion {
					if len(varProof.GetValue()) == 0 {
						return nil, nil
					}
					return C.CString(string(varProof.GetValue())), nil
				}
			}
			return nil, nil
		}
	}

	data, err := ctx.curContract.callState.ctrState.GetData(C.GoBytes(key, keyLen))
	if err != nil {
		return nil, C.CString(err.Error())
	}
	if data == nil {
		return nil, nil
	}
	return C.CString(string(data)), nil
}

//export luaDelDB
func luaDelDB(L *LState, service C.int, key unsafe.Pointer, keyLen C.int) *C.char {
	ctx := contexts[service]
	if ctx == nil {
		return C.CString("[System.LuaDelDB] contract state not found")
	}
	if ctx.isQuery == true || ctx.nestedView > 0 {
		return C.CString("[System.LuaDelDB] delete not permitted in query")
	}

	keyBytes := C.GoBytes(key, keyLen)
	keyStr := string(keyBytes)

	// log the operation with the modified key
	logOperation(ctx, "", "del_variable", convertKey(keyStr))

	// delete the state variable
	if err := ctx.curContract.callState.ctrState.DeleteData(keyBytes); err != nil {
		return C.CString(err.Error())
	}
	if err := ctx.addUpdateSize(int64(32)); err != nil {
		C.luaL_setuncatchablerror(L)
		return C.CString(err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString("[Del]\n")
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("Key=%s Len=%v byte=%v\n",
			string(keyBytes), keyLen, keyBytes))
	}
	return nil
}

func convertKey(keyStr string) string {
	// if the key starts with "_sv_meta-len_", remove it and add the ".length" suffix to the key
	if strings.HasPrefix(keyStr, "_sv_meta-len_") {
		keyStr = keyStr[len("_sv_meta-len_"):]
		keyStr = keyStr + ".length"
	// if the key starts with "_sv_meta-type_", remove it and add the ".type" suffix to the key
	} else if strings.HasPrefix(keyStr, "_sv_meta-type_") {
		keyStr = keyStr[len("_sv_meta-type_"):]
		keyStr = keyStr + ".type"
	// if the key starts with "_sv_", remove it
	} else if strings.HasPrefix(keyStr, "_sv_") {
		keyStr = keyStr[len("_sv_"):]
		// if there is a "-" in the key, get its position, replace it with a `[` and add a `]` to the end
		if idx := strings.Index(keyStr, "-"); idx != -1 {
			keyStr = keyStr[:idx] + "[" + keyStr[idx+1:] + "]"
		}
	// if the key starts with "_", remove it
	} else if strings.HasPrefix(keyStr, "_") {
		keyStr = keyStr[len("_"):]
	}
	return keyStr
}

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

//export luaCallContract
func luaCallContract(L *LState, service C.int, contractId *C.char, fname *C.char, args *C.char,
	amount *C.char, gas uint64) (ret C.int, errormsg *C.char) {
	contractAddress := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)
	amountStr := C.GoString(amount)

	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.LuaCallContract] contract state not found")
	}

	opId := logOperation(ctx, amountStr, "call", contractAddress, fnameStr, argsStr)
	defer func() {
		if errormsg != nil {
			logOperationResult(ctx, opId, C.GoString(errormsg))
		}
	}()

	// get the contract address
	cid, err := getAddressNameResolved(contractAddress, ctx.bs)
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] invalid contractId: " + err.Error())
	}
	aid := types.ToAccountID(cid)
	contractAddress = types.EncodeAddress(cid)

	// read the amount for the contract call
	amountBig, err := transformAmount(amountStr, ctx.blockInfo.ForkVersion)
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] invalid amount: " + err.Error())
	}

	// get the contract state
	cs, err := getContractState(ctx, cid)
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] getAccount error: " + err.Error())
	}

	// check if the contract exists
	bytecode := getContractCode(cs.ctrState, ctx.bs)
	if bytecode == nil {
		return -1, C.CString("[Contract.LuaCallContract] cannot find contract " + contractAddress)
	}

	prevContractInfo := ctx.curContract

	// read the arguments for the contract call
	var ci types.CallInfo
	ci.Name = fnameStr
	err = getCallInfo(&ci.Args, []byte(argsStr), cid)
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] invalid arguments: " + err.Error())
	}

	// get the remaining gas from the parent LState
	ctx.refreshRemainingGas(L)
	// create a new executor with the remaining gas on the child LState
	ce := newExecutor(bytecode, cid, ctx, &ci, amountBig, false, false, cs.ctrState)
	defer func() {
		// save the result if the call was successful
		if ce.preErr == nil && ce.err == nil {
			logOperationResult(ctx, opId, ce.jsonRet)
		}
		// close the executor, closes also the child LState
		ce.close()
		// set the remaining gas on the parent LState
		ctx.setRemainingGas(L)
	}()

	if ce.err != nil {
		return -1, C.CString("[Contract.LuaCallContract] newExecutor error: " + ce.err.Error())
	}

	// send the amount to the contract
	senderState := prevContractInfo.callState.accState
	receiverState := cs.accState
	if amountBig.Cmp(zeroBig) > 0 {
		if ctx.isQuery == true || ctx.nestedView > 0 {
			return -1, C.CString("[Contract.LuaCallContract] send not permitted in query")
		}
		if r := sendBalance(senderState, receiverState, amountBig); r != nil {
			return -1, r
		}
	}

	seq, err := createRecoveryPoint(aid, ctx, senderState, cs, amountBig, false, false)
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CALL Contract %v(%v) %v]\n",
			contractAddress, aid.String(), fnameStr))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("SendBalance: %s\n", amountBig.String()))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
			senderState.Balance().String(), receiverState.Balance().String()))
	}
	if err != nil {
		return -1, C.CString("[System.LuaCallContract] database error: " + err.Error())
	}

	// set the current contract info
	ctx.curContract = newContractInfo(cs, prevContractInfo.contractId, cid,
		receiverState.RP(), amountBig)
	defer func() {
		ctx.curContract = prevContractInfo
	}()

	// execute the contract call
	defer setInstCount(ctx, L, ce.L)
	ret = ce.call(minusCallCount(ctx, C.vm_instcount(L), luaCallCountDeduc), L)

	// check if the contract call failed
	if ce.err != nil {
		err := clearRecoveryPoint(L, ctx, seq, true)
		if err != nil {
			return -1, C.CString("[Contract.LuaCallContract] recovery err: " + err.Error())
		}
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
		}
		switch ceErr := ce.err.(type) {
		case *VmTimeoutError:
			return -1, C.CString(ceErr.Error())
		default:
			return -1, C.CString("[Contract.LuaCallContract] call err: " + ceErr.Error())

		}
	}

	if seq == 1 {
		err := clearRecoveryPoint(L, ctx, seq, false)
		if err != nil {
			return -1, C.CString("[Contract.LuaCallContract] recovery err: " + err.Error())
		}
	}

	return ret, nil
}

//export luaDelegateCallContract
func luaDelegateCallContract(L *LState, service C.int, contractId *C.char,
	fname *C.char, args *C.char, gas uint64) (ret C.int, errormsg *C.char) {
	contractIdStr := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] contract state not found")
	}

	var isMultiCall bool
	var cid []byte
	var err error

	opId := logOperation(ctx, "", "delegate-call", contractIdStr, fnameStr, argsStr)
	defer func() {
		if errormsg != nil {
			logOperationResult(ctx, opId, C.GoString(errormsg))
		}
	}()

	// get the contract address
	if contractIdStr == "multicall" {
		isMultiCall = true
		argsStr = fnameStr
		fnameStr = "execute"
		cid = ctx.curContract.contractId
	} else {
		cid, err = getAddressNameResolved(contractIdStr, ctx.bs)
		if err != nil {
			return -1, C.CString("[Contract.LuaDelegateCallContract] invalid contractId: " + err.Error())
		}
		contractIdStr = types.EncodeAddress(cid)
	}
	aid := types.ToAccountID(cid)

	// get the contract state
	var contractState *statedb.ContractState
	if isMultiCall {
		contractState = statedb.GetMultiCallState(cid, ctx.curContract.callState.ctrState.State)
	} else {
		contractState, err = getOnlyContractState(ctx, cid)
	}
	if err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract]getContractState error" + err.Error())
	}

	// get the contract code
	var bytecode []byte
	if isMultiCall {
		bytecode = getMultiCallContractCode(contractState)
	} else {
		bytecode = getContractCode(contractState, ctx.bs)
	}
	if bytecode == nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] cannot find contract " + contractIdStr)
	}

	// read the arguments for the contract call
	var ci types.CallInfo
	if isMultiCall {
		err = getMultiCallInfo(&ci, []byte(argsStr))
	} else {
		ci.Name = fnameStr
		err = getCallInfo(&ci.Args, []byte(argsStr), cid)
	}
	if err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] invalid arguments: " + err.Error())
	}

	// get the remaining gas from the parent LState
	ctx.refreshRemainingGas(L)
	// create a new executor with the remaining gas on the child LState
	ce := newExecutor(bytecode, cid, ctx, &ci, zeroBig, false, false, contractState)
	defer func() {
		// save the result if the call was successful
		if ce.preErr == nil && ce.err == nil {
			logOperationResult(ctx, opId, ce.jsonRet)
		}
		// close the executor, closes also the child LState
		ce.close()
		// set the remaining gas on the parent LState
		ctx.setRemainingGas(L)
	}()

	if ce.err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] newExecutor error: " + ce.err.Error())
	}

	seq, err := createRecoveryPoint(aid, ctx, nil, ctx.curContract.callState, zeroBig, false, false)
	if err != nil {
		return -1, C.CString("[System.LuaDelegateCallContract] database error: " + err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[DELEGATECALL Contract %v %v]\n", contractIdStr, fnameStr))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
	}

	// execute the contract call
	defer setInstCount(ctx, L, ce.L)
	ret = ce.call(minusCallCount(ctx, C.vm_instcount(L), luaCallCountDeduc), L)

	// check if the contract call failed
	if ce.err != nil {
		err := clearRecoveryPoint(L, ctx, seq, true)
		if err != nil {
			return -1, C.CString("[Contract.LuaDelegateCallContract] recovery error: " + err.Error())
		}
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
		}
		switch ceErr := ce.err.(type) {
		case *VmTimeoutError:
			return -1, C.CString(ceErr.Error())
		default:
			return -1, C.CString("[Contract.LuaDelegateCallContract] call error: " + ce.err.Error())
		}
	}

	if seq == 1 {
		err := clearRecoveryPoint(L, ctx, seq, false)
		if err != nil {
			return -1, C.CString("[Contract.LuaDelegateCallContract] recovery error: " + err.Error())
		}
	}

	return ret, nil
}

func getAddressNameResolved(account string, bs *state.BlockState) ([]byte, error) {
	accountLen := len(account)
	if accountLen == types.EncodedAddressLength {
		return types.DecodeAddress(account)
	} else if accountLen == types.NameLength {
		cid, err := name.Resolve(bs, []byte(account), false)
		if err != nil {
			return nil, err
		}
		if cid == nil {
			return nil, errors.New("name not founded :" + account)
		}
		return cid, nil
	}
	return nil, errors.New("invalid account length:" + account)
}

//export luaSendAmount
func luaSendAmount(L *LState, service C.int, contractId *C.char, amount *C.char) (errormsg *C.char) {
	contractAddress := C.GoString(contractId)
	amountStr := C.GoString(amount)

	ctx := contexts[service]
	if ctx == nil {
		return C.CString("[Contract.LuaSendAmount] contract state not found")
	}

	opId := logOperation(ctx, amountStr, "send", contractAddress)
	defer func() {
		if errormsg != nil {
			logOperationResult(ctx, opId, C.GoString(errormsg))
		}
	}()

	// read the amount to be sent
	amountBig, err := transformAmount(amountStr, ctx.blockInfo.ForkVersion)
	if err != nil {
		return C.CString("[Contract.LuaSendAmount] invalid amount: " + err.Error())
	}

	// cannot send amount in query
	if (ctx.isQuery == true || ctx.nestedView > 0) && amountBig.Cmp(zeroBig) > 0 {
		return C.CString("[Contract.LuaSendAmount] send not permitted in query")
	}

	// get the receiver account
	cid, err := getAddressNameResolved(contractAddress, ctx.bs)
	if err != nil {
		return C.CString("[Contract.LuaSendAmount] invalid contractId: " + err.Error())
	}
	contractAddress = types.EncodeAddress(cid)

	// get the receiver state
	aid := types.ToAccountID(cid)
	cs, err := getCallState(ctx, cid)
	if err != nil {
		return C.CString("[Contract.LuaSendAmount] getAccount error: " + err.Error())
	}

	// get the sender state
	senderState := ctx.curContract.callState.accState
	receiverState := cs.accState

	// check if the receiver is a contract
	if len(receiverState.CodeHash()) > 0 {

		// get the contract state
		if cs.ctrState == nil {
			cs.ctrState, err = statedb.OpenContractState(cid, receiverState.State(), ctx.bs.StateDB)
			if err != nil {
				return C.CString("[Contract.LuaSendAmount] getContractState error: " + err.Error())
			}
		}

		// set the function to be called
		var ci types.CallInfo
		ci.Name = "default"

		// get the contract code
		bytecode := getContractCode(cs.ctrState, ctx.bs)
		if bytecode == nil {
			return C.CString("[Contract.LuaSendAmount] cannot find contract:" + contractAddress)
		}

		// get the remaining gas from the parent LState
		ctx.refreshRemainingGas(L)
		// create a new executor with the remaining gas on the child LState
		ce := newExecutor(bytecode, cid, ctx, &ci, amountBig, false, false, cs.ctrState)
		defer func() {
			// save the result if the call was successful
			if ce.preErr == nil && ce.err == nil {
				logOperationResult(ctx, opId, ce.jsonRet)
			}
			// close the executor, closes also the child LState
			ce.close()
			// set the remaining gas on the parent LState
			ctx.setRemainingGas(L)
		}()

		if ce.err != nil {
			return C.CString("[Contract.LuaSendAmount] newExecutor error: " + ce.err.Error())
		}

		// send the amount to the contract
		if amountBig.Cmp(zeroBig) > 0 {
			if r := sendBalance(senderState, receiverState, amountBig); r != nil {
				return r
			}
		}

		// create a recovery point
		seq, err := createRecoveryPoint(aid, ctx, senderState, cs, amountBig, false, false)
		if err != nil {
			return C.CString("[System.LuaSendAmount] database error: " + err.Error())
		}

		// log some info
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(
				fmt.Sprintf("[Send Call default] %s(%s) : %s\n", types.EncodeAddress(cid), aid.String(), amountBig.String()))
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
				senderState.Balance().String(), receiverState.Balance().String()))
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
		}

		// set the current contract info
		prevContractInfo := ctx.curContract
		ctx.curContract = newContractInfo(cs, prevContractInfo.contractId, cid,
			receiverState.RP(), amountBig)
		defer func() {
			ctx.curContract = prevContractInfo
		}()

		// execute the contract call
		defer setInstCount(ctx, L, ce.L)
		ce.call(minusCallCount(ctx, C.vm_instcount(L), luaCallCountDeduc), L)

		// check if the contract call failed
		if ce.err != nil {
			// recover to the previous state
			err := clearRecoveryPoint(L, ctx, seq, true)
			if err != nil {
				return C.CString("[Contract.LuaSendAmount] recovery err: " + err.Error())
			}
			// log some info
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
			}
			// return the error message
			return C.CString("[Contract.LuaSendAmount] call err: " + ce.err.Error())
		}

		if seq == 1 {
			err := clearRecoveryPoint(L, ctx, seq, false)
			if err != nil {
				return C.CString("[Contract.LuaSendAmount] recovery err: " + err.Error())
			}
		}

		// the transfer and contract call succeeded
		return nil
	}

	// the receiver is not a contract, just send the amount

	// if amount is zero, do nothing
	if amountBig.Cmp(zeroBig) == 0 {
		return nil
	}

	// send the amount to the receiver
	if r := sendBalance(senderState, receiverState, amountBig); r != nil {
		return r
	}

	// update the recovery point
	if ctx.lastRecoveryPoint != nil {
		_, _ = createRecoveryPoint(aid, ctx, senderState, cs, amountBig, true, false)
	}

	// log some info
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[Send] %s(%s) : %s\n",
			types.EncodeAddress(cid), aid.String(), amountBig.String()))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
			senderState.Balance().String(), receiverState.Balance().String()))
	}

	return nil
}

//export luaPrint
func luaPrint(L *LState, service C.int, args *C.char) {
	ctx := contexts[service]
	setInstMinusCount(ctx, L, 1000)
	ctrLgr.Info().Str("Contract SystemPrint", types.EncodeAddress(ctx.curContract.contractId)).Msg(C.GoString(args))
}

//export luaSetRecoveryPoint
func luaSetRecoveryPoint(L *LState, service C.int) (C.int, *C.char) {
	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.pcall] contract state not found")
	}
	if ctx.isQuery == true || ctx.nestedView > 0 {
		return 0, nil
	}
	curContract := ctx.curContract
	// if it is the multicall code, ignore
	if curContract.callState.ctrState.IsMultiCall() {
		return 0, nil
	}
	seq, err := createRecoveryPoint(types.ToAccountID(curContract.contractId), ctx, nil,
		curContract.callState, zeroBig, false, false)
	if err != nil {
		return -1, C.CString("[Contract.pcall] database error: " + err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[Pcall] snapshot set %d\n", seq))
	}
	return C.int(seq), nil
}

func clearRecoveryPoint(L *LState, ctx *vmContext, start int, isError bool) error {
	item := ctx.lastRecoveryPoint
	for {
		if isError {
			if item.revertState(ctx.bs) != nil {
				return errors.New("database error")
			}
		}
		if item.seq == start {
			if isError || item.prev == nil {
				ctx.lastRecoveryPoint = item.prev
			}
			return nil
		}
		item = item.prev
		if item == nil {
			return errors.New("internal error")
		}
	}
}

//export luaClearRecovery
func luaClearRecovery(L *LState, service C.int, start int, isError bool) *C.char {
	ctx := contexts[service]
	if ctx == nil {
		return C.CString("[Contract.pcall] contract state not found")
	}
	err := clearRecoveryPoint(L, ctx, start, isError)
	if err != nil {
		return C.CString(err.Error())
	}
	if ctx.traceFile != nil && isError == true {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("pcall recovery snapshot : %d\n", start))
	}
	return nil
}

//export luaGetBalance
func luaGetBalance(L *LState, service C.int, contractId *C.char) (*C.char, *C.char) {
	ctx := contexts[service]
	if contractId == nil {
		return C.CString(ctx.curContract.callState.ctrState.GetBalanceBigInt().String()), nil
	}
	cid, err := getAddressNameResolved(C.GoString(contractId), ctx.bs)
	if err != nil {
		return nil, C.CString("[Contract.LuaGetBalance] invalid contractId: " + err.Error())
	}
	aid := types.ToAccountID(cid)
	cs := ctx.callState[aid]
	if cs == nil {
		bs := ctx.bs
		as, err := bs.GetAccountState(aid)
		if err != nil {
			return nil, C.CString("[Contract.LuaGetBalance] getAccount error: " + err.Error())
		}
		return C.CString(as.GetBalanceBigInt().String()), nil
	}
	return C.CString(cs.accState.Balance().String()), nil
}

//export luaGetSender
func luaGetSender(L *LState, service C.int) *C.char {
	ctx := contexts[service]
	setInstMinusCount(ctx, L, 1000)
	return C.CString(types.EncodeAddress(ctx.curContract.sender))
}

//export luaGetHash
func luaGetHash(L *LState, service C.int) *C.char {
	ctx := contexts[service]
	return C.CString(base58.Encode(ctx.txHash))
}

//export luaGetBlockNo
func luaGetBlockNo(L *LState, service C.int) C.lua_Integer {
	ctx := contexts[service]
	return C.lua_Integer(ctx.blockInfo.No)
}

//export luaGetTimeStamp
func luaGetTimeStamp(L *LState, service C.int) C.lua_Integer {
	ctx := contexts[service]
	return C.lua_Integer(ctx.blockInfo.Ts / 1e9)
}

//export luaGetContractId
func luaGetContractId(L *LState, service C.int) *C.char {
	ctx := contexts[service]
	setInstMinusCount(ctx, L, 1000)
	return C.CString(types.EncodeAddress(ctx.curContract.contractId))
}

//export luaGetAmount
func luaGetAmount(L *LState, service C.int) *C.char {
	ctx := contexts[service]
	return C.CString(ctx.curContract.amount.String())
}

//export luaGetOrigin
func luaGetOrigin(L *LState, service C.int) *C.char {
	ctx := contexts[service]
	setInstMinusCount(ctx, L, 1000)
	return C.CString(types.EncodeAddress(ctx.origin))
}

//export luaGetPrevBlockHash
func luaGetPrevBlockHash(L *LState, service C.int) *C.char {
	ctx := contexts[service]
	return C.CString(base58.Encode(ctx.blockInfo.PrevBlockHash))
}

//export luaGetDbHandle
func luaGetDbHandle(service C.int) *C.sqlite3 {
	ctx := contexts[service]
	curContract := ctx.curContract
	cs := curContract.callState
	if cs.tx != nil {
		return cs.tx.getHandle()
	}
	var tx sqlTx
	var err error

	aid := types.ToAccountID(curContract.contractId)
	if ctx.isQuery == true {
		tx, err = beginReadOnly(aid.String(), curContract.rp)
	} else {
		tx, err = beginTx(aid.String(), curContract.rp)
	}
	if err != nil {
		sqlLgr.Error().Err(err).Msg("Begin SQL Transaction")
		return nil
	}
	if ctx.isQuery == false {
		err = tx.savepoint()
		if err != nil {
			sqlLgr.Error().Err(err).Msg("Begin SQL Transaction")
			return nil
		}
	}
	cs.tx = tx
	return cs.tx.getHandle()
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
	if checkHexString(string(data)) {
		dataStr := data[2:]
		var err error
		data, err = hex.Decode(string(dataStr))
		if err != nil {
			return nil, C.CString("[Contract.LuaCryptoSha256] hex decoding error: " + err.Error())
		}
	}
	h := sha256.New()
	h.Write(data)
	resultHash := h.Sum(nil)

	return C.CString("0x" + hex.Encode(resultHash)), nil
}

func decodeHex(hexStr string) ([]byte, error) {
	if checkHexString(hexStr) {
		hexStr = hexStr[2:]
	}
	return hex.Decode(hexStr)
}

//export luaECVerify
func luaECVerify(L *LState, service C.int, msg *C.char, sig *C.char, addr *C.char) (C.int, *C.char) {
	bMsg, err := decodeHex(C.GoString(msg))
	if err != nil {
		return -1, C.CString("[Contract.LuaEcVerify] invalid message format: " + err.Error())
	}
	bSig, err := decodeHex(C.GoString(sig))
	if err != nil {
		return -1, C.CString("[Contract.LuaEcVerify] invalid signature format: " + err.Error())
	}
	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.LuaEcVerify]not found contract state")
	}
	setInstMinusCount(ctx, L, 10000)

	var pubKey *btcec.PublicKey
	var verifyResult bool
	address := C.GoString(addr)
	isAergo := len(address) == types.EncodedAddressLength

	/*Aergo Address*/
	if isAergo {
		bAddress, err := types.DecodeAddress(address)
		if err != nil {
			return -1, C.CString("[Contract.LuaEcVerify] invalid aergo address: " + err.Error())
		}
		pubKey, err = btcec.ParsePubKey(bAddress)
		if err != nil {
			return -1, C.CString("[Contract.LuaEcVerify] error parsing pubKey: " + err.Error())
		}
	}

	// CompactSign
	if len(bSig) == 65 {
		// ethereum
		if !isAergo {
			btcsig := make([]byte, 65)
			btcsig[0] = bSig[64] + 27
			copy(btcsig[1:], bSig)
			bSig = btcsig
		}
		pub, _, err := ecdsa.RecoverCompact(bSig, bMsg)
		if err != nil {
			return -1, C.CString("[Contract.LuaEcVerify] error recoverCompact: " + err.Error())
		}
		if pubKey != nil {
			verifyResult = pubKey.IsEqual(pub)
		} else {
			bAddress, err := decodeHex(address)
			if err != nil {
				return -1, C.CString("[Contract.LuaEcVerify] invalid Ethereum address: " + err.Error())
			}
			bPub := pub.SerializeUncompressed()
			h := sha256.New()
			h.Write(bPub[1:])
			signAddress := h.Sum(nil)[12:]
			verifyResult = bytes.Equal(bAddress, signAddress)
		}
	} else {
		sign, err := ecdsa.ParseSignature(bSig)
		if err != nil {
			return -1, C.CString("[Contract.LuaEcVerify] error parsing signature: " + err.Error())
		}
		if pubKey == nil {
			return -1, C.CString("[Contract.LuaEcVerify] error recovering pubKey")
		}
		verifyResult = sign.Verify(bMsg, pubKey)
	}
	if verifyResult {
		return C.int(1), nil
	}
	return C.int(0), nil
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

func luaCryptoRlpToBytes(data unsafe.Pointer) rlpObject {
	x := (*C.struct_rlp_obj)(data)
	if x.rlp_obj_type == C.RLP_TSTRING {
		b, _ := luaCryptoToBytes(x.data, C.int(x.size))
		return rlpString(b)
	}
	var l rlpList
	elems := (*[1 << 30]C.struct_rlp_obj)(unsafe.Pointer(x.data))[:C.int(x.size):C.int(x.size)]
	for _, elem := range elems {
		b, _ := luaCryptoToBytes(elem.data, C.int(elem.size))
		l = append(l, rlpString(b))
	}
	return l
}

//export luaCryptoVerifyProof
func luaCryptoVerifyProof(
	key unsafe.Pointer, keyLen C.int,
	value unsafe.Pointer,
	hash unsafe.Pointer, hashLen C.int,
	proof unsafe.Pointer, nProof C.int,
) C.int {
	k, _ := luaCryptoToBytes(key, keyLen)
	v := luaCryptoRlpToBytes(value)
	h, _ := luaCryptoToBytes(hash, hashLen)
	cProof := (*[1 << 30]C.struct_proof)(proof)[:nProof:nProof]
	bProof := make([][]byte, int(nProof))
	for i, p := range cProof {
		bProof[i], _ = luaCryptoToBytes(p.data, C.int(p.len))
	}
	if verifyEthStorageProof(k, v, h, bProof) {
		return C.int(1)
	}
	return C.int(0)
}

//export luaCryptoKeccak256
func luaCryptoKeccak256(data unsafe.Pointer, dataLen C.int) (unsafe.Pointer, int) {
	d, isHex := luaCryptoToBytes(data, dataLen)
	h := keccak256(d)
	if isHex {
		hexb := []byte("0x" + hex.Encode(h))
		return C.CBytes(hexb), len(hexb)
	} else {
		return C.CBytes(h), len(h)
	}
}

// transformAmount processes the input string to calculate the total amount,
// taking into account the different units ("aergo", "gaer", "aer")
func transformAmount(amountStr string, forkVersion int32) (*big.Int, error) {
	if len(amountStr) == 0 {
		return zeroBig, nil
	}

	if forkVersion >= 4 {
		// Check for amount in decimal format
		if strings.Contains(amountStr,".") && strings.HasSuffix(strings.ToLower(amountStr),"aergo") {
			// Extract the part before the unit
			decimalAmount := amountStr[:len(amountStr)-5]
			decimalAmount = strings.TrimRight(decimalAmount, " ")
			// Parse the decimal amount
			decimalAmount = parseDecimalAmount(decimalAmount, 18)
			if decimalAmount == "error" {
				return nil, errors.New("converting error for BigNum: " + amountStr)
			}
			amount, valid := new(big.Int).SetString(decimalAmount, 10)
			if !valid {
				return nil, errors.New("converting error for BigNum: " + amountStr)
			}
			if forkVersion >= 5 {
				// Check for negative amounts
				if amount.Cmp(zeroBig) < 0 {
					return nil, errors.New("negative amount not allowed")
				}
			}
			return amount, nil
		}
	}

	totalAmount := new(big.Int)
	remainingStr := amountStr

	// Define the units and corresponding multipliers
	for _, data := range []struct {
		unit       string
		multiplier *big.Int
	}{
		{"aergo", mulAergo},
		{"gaer", mulGaer},
		{"aer", zeroBig},
	} {
		idx := strings.Index(strings.ToLower(remainingStr), data.unit)
		if idx != -1 {
			// Extract the part before the unit
			subStr := remainingStr[:idx]

			// Parse and convert the amount
			partialAmount, err := parseAndConvert(subStr, data.unit, data.multiplier, amountStr)
			if err != nil {
				return nil, err
			}

			// Add to the total amount
			totalAmount.Add(totalAmount, partialAmount)

			// Adjust the remaining string to process
			remainingStr = remainingStr[idx+len(data.unit):]
		}
	}

	// Process the rest of the string, if there is some
	if len(remainingStr) > 0 {
		partialAmount, err := parseAndConvert(remainingStr, "", zeroBig, amountStr)
		if err != nil {
			return nil, err
		}

		// Add to the total amount
		totalAmount.Add(totalAmount, partialAmount)
	}

	return totalAmount, nil
}

// convert decimal amount into big integer string
func parseDecimalAmount(str string, num_decimals int) string {
	// Get the integer and decimal parts
	idx := strings.Index(str, ".")
	if idx == -1 {
		return str
	}
	p1 := str[0:idx]
	p2 := str[idx+1:]

	// Check for another decimal point
	if strings.Index(p2, ".") != -1 {
		return "error"
	}

	// Compute the amount of zero digits to add
	to_add := num_decimals - len(p2)
	if to_add > 0 {
		p2 = p2 + strings.Repeat("0", to_add)
	} else if to_add < 0 {
		// Do not truncate decimal amounts
		return "error"
	}

	// Join the integer and decimal parts
	str = p1 + p2

	// Remove leading zeros
	str = strings.TrimLeft(str, "0")
	if str == "" {
		str = "0"
	}
	return str
}

// parseAndConvert is a helper function to parse the substring as a big integer
// and apply the necessary multiplier based on the unit.
func parseAndConvert(subStr, unit string, mulUnit *big.Int, fullStr string) (*big.Int, error) {
	subStr = strings.TrimSpace(subStr)

	// Convert the string to a big integer
	amountBig, valid := new(big.Int).SetString(subStr, 10)
	if !valid {
		// Emits a backwards compatible error message
		// the same as: dataType := len(unit) > 0 ? "BigNum" : "Integer"
		dataType := map[bool]string{true: "BigNum", false: "Integer"}[len(unit) > 0]
		return nil, errors.New("converting error for " + dataType + ": " + strings.TrimSpace(fullStr))
	}

	// Check for negative amounts
	if amountBig.Cmp(zeroBig) < 0 {
		return nil, errors.New("negative amount not allowed")
	}

	// Apply multiplier based on unit
	if mulUnit != zeroBig {
		amountBig.Mul(amountBig, mulUnit)
	}

	return amountBig, nil
}

//export luaDeployContract
func luaDeployContract(
	L *LState,
	service C.int,
	contract *C.char,
	args *C.char,
	amount *C.char,
) (ret C.int, errormsg *C.char) {

	contractStr := C.GoString(contract)
	argsStr := C.GoString(args)
	amountStr := C.GoString(amount)

	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.LuaDeployContract]not found contract state")
	}
	if ctx.isQuery == true || ctx.nestedView > 0 {
		return -1, C.CString("[Contract.LuaDeployContract]send not permitted in query")
	}
	bs := ctx.bs

	opId := logOperation(ctx, amountStr, "deploy", contractStr, argsStr)
	defer func() {
		if errormsg != nil {
			logOperationResult(ctx, opId, C.GoString(errormsg))
		}
	}()

	// contract code
	var codeABI []byte
	var sourceCode []byte

	// check if contract name or address is given
	cid, err := getAddressNameResolved(contractStr, bs)
	if err == nil {
		// check if contract exists
		contractState, err := getOnlyContractState(ctx, cid)
		if err != nil {
			return -1, C.CString("[Contract.LuaDeployContract]" + err.Error())
		}
		// check if contract is blacklisted
		if blacklist.Check(contractStr) {
			ctrLgr.Warn().Msg("attempt to deploy clone of blacklisted contract: " + contractStr)
			return -1, C.CString("[Contract.LuaDeployContract] contract not available")
		}
		// read the contract code
		codeABI, err = contractState.GetCode()
		if err != nil {
			return -1, C.CString("[Contract.LuaDeployContract]" + err.Error())
		} else if len(codeABI) == 0 {
			return -1, C.CString("[Contract.LuaDeployContract]: not found code")
		}
		if ctx.blockInfo.ForkVersion >= 4 {
			sourceCode = contractState.GetSourceCode()
		}
	}

	// compile contract code if not found
	if len(codeABI) == 0 {
		if ctx.blockInfo.ForkVersion >= 2 {
			codeABI, err = Compile(contractStr, L)
		} else {
			codeABI, err = Compile(contractStr, nil)
		}
		if err != nil {
			if C.luaL_hasuncatchablerror(L) != C.int(0) &&
				C.ERR_BF_TIMEOUT == err.Error() {
				return -1, C.CString(C.ERR_BF_TIMEOUT)
			} else if err == ErrVmStart {
				return -1, C.CString("[Contract.LuaDeployContract] get luaState error")
			}
			return -1, C.CString("[Contract.LuaDeployContract]compile error:" + err.Error())
		}
		if ctx.blockInfo.ForkVersion >= 4 {
			sourceCode = []byte(contractStr)
		}
	}

	err = ctx.addUpdateSize(int64(len(codeABI) + len(sourceCode)))
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	// create account for the contract
	prevContractInfo := ctx.curContract
	creator := prevContractInfo.callState.accState
	newContract, err := state.CreateAccountState(CreateContractID(prevContractInfo.contractId, creator.Nonce()), bs.StateDB)
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}
	contractState, err := statedb.OpenContractState(newContract.ID(), newContract.State(), bs.StateDB)
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	cs := &callState{isCallback: true, isDeploy: true, ctrState: contractState, accState: newContract}
	ctx.callState[newContract.AccountID()] = cs

	// read the amount transferred to the contract
	amountBig, err := transformAmount(amountStr, ctx.blockInfo.ForkVersion)
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]value not proper format:" + err.Error())
	}

	// read the arguments for the constructor call
	var ci types.CallInfo
	err = getCallInfo(&ci.Args, []byte(argsStr), newContract.ID())
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract] invalid args:" + err.Error())
	}

	// send the amount to the contract
	senderState := prevContractInfo.callState.accState
	receiverState := cs.accState
	if amountBig.Cmp(zeroBig) > 0 {
		if rv := sendBalance(senderState, receiverState, amountBig); rv != nil {
			return -1, rv
		}
	}

	// create a recovery point
	seq, err := createRecoveryPoint(newContract.AccountID(), ctx, senderState, cs, amountBig, false, true)
	if err != nil {
		return -1, C.CString("[System.LuaDeployContract] DB err:" + err.Error())
	}

	// log some info
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[DEPLOY] %s(%s)\n",
			types.EncodeAddress(newContract.ID()), newContract.AccountID().String()))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("deploy snapshot set %d\n", seq))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("SendBalance : %s\n", amountBig.String()))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
			senderState.Balance().String(), receiverState.Balance().String()))
	}

	// set the contract info
	ctx.curContract = newContractInfo(cs, prevContractInfo.contractId, newContract.ID(),
		receiverState.RP(), amountBig)
	defer func() {
		ctx.curContract = prevContractInfo
	}()

	bytecode := util.LuaCode(codeABI).ByteCode()

	// save the contract code
	err = contractState.SetCode(sourceCode, codeABI)
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	// save the contract creator
	err = contractState.SetData(dbkey.CreatorMeta(), []byte(types.EncodeAddress(prevContractInfo.contractId)))
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	// get the remaining gas from the parent LState
	ctx.refreshRemainingGas(L)
	// create a new executor with the remaining gas on the child LState
	ce := newExecutor(bytecode, newContract.ID(), ctx, &ci, amountBig, true, false, contractState)
	defer func() {
		// save the result if the call was successful
		if ce.preErr == nil && ce.err == nil {
			logOperationResult(ctx, opId, ce.jsonRet)
		}
		// close the executor, which will close the child LState
		ce.close()
		// set the remaining gas on the parent LState
		ctx.setRemainingGas(L)
	}()

	if ce.err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]newExecutor Error :" + ce.err.Error())
	}

	if ctx.blockInfo.ForkVersion < 2 {
		// create a sql database for the contract
		if db := luaGetDbHandle(ctx.service); db == nil {
			return -1, C.CString("[System.LuaDeployContract] DB err: cannot open a database")
		}
	}

	// increment the nonce of the creator
	senderState.SetNonce(senderState.Nonce() + 1)

	addr := C.CString(types.EncodeAddress(newContract.ID()))
	ret = C.int(1)

	if ce != nil {
		// run the constructor
		defer setInstCount(ce.ctx, L, ce.L)
		ret += ce.call(minusCallCount(ctx, C.vm_instcount(L), luaCallCountDeduc), L)

		// check if the execution was successful
		if ce.err != nil {
			// rollback the recovery point
			err := clearRecoveryPoint(L, ctx, seq, true)
			if err != nil {
				return -1, C.CString("[Contract.LuaDeployContract] recovery error: " + err.Error())
			}
			// log some info
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
			}
			// return the error message
			return -1, C.CString("[Contract.LuaDeployContract] call err:" + ce.err.Error())
		}
	}

	if seq == 1 {
		err := clearRecoveryPoint(L, ctx, seq, false)
		if err != nil {
			return -1, C.CString("[Contract.LuaDeployContract] recovery error: " + err.Error())
		}
	}

	return ret, addr
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
func luaRandomInt(min, max, service C.int) C.int {
	ctx := contexts[service]
	if ctx.seed == nil {
		setRandomSeed(ctx)
	}
	return C.int(ctx.seed.Intn(int(max+C.int(1)-min)) + int(min))
}

//export luaEvent
func luaEvent(L *LState, service C.int, name *C.char, args *C.char) *C.char {
	eventName := C.GoString(name)
	eventArgs := C.GoString(args)
	ctx := contexts[service]
	if ctx.isQuery == true || ctx.nestedView > 0 {
		return C.CString("[Contract.Event] event not permitted in query")
	}
	if ctx.eventCount >= maxEventCnt(ctx) {
		return C.CString(fmt.Sprintf("[Contract.Event] exceeded the maximum number of events(%d)", maxEventCnt(ctx)))
	}
	if len(eventName) > maxEventNameSize {
		return C.CString(fmt.Sprintf("[Contract.Event] exceeded the maximum length of event name(%d)", maxEventNameSize))
	}
	if len(eventArgs) > maxEventArgSize {
		return C.CString(fmt.Sprintf("[Contract.Event] exceeded the maximum length of event args(%d)", maxEventArgSize))
	}
	ctx.events = append(
		ctx.events,
		&types.Event{
			ContractAddress: ctx.curContract.contractId,
			EventIdx:        ctx.eventCount,
			EventName:       eventName,
			JsonArgs:        eventArgs,
		},
	)
	ctx.eventCount++
	logOperation(ctx, "", "event", eventName, eventArgs)
	return nil
}

//export luaGetEventCount
func luaGetEventCount(L *LState, service C.int) C.int {
	eventCount := contexts[service].eventCount
	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Int32("eventCount", eventCount).Msg("get event count")
	}
	return C.int(eventCount)
}

//export luaDropEvent
func luaDropEvent(L *LState, service C.int, from C.int) {
	// Drop all the events after the given index.
	ctx := contexts[service]
	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Int32("from", int32(from)).Int("len", len(ctx.events)).Msg("drop events")
	}
	if from >= 0 {
		ctx.events = ctx.events[:from]
		ctx.eventCount = int32(len(ctx.events))
	}
}

//export luaToPubkey
func luaToPubkey(L *LState, address *C.char) *C.char {
	// check the length of address
	if len(C.GoString(address)) != types.EncodedAddressLength {
		return C.CString("[Contract.LuaToPubkey] invalid address length")
	}
	// decode the address in string format to bytes (public key)
	pubkey, err := types.DecodeAddress(C.GoString(address))
	if err != nil {
		return C.CString("[Contract.LuaToPubkey] invalid address")
	}
	// return the public key in hex format
	return C.CString("0x" + hex.Encode(pubkey))
}

//export luaToAddress
func luaToAddress(L *LState, pubkey *C.char) *C.char {
	// decode the pubkey in hex format to bytes
	pubkeyBytes, err := decodeHex(C.GoString(pubkey))
	if err != nil {
		return C.CString("[Contract.LuaToAddress] invalid public key")
	}
	// check the length of pubkey
	if len(pubkeyBytes) != types.AddressLength {
		return C.CString("[Contract.LuaToAddress] invalid public key length")
		// or convert the pubkey to compact format - SerializeCompressed()
	}
	// encode the pubkey in bytes to an address in string format
	address := types.EncodeAddress(pubkeyBytes)
	// return the address
	return C.CString(address)
}

//export luaIsContract
func luaIsContract(L *LState, service C.int, contractId *C.char) (C.int, *C.char) {

	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.LuaIsContract] contract state not found")
	}

	cid, err := getAddressNameResolved(C.GoString(contractId), ctx.bs)
	if err != nil {
		return -1, C.CString("[Contract.LuaIsContract] invalid contractId: " + err.Error())
	}

	cs, err := getCallState(ctx, cid)
	if err != nil {
		return -1, C.CString("[Contract.LuaIsContract] getAccount error: " + err.Error())
	}

	return C.int(len(cs.accState.CodeHash())), nil
}

//export luaNameResolve
func luaNameResolve(L *LState, service C.int, name_or_address *C.char) *C.char {
	ctx := contexts[service]
	if ctx == nil {
		return C.CString("[Contract.LuaNameResolve] contract state not found")
	}
	var addr []byte
	var err error
	account := C.GoString(name_or_address)
	if len(account) == types.EncodedAddressLength {
		// also checks if valid address
		addr, err = types.DecodeAddress(account)
	} else {
		addr, err = name.Resolve(ctx.bs, []byte(account), false)
	}
	if err != nil {
		return C.CString("[Contract.LuaNameResolve] " + err.Error())
	}
	return C.CString(types.EncodeAddress(addr))
}

//export luaGovernance
func luaGovernance(L *LState, service C.int, gType C.char, arg *C.char) (errormsg *C.char) {

	ctx := contexts[service]
	if ctx == nil {
		return C.CString("[Contract.LuaGovernance] contract state not found")
	}

	if ctx.isQuery == true || ctx.nestedView > 0 {
		return C.CString("[Contract.LuaGovernance] governance not permitted in query")
	}

	var amountBig *big.Int
	var payload []byte
	var opId int64

	switch gType {
	case 'S', 'U':
		var err error
		amountBig, err = transformAmount(C.GoString(arg), ctx.blockInfo.ForkVersion)
		if err != nil {
			return C.CString("[Contract.LuaGovernance] invalid amount: " + err.Error())
		}
		if gType == 'S' {
			payload = []byte(fmt.Sprintf(`{"Name":"%s"}`, types.Opstake.Cmd()))
			opId = logOperation(ctx, "", "stake", amountBig.String())
		} else {
			payload = []byte(fmt.Sprintf(`{"Name":"%s"}`, types.Opunstake.Cmd()))
			opId = logOperation(ctx, "", "unstake", amountBig.String())
		}
	case 'V':
		amountBig = zeroBig
		payload = []byte(fmt.Sprintf(`{"Name":"%s","Args":%s}`, types.OpvoteBP.Cmd(), C.GoString(arg)))
		opId = logOperation(ctx, "", "vote", C.GoString(arg))
	case 'D':
		amountBig = zeroBig
		payload = []byte(fmt.Sprintf(`{"Name":"%s","Args":%s}`, types.OpvoteDAO.Cmd(), C.GoString(arg)))
		opId = logOperation(ctx, "", "voteDAO", C.GoString(arg))
	}

	defer func() {
		if errormsg != nil {
			logOperationResult(ctx, opId, C.GoString(errormsg))
		}
	}()

	cid := []byte(types.AergoSystem)
	aid := types.ToAccountID(cid)
	scsState, err := getContractState(ctx, cid)
	if err != nil {
		return C.CString("[Contract.LuaGovernance] getAccount error: " + err.Error())
	}

	curContract := ctx.curContract

	senderState := curContract.callState.accState
	receiverState := scsState.accState

	txBody := types.TxBody{
		Amount:  amountBig.Bytes(),
		Payload: payload,
	}
	if ctx.blockInfo.ForkVersion >= 2 {
		txBody.Account = curContract.contractId
	}

	err = types.ValidateSystemTx(&txBody)
	if err != nil {
		return C.CString("[Contract.LuaGovernance] error: " + err.Error())
	}

	seq, err := createRecoveryPoint(aid, ctx, senderState, scsState, zeroBig, false, false)
	if err != nil {
		return C.CString("[Contract.LuaGovernance] database error: " + err.Error())
	}

	events, err := system.ExecuteSystemTx(scsState.ctrState, &txBody, senderState, receiverState, ctx.blockInfo)
	if err != nil {
		rErr := clearRecoveryPoint(L, ctx, seq, true)
		if rErr != nil {
			return C.CString("[Contract.LuaGovernance] recovery error: " + rErr.Error())
		}
		return C.CString("[Contract.LuaGovernance] error: " + err.Error())
	}

	if seq == 1 {
		err := clearRecoveryPoint(L, ctx, seq, false)
		if err != nil {
			return C.CString("[Contract.LuaGovernance] recovery error: " + err.Error())
		}
	}

	ctx.eventCount += int32(len(events))
	ctx.events = append(ctx.events, events...)

	if ctx.lastRecoveryPoint != nil {
		if gType == 'S' {
			seq, _ = createRecoveryPoint(aid, ctx, senderState, scsState, amountBig, true, false)
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[GOVERNANCE]aid(%s)\n", aid.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("staking : %s\n", amountBig.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
					senderState.Balance().String(), receiverState.Balance().String()))
			}
		} else if gType == 'U' {
			seq, _ = createRecoveryPoint(aid, ctx, receiverState, ctx.curContract.callState, amountBig, true, false)
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[GOVERNANCE]aid(%s)\n", aid.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("unstaking : %s\n", amountBig.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
					senderState.Balance().String(), receiverState.Balance().String()))
			}
		}
	}

	return nil
}

//export luaViewStart
func luaViewStart(service C.int) {
	ctx := contexts[service]
	ctx.nestedView++
}

//export luaViewEnd
func luaViewEnd(service C.int) {
	ctx := contexts[service]
	ctx.nestedView--
}

//export luaCheckView
func luaCheckView(service C.int) C.int {
	ctx := contexts[service]
	return C.int(ctx.nestedView)
}

// luaCheckTimeout checks whether the block creation timeout occurred.
//
//export luaCheckTimeout
func luaCheckTimeout(service C.int) C.int {

	if service < BlockFactory {
		// Originally, MaxVmService was used instead of maxContext. service
		// value can be 2 and decremented by MaxVmService(=2) during VM loading.
		// That means the value of service becomes zero after the latter
		// adjustment.
		//
		// This make the VM check block timeout in a unwanted situation. If that
		// happens during the chain service is connecting block, the block chain
		// becomes out of sync.
		service = service + C.int(maxContext)
	}

	if service != BlockFactory {
		return 0
	}

	ctx := contexts[service]
	select {
	case <-ctx.execCtx.Done():
		return 1
	default:
		return 0
	}
}

//export luaIsFeeDelegation
func luaIsFeeDelegation(L *LState, service C.int) (C.int, *C.char) {
	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.LuaIsContract] contract state not found")
	}
	if ctx.isFeeDelegation {
		return 1, nil
	}
	return 0, nil
}

//export LuaGetDbHandleSnap
func LuaGetDbHandleSnap(service C.int, snap *C.char) *C.char {

	stateSet := contexts[service]
	curContract := stateSet.curContract
	callState := curContract.callState

	if stateSet.isQuery != true {
		return C.CString("[Contract.LuaSetDbSnap] not permitted in transaction")
	}

	if callState.tx != nil {
		return C.CString("[Contract.LuaSetDbSnap] transaction already started")
	}

	rp, err := strconv.ParseUint(C.GoString(snap), 10, 64)
	if err != nil {
		return C.CString("[Contract.LuaSetDbSnap] snapshot is not valid" + C.GoString(snap))
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
func luaGetStaking(service C.int, addr *C.char) (*C.char, C.lua_Integer, *C.char) {

	var (
		ctx          *vmContext
		scs, namescs *statedb.ContractState
		err          error
		staking      *types.Staking
	)

	ctx = contexts[service]
	scs, err = statedb.GetSystemAccountState(ctx.bs.StateDB)
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}

	namescs, err = statedb.GetNameAccountState(ctx.bs.StateDB)
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}

	staking, err = system.GetStaking(scs, name.GetAddress(namescs, types.ToAddress(C.GoString(addr))))
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}

	return C.CString(staking.GetAmountBigInt().String()), C.lua_Integer(staking.When), nil
}

func sendBalance(sender *state.AccountState, receiver *state.AccountState, amount *big.Int) *C.char {
	if forkVersion >= 5 {
		if amount.Cmp(zeroBig) < 0 {
			return C.CString("[Contract.sendBalance] negative amount not allowed")
		}
	}
	if err := state.SendBalance(sender, receiver, amount); err != nil {
		return C.CString("[Contract.sendBalance] insufficient balance: " +
			sender.Balance().String() + " : " + amount.String())
	}
	return nil
}
