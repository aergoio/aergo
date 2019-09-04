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
	"os"
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
	maxContext           = 20
	callMaxInstLimit     = C.int(5000000)
	queryMaxInstLimit    = callMaxInstLimit * C.int(10)
	dbUpdateMaxLimit     = fee.StateDbMaxUpdateSize
	maxCallDepth         = 5
	checkFeeDelegationFn = "check_delegation"
)

var (
	ctrLgr         *log.Logger
	contexts       [maxContext]*vmContext
	lastQueryIndex int
	querySync      sync.Mutex
)

type ChainAccessor interface {
	GetBlockByNo(blockNo types.BlockNo) (*types.Block, error)
	GetBestBlock() (*types.Block, error)
}

type callState struct {
	ctrState  *state.ContractState
	prevState *types.State
	curState  *types.State
	tx        sqlTx
}

type contractInfo struct {
	callState  *callState
	sender     []byte
	contractId []byte
	rp         uint64
	amount     *big.Int
}

type vmContext struct {
	curContract       *contractInfo
	bs                *state.BlockState
	cdb               ChainAccessor
	origin            []byte
	txHash            []byte
	blockInfo         *types.BlockHeaderInfo
	node              string
	confirmed         bool
	isQuery           bool
	nestedView        int32
	isFeeDelegation   bool
	service           C.int
	callState         map[types.AccountID]*callState
	lastRecoveryEntry *recoveryEntry
	dbUpdateTotalSize int64
	seed              *rand.Rand
	events            []*types.Event
	eventCount        int32
	callDepth         int32
	traceFile         *os.File
	gasLimit          uint64
	remainedGas       uint64
	timeout           <-chan struct{}
}

type recoveryEntry struct {
	seq           int
	amount        *big.Int
	senderState   *types.State
	senderNonce   uint64
	callState     *callState
	onlySend      bool
	sqlSaveName   *string
	stateRevision state.Snapshot
	prev          *recoveryEntry
}

type LState = C.struct_lua_State

type executor struct {
	L       *LState
	code    []byte
	err     error
	numArgs C.int
	ctx     *vmContext
	jsonRet string
	isView  bool
}

func init() {
	ctrLgr = log.NewLogger("contract")
	lastQueryIndex = ChainService
}

func newContractInfo(cs *callState, sender, contractId []byte, rp uint64, amount *big.Int) *contractInfo {
	return &contractInfo{
		cs,
		sender,
		contractId,
		rp,
		amount,
	}
}

