/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package contract

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.1 -I${SRCDIR}/../libtool/include
#cgo !windows CFLAGS: -DLJ_TARGET_POSIX
#cgo darwin LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a ${SRCDIR}/../libtool/lib/libgmp.dylib -lm
#cgo windows LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a ${SRCDIR}/../libtool/bin/libgmp-10.dll -lm
#cgo !darwin,!windows LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a ${SRCDIR}/../libtool/lib/libgmp.so -lm

#include <stdlib.h>
#include <string.h>
#include "vm.h"
#include "lgmp.h"
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/fee"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const (
	maxStateSet       = 20
	callMaxInstLimit  = C.int(5000000)
	queryMaxInstLimit = callMaxInstLimit * C.int(10)
	dbUpdateMaxLimit  = fee.StateDbMaxUpdateSize
	maxCallDepth      = 5
)

var (
	ctrLog         *log.Logger
	curStateSet    [maxStateSet]*StateSet
	lastQueryIndex int
	querySync      sync.Mutex
	zeroFee        *big.Int
)

type CallState struct {
	ctrState  *state.ContractState
	prevState *types.State
	curState  *types.State
	tx        Tx
}

type ContractInfo struct {
	callState  *CallState
	sender     []byte
	contractId []byte
	rp         uint64
	amount     *big.Int
}

type StateSet struct {
	curContract       *ContractInfo
	bs                *state.BlockState
	origin            []byte
	txHash            []byte
	blockHeight       uint64
	timestamp         int64
	node              string
	confirmed         bool
	isQuery           bool
	prevBlockHash     []byte
	service           C.int
	callState         map[types.AccountID]*CallState
	lastRecoveryEntry *recoveryEntry
	dbUpdateTotalSize int64
	seed              *rand.Rand
	events            []*types.Event
	eventCount        int32
	callDepth         int32
}

type recoveryEntry struct {
	seq           int
	amount        *big.Int
	senderState   *types.State
	senderNonce   uint64
	callState     *CallState
	onlySend      bool
	sqlSaveName   *string
	stateRevision state.Snapshot
	prev          *recoveryEntry
}

type LState = C.struct_lua_State

type Executor struct {
	L        *LState
	code     []byte
	err      error
	numArgs  C.int
	stateSet *StateSet
	jsonRet  string
	isView   bool
}

func init() {
	ctrLog = log.NewLogger("contract")
	lastQueryIndex = ChainService
	zeroFee = big.NewInt(0)
}

func newContractInfo(callState *CallState, sender, contractId []byte, rp uint64, amount *big.Int) *ContractInfo {
	return &ContractInfo{
		callState,
		sender,
		contractId,
		rp,
		amount,
	}
}

func NewContext(blockState *state.BlockState, sender, reciever *state.V,
	contractState *state.ContractState, senderID []byte, txHash []byte, blockHeight uint64,
	timestamp int64, prevBlockHash []byte, node string, confirmed bool,
	query bool, rp uint64, service int, amount *big.Int) *StateSet {

	callState := &CallState{ctrState: contractState, curState: reciever.State()}

	stateSet := &StateSet{
		curContract:   newContractInfo(callState, senderID, reciever.ID(), rp, amount),
		bs:            blockState,
		origin:        senderID,
		txHash:        txHash,
		node:          node,
		confirmed:     confirmed,
		isQuery:       query,
		blockHeight:   blockHeight,
		timestamp:     timestamp,
		prevBlockHash: prevBlockHash,
		service:       C.int(service),
	}
	stateSet.callState = make(map[types.AccountID]*CallState)
	stateSet.callState[reciever.AccountID()] = callState
	if sender != nil {
		stateSet.callState[sender.AccountID()] = &CallState{curState: sender.State()}
	}

	return stateSet
}

func NewContextQuery(blockState *state.BlockState, receiverId []byte,
	contractState *state.ContractState, node string, confirmed bool,
	rp uint64) *StateSet {

	callState := &CallState{ctrState: contractState, curState: contractState.State}

	stateSet := &StateSet{
		curContract: newContractInfo(callState, nil, receiverId, rp, big.NewInt(0)),
		bs:          blockState,
		node:        node,
		confirmed:   confirmed,
		timestamp:   time.Now().UnixNano(),
		isQuery:     true,
	}
	stateSet.callState = make(map[types.AccountID]*CallState)
	stateSet.callState[types.ToAccountID(receiverId)] = callState

	return stateSet
}

