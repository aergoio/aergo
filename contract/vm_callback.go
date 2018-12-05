package contract

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.0
#cgo LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <string.h>
#include "vm.h"
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"math/big"
	"unsafe"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func luaPushStr(L *LState, str string) {
	cStr := C.CString(str)
	C.lua_pushstring(L, cStr)
	C.free(unsafe.Pointer(cStr))
}

//export LuaSetDB
func LuaSetDB(L *LState, service *C.int, key *C.char, value *C.char) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[System.LuaSetDB]not found contract state")
		return -1
	}
	if stateSet.isQuery == true {
		luaPushStr(L, "[System.LuaSetDB]set not permitted in query")
		return -1
	}
	err := stateSet.curContract.callState.ctrState.SetData([]byte(C.GoString(key)), []byte(C.GoString(value)))
	if err != nil {
		luaPushStr(L, err.Error())
		return -1
	}
	return 0
}

//export LuaGetDB
func LuaGetDB(L *LState, service *C.int, key *C.char) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[System.LuaGetDB]not found contract state")
		return -1
	}

	data, err := stateSet.curContract.callState.ctrState.GetData([]byte(C.GoString(key)))
	if err != nil {
		luaPushStr(L, err.Error())
		return -1
	}

	if data == nil {
		return 0
	}
	luaPushStr(L, string(data))
	return 1
}

//export LuaDelDB
func LuaDelDB(L *LState, service *C.int, key *C.char) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[System.LuaGetDB]not found contract state")
		return -1
	}
	err := stateSet.curContract.callState.ctrState.DeleteData([]byte(C.GoString(key)))
	if err != nil {
		luaPushStr(L, err.Error())
		return -1
	}
	return 0
}

//export LuaCallContract
func LuaCallContract(L *LState, service *C.int, contractId *C.char, fname *C.char, args *C.char,
	amount uint64, gas uint64) C.int {
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)
	amountBig := new(big.Int).SetUint64(amount)
	cid, err := types.DecodeAddress(C.GoString(contractId))
	if err != nil {
		luaPushStr(L, "[System.LuaCallContract]invalid contractId :"+err.Error())
		return -1
	}
	aid := types.ToAccountID(cid)

	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[System.LuaCallContract]not found contract state")
		return -1
	}
	if stateSet.isQuery == true {
		luaPushStr(L, "[System.LuaCallContract]send not permitted in query")
	}
	callState := stateSet.callState[aid]
	if callState == nil {
		bs := stateSet.bs

		prevState, err := bs.GetAccountState(aid)
		if err != nil {
			luaPushStr(L, "[System.LuaCallContract]getAccount Error :"+err.Error())
			return -1
		}

		curState := types.Clone(*prevState).(types.State)
		contractState, err := bs.OpenContractState(aid, &curState)
		if err != nil {
			luaPushStr(L, "[System.LuaCallContract]getAccount Error"+err.Error())
			return -1
		}
		callState =
			&CallState{ctrState: contractState, prevState: prevState, curState: &curState}
		stateSet.callState[aid] = callState
	}
	if callState.ctrState == nil {
		callState.ctrState, err = stateSet.bs.OpenContractState(aid, callState.curState)
		if err != nil {
			luaPushStr(L, "[System.LuaCallContract]getAccount Error"+err.Error())
			return -1
		}
	}

	callee := getContract(callState.ctrState, nil)
	if callee == nil {
		luaPushStr(L, "[System.LuaCallContract]cannot find contract "+C.GoString(contractId))
		return -1
	}

	prevContractInfo := stateSet.curContract

	ce := newExecutor(callee, stateSet)
	defer ce.close()

	if ce.err != nil {
		luaPushStr(L, "[System.LuaCallContract]newExecutor Error :"+ce.err.Error())
		return -1
	}

	var ci types.CallInfo
	ci.Name = fnameStr
	err = json.Unmarshal([]byte(argsStr), &ci.Args)
	if err != nil {
		luaPushStr(L, "[System.LuaCallContract] invalid args:"+err.Error())
		return -1
	}
	senderState := prevContractInfo.callState.curState
	if amount > 0 {
		if sendBalance(L, senderState, callState.curState, amountBig) == false {
			stateSet.transferFailed = true
			return -1
		}
	}
	if stateSet.lastRecoveryEntry != nil {
		setRecoveryPoint(aid, stateSet, senderState, callState, amountBig, callState.ctrState.Snapshot())
	}
	stateSet.curContract = newContractInfo(callState, prevContractInfo.contractId, cid,
		callState.curState.SqlRecoveryPoint, amountBig)
	ret := ce.call(&ci, L)
	if ce.err != nil {
		stateSet.curContract = prevContractInfo
		luaPushStr(L, "[System.LuaCallContract] call err:"+ce.err.Error())
		return -1
	}
	stateSet.curContract = prevContractInfo
	return ret
}

