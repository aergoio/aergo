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
 #cgo !darwin,!windows LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -L${SRCDIR}/../libtool/lib64 -L${SRCDIR}/../libtool/lib -lgmp -lm


 #include <stdlib.h>
 #include <string.h>
 #include "vm.h"
 #include "bignum_module.h"
*/
import "C"
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/aergoio/aergo-lib/log"
	luacUtil "github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
	jsoniter "github.com/json-iterator/go"
)

const (
	callMaxInstLimit     = C.int(5000000)
	queryMaxInstLimit    = callMaxInstLimit * C.int(10)
	dbUpdateMaxLimit     = fee.StateDbMaxUpdateSize
	maxCallDepthOld      = 5
	maxCallDepth         = 64
	checkFeeDelegationFn = "check_delegation"
	constructor          = "constructor"

	vmTimeoutErrMsg = "contract timeout during vm execution"
)

var (
	maxContext         int
	ctrLgr             *log.Logger
	contexts           []*vmContext
	lastQueryIndex     int
	querySync          sync.Mutex
	currentForkVersion int32
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

// vmContext contains context datas during execution of smart contract.
// It has both call infos which are immutable, and real time statuses
// which are mutable during execution
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
	nestedView        int32 // indicates which parent called the contract in view (read-only mode)
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
	execCtx           context.Context
}

type recoveryEntry struct {
	seq           int
	amount        *big.Int
	senderState   *types.State
	senderNonce   uint64
	callState     *callState
	onlySend      bool
	isDeploy      bool
	sqlSaveName   *string
	stateRevision state.Snapshot
	prev          *recoveryEntry
}

type LState = C.struct_lua_State
type executor struct {
	L          *LState
	code       []byte
	err        error
	numArgs    C.int
	ci         *types.CallInfo
	fname      string
	ctx        *vmContext
	jsonRet    string
	isView     bool
	isAutoload bool
	preErr     error
}

func MaxCallDepth(version int32) int32 {
	if version >= 3 {
		return maxCallDepth
	}
	return maxCallDepthOld
}

func init() {
	ctrLgr = log.NewLogger("contract")
	lastQueryIndex = ChainService
}

func InitContext(numCtx int) {
	maxContext = numCtx
	contexts = make([]*vmContext, maxContext)
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
		_, _ = f.WriteString(fmt.Sprintf("[START TX]: %s\n", enc.B58Encode(tx)))
	}
	return f
}