func (s *StateSet) usedFee() *big.Int {
	if fee.IsZeroFee() {
		return zeroFee
	}
	size := fee.PaymentDataSize(s.dbUpdateTotalSize)
	return new(big.Int).Mul(big.NewInt(size), fee.AerPerByte)
}

func NewLState() *LState {
	return C.vm_newstate()
}

func (L *LState) Close() {
	if L != nil {
		C.lua_close(L)
	}
}

func newExecutor(contract []byte, contractId []byte, stateSet *StateSet, ci *types.CallInfo, amount *big.Int, isCreate bool) *Executor {
	if stateSet.callDepth > maxCallDepth {
		ce := &Executor{
			code:     contract,
			stateSet: stateSet,
		}
		ce.err = errors.New(fmt.Sprintf("exceeded the maximum call depth(%d)", maxCallDepth))
		return ce
	}
	stateSet.callDepth++
	ce := &Executor{
		code:     contract,
		L:        GetLState(),
		stateSet: stateSet,
	}
	if ce.L == nil {
		ce.err = newVmStartError()
		ctrLog.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("new AergoLua executor")
		return ce
	}
	bakupService := stateSet.service
	stateSet.service = -1
	hexId := C.CString(hex.EncodeToString(contractId))
	defer C.free(unsafe.Pointer(hexId))
	if cErrMsg := C.vm_loadbuff(
		ce.L,
		(*C.char)(unsafe.Pointer(&contract[0])),
		C.size_t(len(contract)),
		hexId,
		&stateSet.service,
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		ce.err = errors.New(errMsg)
		ctrLog.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("failed to load code")
		return ce
	}
	stateSet.service = bakupService

	if isCreate == false {
		C.vm_remove_constructor(ce.L)
		fname := C.CString(ci.Name)
		defer C.free(unsafe.Pointer(fname))

		var viewFlag, payFlag C.int
		resolvedName := C.vm_resolve_function(ce.L, fname, &viewFlag, &payFlag)
		if resolvedName == nil {
			ce.err = fmt.Errorf("attempt to call global '%s' (a nil value)", ci.Name)
			ctrLog.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("not found function")
			return ce
		}

		if err := checkPayable(ce.L, resolvedName, &payFlag, amount); err != nil {
			ce.err = err
			ctrLog.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("check payable function")
			return ce
		}
		if viewFlag != C.int(0) {
			ce.isView = true
		}
		C.vm_get_abi_function(ce.L, resolvedName)
		ce.numArgs = C.int(len(ci.Args) + 1)
	} else {
		if err := checkPayable(ce.L, C.construct_name, nil, amount); err != nil {
			ce.err = errVmConstructorIsNotPayable
			ctrLog.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("check payable function")
			return ce
		}
		C.vm_get_constructor(ce.L)
		if C.vm_isnil(ce.L, C.int(-1)) == 1 {
			ce.close()
			return nil
		}
		ce.numArgs = C.int(len(ci.Args))
	}
	ce.processArgs(ci)
	if ce.err != nil {
		ctrLog.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("invalid argument")
	}
	return ce
}

func (ce *Executor) processArgs(ci *types.CallInfo) {
	for _, v := range ci.Args {
		if err := pushValue(ce.L, v); err != nil {
			ce.err = err
			return
		}
	}
}

func (ce *Executor) getEvents() []*types.Event {
	if ce == nil || ce.stateSet == nil {
		return nil
	}
	return ce.stateSet.events
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
	for k, v := range tab {
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

func checkPayable(L *LState, fname *C.char, flag *C.int, amount *big.Int) error {
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil
	}
	var payableFlag C.int
	if flag == nil {
		payableFlag = C.vm_is_payable_function(L, fname)
	} else {
		payableFlag = *flag
	}
	if payableFlag == C.int(0) {
		return fmt.Errorf("'%v' is not payable", C.GoString(fname))
	}
	return nil
}

func (ce *Executor) call(target *LState) C.int {
	if ce.err != nil {
		return 0
	}
	if ce.isView == true {
		oldIsQuery := ce.stateSet.isQuery
		ce.stateSet.isQuery = true
		defer func() {
			ce.stateSet.isQuery = oldIsQuery
		}()
	}
	nret := C.int(0)
	if cErrMsg := C.vm_pcall(ce.L, ce.numArgs, &nret); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		if target != nil {
			if C.luaL_hasuncatchablerror(ce.L) != C.int(0) {
				C.luaL_setuncatchablerror(target)
			}
			if C.luaL_hassyserror(ce.L) != C.int(0) {
				C.luaL_setsyserror(target)
			}
		}
		if C.luaL_hassyserror(ce.L) != C.int(0) {
			ce.err = newVmSystemError(errors.New(errMsg))
		} else {
			ce.err = errors.New(errMsg)
		}
		ctrLog.Warn().Err(ce.err).Str(
			"contract",
			types.EncodeAddress(ce.stateSet.curContract.contractId),
		).Msg("contract is failed")
		return 0
	}
	if target == nil {
		ce.jsonRet = C.GoString(C.vm_get_json_ret(ce.L, nret))
	} else {
		C.luaL_disablemaxmem(target)
		if cErrMsg := C.vm_copy_result(ce.L, target, nret); cErrMsg != nil {
			errMsg := C.GoString(cErrMsg)
			ce.err = errors.New(errMsg)
			ctrLog.Warn().Err(ce.err).Str(
				"contract",
				types.EncodeAddress(ce.stateSet.curContract.contractId),
			).Msg("failed to move results")
		}
		C.luaL_enablemaxmem(target)
	}
	return nret
}