//export LuaDelegateCallContract
func LuaDelegateCallContract(L *LState, service *C.int, contractId *C.char,
	fname *C.char, args *C.char, gas uint64) C.int {
	contractIdStr := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	cid, err := types.DecodeAddress(contractIdStr)
	if err != nil {
		luaPushStr(L, "[System.LuaDelegateCallContract]invalid contractId :"+err.Error())
		return -1
	}

	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[System.LuaDelegateCallContract]not found contract state")
		return -1
	}
	bs := stateSet.bs
	aid := types.ToAccountID(cid)
	contractState, err := bs.OpenContractStateAccount(aid)
	contract := getContract(contractState, nil)
	if contract == nil {
		luaPushStr(L, "[System.LuaDelegateCallContract]cannot find contract "+contractIdStr)
		return -1
	}
	ce := newExecutor(contract, stateSet)
	defer ce.close()

	if ce.err != nil {
		luaPushStr(L, "[System.LuaDelegateCallContract]newExecutor Error :"+ce.err.Error())
		return -1
	}

	var ci types.CallInfo
	ci.Name = fnameStr
	err = json.Unmarshal([]byte(argsStr), &ci.Args)
	if err != nil {
		luaPushStr(L, "[System.LuaDelegateCallContract] invalid args:"+err.Error())
		return -1
	}

	if stateSet.lastRecoveryEntry != nil {
		callState := stateSet.curContract.callState
		setRecoveryPoint(aid, stateSet, nil, callState, big.NewInt(0), callState.ctrState.Snapshot())
	}
	ret := ce.call(&ci, L)
	if ce.err != nil {
		luaPushStr(L, "[System.LuaDelegateCallContract] call err:"+ce.err.Error())
		return -1
	}
	return ret
}

//export LuaSendAmount
func LuaSendAmount(L *LState, service *C.int, contractId *C.char, amount uint64) C.int {
	amountBig := new(big.Int).SetUint64(amount)
	cid, err := types.DecodeAddress(C.GoString(contractId))
	if err != nil {
		luaPushStr(L, "[Contract.LuaSendAmount]invalid contractId :"+err.Error())
		return -1
	}

	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[Contract.LuaSendAmount]not found contract state")
		return -1
	}
	if stateSet.isQuery == true {
		luaPushStr(L, "[Contract.LuaSendAmount]send not permitted in query")
	}

	aid := types.ToAccountID(cid)
	callState := stateSet.callState[aid]
	if callState == nil {
		bs := stateSet.bs

		prevState, err := bs.GetAccountState(aid)
		if err != nil {
			luaPushStr(L, "[System.LuaGetContract]getAccount Error :"+err.Error())
			return -1
		}

		curState := types.Clone(*prevState).(types.State)
		callState =
			&CallState{prevState: prevState, curState: &curState}
		stateSet.callState[aid] = callState
	}
	senderState := stateSet.curContract.callState.curState
	if sendBalance(L, senderState, callState.curState, amountBig) == false {
		stateSet.transferFailed = true
		return -1
	}
	if stateSet.lastRecoveryEntry != nil {
		setRecoveryPoint(aid, stateSet, senderState, callState, amountBig, 0)
	}
	return 0
}

func sendBalance(L *LState, sender *types.State, receiver *types.State, amount *big.Int) bool {
	if sender == receiver {
		return true
	}
	if sender.GetBalanceBigInt().Cmp(amount) < 0 {
		luaPushStr(L, "[Contract.call]insuficient balance"+
			sender.GetBalanceBigInt().String()+" : "+amount.String())
		return false
	} else {
		sender.Balance = new(big.Int).Sub(sender.GetBalanceBigInt(), amount).Bytes()
	}
	receiver.Balance = new(big.Int).Add(receiver.GetBalanceBigInt(), amount).Bytes()

	return true
}

//export LuaPrint
func LuaPrint(service *C.int, args *C.char) {
	stateSet := curStateSet[*service]
	logger.Info().Str("Contract SystemPrint", types.EncodeAddress(stateSet.curContract.contractId)).Msg(C.GoString(args))
}