func NewVmContext(execCtx context.Context, blockState *state.BlockState, cdb ChainAccessor, sender, reciever *state.V, contractState *state.ContractState, senderID, txHash []byte, bi *types.BlockHeaderInfo, node string, confirmed, query bool, rp uint64, service int, amount *big.Int, gasLimit uint64, feeDelegation bool) *vmContext {

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
		isFeeDelegation: feeDelegation,
		execCtx:         execCtx,
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

func NewVmContextQuery(
	blockState *state.BlockState,
	cdb ChainAccessor,
	receiverId []byte,
	contractState *state.ContractState,
	rp uint64,
) (*vmContext, error) {
	cs := &callState{ctrState: contractState, curState: contractState.State}
	bb, err := cdb.GetBestBlock()
	if err != nil {
		return nil, err
	}
	ctx := &vmContext{
		curContract: newContractInfo(cs, nil, receiverId, rp, big.NewInt(0)),
		bs:          blockState,
		cdb:         cdb,
		confirmed:   true,
		blockInfo:   types.NewBlockHeaderInfo(bb),
		isQuery:     true,
		execCtx:     context.Background(), // FIXME query also should cancel if query is too long
	}

	ctx.callState = make(map[types.AccountID]*callState)
	ctx.callState[types.ToAccountID(receiverId)] = cs
	return ctx, nil
}

func (ctx *vmContext) IsGasSystem() bool {
	return !ctx.isQuery && PubNet && ctx.blockInfo.ForkVersion >= 2
}

// get the remaining gas from the given LState
func (ctx *vmContext) refreshRemainingGas(L *LState) {
	if ctx.IsGasSystem() {
		ctx.remainedGas = uint64(C.lua_gasget(L))
	}
}

// set the remaining gas on the given LState
func (ctx *vmContext) setRemainingGas(L *LState) {
	if ctx.IsGasSystem() {
		C.lua_gasset(L, C.ulonglong(ctx.remainedGas))
	}
}

func (ctx *vmContext) usedFee() *big.Int {
	if fee.IsZeroFee() {
		return fee.NewZeroFee()
	}
	if ctx.IsGasSystem() {
		usedGas := ctx.usedGas()
		if ctrLgr.IsDebugEnabled() {
			ctrLgr.Debug().Uint64("gas used", usedGas).Str("lua vm", "executed").Msg("gas information")
		}
		return new(big.Int).Mul(ctx.bs.GasPrice, new(big.Int).SetUint64(usedGas))
	}
	return fee.PaymentDataFee(ctx.dbUpdateTotalSize)
}

func (ctx *vmContext) usedGas() uint64 {
	if fee.IsZeroFee() || !ctx.IsGasSystem() {
		return 0
	}
	return ctx.gasLimit - ctx.remainedGas
}

func newLState() *LState {
	ctrLgr.Debug().Msg("LState created")
	return C.vm_newstate(C.int(currentForkVersion))
}

func (L *LState) close() {
	if L != nil {
		C.lua_close(L)
	}
}

type lStatesBuffer struct {
	s     []*LState
	limit int
}

func newLStatesBuffer(limit int) *lStatesBuffer {
	return &lStatesBuffer{
		s:     make([]*LState, 0),
		limit: limit,
	}
}

func (Ls *lStatesBuffer) len() int {
	return len(Ls.s)
}

func (Ls *lStatesBuffer) append(s *LState) {
	Ls.s = append(Ls.s, s)
	if Ls.len() == Ls.limit {
		Ls.close()
	}
}

func (Ls *lStatesBuffer) close() {
	C.vm_closestates(&Ls.s[0], C.int(len(Ls.s)))
	Ls.s = Ls.s[:0]
}

func resolveFunction(contractState *state.ContractState, bs *state.BlockState, name string, constructor bool) (*types.Function, error) {
	abi, err := GetABI(contractState, bs)
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

	if ctx.blockInfo.ForkVersion != currentForkVersion {
		// force the StatePool to regenerate the LStates
		// using the new hardfork version
		currentForkVersion = ctx.blockInfo.ForkVersion
		FlushLStates()
	}

	if ctx.callDepth > MaxCallDepth(ctx.blockInfo.ForkVersion) {
		ce := &executor{
			code: contract,
			ctx:  ctx,
		}
		ce.err = fmt.Errorf("exceeded the maximum call depth(%d)", MaxCallDepth(ctx.blockInfo.ForkVersion))
		return ce
	}
	ctx.callDepth++

	ce := &executor{
		code: contract,
		L:    GetLState(),
		ctx:  ctx,
	}
	if ce.L == nil {
		ce.err = ErrVmStart
		ctrLgr.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("new AergoLua executor")
		return ce
	}
	if ctx.blockInfo.ForkVersion >= 2 {
		C.luaL_set_hardforkversion(ce.L, C.int(ctx.blockInfo.ForkVersion))
	}

	if ctx.IsGasSystem() {
		ce.setGas()
		defer func() {
			ce.refreshRemainingGas()
			if ctrLgr.IsDebugEnabled() {
				ctrLgr.Debug().Uint64("gas used", ce.ctx.usedGas()).Str("lua vm", "loaded").Msg("gas information")
			}
		}()
	}

	ce.vmLoadCode(contractId)
	if ce.err != nil {
		return ce
	}

	if isCreate {
		f, err := resolveFunction(ctrState, ctx.bs, constructor, isCreate)
		if err != nil {
			ce.preErr = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("not found function")
			return ce
		}
		if f == nil {
			f = &types.Function{
				Name:    constructor,
				Payable: false,
			}
		}
		err = checkPayable(f, amount)
		if err != nil {
			ce.preErr = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("check payable function")
			return ce
		}
		ce.isView = f.View
		ce.fname = constructor
		ce.isAutoload = true
		ce.numArgs = C.int(len(ci.Args))
	} else if isDelegation {
		_, err := resolveFunction(ctrState, ctx.bs, checkFeeDelegationFn, false)
		if err != nil {
			ce.preErr = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("not found function")
			return ce
		}
		ce.isView = true
		ce.fname = checkFeeDelegationFn
		ce.isAutoload = true
		ce.numArgs = C.int(len(ci.Args))
	} else {
		f, err := resolveFunction(ctrState, ctx.bs, ci.Name, isCreate)
		if err != nil {
			ce.preErr = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("not found function")
			return ce
		}
		err = checkPayable(f, amount)
		if err != nil {
			ce.preErr = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("check payable function")
			return ce
		}
		ce.isView = f.View
		ce.fname = f.Name
		ce.numArgs = C.int(len(ci.Args) + 1)
	}
	ce.ci = ci

	return ce
}

func (ce *executor) processArgs() {
	for _, v := range ce.ci.Args {
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

func checkPayable(callee *types.Function, amount *big.Int) error {
	if amount.Cmp(big.NewInt(0)) <= 0 || callee.Payable {
		return nil
	}
	return fmt.Errorf("'%s' is not payable", callee.Name)
}

func (ce *executor) call(instLimit C.int, target *LState) C.int {
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
	ce.processArgs()
	if ce.err != nil {
		ctrLgr.Debug().Err(ce.err).Stringer("contract",
			types.LogAddr(ce.ctx.curContract.contractId)).Msg("invalid argument")
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
				return newVmError(err)
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
	for id, v := range ctx.callState {
		if v.prevState != nil && len(v.prevState.GetCodeHash()) == 0 &&
			len(v.curState.GetCodeHash()) != 0 {
			ctx.bs.RemoveCache(id)
		}
		if v.tx == nil {
			continue
		}
		err = v.tx.rollbackToSavepoint()
		if err != nil {
			if strings.HasPrefix(err.Error(), "no such savepoint") {
				_ = v.tx.begin()
			}
			return newVmError(err)
		}
	}
	return nil
}

func (ce *executor) closeQuerySql() error {
	ctx := ce.ctx

	if ctx == nil || ctx.callState == nil {
		return nil
	}

	var err error
	for _, v := range ctx.callState {
		if v.tx == nil {
			continue
		}
		err = v.tx.close()
		if err != nil {
			return newVmError(err)
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
			if ce.ctx.traceFile != nil {
				ce.ctx.traceFile.Close()
				ce.ctx.traceFile = nil
			}
		}
		if ce.L != nil {
			FreeLState(ce.L)
		}
	}
}

func (ce *executor) refreshRemainingGas() {
	ce.ctx.refreshRemainingGas(ce.L)
}

func (ce *executor) gas() uint64 {
	return uint64(C.lua_gasget(ce.L))
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
	payload, contractAddress []byte,
	ctx *vmContext,
) (string, []*types.Event, *big.Int, error) {

	var err error
	var ci types.CallInfo

	// get contract
	contract := getContract(contractState, ctx.bs)
	if contract != nil {
		if len(payload) > 0 {
			// get call arguments
			err = getCallInfo(&ci, payload, contractAddress)
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
		ctrLgr.Debug().Str("abi", string(payload)).Str("contract", types.EncodeAddress(contractAddress)).Msg("call")
	}

	// create a new executor
	contexts[ctx.service] = ctx
	ce := newExecutor(contract, contractAddress, ctx, &ci, ctx.curContract.amount, false, false, contractState)
	defer ce.close()

	startTime := time.Now()
	// execute the contract call
	ce.call(callMaxInstLimit, nil)
	vmExecTime := time.Now().Sub(startTime).Microseconds()
	vmLogger.Trace().Int64("execµs", vmExecTime).Stringer("txHash", types.LogBase58(ce.ctx.txHash)).Msg("tx execute time in vm")

	// check if there is an error
	err = ce.err
	if err != nil {
		// rollback the state of the contract
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			ctrLgr.Error().Err(dbErr).Str("contract", types.EncodeAddress(contractAddress)).Msg("rollback state")
		}
		// log the error
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[error] : %s\n", err))
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
			events := ce.getEvents()
			if events != nil {
				_, _ = ctx.traceFile.WriteString("[Event]\n")
				for _, event := range events {
					eventJson, _ := event.MarshalJSON()
					_, _ = ctx.traceFile.Write(eventJson)
					_, _ = ctx.traceFile.WriteString("\n")
				}
			}
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CALL END] : %s(%s)\n",
				types.EncodeAddress(contractAddress), types.ToAccountID(contractAddress)))
		}
		// return the error
		return "", ce.getEvents(), ctx.usedFee(), err
	}

	// save the state of the contract
	err = ce.commitCalledContract()
	if err != nil {
		ctrLgr.Error().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("commit state")
		return "", ce.getEvents(), ctx.usedFee(), err
	}

	// log the result
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[ret] : %s\n", ce.jsonRet))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
		events := ce.getEvents()
		if events != nil {
			_, _ = ctx.traceFile.WriteString("[Event]\n")
			for _, event := range events {
				eventJson, _ := event.MarshalJSON()
				_, _ = ctx.traceFile.Write(eventJson)
				_, _ = ctx.traceFile.WriteString("\n")
			}
		}
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CALL END] : %s(%s)\n",
			types.EncodeAddress(contractAddress), types.ToAccountID(contractAddress)))
	}

	// return the result
	return ce.jsonRet, ce.getEvents(), ctx.usedFee(), nil
}