func (ce *Executor) commitCalledContract() error {
	stateSet := ce.stateSet

	if stateSet == nil || stateSet.callState == nil {
		return nil
	}

	bs := stateSet.bs
	rootContract := stateSet.curContract.callState.ctrState

	var err error
	for k, v := range stateSet.callState {
		if v.tx != nil {
			err = v.tx.Release()
			if err != nil {
				return newDbSystemError(err)
			}
		}
		if v.ctrState == rootContract {
			continue
		}
		if v.ctrState != nil {
			err = bs.StageContractState(v.ctrState)
			if err != nil {
				return newDbSystemError(err)
			}
		}
		/* For Sender */
		if v.prevState == nil {
			continue
		}
		err = bs.PutState(k, v.curState)
		if err != nil {
			return newDbSystemError(err)
		}
	}
	return nil
}

func (ce *Executor) rollbackToSavepoint() error {
	stateSet := ce.stateSet

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
			return newDbSystemError(err)
		}
	}
	return nil
}

func (ce *Executor) close() {
	if ce != nil {
		if ce.stateSet != nil {
			ce.stateSet.callDepth--
		}
		FreeLState(ce.L)
	}
}

func getCallInfo(ci interface{}, args []byte, contractAddress []byte) error {
	d := json.NewDecoder(bytes.NewReader(args))
	d.UseNumber()
	d.DisallowUnknownFields()
	err := d.Decode(ci)
	if err != nil {
		ctrLog.Warn().AnErr("error", err).Str(
			"contract",
			types.EncodeAddress(contractAddress),
		).Msg("invalid calling information")
	}
	return err
}

func Call(contractState *state.ContractState, code, contractAddress []byte,
	stateSet *StateSet) (string, []*types.Event, *big.Int, error) {

	var err error
	var ci types.CallInfo
	contract := getContract(contractState, nil)
	if contract != nil {
		if len(code) > 0 {
			err = getCallInfo(&ci, code, contractAddress)
		}
	} else {
		addr := types.EncodeAddress(contractAddress)
		ctrLog.Warn().Str("error", "not found contract").Str("contract", addr).Msg("call")
		err = fmt.Errorf("not found contract %s", addr)
	}
	if err != nil {
		return "", nil, stateSet.usedFee(), err
	}
	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("abi", string(code)).Str("contract", types.EncodeAddress(contractAddress)).Msg("call")
	}

	curStateSet[stateSet.service] = stateSet
	ce := newExecutor(contract, contractAddress, stateSet, &ci, stateSet.curContract.amount, false)
	defer ce.close()

	ce.setCountHook(callMaxInstLimit)

	ce.call(nil)
	err = ce.err
	if err == nil {
		err = ce.commitCalledContract()
		if err != nil {
			logger.Error().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("commit state")
		}
	} else {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			logger.Error().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("rollback state")
			err = dbErr
		}
	}

	return ce.jsonRet, ce.getEvents(), stateSet.usedFee(), err
}

func setRandomSeed(stateSet *StateSet) {
	var randSrc rand.Source
	if stateSet.isQuery {
		randSrc = rand.NewSource(stateSet.timestamp)
	} else {
		b, _ := new(big.Int).SetString(enc.ToString(stateSet.prevBlockHash[:7]), 62)
		t, _ := new(big.Int).SetString(enc.ToString(stateSet.txHash[:7]), 62)
		b.Add(b, t)
		randSrc = rand.NewSource(b.Int64())
	}
	stateSet.seed = rand.New(randSrc)
}