func setRecoveryPoint(aid types.AccountID, stateSet *StateSet, senderState *types.State,
	callState *CallState, amount *big.Int, snapshot state.Snapshot) {
	var seq int
	prev := stateSet.lastRecoveryEntry
	if prev != nil {
		seq = prev.seq + 1
	} else {
		seq = 1
	}
	recoveryEntry := &recoveryEntry{
		seq,
		amount,
		senderState,
		callState,
		nil,
		snapshot,
		prev,
	}
	tx := callState.tx
	if tx != nil {
		saveName := fmt.Sprintf("%s_%p", aid.String(), &recoveryEntry)
		tx.SubSavepoint(saveName)
		recoveryEntry.sqlSaveName = &saveName
	}
	stateSet.lastRecoveryEntry = recoveryEntry
}

//export LuaSetRecoveryPoint
func LuaSetRecoveryPoint(L *LState, service *C.int) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[Contract.pcall]not found contract state")
		return -1
	}
	if stateSet.isQuery == true {
		return 0
	}
	curContract := stateSet.curContract
	setRecoveryPoint(types.ToAccountID(curContract.contractId), stateSet, nil,
		curContract.callState, big.NewInt(0), curContract.callState.ctrState.Snapshot())
	return C.int(stateSet.lastRecoveryEntry.seq)
}

//export LuaClearRecovery
func LuaClearRecovery(L *LState, service *C.int, start int, error bool) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[Contract.pcall]not found contract state")
		return -1
	}
	item := stateSet.lastRecoveryEntry
	for {
		if error {
			item.recovery()
		}
		if item.seq == start {
			if error || item.prev == nil {
				stateSet.lastRecoveryEntry = item.prev
			}
			return 0
		}
		item = item.prev
	}
	return 0
}

//export LuaGetBalance
func LuaGetBalance(L *LState, service *C.int, contractId *C.char) C.int {
	stateSet := curStateSet[*service]
	if contractId == nil {
		luaPushStr(L, stateSet.curContract.callState.ctrState.GetBalanceBigInt().String())
		return 0
	}
	cid, err := types.DecodeAddress(C.GoString(contractId))
	if err != nil {
		luaPushStr(L, "[Contract.LuaGetBalance]invalid contractId :"+err.Error())
		return -1
	}

	aid := types.ToAccountID(cid)
	callState := stateSet.callState[aid]
	if callState == nil {
		bs := stateSet.bs

		as, err := bs.GetAccountState(aid)
		if err != nil {
			luaPushStr(L, "[Contract.LuaGetBalance]getAccount Error :"+err.Error())
			return -1
		}
		luaPushStr(L, as.GetBalanceBigInt().String())
	} else {
		luaPushStr(L, callState.curState.GetBalanceBigInt().String())
	}

	return 0
}

//export LuaGetSender
func LuaGetSender(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, types.EncodeAddress(stateSet.curContract.sender))
}

//export LuaGetHash
func LuaGetHash(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, enc.ToString(stateSet.txHash))
}

//export LuaGetBlockNo
func LuaGetBlockNo(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	C.lua_pushinteger(L, C.lua_Integer(stateSet.blockHeight))
}

//export LuaGetTimeStamp
func LuaGetTimeStamp(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	C.lua_pushinteger(L, C.lua_Integer(stateSet.timestamp/1e9))
}

//export LuaGetContractId
func LuaGetContractId(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, types.EncodeAddress(stateSet.curContract.contractId))
}

//export LuaGetAmount
func LuaGetAmount(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, stateSet.curContract.amount.String())
}

//export LuaGetOrigin
func LuaGetOrigin(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, types.EncodeAddress(stateSet.origin))
}

//export LuaGetDbHandle
func LuaGetDbHandle(service *C.int) *C.sqlite3 {
	stateSet := curStateSet[*service]
	curContract := stateSet.curContract
	callState := curContract.callState
	if callState.tx != nil {
		return callState.tx.GetHandle()
	}
	var tx Tx
	var err error

	aid := types.ToAccountID(curContract.contractId)
	if stateSet.isQuery == true {
		tx, err = BeginReadOnly(aid.String(), curContract.rp)
	} else {
		tx, err = BeginTx(aid.String(), curContract.rp)
	}
	if err != nil {
		stateSet.dbSystemError = true
		logger.Error().Err(err).Msg("Begin SQL Transaction")
		return nil
	}
	if stateSet.isQuery == false {
		err = tx.Savepoint()
		if err != nil {
			stateSet.dbSystemError = true
			logger.Error().Err(err).Msg("Begin SQL Transaction")
			return nil
		}
	}
	callState.tx = tx
	return callState.tx.GetHandle()
}