func setRandomSeed(ctx *vmContext) {
	var randSrc rand.Source
	if ctx.isQuery {
		randSrc = rand.NewSource(ctx.blockInfo.Ts)
	} else {
		b, _ := new(big.Int).SetString(enc.B58Encode(ctx.blockInfo.PrevBlockHash[:7]), 62)
		t, _ := new(big.Int).SetString(enc.B58Encode(ctx.txHash[:7]), 62)
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
	rp, gasLimit uint64,
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

	ctx.gasLimit = gasLimit
	ctx.remainedGas = gasLimit
	if ctx.IsGasSystem() {
		ce.setGas()
	}

	contexts[ctx.service] = ctx

	// execute the contract call
	ce.call(callMaxInstLimit, nil)

	err = ce.err
	if err == nil {
		// save the state of the contract
		err = ce.commitCalledContract()
		if err != nil {
			ctrLgr.Error().Err(err).Str(
				"contract",
				types.EncodeAddress(ctx.curContract.contractId),
			).Msg("pre-call")
		}
	} else {
		// rollback the state of the contract
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			ctrLgr.Error().Err(dbErr).Str(
				"contract",
				types.EncodeAddress(ctx.curContract.contractId),
			).Msg("pre-call")
		}
	}

	if ctx.traceFile != nil {
		contractId := ctx.curContract.contractId
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[ret] : %s\n", ce.jsonRet))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
		events := ce.getEvents()
		if events != nil {
			_, _ = ctx.traceFile.WriteString("[Event]\n")
			for _, event := range events {
				eventJson, _ := event.MarshalJSON()
				_, _ = ctx.traceFile.Write(eventJson)
				_, _ = ctx.traceFile.WriteString("\n")
			}
		}
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[PRECALL END] : %s(%s)\n",
			types.EncodeAddress(contractId), types.ToAccountID(contractId)))
	}

	return ce.jsonRet, ce.getEvents(), ctx.usedFee(), err
}