func getTraceFile(blkno uint64, tx []byte) *os.File {
	f, _ := os.OpenFile(fmt.Sprintf("%s%s%d.trace", os.TempDir(), string(os.PathSeparator), blkno), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if f != nil {
		_, _ = f.WriteString(fmt.Sprintf("[START TX]: %s\n", enc.ToString(tx)))
	}
	return f
}

func newVmContext(blockState *state.BlockState, cdb ChainAccessor, sender, reciever *state.V,
	contractState *state.ContractState, senderID []byte, txHash []byte, bi *types.BlockHeaderInfo, node string, confirmed bool,
	query bool, rp uint64, service int, amount *big.Int, gasLimit uint64, timeout <-chan struct{}, feeDelegation bool) *vmContext {

	cs := &callState{ctrState: contractState, curState: reciever.State()}

	ctx := &vmContext{
		curContract:     newContractInfo(cs, senderID, reciever.ID(), rp, amount),
		bs:              blockState,
		cdb:             cdb,
		origin:          senderID,
		txHash:          txHash,
		node:            node,
		confirmed:       confirmed,
		isQuery:         query,
		blockInfo:       bi,
		service:         C.int(service),
		gasLimit:        gasLimit,
		remainedGas:     gasLimit,
		timeout:         timeout,
		isFeeDelegation: feeDelegation,
	}
	ctx.callState = make(map[types.AccountID]*callState)
	ctx.callState[reciever.AccountID()] = cs
	if sender != nil {
		ctx.callState[sender.AccountID()] = &callState{curState: sender.State()}
	}
	if TraceBlockNo != 0 && TraceBlockNo == ctx.blockInfo.No {
		ctx.traceFile = getTraceFile(ctx.blockInfo.No, txHash)
	}

	return ctx
}

func newVmContextQuery(
	blockState *state.BlockState,
	cdb ChainAccessor,
	receiverId []byte,
	contractState *state.ContractState,
	rp uint64,
) *vmContext {
	cs := &callState{ctrState: contractState, curState: contractState.State}
	ctx := &vmContext{
		curContract: newContractInfo(cs, nil, receiverId, rp, big.NewInt(0)),
		bs:          blockState,
		cdb:         cdb,
		confirmed:   true,
		blockInfo:   &types.BlockHeaderInfo{Ts: time.Now().UnixNano()},
		isQuery:     true,
	}
	ctx.callState = make(map[types.AccountID]*callState)
	ctx.callState[types.ToAccountID(receiverId)] = cs
	return ctx
}

func (s *vmContext) usedFee() *big.Int {
	if fee.IsZeroFee() {
		return fee.NewZeroFee()
	}
	if vmIsGasSystem(s) {
		return new(big.Int).Mul(s.bs.GasPrice, new(big.Int).SetUint64(s.gasLimit - s.remainedGas))
	}
	return fee.PaymentDataFee(s.dbUpdateTotalSize)
}

func newLState() *LState {
	return C.vm_newstate()
}

func (L *LState) close() {
	if L != nil {
		C.lua_close(L)
	}
}

func resolveFunction(contractState *state.ContractState, name string, constructor bool) (*types.Function, error) {
	abi, err := GetABI(contractState)
	if err != nil {
		return nil, err
	}
	var defaultFunc *types.Function
	for _, f := range abi.Functions {
		if f.Name == name {
			return f, nil
		}
		if f.Name == "default" {
			defaultFunc = f
		}
	}
	if constructor {
		return nil, nil
	}
	if len(name) == 0 && defaultFunc != nil {
		return defaultFunc, nil
	}
	return nil, errors.New("not found function: " + name)
}

func newExecutor(
	contract []byte,
	contractId []byte,
	ctx *vmContext,
	ci *types.CallInfo,
	amount *big.Int,
	isCreate bool,
	isDelegation bool,
	ctrState *state.ContractState,
) *executor {

	if ctx.callDepth > maxCallDepth {
		ce := &executor{
			code: contract,
			ctx:  ctx,
		}
		ce.err = errors.New(fmt.Sprintf("exceeded the maximum call depth(%d)", maxCallDepth))
		return ce
	}
	ctx.callDepth++
	ce := &executor{
		code: contract,
		L:    getLState(),
		ctx:  ctx,
	}
	if ce.L == nil {
		ce.err = newVmStartError()
		ctrLgr.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("new AergoLua executor")
		return ce
	}
	if HardforkConfig.IsV2Fork(ctx.blockInfo.No) {
		C.setHardforkV2(ce.L)
		C.vm_set_timeout_hook(ce.L)
	}
	if vmIsGasSystem(ctx) {
		ce.setGas()
	} else {
		C.vm_set_resource_limit(ce.L)
	}
	backupService := ctx.service
	ctx.service = -1
	hexId := C.CString(hex.EncodeToString(contractId))
	defer C.free(unsafe.Pointer(hexId))
	if cErrMsg := C.vm_loadbuff(
		ce.L,
		(*C.char)(unsafe.Pointer(&contract[0])),
		C.size_t(len(contract)),
		hexId,
		&ctx.service,
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		ce.err = errors.New(errMsg)
		ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("failed to load code")
		return ce
	}
	ctx.service = backupService

	if isCreate {
		f, err := resolveFunction(ctrState, "constructor", isCreate)
		if err != nil {
			ce.err = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("not found function")
			return ce
		}
		if f == nil {
			f = &types.Function{
				Name:    "constructor",
				Payable: false,
			}
		}
		err = checkPayable(f, amount)
		if err != nil {
			ce.err = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("check payable function")
			return ce
		}
		ce.isView = f.View
		fName := C.CString("constructor")
		C.vm_get_autoload(ce.L, fName)
		C.free(unsafe.Pointer(fName))
		if C.vm_isnil(ce.L, C.int(-1)) == 1 {
			ce.close()
			return nil
		}
		ce.numArgs = C.int(len(ci.Args))
	} else if isDelegation {
		_, err := resolveFunction(ctrState, checkFeeDelegationFn, false)
		if err != nil {
			ce.err = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("not found function")
			return ce
		}
		ce.isView = true
		fName := C.CString(checkFeeDelegationFn)
		C.vm_get_autoload(ce.L, fName)
		C.free(unsafe.Pointer(fName))
		if C.vm_isnil(ce.L, C.int(-1)) == 1 {
			ce.err = errors.New("not found check_delegation function")
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("not found function")
			return ce
		}
		ce.numArgs = C.int(len(ci.Args))
	} else {
		C.vm_remove_constructor(ce.L)
		f, err := resolveFunction(ctrState, ci.Name, isCreate)
		if err != nil {
			ce.err = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("not found function")
			return ce
		}
		err = checkPayable(f, amount)
		if err != nil {
			ce.err = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("check payable function")
			return ce
		}
		ce.isView = f.View
		resolvedName := C.CString(f.Name)
		C.vm_get_abi_function(ce.L, resolvedName)
		C.free(unsafe.Pointer(resolvedName))
		ce.numArgs = C.int(len(ci.Args) + 1)
	}
	ce.processArgs(ci)
	if ce.err != nil {
		ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("invalid argument")
	}
	return ce
}

func (ce *executor) processArgs(ci *types.CallInfo) {
	for _, v := range ci.Args {
		if err := pushValue(ce.L, v); err != nil {
			ce.err = err
			return
		}
	}
}

func (ce *executor) getEvents() []*types.Event {
	if ce == nil || ce.ctx == nil {
		return nil
	}
	return ce.ctx.events
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

func checkPayable(callee *types.Function, amount *big.Int) error {
	if amount.Cmp(big.NewInt(0)) <= 0 || callee.Payable {
		return nil
	}
	return fmt.Errorf("'%s' is not payable", callee.Name)
}

func (ce *executor) call(target *LState) C.int {
	if ce.err != nil {
		return 0
	}
	if ce.isView == true {
		ce.ctx.nestedView++
		defer func() {
			ce.ctx.nestedView--
		}()
	}
	nret := C.int(0)
	if cErrMsg := C.vm_pcall(ce.L, ce.numArgs, &nret); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		if C.luaL_hassyserror(ce.L) != C.int(0) {
			ce.err = newVmSystemError(errors.New(errMsg))
		} else {
			if C.luaL_hasuncatchablerror(ce.L) != C.int(0) &&
				C.ERR_BF_TIMEOUT == errMsg {
				ce.err = &VmTimeoutError{}
			} else {
				ce.err = errors.New(errMsg)
			}
		}
		ctrLgr.Debug().Err(ce.err).Str(
			"contract",
			types.EncodeAddress(ce.ctx.curContract.contractId),
		).Msg("contract is failed")
		if target != nil {
			if C.luaL_hasuncatchablerror(ce.L) != C.int(0) {
				C.luaL_setuncatchablerror(target)
			}
			if C.luaL_hassyserror(ce.L) != C.int(0) {
				C.luaL_setsyserror(target)
			}
		}
		return 0
	}
	if target == nil {
		var errRet C.int
		retMsg := C.GoString(C.vm_get_json_ret(ce.L, nret, &errRet))
		if errRet == 1 {
			ce.err = errors.New(retMsg)
		} else {
			ce.jsonRet = retMsg
		}
	} else {
		if !vmIsGasSystem(ce.ctx) {
			C.luaL_disablemaxmem(target)
		}
		if cErrMsg := C.vm_copy_result(ce.L, target, nret); cErrMsg != nil {
			errMsg := C.GoString(cErrMsg)
			ce.err = errors.New(errMsg)
			ctrLgr.Debug().Err(ce.err).Str(
				"contract",
				types.EncodeAddress(ce.ctx.curContract.contractId),
			).Msg("failed to move results")
		}
		if !vmIsGasSystem(ce.ctx) {
			C.luaL_enablemaxmem(target)
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
	return nret
}

func (ce *executor) commitCalledContract() error {
	ctx := ce.ctx

	if ctx == nil || ctx.callState == nil {
		return nil
	}

	bs := ctx.bs
	rootContract := ctx.curContract.callState.ctrState

	var err error
	for k, v := range ctx.callState {
		if v.tx != nil {
			err = v.tx.release()
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

	if ctx.traceFile != nil {
		_, _ = ce.ctx.traceFile.WriteString("[Put State Balance]\n")
		for k, v := range ctx.callState {
			_, _ = ce.ctx.traceFile.WriteString(fmt.Sprintf("%s : nonce=%d ammount=%s\n",
				k.String(), v.curState.GetNonce(), v.curState.GetBalanceBigInt().String()))
		}
	}

	return nil
}

func (ce *executor) rollbackToSavepoint() error {
	ctx := ce.ctx

	if ctx == nil || ctx.callState == nil {
		return nil
	}

	var err error
	for _, v := range ctx.callState {
		if v.tx == nil {
			continue
		}
		err = v.tx.rollbackToSavepoint()
		if err != nil {
			return newDbSystemError(err)
		}
	}
	return nil
}

func (ce *executor) setGas() {
	if ce == nil || ce.L == nil || ce.err != nil {
		return
	}
	C.lua_gasset(ce.L, C.ulonglong(ce.ctx.remainedGas))
}

func (ce *executor) close() {
	if ce != nil {
		if ce.ctx != nil {
			ce.ctx.callDepth--
			ce.ctx.remainedGas -= uint64(C.lua_gasget(ce.L))
		}
		freeLState(ce.L)
	}
}

func getCallInfo(ci interface{}, args []byte, contractAddress []byte) error {
	d := json.NewDecoder(bytes.NewReader(args))
	d.UseNumber()
	d.DisallowUnknownFields()
	err := d.Decode(ci)
	if err != nil {
		ctrLgr.Debug().AnErr("error", err).Str(
			"contract",
			types.EncodeAddress(contractAddress),
		).Msg("invalid calling information")
	}
	return err
}

func Call(
	contractState *state.ContractState,
	code, contractAddress []byte,
	ctx *vmContext,
) (string, []*types.Event, *big.Int, error) {

	var err error
	var ci types.CallInfo
	contract := getContract(contractState, nil)
	if contract != nil {
		if len(code) > 0 {
			err = getCallInfo(&ci, code, contractAddress)
		}
	} else {
		addr := types.EncodeAddress(contractAddress)
		ctrLgr.Warn().Str("error", "not found contract").Str("contract", addr).Msg("call")
		err = fmt.Errorf("not found contract %s", addr)
	}
	if err != nil {
		return "", nil, ctx.usedFee(), err
	}
	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Str("abi", string(code)).Str("contract", types.EncodeAddress(contractAddress)).Msg("call")
	}

	contexts[ctx.service] = ctx
	ce := newExecutor(contract, contractAddress, ctx, &ci, ctx.curContract.amount, false, false, contractState)
	defer ce.close()
	ce.setCountHook(callMaxInstLimit)

	ce.call(nil)
	err = ce.err
	if err != nil {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			ctrLgr.Error().Err(dbErr).Str("contract", types.EncodeAddress(contractAddress)).Msg("rollback state")
			err = dbErr
		}
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[error] : %s\n", err))
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
			evs := ce.getEvents()
			if evs != nil {
				_, _ = ctx.traceFile.WriteString("[Event]\n")
				for _, ev := range evs {
					eb, _ := ev.MarshalJSON()
					_, _ = ctx.traceFile.Write(eb)
					_, _ = ctx.traceFile.WriteString("\n")
				}
			}
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CALL END] : %s(%s)\n",
				types.EncodeAddress(contractAddress), types.ToAccountID(contractAddress)))
		}
		return "", ce.getEvents(), ctx.usedFee(), err
	}
	err = ce.commitCalledContract()
	if err != nil {
		ctrLgr.Error().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("commit state")
		return "", ce.getEvents(), ctx.usedFee(), err
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[ret] : %s\n", ce.jsonRet))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
		evs := ce.getEvents()
		if evs != nil {
			_, _ = ctx.traceFile.WriteString("[Event]\n")
			for _, ev := range evs {
				eb, _ := ev.MarshalJSON()
				_, _ = ctx.traceFile.Write(eb)
				_, _ = ctx.traceFile.WriteString("\n")
			}
		}
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CALL END] : %s(%s)\n",
			types.EncodeAddress(contractAddress), types.ToAccountID(contractAddress)))
	}
	return ce.jsonRet, ce.getEvents(), ctx.usedFee(), nil
}