func PreCall(ce *Executor, bs *state.BlockState, sender *state.V, contractState *state.ContractState,
	blockNo uint64, ts int64, rp uint64, prevBlockHash []byte) (string, []*types.Event, *big.Int, error) {
	var err error

	defer ce.close()

	stateSet := ce.stateSet
	stateSet.bs = bs
	callState := stateSet.curContract.callState
	callState.ctrState = contractState
	callState.curState = contractState.State
	stateSet.callState[sender.AccountID()] = &CallState{curState: sender.State()}

	stateSet.blockHeight = blockNo
	stateSet.timestamp = ts
	stateSet.curContract.rp = rp
	stateSet.prevBlockHash = prevBlockHash

	curStateSet[stateSet.service] = stateSet
	ce.call(nil)
	err = ce.err
	if err == nil {
		err = ce.commitCalledContract()
		if err != nil {
			ctrLog.Error().Err(err).Str(
				"contract",
				types.EncodeAddress(stateSet.curContract.contractId),
			).Msg("pre-call")
		}
	} else {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			ctrLog.Error().Err(dbErr).Str(
				"contract",
				types.EncodeAddress(stateSet.curContract.contractId),
			).Msg("pre-call")
			err = dbErr
		}
	}
	return ce.jsonRet, ce.getEvents(), stateSet.usedFee(), err
}

func PreloadEx(bs *state.BlockState, contractState *state.ContractState, contractAid types.AccountID, code, contractAddress []byte,
	stateSet *StateSet) (*Executor, error) {

	var err error
	var ci types.CallInfo
	var contractCode []byte

	if bs != nil {
		contractCode = bs.CodeMap[contractAid]
	}
	if contractCode == nil {
		contractCode = getContract(contractState, nil)
		if contractCode != nil && bs != nil {
			bs.CodeMap[contractAid] = contractCode
		}
	}

	if contractCode != nil {
		if len(code) > 0 {
			err = getCallInfo(&ci, code, contractAddress)
		}
	} else {
		addr := types.EncodeAddress(contractAddress)
		ctrLog.Warn().Str("error", "not found contract").Str("contract", addr).Msg("preload")
		err = fmt.Errorf("not found contract %s", addr)
	}
	if err != nil {
		return nil, err
	}
	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("abi", string(code)).Str("contract", types.EncodeAddress(contractAddress)).Msg("preload")
	}
	ce := newExecutor(contractCode, contractAddress, stateSet, &ci, stateSet.curContract.amount, false)
	ce.setCountHook(callMaxInstLimit)

	return ce, nil

}

func setContract(contractState *state.ContractState, contractAddress, code []byte) ([]byte, uint32, error) {
	if len(code) <= 4 {
		err := fmt.Errorf("invalid code (%d bytes is too short)", len(code))
		ctrLog.Warn().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
		return nil, 0, err
	}
	codeLen := codeLength(code[0:])
	if uint32(len(code)) < codeLen {
		err := fmt.Errorf("invalid code (expected %d bytes, actual %d bytes)", codeLen, len(code))
		ctrLog.Warn().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
		return nil, 0, err
	}
	sCode := code[4:codeLen]

	err := contractState.SetCode(sCode)
	if err != nil {
		return nil, 0, err
	}
	contract := getContract(contractState, sCode)
	if contract == nil {
		err = fmt.Errorf("cannot deploy contract %s", types.EncodeAddress(contractAddress))
		ctrLog.Warn().Str("error", "cannot load contract").Str(
			"contract",
			types.EncodeAddress(contractAddress),
		).Msg("deploy")
		return nil, 0, err
	}

	return contract, codeLen, nil
}