// loads a contract and prepares it for execution
func PreloadExecutor(bs *state.BlockState, contractState *state.ContractState, payload, contractAddress []byte,
	ctx *vmContext) (*executor, error) {

	var err error
	var ci types.CallInfo

	// read contract code
	contractCode := getContract(contractState, bs)

	// get the arguments for the call
	if contractCode != nil {
		if len(payload) > 0 {
			err = getCallInfo(&ci, payload, contractAddress)
		}
	} else {
		addr := types.EncodeAddress(contractAddress)
		ctrLgr.Warn().Str("error", "not found contract").Str("contract", addr).Msg("preload")
		err = fmt.Errorf("not found contract %s", addr)
	}
	if err != nil {
		return nil, err
	}

	// log some information
	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Str("abi", string(payload)).Str("contract", types.EncodeAddress(contractAddress)).Msg("preload")
	}

	// create a new executor for the call
	ce := newExecutor(contractCode, contractAddress, ctx, &ci, ctx.curContract.amount, false, false, contractState)

	// return the executor
	return ce, ce.err
}

func setContract(contractState *state.ContractState, contractAddress, payload []byte) ([]byte, []byte, error) {
	codePayload := luacUtil.LuaCodePayload(payload)
	if _, err := codePayload.IsValidFormat(); err != nil {
		ctrLgr.Warn().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
		return nil, nil, err
	}
	code := codePayload.Code()
	err := contractState.SetCode(code.Bytes())
	if err != nil {
		return nil, nil, err
	}
	contract := getContract(contractState, nil)
	if contract == nil {
		err = fmt.Errorf("cannot deploy contract %s", types.EncodeAddress(contractAddress))
		ctrLgr.Warn().Str("error", "cannot load contract").Str(
			"contract",
			types.EncodeAddress(contractAddress),
		).Msg("deploy")
		return nil, nil, err
	}

	return contract, codePayload.Args(), nil
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

	// save the contract code
	contract, args, err := setContract(contractState, contractAddress, code)
	if err != nil {
		return "", nil, ctx.usedFee(), err
	}

	// set the creator
	err = contractState.SetData(dbkey.CreatorMeta(), []byte(types.EncodeAddress(ctx.curContract.sender)))
	if err != nil {
		return "", nil, ctx.usedFee(), err
	}

	// get the arguments for the constructor
	var ci types.CallInfo
	if len(args) > 0 {
		err = getCallInfo(&ci.Args, args, contractAddress)
		if err != nil {
			errMsg, _ := json.Marshal("constructor call error:" + err.Error())
			return string(errMsg), nil, ctx.usedFee(), nil
		}
	}

	contexts[ctx.service] = ctx

	if ctx.blockInfo.ForkVersion < 2 {
		// create a sql database for the contract
		if db := luaGetDbHandle(ctx.service); db == nil {
			return "", nil, ctx.usedFee(), newVmError(errors.New("can't open a database connection"))
		}
	}

	// create a new executor for the constructor
	ce := newExecutor(contract, contractAddress, ctx, &ci, ctx.curContract.amount, true, false, contractState)
	defer ce.close()

	// call the constructor
	ce.call(callMaxInstLimit, nil)

	// check if the call failed
	err = ce.err
	if err != nil {
		ctrLgr.Debug().Msg("constructor is failed")
		// rollback the state
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			ctrLgr.Error().Err(dbErr).Msg("rollback state")
		}
		// write the trace
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[error] : %s\n", err))
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
			events := ce.getEvents()
			if events != nil {
				_, _ = ctx.traceFile.WriteString("[Event]\n")
				for _, event := range events {
					eventJson, _ := event.MarshalJSON()
					_, _ = ctx.traceFile.Write(eventJson)
					_, _ = ctx.traceFile.WriteString("\n")
				}
			}
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CREATE END] : %s(%s)\n",
				types.EncodeAddress(contractAddress), types.ToAccountID(contractAddress)))
		}
		// return the error
		return "", ce.getEvents(), ctx.usedFee(), err
	}

	// commit the state
	err = ce.commitCalledContract()
	if err != nil {
		ctrLgr.Debug().Msg("constructor is failed")
		ctrLgr.Error().Err(err).Msg("commit state")
		return "", ce.getEvents(), ctx.usedFee(), err
	}

	// write the trace
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[ret] : %s\n", ce.jsonRet))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[usedFee] : %s\n", ctx.usedFee().String()))
		events := ce.getEvents()
		if events != nil {
			_, _ = ctx.traceFile.WriteString("[Event]\n")
			for _, event := range events {
				eventJson, _ := event.MarshalJSON()
				_, _ = ctx.traceFile.Write(eventJson)
				_, _ = ctx.traceFile.WriteString("\n")
			}
		}
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CREATE END] : %s(%s)\n",
			types.EncodeAddress(contractAddress), types.ToAccountID(contractAddress)))
	}

	// return the result
	return ce.jsonRet, ce.getEvents(), ctx.usedFee(), nil
}