func setRandomSeed(ctx *vmContext) {
	var randSrc rand.Source
	if ctx.isQuery {
		randSrc = rand.NewSource(ctx.blockInfo.Ts)
	} else {
		b, _ := new(big.Int).SetString(enc.ToString(ctx.blockInfo.PrevBlockHash[:7]), 62)
		t, _ := new(big.Int).SetString(enc.ToString(ctx.txHash[:7]), 62)
		b.Add(b, t)
		randSrc = rand.NewSource(b.Int64())
	}
	ctx.seed = rand.New(randSrc)
}

func PreCall(
	ce *executor,
	bs *state.BlockState,
	sender *state.V,
	contractState *state.ContractState,
	rp uint64,
	timeout <-chan struct{},
) (string, []*types.Event, *big.Int, error) {
	var err error

	defer ce.close()

	ctx := ce.ctx
	ctx.bs = bs
	cs := ctx.curContract.callState
	cs.ctrState = contractState
	cs.curState = contractState.State
	ctx.callState[sender.AccountID()] = &callState{curState: sender.State()}

	ctx.curContract.rp = rp
	ctx.timeout = timeout

	if TraceBlockNo != 0 && TraceBlockNo == ctx.blockInfo.No {
		ctx.traceFile = getTraceFile(ctx.blockInfo.No, ctx.txHash)
		if ctx.traceFile != nil {
			defer func() {
				_ = ctx.traceFile.Close()
			}()
		}
	}
	contexts[ctx.service] = ctx
	ce.call(nil)
	err = ce.err
	if err == nil {
		err = ce.commitCalledContract()
		if err != nil {
			ctrLgr.Error().Err(err).Str(
				"contract",
				types.EncodeAddress(ctx.curContract.contractId),
			).Msg("pre-call")
		}
	} else {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			ctrLgr.Error().Err(dbErr).Str(
				"contract",
				types.EncodeAddress(ctx.curContract.contractId),
			).Msg("pre-call")
			err = dbErr
		}
	}
	if ctx.traceFile != nil {
		contractId := ctx.curContract.contractId
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[ret] : %s\n", ce.jsonRet))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
		evs := ce.getEvents()
		if evs != nil {
			_, _ = ctx.traceFile.WriteString("[Event]\n")
			for _, ev := range evs {
				eb, _ := ev.MarshalJSON()
				_, _ = ctx.traceFile.Write(eb)
				_, _ = ctx.traceFile.WriteString("\n")
			}
		}
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[PRECALL END] : %s(%s)\n",
			types.EncodeAddress(contractId), types.ToAccountID(contractId)))
	}
	return ce.jsonRet, ce.getEvents(), ctx.usedFee(), err
}

