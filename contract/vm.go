/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

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
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const DbName = "contracts.db"
const constructorName = "constructor"

var (
	ctrLog      *log.Logger
	contractMap stateMap
)

type Contract struct {
	code    []byte
	address []byte
}

type CallState struct {
	ctrState  *state.ContractState
	prevState *types.State
	curState  *types.State
}

type StateSet struct {
	contract  *state.ContractState
	bs        *state.BlockState
	callState map[string]*CallState
	rootState *StateSet
	refCnt    uint
}

type stateMap struct {
	states map[string]*StateSet
	mu     sync.Mutex
}

type LState = C.struct_lua_State
type LBlockchainCtx = C.struct_blockchain_ctx

type Executor struct {
	L             *LState
	contract      *Contract
	args          *types.CallInfo
	err           error
	blockchainCtx *LBlockchainCtx
	jsonRet       string
}

func init() {
	ctrLog = log.NewLogger("contract")
	contractMap.init()
}

func NewContext(blockState *state.BlockState, senderState *types.State,
	contractState *state.ContractState, Sender string,
	txHash string, blockHeight uint64, timestamp int64, node string, confirmed int,
	contractId string, query int, root *StateSet, dbHandle *C.sqlite3, service int) *LBlockchainCtx {

	stateKey := fmt.Sprintf("%d%s%s", service, contractId, txHash)
	stateSet := &StateSet{contract: contractState, bs: blockState, rootState: root}
	if root == nil {
		stateSet.callState = make(map[string]*CallState)
		stateSet.callState[contractId] = &CallState{ctrState: contractState, curState: contractState.State}
		stateSet.callState[Sender] = &CallState{curState: senderState}
		stateSet.rootState = stateSet
	}
	contractMap.register(stateKey, stateSet)

	return &LBlockchainCtx{
		stateKey:    C.CString(stateKey),
		sender:      C.CString(Sender),
		txHash:      C.CString(txHash),
		blockHeight: C.ulonglong(blockHeight),
		timestamp:   C.longlong(timestamp),
		node:        C.CString(node),
		confirmed:   C.int(confirmed),
		contractId:  C.CString(contractId),
		isQuery:     C.int(query),
		db:          dbHandle,
		service:     C.int(service),
	}
}

func NewLState() *LState {
	return C.vm_newstate()
}

func (L *LState) Close() {
	if L != nil {
		C.lua_close(L)
	}
}

func newExecutor(contract *Contract, bcCtx *LBlockchainCtx) *Executor {
	ce := &Executor{
		contract:      contract,
		L:             GetLState(),
		blockchainCtx: bcCtx,
	}
	if ce.L == nil {
		ctrLog.Error().Str("error", "failed: create lua state")
		ce.err = errors.New("failed: create lua state")
		return ce
	}
	if cErrMsg := C.vm_loadbuff(
		ce.L,
		(*C.char)(unsafe.Pointer(&contract.code[0])),
		C.size_t(len(contract.code)),
		bcCtx,
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Error().Str("error", errMsg)
		ce.err = errors.New(errMsg)
	}
	return ce
}

func (ce *Executor) processArgs(ci *types.CallInfo) {
	for _, v := range ci.Args {
		switch arg := v.(type) {
		case string:
			argC := C.CString(arg)
			C.lua_pushstring(ce.L, argC)
			C.free(unsafe.Pointer(argC))
		case int:
			C.lua_pushinteger(ce.L, C.lua_Integer(arg))
		case float64:
			C.lua_pushnumber(ce.L, C.double(arg))
		case bool:
			var b int
			if arg {
				b = 1
			}
			C.lua_pushboolean(ce.L, C.int(b))
		default:
			fmt.Println("unsupported type:" + reflect.TypeOf(v).Name())
			ce.err = errors.New("unsupported type:" + reflect.TypeOf(v).Name())
			return
		}
	}
}