func allocContextSlot(ctx *vmContext) {
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

func freeContextSlot(ctx *vmContext) {
	querySync.Lock()
	defer querySync.Unlock()
	contexts[ctx.service] = nil
}

func Query(contractAddress []byte, bs *state.BlockState, cdb ChainAccessor, contractState *state.ContractState, queryInfo []byte) (res []byte, err error) {
	var ci types.CallInfo

	contract := getContract(contractState, bs)
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

	var ctx *vmContext
	ctx, err = NewVmContextQuery(bs, cdb, contractAddress, contractState, contractState.SqlRecoveryPoint)
	if err != nil {
		return
	}

	allocContextSlot(ctx)
	defer freeContextSlot(ctx)

	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Str("abi", string(queryInfo)).Str("contract", types.EncodeAddress(contractAddress)).Msg("query")
	}

	ce := newExecutor(contract, contractAddress, ctx, &ci, ctx.curContract.amount, false, false, contractState)
	defer ce.close()
	defer func() {
		if dbErr := ce.closeQuerySql(); dbErr != nil {
			err = dbErr
		}
	}()

	ce.call(queryMaxInstLimit, nil)

	return []byte(ce.jsonRet), ce.err
}

func CheckFeeDelegation(contractAddress []byte, bs *state.BlockState, bi *types.BlockHeaderInfo, cdb ChainAccessor,
	contractState *state.ContractState, payload, txHash, sender, amount []byte) (err error) {
	var ci types.CallInfo

	err = getCallInfo(&ci, payload, contractAddress)
	if err != nil {
		return
	}

	abi, err := GetABI(contractState, bs)
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

	contract := getContract(contractState, bs)
	if contract == nil {
		addr := types.EncodeAddress(contractAddress)
		ctrLgr.Warn().Str("error", "not found contract").Str("contract", addr).Msg("checkFeeDelegation")
		err = fmt.Errorf("not found contract %s", addr)
	}
	if err != nil {
		return
	}

	var ctx *vmContext
	ctx, err = NewVmContextQuery(bs, cdb, contractAddress, contractState, contractState.SqlRecoveryPoint)
	if err != nil {
		return
	}
	ctx.origin = sender
	ctx.txHash = txHash
	ctx.curContract.amount = new(big.Int).SetBytes(amount)
	ctx.curContract.sender = sender
	if bi != nil {
		ctx.blockInfo = bi
	}

	allocContextSlot(ctx)
	defer freeContextSlot(ctx)

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

	ce.call(queryMaxInstLimit, nil)

	if ce.err != nil {
		return ce.err
	}
	if ce.jsonRet != "true" {
		return types.ErrNotAllowedFeeDelegation
	}
	return nil
}