func PreloadEx(bs *state.BlockState, contractState *state.ContractState, contractAid types.AccountID, code, contractAddress []byte,
	ctx *vmContext) (*executor, error) {

	var err error
	var ci types.CallInfo
	var contractCode []byte

	if bs != nil {
		contractCode = bs.CodeMap.Get(contractAid)
	}
	if contractCode == nil {
		contractCode = getContract(contractState, nil)
		if contractCode != nil && bs != nil {
			bs.CodeMap.Add(contractAid, contractCode)
		}
	}

	if contractCode != nil {
		if len(code) > 0 {
			err = getCallInfo(&ci, code, contractAddress)
		}
	} else {
		addr := types.EncodeAddress(contractAddress)
		ctrLgr.Warn().Str("error", "not found contract").Str("contract", addr).Msg("preload")
		err = fmt.Errorf("not found contract %s", addr)
	}
	if err != nil {
		return nil, err
	}
	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Str("abi", string(code)).Str("contract", types.EncodeAddress(contractAddress)).Msg("preload")
	}
	ce := newExecutor(contractCode, contractAddress, ctx, &ci, ctx.curContract.amount, false, false, contractState)
	ce.setCountHook(callMaxInstLimit)

	return ce, nil

}

func setContract(contractState *state.ContractState, contractAddress, code []byte) ([]byte, uint32, error) {
	if len(code) <= 4 {
		err := fmt.Errorf("invalid code (%d bytes is too short)", len(code))
		ctrLgr.Warn().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
		return nil, 0, err
	}
	codeLen := codeLength(code[0:])
	if uint32(len(code)) < codeLen {
		err := fmt.Errorf("invalid code (expected %d bytes, actual %d bytes)", codeLen, len(code))
		ctrLgr.Warn().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
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
		ctrLgr.Warn().Str("error", "cannot load contract").Str(
			"contract",
			types.EncodeAddress(contractAddress),
		).Msg("deploy")
		return nil, 0, err
	}

	return contract, codeLen, nil
}