func (ce *Executor) call(ci *types.CallInfo, target *LState) C.int {
	if ce.err != nil {
		return 0
	}
	abiStr := C.CString("abi")
	callStr := C.CString("call")
	abiName := C.CString(ci.Name)

	defer C.free(unsafe.Pointer(abiStr))
	defer C.free(unsafe.Pointer(callStr))
	defer C.free(unsafe.Pointer(abiName))

	C.vm_getfield(ce.L, abiStr)
	C.lua_getfield(ce.L, -1, callStr)
	C.lua_pushstring(ce.L, abiName)

	ce.processArgs(ci)
	nret := C.int(0)
	if cErrMsg := C.vm_pcall(ce.L, C.int(len(ci.Args)+1), &nret); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Warn().Str("error", errMsg).Msgf("contract %s", types.EncodeAddress(ce.contract.address))
		if ce.blockchainCtx.transferFailed == C.int(1) {
			ce.err = types.ErrInsufficientBalance
		} else {
			ce.err = errors.New(errMsg)
		}
		return 0
	}

	if target == nil {
		ce.jsonRet = C.GoString(C.vm_get_json_ret(ce.L, nret))
	} else {
		if cErrMsg := C.vm_copy_result(ce.L, target, nret); cErrMsg != nil {
			errMsg := C.GoString(cErrMsg)
			ce.err = errors.New(errMsg)
		}
	}
	return nret
}

func (ce *Executor) constructCall(ci *types.CallInfo) {
	if ce.err != nil {
		return
	}
	initName := C.CString(constructorName)
	defer C.free(unsafe.Pointer(initName))

	C.vm_getfield(ce.L, initName)
	if C.vm_isnil(ce.L, C.int(-1)) == 1 {
		return
	}

	ce.processArgs(ci)
	if ce.err != nil {
		return
	}
	nret := C.int(0)
	if cErrMsg := C.vm_pcall(ce.L, C.int(len(ci.Args)), &nret); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Warn().Str("error", errMsg).Msgf("contract %s constructor call", types.EncodeAddress(ce.contract.address))
		if ce.blockchainCtx.transferFailed == C.int(1) {
			ce.err = types.ErrInsufficientBalance
		} else {
			ce.err = errors.New(errMsg)
		}
		return
	}

	ce.jsonRet = C.GoString(C.vm_get_json_ret(ce.L, nret))
}

func (ce *Executor) commitCalledContract() error {
	if ce.blockchainCtx == nil {
		return nil
	}
	stateKey := C.GoString(ce.blockchainCtx.stateKey)
	stateSet := contractMap.lookup(stateKey)

	if stateSet == nil || stateSet.callState == nil {
		return nil
	}

	bs := stateSet.bs

	var err error
	for k, v := range stateSet.callState {
		if v.ctrState == stateSet.contract {
			continue
		}
		if v.ctrState != nil {
			err = bs.CommitContractState(v.ctrState)
			if err != nil {
				return err
			}
		}
		/* For Sender */
		if v.prevState == nil {
			continue
		}
		aid, _ := types.DecodeAddress(k)
		err = bs.PutState(types.ToAccountID(aid), v.ctrState.State)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ce *Executor) close(bcCtxFree bool) {
	if ce != nil {
		FreeLState(ce.L)
		if ce.blockchainCtx != nil && bcCtxFree {
			context := ce.blockchainCtx
			contractMap.unregister(C.GoString(context.stateKey))
			C.bc_ctx_delete(context)
		}
	}
}

func Call(contractState *state.ContractState, code, contractAddress []byte,
	bcCtx *LBlockchainCtx) (string, error) {

	var err error
	var ci types.CallInfo
	contract := getContract(contractState, contractAddress, nil)
	if contract != nil {
		err = json.Unmarshal(code, &ci)
		if err != nil {
			ctrLog.Warn().AnErr("error", err).Msgf("contract %s", types.EncodeAddress(contractAddress))
		}
	} else {
		err = fmt.Errorf("cannot find contract %s", string(contractAddress))
		ctrLog.Warn().AnErr("err", err)
	}
	if err != nil {
		return "", err
	}
	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("abi", string(code)).Msgf("contract %s", types.EncodeAddress(contractAddress))
	}

	ce := newExecutor(contract, bcCtx)
	defer ce.close(true)
	ce.call(&ci, nil)
	err = ce.err
	if err == nil {
		err = ce.commitCalledContract()
	} else {
		ctrLog.Warn().AnErr("err", err).Msgf("contract call is failed")
	}
	return ce.jsonRet, err
}