func getCode(contractState *state.ContractState, bs *state.BlockState) ([]byte, error) {
	var code []byte
	var err error

	code = bs.GetCode(contractState.GetAccountID())
	if code != nil {
		return code, nil
	}
	code, err = contractState.GetCode()
	if err != nil {
		return nil, err
	}
	bs.AddCode(contractState.GetAccountID(), code)

	return code, nil
}

func getContract(contractState *state.ContractState, bs *state.BlockState) []byte {
	code, err := getCode(contractState, bs)
	if err != nil {
		return nil
	}
	return luacUtil.LuaCode(code).ByteCode()
}

func GetABI(contractState *state.ContractState, bs *state.BlockState) (*types.ABI, error) {
	var abi *types.ABI

	abi = bs.GetABI(contractState.GetAccountID())
	if abi != nil {
		return abi, nil
	}
	code, err := getCode(contractState, bs)
	if err != nil {
		return nil, err
	}
	luaCode := luacUtil.LuaCode(code)
	if luaCode.Len() == 0 {
		return nil, errors.New("cannot find contract")
	}
	rawAbi := luaCode.ABI()
	if len(rawAbi) == 0 {
		return nil, errors.New("cannot find abi")
	}
	abi = new(types.ABI)
	var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary
	if err = jsonIter.Unmarshal(rawAbi, abi); err != nil {
		return nil, err
	}
	bs.AddABI(contractState.GetAccountID(), abi)
	return abi, nil
}

func (re *recoveryEntry) recovery(bs *state.BlockState) error {
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
		if re.isDeploy {
			err := cs.ctrState.SetCode(nil)
			if err != nil {
				return newDbSystemError(err)
			}
			bs.RemoveCache(cs.ctrState.GetAccountID())
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

func Compile(code string, parent *LState) (luacUtil.LuaCode, error) {
	L := luacUtil.NewLState()
	if L == nil {
		return nil, ErrVmStart
	}
	defer luacUtil.CloseLState(L)
	if parent != nil {
		var lState = (*LState)(L)
		if cErrMsg := C.vm_copy_service(lState, parent); cErrMsg != nil {
			if C.luaL_hasuncatchablerror(lState) != C.int(0) {
				C.luaL_setuncatchablerror(parent)
			}
			errMsg := C.GoString(cErrMsg)
			return nil, errors.New(errMsg)
		}
		C.luaL_set_hardforkversion(lState, C.luaL_hardforkversion(parent))
		C.vm_set_timeout_hook(lState)
	}
	byteCodeAbi, err := luacUtil.Compile(L, code)
	if err != nil {
		if parent != nil && C.luaL_hasuncatchablerror((*LState)(L)) != C.int(0) {
			C.luaL_setuncatchablerror(parent)
		}
		return nil, err
	}
	return byteCodeAbi, nil
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
		chunkId = C.CString(enc.HexEncode(id))
	}
	defer C.free(unsafe.Pointer(chunkId))
	if cErrMsg := C.vm_loadbuff(
		ce.L,
		(*C.char)(unsafe.Pointer(&ce.code[0])),
		C.size_t(len(ce.code)),
		chunkId,
		ce.ctx.service-C.int(maxContext),
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		ce.err = errors.New(errMsg)
		ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(id)).Msg("failed to load code")
	}
}

func (ce *executor) vmLoadCall() {
	if cErrMsg := C.vm_loadcall(
		ce.L,
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		ce.err = errors.New(errMsg)
	}
	C.luaL_set_service(ce.L, ce.ctx.service)
}