func Create(
	contractState *state.ContractState,
	code, contractAddress []byte,
	ctx *vmContext,
) (string, []*types.Event, *big.Int, error) {
	if len(code) == 0 {
		return "", nil, ctx.usedFee(), errors.New("contract code is required")
	}

	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
	}
	contract, codeLen, err := setContract(contractState, contractAddress, code)
	if err != nil {
		return "", nil, ctx.usedFee(), err
	}
	err = contractState.SetData(creatorMetaKey, []byte(types.EncodeAddress(ctx.curContract.sender)))
	if err != nil {
		return "", nil, ctx.usedFee(), err
	}
	var ci types.CallInfo
	if len(code) != int(codeLen) {
		err = getCallInfo(&ci.Args, code[codeLen:], contractAddress)
		if err != nil {
			errMsg, _ := json.Marshal("constructor call error:" + err.Error())
			return string(errMsg), nil, ctx.usedFee(), nil
		}
	}

	contexts[ctx.service] = ctx

	// create a sql database for the contract
	if !HardforkConfig.IsV2Fork(ctx.blockInfo.No) {
		if db := luaGetDbHandle(&ctx.service); db == nil {
			return "", nil, ctx.usedFee(), newDbSystemError(errors.New("can't open a database connection"))
		}
	}

	ce := newExecutor(contract, contractAddress, ctx, &ci, ctx.curContract.amount, true, false, contractState)
	if ce == nil {
		return "", nil, ctx.usedFee(), nil
	}
	defer ce.close()
	ce.setCountHook(callMaxInstLimit)

	ce.call(nil)
	err = ce.err
	if err != nil {
		ctrLgr.Debug().Msg("constructor is failed")
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			ctrLgr.Error().Err(dbErr).Msg("rollback state")
			err = dbErr
		}

		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[error] : %s\n", err))
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
			evs := ce.getEvents()
			if evs != nil {
				_, _ = ctx.traceFile.WriteString("[Event]\n")
				for _, ev := range evs {
					eb, _ := ev.MarshalJSON()
					_, _ = ctx.traceFile.Write(eb)
					_, _ = ctx.traceFile.WriteString("\n")
				}
			}
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CREATE END] : %s(%s)\n",
				types.EncodeAddress(contractAddress), types.ToAccountID(contractAddress)))
		}
		return "", ce.getEvents(), ctx.usedFee(), err
	}
	err = ce.commitCalledContract()
	if err != nil {
		ctrLgr.Debug().Msg("constructor is failed")
		ctrLgr.Error().Err(err).Msg("commit state")
		return "", ce.getEvents(), ctx.usedFee(), err
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[ret] : %s\n", ce.jsonRet))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
		evs := ce.getEvents()
		if evs != nil {
			_, _ = ctx.traceFile.WriteString("[Event]\n")
			for _, ev := range evs {
				eb, _ := ev.MarshalJSON()
				_, _ = ctx.traceFile.Write(eb)
				_, _ = ctx.traceFile.WriteString("\n")
			}
		}
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CREATE END] : %s(%s)\n",
			types.EncodeAddress(contractAddress), types.ToAccountID(contractAddress)))
	}
	return ce.jsonRet, ce.getEvents(), ctx.usedFee(), nil
}