func PreCall(ce *Executor, bs *state.BlockState, senderState *types.State, contractState *state.ContractState,
	blockNo uint64, ts int64, dbHandle *C.sqlite3) (string, error) {
	var err error

	defer ce.close(true)

	bcCtx := ce.blockchainCtx

	contractId := C.GoString(bcCtx.contractId)
	sender := C.GoString(bcCtx.sender)
	stateKey := C.GoString(bcCtx.stateKey)

	stateSet := contractMap.lookup(stateKey)
	stateSet.contract = contractState
	stateSet.bs = bs
	stateSet.callState[contractId].ctrState = contractState
	stateSet.callState[contractId].curState = contractState.State
	stateSet.callState[sender].curState = senderState

	bcCtx.blockHeight = C.ulonglong(blockNo)
	bcCtx.timestamp = C.longlong(ts)
	bcCtx.db = dbHandle

	ce.call(ce.args, nil)
	err = ce.err
	if err == nil {
		err = ce.commitCalledContract()
	} else {
		ctrLog.Warn().AnErr("err", err).Msgf("contract call is failed")
	}
	return ce.jsonRet, err
}

func PreloadEx(contractState *state.ContractState, code, contractAddress []byte,
	bcCtx *LBlockchainCtx) (*Executor, error) {

	var err error
	var ci types.CallInfo
	contract := getContract(contractState, contractAddress, nil)
	if contract != nil {
		err = json.Unmarshal(code, &ci)
		if err != nil {
			ctrLog.Warn().AnErr("error", err).Msgf("contract %s", types.EncodeAddress(contractAddress))
		}
	} else {
		err = fmt.Errorf("cannot find contract %s", string(contractAddress))
		ctrLog.Warn().AnErr("err", err)
	}
	if err != nil {
		return nil, err
	}
	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("abi", string(code)).Msgf("contract %s", types.EncodeAddress(contractAddress))
	}
	ce := newExecutor(contract, bcCtx)
	ce.args = &ci

	return ce, nil

}

func Create(contractState *state.ContractState, code, contractAddress []byte,
	bcCtx *LBlockchainCtx) (string, error) {

	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("contractAddress", types.EncodeAddress(contractAddress)).Msg("new contract is deployed")
	}
	if len(code) <= 4 {
		err := fmt.Errorf("code length is short %d", len(code))
		ctrLog.Warn().AnErr("err", err)
		return "", err
	}
	codeLen := codeLength(code[0:])
	if uint32(len(code)) < codeLen {
		err := fmt.Errorf("code length does not match (%d > %d)", codeLen, len(code))
		ctrLog.Warn().AnErr("err", err)
		return "", err
	}
	sCode := code[4:codeLen]

	err := contractState.SetCode(sCode)
	if err != nil {
		return "", err
	}
	contract := getContract(contractState, contractAddress, sCode)
	if contract == nil {
		err = fmt.Errorf("cannot deploy contract %s", types.EncodeAddress(contractAddress))
		ctrLog.Warn().AnErr("err", err)
		return "", err
	}
	contractState.SetData([]byte("Creator"), []byte(C.GoString(bcCtx.sender)))

	var ce *Executor
	ce = newExecutor(contract, bcCtx)
	defer ce.close(true)

	var ci types.CallInfo
	if len(code) != int(codeLen) {
		err = json.Unmarshal(code[codeLen:], &ci.Args)
	}
	var errMsg string
	if err != nil {
		logger.Warn().Err(err).Msg("constructor's argument is invalid")
		errMsg = err.Error()
	}
	ce.constructCall(&ci)
	if ce.err != nil {
		logger.Warn().Err(err).Msg("constructor is failed")
		errMsg += "\", \"constructor call error:" + ce.err.Error()
	}

	return `{""` + errMsg + `"}`, nil
}