func Create(contractState *state.ContractState, code, contractAddress []byte,
	stateSet *StateSet) (string, []*types.Event, *big.Int, error) {
	if len(code) == 0 {
		return "", nil, stateSet.usedFee(), errors.New("contract code is required")
	}

	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
	}
	contract, codeLen, err := setContract(contractState, contractAddress, code)
	if err != nil {
		return "", nil, stateSet.usedFee(), err
	}
	err = contractState.SetData([]byte("Creator"), []byte(types.EncodeAddress(stateSet.curContract.sender)))
	if err != nil {
		return "", nil, stateSet.usedFee(), err
	}
	var ci types.CallInfo
	if len(code) != int(codeLen) {
		err = getCallInfo(&ci.Args, code[codeLen:], contractAddress)
		if err != nil {
			logger.Warn().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("invalid constructor argument")
			errMsg, _ := json.Marshal("constructor call error:" + err.Error())
			return string(errMsg), nil, stateSet.usedFee(), nil
		}
	}

	curStateSet[stateSet.service] = stateSet

	// create a sql database for the contract
	db := LuaGetDbHandle(&stateSet.service)
	if db == nil {
		return "", nil, stateSet.usedFee(), newDbSystemError(errors.New("can't open a database connection"))
	}

	var ce *Executor
	ce = newExecutor(contract, contractAddress, stateSet, &ci, stateSet.curContract.amount, true)
	if ce == nil {
		return "", nil, stateSet.usedFee(), nil
	}
	defer ce.close()

	ce.setCountHook(callMaxInstLimit)
	ce.call(nil)
	err = ce.err

	if err != nil {
		logger.Warn().Err(err).Msg("constructor is failed")
		ret, _ := json.Marshal("constructor call error:" + err.Error())
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			logger.Error().Err(dbErr).Msg("constructor is failed")
			return string(ret), ce.getEvents(), stateSet.usedFee(), dbErr
		}
		return string(ret), ce.getEvents(), stateSet.usedFee(), err
	}
	err = ce.commitCalledContract()
	if err != nil {
		ret, _ := json.Marshal("constructor call error:" + err.Error())
		logger.Error().Err(err).Msg("constructor is failed")
		return string(ret), ce.getEvents(), stateSet.usedFee(), err
	}

	return ce.jsonRet, ce.getEvents(), stateSet.usedFee(), err
}

func setQueryContext(stateSet *StateSet) {
	querySync.Lock()
	defer querySync.Unlock()
	startIndex := lastQueryIndex
	index := startIndex
	for {
		index++
		if index == maxStateSet {
			index = ChainService + 1
		}
		if curStateSet[index] == nil {
			stateSet.service = C.int(index)
			curStateSet[index] = stateSet
			lastQueryIndex = index
			return
		}
		if index == startIndex {
			querySync.Unlock()
			time.Sleep(100 * time.Millisecond)
			querySync.Lock()
		}
	}
}

func Query(contractAddress []byte, bs *state.BlockState, contractState *state.ContractState, queryInfo []byte) (res []byte, err error) {
	var ci types.CallInfo
	contract := getContract(contractState, nil)
	if contract != nil {
		err = getCallInfo(&ci, queryInfo, contractAddress)
	} else {
		addr := types.EncodeAddress(contractAddress)
		ctrLog.Warn().Str("error", "not found contract").Str("contract", addr).Msg("query")
		err = fmt.Errorf("not found contract %s", addr)
	}
	if err != nil {
		return
	}

	var ce *Executor

	stateSet := NewContextQuery(bs, contractAddress, contractState, "", true,
		contractState.SqlRecoveryPoint)

	setQueryContext(stateSet)
	if ctrLog.IsDebugEnabled() {
		ctrLog.Debug().Str("abi", string(queryInfo)).Str("contract", types.EncodeAddress(contractAddress)).Msg("query")
	}
	ce = newExecutor(contract, contractAddress, stateSet, &ci, stateSet.curContract.amount, false)
	defer ce.close()
	defer func() {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			err = dbErr
		}
	}()
	ce.setCountHook(queryMaxInstLimit)
	ce.call(nil)

	curStateSet[stateSet.service] = nil
	return []byte(ce.jsonRet), ce.err
}

func getContract(contractState *state.ContractState, code []byte) []byte {
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
	return val[4 : 4+l]
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

func (re *recoveryEntry) recovery() error {
	var zero big.Int
	callState := re.callState
	if re.amount.Cmp(&zero) > 0 {
		if re.senderState != nil {
			re.senderState.Balance = new(big.Int).Add(re.senderState.GetBalanceBigInt(), re.amount).Bytes()
		}
		if callState != nil {
			callState.curState.Balance = new(big.Int).Sub(callState.curState.GetBalanceBigInt(), re.amount).Bytes()
		}
	}
	if re.onlySend {
		return nil
	}
	if re.senderState != nil {
		re.senderState.Nonce = re.senderNonce
	}

	if callState == nil {
		return nil
	}
	if re.stateRevision != -1 {
		err := callState.ctrState.Rollback(re.stateRevision)
		if err != nil {
			return newDbSystemError(err)
		}
	}
	if callState.tx != nil {
		if re.sqlSaveName == nil {
			err := callState.tx.RollbackToSavepoint()
			if err != nil {
				return newDbSystemError(err)
			}
			callState.tx = nil
		} else {
			err := callState.tx.RollbackToSubSavepoint(*re.sqlSaveName)
			if err != nil {
				return newDbSystemError(err)
			}
		}
	}
	return nil
}
