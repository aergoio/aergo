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
	tx        Tx
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

func registerMap(bcCtx *LBlockchainCtx, blockState *state.BlockState, senderState *types.State,
	contractState *state.ContractState, root *StateSet) {
	contractId := C.GoString(bcCtx.contractId)
	sender := C.GoString(bcCtx.sender)
	stateKey := C.GoString(bcCtx.stateKey)
	stateSet := &StateSet{contract: contractState, bs: blockState, rootState: root}
	if root == nil {
		stateSet.callState = make(map[string]*CallState)
		stateSet.callState[contractId] = &CallState{ctrState: contractState, curState: contractState.State}
		stateSet.callState[sender] = &CallState{curState: senderState}
		stateSet.rootState = stateSet
	}
	contractMap.register(stateKey, stateSet)
}

func NewContext(blockState *state.BlockState, senderState *types.State,
	contractState *state.ContractState, Sender string,
	txHash string, blockHeight uint64, timestamp int64, node string, confirmed int,
	contractId string, query int, root *StateSet, rp uint64, service int, amount uint64) *LBlockchainCtx {

	stateKey := fmt.Sprintf("%d%s%s", service, contractId, txHash)

	bcCtx := &LBlockchainCtx{
		stateKey:    C.CString(stateKey),
		sender:      C.CString(Sender),
		txHash:      C.CString(txHash),
		blockHeight: C.ulonglong(blockHeight),
		timestamp:   C.longlong(timestamp),
		node:        C.CString(node),
		confirmed:   C.int(confirmed),
		contractId:  C.CString(contractId),
		isQuery:     C.int(query),
		rp:          C.ulonglong(rp),
		service:     C.int(service),
		amount:      C.ulonglong(amount),
	}
	registerMap(bcCtx, blockState, senderState, contractState, root)

	return bcCtx
}

func (bcCtx *LBlockchainCtx) Close() {
	if bcCtx == nil {
		return
	}
	if bcCtx.stateKey != nil {
		contractMap.unregister(C.GoString(bcCtx.stateKey))
	}
	C.bc_ctx_delete(bcCtx)
}

func (bcCtx *LBlockchainCtx) Del() {
	C.bc_ctx_delete(bcCtx)
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
		case nil:
			C.lua_pushnil(ce.L)
		default:
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
		} else if ce.blockchainCtx.dbSystemError == C.int(1) {
			ce.err = newDbSystemError(errMsg)
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
		} else if ce.blockchainCtx.dbSystemError == C.int(1) {
			ce.err = newDbSystemError(errMsg)
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
		if v.tx != nil {
			err = v.tx.Release()
			if err != nil {
				return DbSystemError(err)
			}
		}
		if v.ctrState == stateSet.contract {
			continue
		}
		if v.ctrState != nil {
			err = bs.StageContractState(v.ctrState)
			if err != nil {
				return DbSystemError(err)
			}
		}
		/* For Sender */
		if v.prevState == nil {
			continue
		}
		aid, _ := types.DecodeAddress(k)
		err = bs.PutState(types.ToAccountID(aid), v.ctrState.State)
		if err != nil {
			return DbSystemError(err)
		}
	}
	return nil
}

func (ce *Executor) rollbackToSavepoint() error {
	if ce.blockchainCtx == nil {
		return nil
	}
	stateKey := C.GoString(ce.blockchainCtx.stateKey)
	stateSet := contractMap.lookup(stateKey)

	if stateSet == nil || stateSet.callState == nil {
		return nil
	}

	var err error
	for _, v := range stateSet.callState {
		if v.tx == nil {
			continue
		}
		err = v.tx.RollbackToSavepoint()
		if err != nil {
			return DbSystemError(err)
		}
	}
	return nil
}