func Query(contractAddress []byte, bs *state.BlockState, contractState *state.ContractState, queryInfo []byte) ([]byte, error) {
	var ci types.CallInfo
	var err error
	contract := getContract(contractState, contractAddress, nil)
	if contract != nil {
		err = json.Unmarshal(queryInfo, &ci)
		if err != nil {
			ctrLog.Warn().AnErr("error", err).Msgf("contract %s", types.EncodeAddress(contractAddress))
		}
	} else {
		err = fmt.Errorf("cannot find contract %s", types.EncodeAddress(contractAddress))
		ctrLog.Warn().AnErr("err", err)
	}
	if err != nil {
		return nil, err
	}
	var ce *Executor

	tx, err := BeginReadOnly(types.ToAccountID(contractAddress), contractState.SqlRecoveryPoint)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bcCtx := NewContext(bs, nil, contractState, "", "",
		0, 0, "", 0, types.EncodeAddress(contractAddress),
		1, nil, tx.GetHandle(), ChainService)

	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("abi", string(queryInfo)).Msgf("contract %s", types.EncodeAddress(contractAddress))
	}
	ce = newExecutor(contract, bcCtx)
	defer ce.close(true)
	ce.call(&ci, nil)
	err = ce.err

	return []byte(ce.jsonRet), err
}

func getContract(contractState *state.ContractState, contractAddress []byte, code []byte) *Contract {
	var val []byte
	val = code
	if val == nil {
		var err error
		val, err = contractState.GetCode()

		if err != nil {
			return nil
		}
	}
	valLen := len(val)
	if valLen <= 4 {
		return nil
	}
	l := codeLength(val[0:])
	if 4+l > uint32(valLen) {
		return nil
	}
	return &Contract{
		code:    val[4 : 4+l],
		address: contractAddress[:],
	}
}

func GetABI(contractState *state.ContractState) (*types.ABI, error) {
	val, err := contractState.GetCode()
	if err != nil {
		return nil, err
	}
	valLen := len(val)
	if valLen == 0 {
		return nil, errors.New("cannot find contract")
	}
	if valLen <= 4 {
		return nil, errors.New("cannot find abi")
	}
	l := codeLength(val)
	if 4+l >= uint32(len(val)) {
		return nil, errors.New("cannot find abi")
	}
	abi := new(types.ABI)
	if err := json.Unmarshal(val[4+l:], abi); err != nil {
		return nil, err
	}
	return abi, nil
}

func codeLength(val []byte) uint32 {
	return binary.LittleEndian.Uint32(val[0:])
}

func (sm *stateMap) init() {
	sm.states = make(map[string]*StateSet)
}

func (sm *stateMap) register(key string, item *StateSet) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.states[key] != nil {
		item.refCnt++
		return
	}
	item.refCnt++
	sm.states[key] = item
}

func (sm *stateMap) unregister(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	item := sm.states[key]

	if item == nil {
		err := fmt.Errorf("cannot find contract state: %s", key)
		ctrLog.Warn().AnErr("err", err)
		return
	}
	item.refCnt--
	if item.refCnt == 0 {
		delete(sm.states, key)
	}
}

func (sm *stateMap) lookup(key string) *StateSet {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.states[key]
}

func luaPushStr(L *LState, str string) {
	cStr := C.CString(str)
	C.lua_pushstring(L, cStr)
	C.free(unsafe.Pointer(cStr))
}