func setQueryContext(ctx *vmContext) {
	querySync.Lock()
	defer querySync.Unlock()
	startIndex := lastQueryIndex
	index := startIndex
	for {
		index++
		if index == maxContext {
			index = ChainService + 1
		}
		if contexts[index] == nil {
			ctx.service = C.int(index)
			contexts[index] = ctx
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

func Query(contractAddress []byte, bs *state.BlockState, cdb ChainAccessor, contractState *state.ContractState, queryInfo []byte) (res []byte, err error) {
	var ci types.CallInfo
	contract := getContract(contractState, nil)
	if contract != nil {
		err = getCallInfo(&ci, queryInfo, contractAddress)
	} else {
		addr := types.EncodeAddress(contractAddress)
		if ctrLgr.IsDebugEnabled() {
			ctrLgr.Debug().Str("error", "not found contract").Str("contract", addr).Msg("query")
		}
		err = fmt.Errorf("not found contract %s", addr)
	}
	if err != nil {
		return
	}

	ctx := newVmContextQuery(bs, cdb, contractAddress, contractState, contractState.SqlRecoveryPoint)

	setQueryContext(ctx)
	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Str("abi", string(queryInfo)).Str("contract", types.EncodeAddress(contractAddress)).Msg("query")
	}
	ce := newExecutor(contract, contractAddress, ctx, &ci, ctx.curContract.amount, false, false, contractState)
	defer ce.close()
	defer func() {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			err = dbErr
		}
	}()
	ce.setCountHook(queryMaxInstLimit)
	ce.call(nil)

	contexts[ctx.service] = nil
	return []byte(ce.jsonRet), ce.err
}

func CheckFeeDelegation(contractAddress []byte, bs *state.BlockState, cdb ChainAccessor,
	contractState *state.ContractState, payload, txHash, sender, amount []byte) (err error) {
	var ci types.CallInfo

	err = getCallInfo(&ci, payload, contractAddress)
	if err != nil {
		return
	}
	abi, err := GetABI(contractState)
	if err != nil {
		return err
	}
	var found *types.Function
	for _, f := range abi.Functions {
		if f.Name == ci.Name {
			found = f
			break
		}
	}
	if found == nil {
		return fmt.Errorf("not found function %s", ci.Name)
	}
	if found.FeeDelegation == false {
		return fmt.Errorf("%s function is not declared of fee delegation", ci.Name)
	}

	contract := getContract(contractState, nil)
	if contract == nil {
		addr := types.EncodeAddress(contractAddress)
		ctrLgr.Warn().Str("error", "not found contract").Str("contract", addr).Msg("checkFeeDelegation")
		err = fmt.Errorf("not found contract %s", addr)
	}
	if err != nil {
		return
	}
	ctx := newVmContextQuery(bs, cdb, contractAddress, contractState, contractState.SqlRecoveryPoint)
	ctx.origin = sender
	ctx.txHash = txHash
	ctx.curContract.amount = new(big.Int).SetBytes(amount)
	ctx.curContract.sender = sender

	setQueryContext(ctx)
	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Str("abi", string(checkFeeDelegationFn)).Str("contract", types.EncodeAddress(contractAddress)).Msg("checkFeeDelegation")
	}
	ci.Args = append([]interface{}{ci.Name}, ci.Args...)
	ci.Name = checkFeeDelegationFn
	ce := newExecutor(contract, contractAddress, ctx, &ci, ctx.curContract.amount, false, true, contractState)
	defer ce.close()
	defer func() {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			err = dbErr
		}
	}()
	ce.setCountHook(queryMaxInstLimit)
	ce.call(nil)

	contexts[ctx.service] = nil
	if ce.err != nil {
		return ce.err
	}
	if ce.jsonRet != "true" {
		return types.ErrNotAllowedFeeDelegation
	}
	return nil
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
	cs := re.callState
	if re.amount.Cmp(&zero) > 0 {
		if re.senderState != nil {
			re.senderState.Balance = new(big.Int).Add(re.senderState.GetBalanceBigInt(), re.amount).Bytes()
		}
		if cs != nil {
			cs.curState.Balance = new(big.Int).Sub(cs.curState.GetBalanceBigInt(), re.amount).Bytes()
		}
	}
	if re.onlySend {
		return nil
	}
	if re.senderState != nil {
		re.senderState.Nonce = re.senderNonce
	}

	if cs == nil {
		return nil
	}
	if re.stateRevision != -1 {
		err := cs.ctrState.Rollback(re.stateRevision)
		if err != nil {
			return newDbSystemError(err)
		}
	}
	if cs.tx != nil {
		if re.sqlSaveName == nil {
			err := cs.tx.rollbackToSavepoint()
			if err != nil {
				return newDbSystemError(err)
			}
			cs.tx = nil
		} else {
			err := cs.tx.rollbackToSubSavepoint(*re.sqlSaveName)
			if err != nil {
				return newDbSystemError(err)
			}
		}
	}
	return nil
}