func (ce *Executor) close(bcCtxFree bool) {
	if ce != nil {
		FreeLState(ce.L)
		if bcCtxFree {
			ce.blockchainCtx.Close()
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
		err = fmt.Errorf("cannot find contract %s", types.EncodeAddress(contractAddress))
		ctrLog.Warn().AnErr("err", err)
	}
	if err != nil {
		bcCtx.Close()
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
		if err != nil {
			logger.Error().Err(err).Msg("contract call is failed")
		}
	} else {
		ctrLog.Warn().Err(err).Msg("contract call is failed")
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			ctrLog.Error().Err(dbErr).Msg("contract call is failed")
			err = dbErr
		}
	}
	return ce.jsonRet, err
}

func PreCall(ce *Executor, bs *state.BlockState, senderState *types.State, contractState *state.ContractState,
	blockNo uint64, ts int64, rp uint64) (string, error) {
	var err error

	defer ce.close(true)

	bcCtx := ce.blockchainCtx

	stateKey := fmt.Sprintf("%d%s%s", C.int(bcCtx.service),
		C.GoString(bcCtx.contractId), C.GoString(bcCtx.txHash))
	bcCtx.stateKey = C.CString(stateKey)
	registerMap(bcCtx, bs, senderState, contractState, nil)
	bcCtx.blockHeight = C.ulonglong(blockNo)
	bcCtx.timestamp = C.longlong(ts)
	bcCtx.rp = C.ulonglong(rp)

	ce.call(ce.args, nil)
	err = ce.err
	if err == nil {
		err = ce.commitCalledContract()
		if err != nil {
			ctrLog.Error().Err(err).Msg("contract call is failed")
		}
	} else {
		ctrLog.Warn().Err(err).Msg("contract call is failed")
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			ctrLog.Error().Err(dbErr).Msg("contract call is failed")
			err = dbErr
		}
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
		err = fmt.Errorf("cannot find contract %s", types.EncodeAddress(contractAddress))
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

func setContract(contractState *state.ContractState, code, contractAddress []byte) (*Contract, uint32, error) {
	if len(code) <= 4 {
		err := fmt.Errorf("invalid code (%d bytes is too short)", len(code))
		ctrLog.Warn().AnErr("err", err)
		return nil, 0, err
	}
	codeLen := codeLength(code[0:])
	if uint32(len(code)) < codeLen {
		err := fmt.Errorf("invalid code (expected %d bytes, actual %d bytes)", codeLen, len(code))
		ctrLog.Warn().AnErr("err", err)
		return nil, 0, err
	}
	sCode := code[4:codeLen]

	err := contractState.SetCode(sCode)
	if err != nil {
		return nil, 0, err
	}
	contract := getContract(contractState, contractAddress, sCode)
	if contract == nil {
		err = fmt.Errorf("cannot deploy contract %s", types.EncodeAddress(contractAddress))
		ctrLog.Warn().AnErr("err", err)
		return nil, 0, err
	}

	return contract, codeLen, nil
}

func Create(contractState *state.ContractState, code, contractAddress []byte,
	bcCtx *LBlockchainCtx) (string, error) {

	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("contractAddress", types.EncodeAddress(contractAddress)).Msg("new contract is deployed")
	}
	contract, codeLen, err := setContract(contractState, code, contractAddress)
	if err != nil {
		bcCtx.Close()
		return "", err
	}
	contractState.SetData([]byte("Creator"), []byte(C.GoString(bcCtx.sender)))
	var ci types.CallInfo
	if len(code) != int(codeLen) {
		err = json.Unmarshal(code[codeLen:], &ci.Args)
	}
	if err != nil {
		bcCtx.Close()
		logger.Warn().Err(err).Msg("invalid constructor argument")
		errMsg, _ := json.Marshal("constructor call error:" + err.Error())
		return string(errMsg), nil
	}

	var ce *Executor
	ce = newExecutor(contract, bcCtx)
	defer ce.close(true)

	// create a sql database for the contract
	db := LuaGetDbHandle(bcCtx.stateKey, bcCtx.contractId, bcCtx.rp, bcCtx.isQuery)
	if db == nil {
		return "", newDbSystemError("can't open a database connection")
	}

	ce.constructCall(&ci)
	err = ce.err

	if err != nil {
		logger.Warn().Err(err).Msg("constructor is failed")
		ret, _ := json.Marshal("constructor call error:" + err.Error())
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			logger.Error().Err(dbErr).Msg("constructor is failed")
			return string(ret), dbErr
		}
		return string(ret), nil
	}
	err = ce.commitCalledContract()
	if err != nil {
		ret, _ := json.Marshal("constructor call error:" + err.Error())
		logger.Error().Err(err).Msg("constructor is failed")
		return string(ret), err
	}
	return ce.jsonRet, nil

}