//export LuaSetDB
func LuaSetDB(L *LState, stateKey *C.char, key *C.char, value *C.char) C.int {
	stateKeyString := C.GoString(stateKey)
	keyString := C.GoString(key)
	valueString := C.GoString(value)

	stateSet := contractMap.lookup(stateKeyString)
	if stateSet == nil {
		luaPushStr(L, "[System.LuaSetDB]not found contract state")
		return -1
	}

	err := stateSet.contract.SetData([]byte(keyString), []byte(valueString))
	if err != nil {
		luaPushStr(L, err.Error())
		return -1
	}
	return 0
}

//export LuaGetDB
func LuaGetDB(L *LState, stateKey *C.char, key *C.char) C.int {
	stateKeyString := C.GoString(stateKey)
	keyString := C.GoString(key)

	stateSet := contractMap.lookup(stateKeyString)
	if stateSet == nil {
		luaPushStr(L, "[System.LuaGetDB]not found contract state")
		return -1
	}

	data, err := stateSet.contract.GetData([]byte(keyString))
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

//export LuaCallContract
func LuaCallContract(L *LState, bcCtx *LBlockchainCtx, contractId *C.char, fname *C.char, args *C.char,
	amount uint64, gas uint64) C.int {
	stateKeyStr := C.GoString(bcCtx.stateKey)
	contractIdStr := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	cid, err := types.DecodeAddress(contractIdStr)
	if err != nil {
		luaPushStr(L, "[System.LuaGetContract]invalid contractId :"+err.Error())
		return -1
	}

	stateSet := contractMap.lookup(stateKeyStr)
	if stateSet == nil {
		luaPushStr(L, "[System.LuaCallContract]not found contract state")
		return -1
	}

	rootState := stateSet.rootState
	callState := rootState.callState[contractIdStr]
	if callState == nil {
		bs := rootState.bs

		prevState, err := bs.GetAccountState(types.ToAccountID(cid))
		if err != nil {
			luaPushStr(L, "[System.LuaGetContract]getAccount Error :"+err.Error())
			return -1
		}

		curState := types.Clone(*prevState).(types.State)
		contractState, err := bs.OpenContractState(&curState)
		if err != nil {
			luaPushStr(L, "[System.LuaGetContract]getAccount Error"+err.Error())
			return -1
		}
		callState =
			&CallState{ctrState: contractState, prevState: prevState, curState: &curState}
		rootState.callState[contractIdStr] = callState
	}
	if callState.ctrState == nil {
		callState.ctrState, err = rootState.bs.OpenContractState(callState.curState)
		if err != nil {
			luaPushStr(L, "[System.LuaGetContract]getAccount Error"+err.Error())
			return -1
		}
	}
	if sendBalance(L, stateSet.contract.State, callState.curState, amount) == false {
		bcCtx.transferFailed = 1
		return -1
	}
	callee := getContract(callState.ctrState, cid, nil)
	if callee == nil {
		luaPushStr(L, "[System.LuaGetContract]cannot find contract "+string(contractIdStr))
		return -1
	}
	sqlTx, err := BeginTx(types.ToAccountID(callee.address), callState.curState.SqlRecoveryPoint)
	if err != nil {
		luaPushStr(L, "[System.LuaGetContract] begin tx:"+err.Error())
		return -1
	}
	sqlTx.Savepoint()
	newBcCtx := NewContext(nil, nil, callState.ctrState,
		C.GoString(bcCtx.contractId), C.GoString(bcCtx.txHash), uint64(bcCtx.blockHeight), int64(bcCtx.timestamp),
		"", int(bcCtx.confirmed), contractIdStr, int(bcCtx.isQuery), rootState, sqlTx.GetHandle(),
		int(bcCtx.service))
	ce := newExecutor(callee, newBcCtx)
	defer ce.close(true)

	if ce.err != nil {
		luaPushStr(L, "[System.LuaGetContract]newExecutor Error :"+ce.err.Error())
		return -1
	}

	var ci types.CallInfo
	ci.Name = fnameStr
	err = json.Unmarshal([]byte(argsStr), &ci.Args)
	if err != nil {
		luaPushStr(L, "[System.LuaCallContract] invalid args:"+err.Error())
		return -1
	}
	ret := ce.call(&ci, L)
	if ce.err != nil {
		sqlTx.RollbackToSavepoint()
		luaPushStr(L, "[System.LuaCallContract] call err:"+ce.err.Error())
		return -1
	}
	sqlTx.Release()

	return ret
}

//export LuaDelegateCallContract
func LuaDelegateCallContract(L *LState, bcCtx *LBlockchainCtx, contractId *C.char,
	fname *C.char, args *C.char, gas uint64) C.int {
	stateKeyStr := C.GoString(bcCtx.stateKey)
	contractIdStr := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	cid, err := types.DecodeAddress(contractIdStr)
	if err != nil {
		luaPushStr(L, "[System.LuaGetContract]invalid contractId :"+err.Error())
		return -1
	}

	stateSet := contractMap.lookup(stateKeyStr)
	if stateSet == nil {
		luaPushStr(L, "[System.LuaCallContract]not found contract state")
		return -1
	}
	bs := stateSet.rootState.bs
	contractState, err := bs.OpenContractStateAccount(types.ToAccountID(cid))
	contract := getContract(contractState, cid, nil)
	if contract == nil {
		luaPushStr(L, "[System.LuaGetContract]cannot find contract "+string(contractIdStr))
		return -1
	}
	ce := newExecutor(contract, bcCtx)
	defer ce.close(false)

	if ce.err != nil {
		luaPushStr(L, "[System.LuaGetContract]newExecutor Error :"+ce.err.Error())
		return -1
	}

	var ci types.CallInfo
	ci.Name = fnameStr
	err = json.Unmarshal([]byte(argsStr), &ci.Args)
	if err != nil {
		luaPushStr(L, "[System.LuaCallContract] invalid args:"+err.Error())
		return -1
	}
	ret := ce.call(&ci, L)
	if ce.err != nil {
		luaPushStr(L, "[System.LuaCallContract] call err:"+ce.err.Error())
		return -1
	}
	return ret
}

//export LuaSendAmount
func LuaSendAmount(L *LState, bcCtx *LBlockchainCtx, contractId *C.char, amount uint64) C.int {
	stateKeyStr := C.GoString(bcCtx.stateKey)
	contractIdStr := C.GoString(contractId)

	cid, err := types.DecodeAddress(contractIdStr)
	if err != nil {
		luaPushStr(L, "[System.LuaGetContract]invalid contractId :"+err.Error())
		return -1
	}

	stateSet := contractMap.lookup(stateKeyStr)
	if stateSet == nil {
		luaPushStr(L, "[System.LuaCallContract]not found contract state")
		return -1
	}

	rootState := stateSet.rootState
	callState := rootState.callState[contractIdStr]
	if callState == nil {
		bs := rootState.bs

		prevState, err := bs.GetAccountState(types.ToAccountID(cid))
		if err != nil {
			luaPushStr(L, "[System.LuaGetContract]getAccount Error :"+err.Error())
			return -1
		}

		curState := types.Clone(*prevState).(types.State)
		callState =
			&CallState{prevState: prevState, curState: &curState}
		rootState.callState[contractIdStr] = callState
	}
	if sendBalance(L, stateSet.contract.State, callState.curState, amount) == false {
		bcCtx.transferFailed = 1
		return -1
	}
	return 0
}

func sendBalance(L *LState, sender *types.State, receiver *types.State, amount uint64) bool {
	if sender == receiver {
		return true
	}
	if sender.Balance < amount {
		luaPushStr(L, "[Contract.call]insuficient balance"+
			string(sender.Balance)+" : "+string(amount))
		return false
	} else {
		sender.Balance = sender.Balance - amount
	}
	receiver.Balance = receiver.Balance + amount

	return true
}

//export LuaPrint
func LuaPrint(contractId *C.char, args *C.char) {
	logger.Info().Str("Contract SystemPrint", C.GoString(contractId)).Msg(C.GoString(args))
}