func Query(contractAddress []byte, bs *state.BlockState, contractState *state.ContractState, queryInfo []byte) (res []byte, err error) {
	var ci types.CallInfo
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
		return
	}

	var ce *Executor

	bcCtx := NewContext(bs, nil, contractState, "", "",
		0, 0, "", 0, types.EncodeAddress(contractAddress),
		1, nil, contractState.SqlRecoveryPoint, ChainService, 0)

	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("abi", string(queryInfo)).Msgf("contract %s", types.EncodeAddress(contractAddress))
	}
	ce = newExecutor(contract, bcCtx)
	defer ce.close(true)
	defer func() {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			err = dbErr
		}
	}()
	ce.call(&ci, nil)
	return []byte(ce.jsonRet), ce.err
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
	aid := types.ToAccountID(cid)

	stateSet := contractMap.lookup(stateKeyStr)
	if stateSet == nil {
		luaPushStr(L, "[System.LuaCallContract]not found contract state")
		return -1
	}

	rootState := stateSet.rootState
	callState := rootState.callState[contractIdStr]
	if callState == nil {
		bs := rootState.bs

		prevState, err := bs.GetAccountState(aid)
		if err != nil {
			luaPushStr(L, "[System.LuaGetContract]getAccount Error :"+err.Error())
			return -1
		}

		curState := types.Clone(*prevState).(types.State)
		contractState, err := bs.OpenContractState(aid, &curState)
		if err != nil {
			luaPushStr(L, "[System.LuaGetContract]getAccount Error"+err.Error())
			return -1
		}
		callState =
			&CallState{ctrState: contractState, prevState: prevState, curState: &curState}
		rootState.callState[contractIdStr] = callState
	}
	if callState.ctrState == nil {
		callState.ctrState, err = rootState.bs.OpenContractState(aid, callState.curState)
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
	newBcCtx := NewContext(nil, nil, callState.ctrState,
		C.GoString(bcCtx.contractId), C.GoString(bcCtx.txHash), uint64(bcCtx.blockHeight), int64(bcCtx.timestamp),
		"", int(bcCtx.confirmed), contractIdStr, int(bcCtx.isQuery), rootState, callState.curState.SqlRecoveryPoint,
		int(bcCtx.service), amount)
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
		luaPushStr(L, "[System.LuaCallContract] call err:"+ce.err.Error())
		return -1
	}

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

//export LuaGetDbHandle
func LuaGetDbHandle(ctxKey *C.char, contract *C.char, rp C.ulonglong, readOnly C.int) *C.sqlite3 {
	ctx := contractMap.lookup(C.GoString(ctxKey))
	callState := ctx.rootState.callState[C.GoString(contract)]
	if callState.tx != nil {
		return callState.tx.GetHandle()
	}
	var tx Tx
	var err error
	if int(readOnly) == 1 {
		tx, err = BeginReadOnly(C.GoString(contract), uint64(rp))
	} else {
		tx, err = BeginTx(C.GoString(contract), uint64(rp))
	}
	if err != nil {
		logger.Error().Err(err).Msg("Begin SQL Transaction")
		return nil
	}
	if int(readOnly) != 1 {
		err = tx.Savepoint()
		if err != nil {
			logger.Error().Err(err).Msg("Begin SQL Transaction")
			return nil
		}
	}
	callState.tx = tx
	return callState.tx.GetHandle()
}
