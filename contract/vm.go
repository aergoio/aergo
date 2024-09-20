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
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
	"github.com/aergoio/aergo/v2/blacklist"
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
	isMultiCall       bool
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

func NewVmContext(
	execCtx context.Context,
	blockState *state.BlockState,
	cdb ChainAccessor,
	sender, receiver *state.AccountState,
	contractState *statedb.ContractState,
	senderID,
	txHash []byte,
	bi *types.BlockHeaderInfo,
	node string,
	confirmed, query bool,
	rp uint64,
	executionMode int,
	amount *big.Int,
	gasLimit uint64,
	feeDelegation, isMultiCall bool,
) *vmContext {

	csReceiver := &callState{ctrState: contractState, accState: receiver}
	csSender := &callState{accState: sender}

	ctx := &vmContext{
		curContract:     newContractInfo(csReceiver, senderID, receiver.ID(), rp, amount),
		bs:              blockState,
		cdb:             cdb,
		origin:          senderID,
		txHash:          txHash,
		node:            node,
		confirmed:       confirmed,
		isQuery:         query,
		blockInfo:       bi,
		service:         C.int(executionMode),
		gasLimit:        gasLimit,
		remainedGas:     gasLimit,
		isFeeDelegation: feeDelegation,
		isMultiCall:     isMultiCall,
		execCtx:         execCtx,
	}

	// init call state
	ctx.callState = make(map[types.AccountID]*callState)
	ctx.callState[receiver.AccountID()] = csReceiver
	if sender != nil && sender != receiver {
		ctx.callState[sender.AccountID()] = csSender
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
	contractState *statedb.ContractState,
	rp uint64,
) (*vmContext, error) {
	cs := &callState{
		ctrState: contractState,
		accState: state.InitAccountState(contractState.GetID(), blockState.StateDB, contractState.State, contractState.State),
	}

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

func (ctx *vmContext) IsMultiCall() bool {
	return ctx.isMultiCall
}

func (ctx *vmContext) IsGasSystem() bool {
	return fee.GasEnabled(ctx.blockInfo.ForkVersion) && !ctx.isQuery
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
	return fee.TxExecuteFee(ctx.blockInfo.ForkVersion, ctx.bs.GasPrice, ctx.usedGas(), ctx.dbUpdateTotalSize)
}

func (ctx *vmContext) usedGas() uint64 {
	if fee.IsZeroFee() || !ctx.IsGasSystem() {
		return 0
	}
	return ctx.gasLimit - ctx.remainedGas
}

func (ctx *vmContext) addUpdateSize(updateSize int64) error {
	if ctx.IsGasSystem() {
		return nil
	}
	if ctx.dbUpdateTotalSize+updateSize > dbUpdateMaxLimit {
		return errors.New("exceeded size of updates in the state database")
	}
	ctx.dbUpdateTotalSize += updateSize
	return nil
}

func resolveFunction(contractState *statedb.ContractState, bs *state.BlockState, name string, constructor bool) (*types.Function, error) {
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
	ctrState *statedb.ContractState,
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

	if blacklist.Check(types.EncodeAddress(contractId)) {
		ce := &executor{
			code: contract,
			ctx:  ctx,
		}
		ce.err = fmt.Errorf("contract not available")
		ctrLgr.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("blocked contract")
		return ce
	}

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

func (ce *executor) commitCalledContract() error {
	ctx := ce.ctx

	if ctx == nil || ctx.callState == nil {
		return nil
	}

	bs := ctx.bs
	rootContract := ctx.curContract.callState.ctrState

	var err error
	for _, v := range ctx.callState {
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
			err = statedb.StageContractState(v.ctrState, bs.StateDB)
			if err != nil {
				return newDbSystemError(err)
			}
		}
		/* Put account state only for callback */
		if v.isCallback {
			err = v.accState.PutState()
			if err != nil {
				return newDbSystemError(err)
			}
		}

	}

	if ctx.traceFile != nil {
		_, _ = ce.ctx.traceFile.WriteString("[Put State Balance]\n")
		for k, v := range ctx.callState {
			_, _ = ce.ctx.traceFile.WriteString(fmt.Sprintf("%s : nonce=%d ammount=%s\n",
				k.String(), v.accState.Nonce(), v.accState.Balance().String()))
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
		// remove code cache in block ( revert new deploy )
		if v.isDeploy && len(v.accState.CodeHash()) != 0 {
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
		ce.ctx.service-C.int(maxContext),
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

func getMultiCallInfo(ci *types.CallInfo, payload []byte) error {
	payload = append([]byte{'['}, payload...)
	payload = append(payload, ']')
	ci.Name = "execute"
	return getCallInfo(&ci.Args, payload, []byte("multicall"))
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

type specialTxn struct {
	newAccount types.AccountID
	amount     []byte
	usedGas    uint64
}

var specialTxns = map[[32]byte]specialTxn{
	{129, 234, 154, 23, 151, 13, 246, 68, 56, 193, 192, 55, 227, 81, 82, 222, 19, 28, 228, 121, 236, 133, 219, 108, 71, 179, 142, 120, 88, 174, 222, 81}: {
			newAccount: types.AccountID{0xad, 0x4b, 0x85, 0x8e, 0xda, 0xb4, 0x75, 0xbd, 0x28, 0x71, 0x18, 0x36, 0xff, 0x89, 0x0a, 0xaa, 0x72, 0x06, 0x24, 0x5b, 0x24, 0x9c, 0xbe, 0x88, 0x6d, 0x62, 0x9e, 0xf4, 0x65, 0x4c, 0x12, 0xfa},
			amount:     []byte{0x9e, 0x5d, 0x65, 0xfb, 0x24, 0xdb, 0x40, 0x00, 0x00, 0x00},
			usedGas:    459555,
	},
	{193, 110, 30, 216, 15, 146, 122, 227, 51, 13, 93, 112, 219, 98, 242, 236, 99, 37, 214, 92, 111, 56, 52, 129, 14, 178, 60, 75, 1, 213, 83, 182}: {
			newAccount: types.AccountID{0x11, 0xd8, 0x3c, 0xc8, 0xd5, 0x9e, 0xd8, 0xa6, 0x78, 0xd3, 0x3f, 0xa3, 0x88, 0x72, 0xbf, 0xd4, 0x01, 0x06, 0xc0, 0xd0, 0x94, 0x03, 0x34, 0xb0, 0x30, 0x7c, 0xa4, 0x58, 0x60, 0xe9, 0xf9, 0x09},
			amount:     []byte{0x12, 0x01, 0xb5, 0x15, 0xd0, 0x12, 0x5e, 0x00, 0x00, 0x00},
			usedGas:    466263,
	},
	{184, 128, 171, 123, 34, 86, 139, 235, 113, 109, 186, 164, 192, 196, 23, 226, 252, 243, 186, 222, 223, 189, 116, 146, 109, 69, 14, 16, 122, 42, 178, 9}: {
			newAccount: types.AccountID{0x78, 0x04, 0x87, 0xd8, 0xc1, 0x13, 0xfa, 0xcf, 0x1c, 0x3a, 0x69, 0x4f, 0xe9, 0xcd, 0x72, 0x00, 0x4d, 0xe6, 0x72, 0x4e, 0x03, 0xa8, 0xcb, 0xa9, 0x27, 0x37, 0x4b, 0x4d, 0xf2, 0xe9, 0xb7, 0x71},
			amount:     []byte{0x11, 0xfc, 0x7a, 0xd2, 0x86, 0xfd, 0x89, 0x00, 0x00, 0x00},
			usedGas:    455352,
	},
	{43, 239, 8, 227, 4, 189, 119, 144, 230, 133, 133, 197, 166, 209, 239, 20, 23, 60, 152, 252, 237, 213, 1, 54, 96, 115, 209, 67, 198, 200, 147, 212}: {
			newAccount: types.AccountID{0x25, 0x89, 0xc5, 0x19, 0xbb, 0x59, 0xdf, 0x80, 0x44, 0xd6, 0x2a, 0x5c, 0xeb, 0x83, 0x20, 0xcf, 0xb4, 0x58, 0xb1, 0x19, 0x85, 0x43, 0x4f, 0x66, 0x62, 0x50, 0xf0, 0xcc, 0xac, 0xc1, 0x4f, 0xd9},
			amount:     []byte{0x94, 0x81, 0xaf, 0x01, 0xac, 0x37, 0x08, 0x00, 0x00, 0x00},
			usedGas:    429265,
	},
	{13, 204, 23, 88, 64, 177, 202, 5, 45, 224, 104, 126, 53, 146, 204, 43, 42, 43, 159, 111, 42, 231, 138, 230, 95, 173, 102, 129, 217, 151, 192, 172}: {
			newAccount: types.AccountID{0x4d, 0xd9, 0x4d, 0x5d, 0x00, 0xbe, 0x02, 0xb2, 0x99, 0x35, 0x42, 0xf6, 0x3b, 0x52, 0x2d, 0xad, 0xe2, 0x21, 0x37, 0xb2, 0x89, 0x07, 0xcd, 0x68, 0x8f, 0x29, 0x07, 0x0a, 0x3b, 0xa7, 0xc9, 0x67},
			amount:     []byte{0x17, 0xed, 0x38, 0xcb, 0xa7, 0x34, 0x31, 0x00, 0x00, 0x00},
			usedGas:    445449,
	},
	{158, 42, 176, 20, 157, 122, 160, 96, 53, 110, 21, 91, 247, 6, 72, 217, 0, 246, 121, 227, 89, 126, 100, 164, 142, 117, 225, 182, 32, 78, 233, 65}: {
			newAccount: types.AccountID{0xc1, 0xb8, 0xd2, 0x09, 0x2e, 0x49, 0x81, 0xe2, 0xad, 0x6c, 0x41, 0x0a, 0x47, 0xa3, 0x82, 0x52, 0xc9, 0xca, 0xfc, 0xfa, 0x39, 0x92, 0x80, 0x6f, 0x2b, 0x09, 0xb9, 0x6e, 0x0c, 0x68, 0x69, 0xe0},
			amount:     []byte{0x68, 0x3a, 0x50, 0x54, 0x70, 0xa6, 0x8c, 0x00, 0x00, 0x00, 0x00},
			usedGas:    423622,
	},
	{168, 176, 216, 231, 63, 194, 129, 104, 144, 3, 231, 151, 43, 104, 61, 245, 235, 176, 111, 69, 203, 138, 207, 104, 45, 17, 191, 187, 117, 189, 155, 138}: {
			newAccount: types.AccountID{0x0f, 0x10, 0x64, 0x5d, 0xe1, 0x1b, 0xf8, 0x01, 0xf6, 0xab, 0x98, 0x5c, 0x78, 0x4d, 0x49, 0x89, 0x4f, 0x0e, 0x7a, 0xc4, 0x67, 0x8f, 0xdd, 0xf5, 0x4b, 0xed, 0x98, 0xff, 0x4b, 0xea, 0xaa, 0xc0},
			amount:     []byte{0x7b, 0x07, 0xc7, 0xe3, 0x98, 0x53, 0x84, 0x00, 0x00, 0x00, 0x00},
			usedGas:    467981,
	},
	{202, 56, 0, 47, 108, 198, 163, 231, 225, 0, 215, 20, 198, 145, 145, 110, 145, 238, 17, 157, 225, 76, 81, 225, 215, 253, 87, 139, 242, 173, 163, 63}: {
			newAccount: types.AccountID{0x70, 0x64, 0x0f, 0x87, 0x36, 0xea, 0xd6, 0x39, 0x6a, 0xfa, 0x8b, 0x8e, 0x87, 0xdc, 0x3e, 0xc8, 0x8b, 0xe4, 0x75, 0x36, 0xd6, 0xc3, 0xf3, 0xe3, 0xef, 0x66, 0xc3, 0x0b, 0x26, 0x3d, 0x75, 0x9d},
			amount:     []byte{0xdc, 0x8a, 0x1c, 0x7e, 0x07, 0x4c, 0xb0, 0x00, 0x00, 0x00},
			usedGas:    427595,
	},
	{245, 191, 35, 66, 249, 205, 248, 23, 226, 73, 200, 131, 194, 199, 118, 251, 177, 146, 74, 84, 25, 176, 35, 214, 232, 97, 213, 73, 118, 179, 1, 227}: {
			newAccount: types.AccountID{0x50, 0x36, 0x96, 0xe8, 0x09, 0xb7, 0x9b, 0x42, 0xe0, 0xa1, 0x1a, 0x6f, 0x41, 0x41, 0x21, 0x61, 0x91, 0x66, 0xb0, 0xd9, 0x84, 0x0f, 0x02, 0x9f, 0xf8, 0xf5, 0x3c, 0x79, 0xde, 0xc6, 0xea, 0x54},
			amount:     []byte{0x83, 0xad, 0xc2, 0x48, 0xba, 0x57, 0x10, 0x00, 0x00, 0x00, 0x00},
			usedGas:    433716,
	},
	{225, 226, 128, 139, 56, 60, 5, 183, 199, 53, 234, 75, 55, 4, 106, 208, 245, 216, 62, 246, 232, 253, 104, 191, 191, 226, 221, 38, 103, 102, 66, 186}: {
			newAccount: types.AccountID{0xec, 0xf1, 0x3d, 0x5e, 0xbf, 0xa8, 0x0f, 0x26, 0x64, 0xc8, 0x70, 0x65, 0x8f, 0x05, 0xce, 0x0e, 0xe7, 0xf4, 0xa3, 0x42, 0x95, 0xbe, 0x57, 0x7b, 0xd5, 0x9a, 0x97, 0x41, 0x04, 0xfe, 0x91, 0x63},
			amount:     []byte{0xd9, 0xcb, 0xfc, 0xe9, 0x42, 0x5a, 0x90, 0x00, 0x00, 0x00},
			usedGas:    423567,
	},
	{162, 43, 3, 56, 188, 10, 24, 40, 206, 141, 6, 103, 26, 5, 251, 144, 9, 19, 103, 79, 83, 236, 68, 104, 151, 98, 47, 187, 159, 207, 234, 202}: {
			newAccount: types.AccountID{0xd2, 0x0a, 0x23, 0xc5, 0xc8, 0x23, 0x9c, 0x2d, 0x40, 0xce, 0x2a, 0xf2, 0x76, 0xf9, 0x4d, 0xac, 0x3d, 0x2b, 0x53, 0x5a, 0x59, 0xe4, 0x1b, 0x14, 0x2d, 0x50, 0x05, 0x70, 0x91, 0x67, 0x02, 0x78},
			amount:     []byte{0x8b, 0xca, 0x9e, 0x4c, 0x4e, 0x77, 0x28, 0x00, 0x00, 0x00},
			usedGas:    423788,
	},
	{255, 64, 144, 67, 86, 6, 83, 8, 142, 249, 135, 70, 158, 254, 206, 105, 173, 127, 26, 152, 50, 249, 78, 181, 206, 156, 72, 6, 232, 245, 158, 75}: {
			newAccount: types.AccountID{0xc1, 0xab, 0x0a, 0x6d, 0x7e, 0x1a, 0x7c, 0x1f, 0x5e, 0xe6, 0x68, 0xc5, 0x45, 0x2c, 0xa1, 0x8f, 0xe6, 0xe4, 0xb6, 0xbf, 0xc5, 0x28, 0xb3, 0x0c, 0xd8, 0x59, 0x40, 0x76, 0x7c, 0xd4, 0xbf, 0xac},
			amount:     []byte{0x9f, 0xfb, 0x62, 0x58, 0x2a, 0xf0, 0x50, 0x00, 0x00, 0x00},
			usedGas:    452333,
	},
	{192, 199, 32, 196, 126, 192, 160, 137, 199, 191, 59, 166, 70, 99, 115, 233, 251, 102, 116, 132, 206, 153, 184, 239, 124, 48, 127, 81, 211, 105, 56, 58}: {
			newAccount: types.AccountID{0x9a, 0xb9, 0x43, 0xae, 0xc1, 0x42, 0xbf, 0xe1, 0x0b, 0x49, 0x48, 0x2d, 0x99, 0xa9, 0x67, 0x06, 0x6f, 0x04, 0x28, 0xd8, 0xd1, 0xbe, 0xfd, 0x1a, 0x96, 0xe8, 0x58, 0xa3, 0x2f, 0x0b, 0xd1, 0xea},
			amount:     []byte{0x71, 0x89, 0x82, 0xa3, 0x76, 0x38, 0xac, 0x00, 0x00, 0x00},
			usedGas:    433978,
	},
	{191, 204, 57, 61, 215, 0, 40, 20, 248, 158, 203, 69, 93, 255, 88, 116, 251, 248, 11, 3, 227, 200, 155, 163, 135, 49, 31, 216, 225, 175, 79, 208}: {
			newAccount: types.AccountID{0x64, 0x5d, 0x05, 0xbd, 0x8f, 0xd9, 0x35, 0xe5, 0x9f, 0x84, 0xdb, 0xf2, 0x28, 0x8b, 0x14, 0x2f, 0xe1, 0x05, 0xc1, 0x50, 0xb0, 0x97, 0x53, 0xf9, 0xd8, 0x03, 0x79, 0x5e, 0x59, 0x5c, 0x82, 0xde},
			amount:     []byte{0x58, 0x63, 0xbc, 0xe6, 0xfc, 0x28, 0x58, 0x00, 0x00, 0x00},
			usedGas:    425657,
	},
	{156, 202, 165, 40, 128, 218, 153, 0, 241, 182, 108, 8, 216, 249, 52, 227, 153, 210, 130, 114, 67, 9, 145, 7, 206, 243, 174, 78, 21, 248, 162, 65}: {
			newAccount: types.AccountID{0x3a, 0xf3, 0x34, 0xdd, 0x82, 0x40, 0x76, 0x4a, 0x80, 0xcb, 0xb3, 0xe9, 0xd6, 0x26, 0xbe, 0x45, 0x66, 0x84, 0x32, 0x93, 0x9c, 0xee, 0xdc, 0x5d, 0xfa, 0x02, 0x8a, 0x66, 0x21, 0xb5, 0x26, 0x8e},
			amount:     []byte{0x8b, 0x61, 0x28, 0x08, 0x88, 0xcb, 0xa0, 0x00, 0x00, 0x00},
			usedGas:    440491,
	},
	{205, 86, 207, 51, 229, 35, 228, 200, 128, 169, 27, 55, 105, 224, 93, 163, 68, 170, 94, 163, 124, 131, 162, 6, 194, 181, 52, 24, 224, 160, 203, 115}: {
			newAccount: types.AccountID{0x9f, 0xc6, 0xa1, 0x63, 0x0d, 0x2f, 0x36, 0x42, 0xdc, 0xd1, 0x89, 0xd3, 0x45, 0x47, 0x0f, 0x9b, 0x0c, 0xf5, 0x77, 0x63, 0x73, 0x1f, 0x8c, 0x01, 0x4c, 0xa8, 0x92, 0xa9, 0x19, 0x94, 0xe7, 0xcf},
			amount:     []byte{0xdd, 0xaf, 0xfe, 0x7a, 0x5e, 0xa8, 0x30, 0x00, 0x00, 0x00},
			usedGas:    424100,
	},
	{158, 22, 113, 82, 4, 50, 178, 151, 152, 248, 251, 12, 47, 2, 83, 121, 135, 37, 200, 175, 37, 20, 245, 249, 4, 22, 161, 21, 67, 203, 137, 98}: {
			newAccount: types.AccountID{0x9f, 0xc6, 0xa1, 0x63, 0x0d, 0x2f, 0x36, 0x42, 0xdc, 0xd1, 0x89, 0xd3, 0x45, 0x47, 0x0f, 0x9b, 0x0c, 0xf5, 0x77, 0x63, 0x73, 0x1f, 0x8c, 0x01, 0x4c, 0xa8, 0x92, 0xa9, 0x19, 0x94, 0xe7, 0xcf},
			amount:     []byte{0xdd, 0xaf, 0xfe, 0x7a, 0x5e, 0xa8, 0x30, 0x00, 0x00, 0x00},
			usedGas:    455272,
	},
	{148, 124, 97, 4, 121, 165, 180, 245, 160, 13, 123, 216, 83, 177, 202, 3, 167, 40, 46, 173, 187, 202, 84, 126, 146, 176, 10, 153, 201, 246, 42, 41}: {
			newAccount: types.AccountID{0xe1, 0x7a, 0xc6, 0xb0, 0xc9, 0x2a, 0xe5, 0xce, 0xf3, 0x97, 0x9d, 0x25, 0xf4, 0x6b, 0x14, 0x1d, 0x8d, 0xc0, 0x2d, 0xd5, 0x5f, 0xaf, 0x03, 0x17, 0xa3, 0xaa, 0xaf, 0x7a, 0x5d, 0x6e, 0x05, 0x3e},
			amount:     []byte{0x8c, 0x77, 0xcd, 0x1c, 0x96, 0x6e, 0x70, 0x00, 0x00, 0x00},
			usedGas:    468320,
	},
	{15, 163, 7, 104, 16, 126, 231, 59, 9, 237, 64, 41, 197, 57, 58, 134, 45, 242, 164, 9, 186, 39, 39, 52, 215, 137, 155, 220, 249, 234, 106, 92}: {
			newAccount: types.AccountID{0x96, 0xbc, 0x9f, 0x5f, 0x4e, 0x91, 0x10, 0x27, 0xd8, 0x86, 0xf8, 0x1f, 0x60, 0x09, 0xa1, 0x54, 0xa2, 0x4e, 0x8f, 0xee, 0x28, 0x79, 0xfb, 0xf0, 0xec, 0xa0, 0xec, 0x59, 0x4e, 0xae, 0xc4, 0x65},
			amount:     []byte{0xcd, 0xa8, 0xa2, 0x7b, 0x32, 0xa4, 0xc8, 0x00, 0x00, 0x00},
			usedGas:    461638,
	},
	{139, 205, 222, 222, 17, 161, 215, 209, 39, 64, 108, 248, 55, 166, 228, 231, 72, 166, 88, 228, 125, 94, 215, 130, 158, 54, 227, 57, 247, 182, 76, 18}: {
			newAccount: types.AccountID{0xc5, 0x8e, 0x0f, 0x31, 0xab, 0x92, 0x4e, 0xf4, 0xce, 0x5f, 0xc7, 0x5c, 0xcc, 0xa4, 0x7f, 0xcf, 0xe9, 0xa1, 0xb5, 0xd6, 0xf0, 0xfe, 0x9f, 0xce, 0xea, 0xd8, 0x2a, 0x2a, 0xf1, 0xb0, 0x4b, 0x40},
			amount:     []byte{0xcb, 0x9d, 0x12, 0xaa, 0xf5, 0xf9, 0x40, 0x00, 0x00, 0x00},
			usedGas:    458325,
	},
	{193, 212, 47, 16, 53, 7, 139, 6, 98, 72, 216, 146, 249, 177, 201, 204, 98, 90, 37, 31, 203, 225, 189, 103, 136, 60, 118, 213, 82, 159, 215, 250}: {
			newAccount: types.AccountID{0x2a, 0x16, 0xbb, 0x7f, 0x44, 0x78, 0xaf, 0x5d, 0x28, 0x2a, 0x6f, 0x2f, 0x93, 0x0e, 0x2a, 0x8c, 0xeb, 0x8f, 0x9f, 0x82, 0xae, 0x78, 0x3e, 0x67, 0x08, 0x63, 0xf1, 0xdf, 0x97, 0xeb, 0x96, 0xe6},
			amount:     []byte{0xd8, 0xf3, 0x87, 0x99, 0x2c, 0xe8, 0xe0, 0x00, 0x00, 0x00},
			usedGas:    445256,
	},
	{90, 156, 215, 114, 92, 174, 10, 79, 62, 110, 154, 23, 127, 232, 153, 218, 221, 203, 242, 48, 252, 11, 226, 177, 53, 121, 29, 131, 178, 69, 171, 128}: {
			newAccount: types.AccountID{0x9b, 0x8e, 0x11, 0x8a, 0x31, 0x46, 0xad, 0x10, 0x9b, 0x38, 0xb2, 0x3d, 0x6b, 0xd6, 0x94, 0x7f, 0xf2, 0xed, 0x04, 0xe1, 0x04, 0x01, 0x4e, 0x00, 0xb8, 0x00, 0x4b, 0x01, 0x10, 0x1a, 0xeb, 0x28},
			amount:     []byte{0xbc, 0xdf, 0x2b, 0x66, 0x5f, 0x91, 0x70, 0x00, 0x00, 0x00},
			usedGas:    441488,
	},
	{23, 10, 63, 179, 71, 93, 24, 4, 21, 56, 57, 19, 138, 242, 76, 211, 131, 55, 135, 214, 117, 227, 253, 222, 45, 69, 74, 177, 34, 49, 180, 230}: {
			newAccount: types.AccountID{0x91, 0x0a, 0x43, 0x0b, 0x32, 0xa1, 0x2f, 0x95, 0xc6, 0x78, 0x48, 0x92, 0x9b, 0xd4, 0xec, 0x9c, 0x36, 0xb1, 0xd7, 0x6a, 0xb7, 0x87, 0xb4, 0x19, 0x47, 0x77, 0xe3, 0xfd, 0x9d, 0x9d, 0x03, 0x81},
			amount:     []byte{0x80, 0xa3, 0x2f, 0xe4, 0x17, 0x27, 0x00, 0x00, 0x00, 0x00},
			usedGas:    458393,
	},
	{141, 163, 213, 1, 40, 119, 140, 118, 157, 177, 253, 190, 122, 131, 14, 8, 95, 74, 144, 18, 227, 139, 107, 199, 6, 133, 208, 177, 122, 29, 12, 230}: {
			newAccount: types.AccountID{0x14, 0xd7, 0x17, 0x2f, 0xc9, 0x21, 0xbc, 0xc7, 0x3a, 0x58, 0xcb, 0x2e, 0x66, 0x67, 0x2b, 0x0a, 0x73, 0x56, 0xac, 0xca, 0x5d, 0x1b, 0x4d, 0xf8, 0xec, 0x15, 0xfd, 0x4d, 0x86, 0x33, 0xbc, 0x52},
			amount:     []byte{0x60, 0x6d, 0xe2, 0xcc, 0x5d, 0xfa, 0xd4, 0x00, 0x00, 0x00},
			usedGas:    433751,
	},
	{100, 48, 0, 36, 75, 146, 55, 36, 97, 27, 78, 154, 120, 112, 130, 10, 206, 4, 169, 98, 64, 213, 195, 178, 37, 31, 10, 227, 177, 192, 101, 207}: {
			newAccount: types.AccountID{0x49, 0xbf, 0xb9, 0x7e, 0x47, 0xd3, 0x11, 0x74, 0x1c, 0x7a, 0x7e, 0xcf, 0x71, 0xbc, 0xaf, 0xc2, 0x1a, 0x39, 0xbd, 0x8c, 0x5f, 0x24, 0x5b, 0xa5, 0x80, 0xe6, 0x9e, 0xf7, 0x95, 0x77, 0xa7, 0x8f},
			amount:     []byte{0x65, 0xf3, 0x3d, 0x6d, 0xcd, 0xc4, 0xec, 0x00, 0x00, 0x00},
			usedGas:    423845,
	},
	{34, 160, 32, 215, 131, 168, 207, 143, 223, 16, 122, 236, 254, 117, 89, 175, 132, 189, 161, 139, 17, 39, 179, 36, 126, 41, 29, 91, 56, 25, 157, 209}: {
			newAccount: types.AccountID{0xfb, 0xc7, 0x45, 0xcd, 0xfa, 0xcc, 0x9e, 0x05, 0x44, 0x24, 0x81, 0x80, 0xb7, 0x9a, 0x84, 0x37, 0xa5, 0xc7, 0xc3, 0xac, 0xae, 0x08, 0xc5, 0x66, 0x18, 0x41, 0x53, 0xaf, 0x55, 0x24, 0x6c, 0xc4},
			amount:     []byte{0xcd, 0x24, 0x39, 0x26, 0xdf, 0xe1, 0x90, 0x00, 0x00, 0x00},
			usedGas:    432877,
	},
	{35, 129, 46, 56, 122, 118, 146, 5, 72, 75, 166, 66, 249, 65, 208, 80, 115, 125, 129, 95, 155, 9, 41, 25, 107, 62, 140, 11, 150, 170, 34, 94}: {
			newAccount: types.AccountID{0xba, 0x20, 0x8e, 0x17, 0x34, 0xea, 0xf7, 0x53, 0x36, 0x59, 0x9b, 0x1f, 0xaa, 0x95, 0x55, 0x7c, 0xcf, 0xac, 0xfb, 0x37, 0xfc, 0xa8, 0x08, 0xab, 0xae, 0x35, 0x3b, 0x84, 0x84, 0xab, 0xe4, 0x12},
			amount:     []byte{0x51, 0x5a, 0xb0, 0xb8, 0x6b, 0x4d, 0x04, 0x00, 0x00, 0x00, 0x00},
			usedGas:    460381,
	},
	{147, 243, 10, 204, 212, 249, 63, 202, 141, 192, 140, 94, 142, 65, 8, 112, 194, 234, 247, 109, 18, 254, 225, 158, 212, 100, 234, 50, 34, 177, 212, 64}: {
			newAccount: types.AccountID{0xef, 0xcc, 0x06, 0x56, 0x0e, 0x63, 0x7c, 0xc4, 0x47, 0xb7, 0xfe, 0xdd, 0x13, 0x06, 0x70, 0x5c, 0x0a, 0x57, 0x8f, 0x3d, 0xc7, 0x3c, 0x25, 0xa3, 0x92, 0x76, 0xaf, 0xd4, 0xcd, 0xab, 0x6d, 0xc5},
			amount:     []byte{0x39, 0xdd, 0x6b, 0x30, 0x66, 0x92, 0x84, 0x00, 0x00, 0x00, 0x00},
			usedGas:    462894,
	},
	{201, 156, 154, 94, 114, 27, 50, 158, 195, 115, 100, 47, 63, 120, 129, 161, 239, 9, 12, 112, 52, 115, 149, 150, 13, 18, 231, 60, 169, 11, 74, 141}: {
			newAccount: types.AccountID{0xd9, 0xaa, 0xdb, 0x4f, 0xad, 0xa4, 0x10, 0x2e, 0x03, 0xc0, 0x97, 0x3b, 0xde, 0xc8, 0x84, 0x51, 0xa8, 0x5e, 0xe1, 0xb1, 0x20, 0xab, 0x03, 0xa8, 0x9a, 0x97, 0x20, 0xc1, 0x93, 0x92, 0x5f, 0x34},
			amount:     []byte{0x97, 0x56, 0x98, 0xf7, 0x3d, 0x99, 0xf8, 0x00, 0x00, 0x00, 0x00},
			usedGas:    461758,
	},
	{241, 243, 200, 138, 132, 15, 174, 85, 140, 63, 251, 237, 29, 85, 62, 132, 94, 30, 62, 204, 83, 24, 248, 163, 59, 223, 86, 213, 154, 140, 170, 115}: {
			newAccount: types.AccountID{0x70, 0x0e, 0xca, 0x8a, 0xb7, 0xe1, 0xc9, 0xa7, 0x2a, 0x3c, 0x91, 0xe7, 0x76, 0x08, 0xe1, 0x95, 0xd8, 0xcf, 0xb0, 0xae, 0xcf, 0x8b, 0x9e, 0xac, 0x3c, 0xa7, 0xd3, 0x11, 0x76, 0x50, 0xc6, 0xe6},
			amount:     []byte{0x96, 0x9c, 0x34, 0xf5, 0x94, 0x0f, 0xb8, 0x00, 0x00, 0x00, 0x00},
			usedGas:    461319,
	},
	{228, 187, 147, 128, 57, 9, 229, 167, 13, 98, 179, 118, 38, 241, 214, 69, 97, 198, 111, 171, 27, 46, 229, 100, 153, 243, 119, 43, 52, 138, 15, 203}: {
			newAccount: types.AccountID{0xcc, 0x54, 0x8f, 0x39, 0x43, 0x4b, 0x0d, 0x66, 0x6e, 0x0c, 0xe3, 0xeb, 0x36, 0x0e, 0x13, 0xb3, 0x36, 0xf5, 0x15, 0xef, 0x6b, 0xbf, 0x2b, 0x77, 0x1f, 0xfd, 0x46, 0x57, 0x7a, 0x5c, 0x5c, 0x83},
			amount:     []byte{0x97, 0xa3, 0x4f, 0xd2, 0x82, 0x0d, 0x98, 0x00, 0x00, 0x00},
			usedGas:    448601,
	},
	{161, 81, 158, 10, 1, 9, 78, 179, 115, 96, 38, 227, 255, 223, 19, 237, 86, 46, 3, 178, 115, 77, 191, 128, 77, 170, 64, 166, 149, 165, 217, 46}: {
			newAccount: types.AccountID{0x42, 0xe1, 0xad, 0xa7, 0x92, 0x8d, 0xd4, 0x5a, 0x7b, 0x53, 0x4b, 0x81, 0xa7, 0x8b, 0x66, 0xd5, 0x09, 0xe6, 0x2e, 0xe9, 0xa5, 0xa9, 0x3e, 0xac, 0xe2, 0x0e, 0x15, 0x5a, 0xec, 0xc6, 0xae, 0xa1},
			amount:     []byte{0x73, 0xed, 0xf1, 0xf6, 0x5c, 0x53, 0xa0, 0x00, 0x00, 0x00},
			usedGas:    469172,
	},
	{211, 47, 58, 225, 116, 11, 211, 248, 47, 203, 139, 32, 170, 187, 215, 177, 44, 220, 107, 228, 75, 40, 66, 203, 69, 64, 85, 161, 85, 157, 155, 62}: {
			newAccount: types.AccountID{0xe2, 0xae, 0xe7, 0xb3, 0x15, 0xed, 0x4a, 0x2e, 0x94, 0xdd, 0xd0, 0x78, 0xdb, 0x9a, 0x9b, 0x6e, 0x41, 0xfd, 0x00, 0x24, 0x2b, 0x3c, 0xc8, 0x39, 0x21, 0x3b, 0x01, 0x30, 0x58, 0x41, 0xac, 0x0b},
			amount:     []byte{0xb5, 0x01, 0x40, 0x87, 0x9b, 0xde, 0x98, 0x00, 0x00, 0x00},
			usedGas:    434022,
	},
	{190, 95, 117, 79, 98, 34, 190, 125, 183, 171, 43, 113, 27, 180, 103, 8, 183, 207, 62, 7, 70, 52, 110, 145, 26, 24, 111, 243, 201, 148, 111, 181}: {
			newAccount: types.AccountID{0x7d, 0xd9, 0xab, 0x21, 0xd3, 0x0d, 0x08, 0xae, 0x32, 0x6b, 0x8d, 0x09, 0x5f, 0x30, 0xa5, 0x9e, 0x4a, 0xf6, 0xa8, 0xa8, 0xf0, 0xbe, 0x27, 0x74, 0x47, 0x81, 0x20, 0x7c, 0x2e, 0x3a, 0x4d, 0xe6},
			amount:     []byte{0xcb, 0x11, 0x06, 0xd1, 0x8c, 0x67, 0x20, 0x00, 0x00, 0x00},
			usedGas:    445276,
	},
	{81, 86, 145, 236, 173, 19, 33, 53, 172, 26, 234, 101, 180, 85, 119, 88, 169, 44, 117, 131, 131, 25, 208, 222, 181, 101, 145, 144, 146, 95, 176, 187}: {
			newAccount: types.AccountID{0x78, 0x9b, 0xb3, 0x38, 0xc3, 0xe5, 0xe0, 0x87, 0x64, 0x54, 0xe0, 0xf4, 0x41, 0x6e, 0x94, 0x22, 0x84, 0xe9, 0x7d, 0xfa, 0x09, 0xce, 0x72, 0x4c, 0x13, 0x28, 0x00, 0xfd, 0x9e, 0xf6, 0xb5, 0xd0},
			amount:     []byte{0x5a, 0x60, 0xb0, 0xca, 0x69, 0xa5, 0x90, 0x00, 0x00, 0x00},
			usedGas:    461294,
	},
	{236, 51, 45, 81, 235, 114, 145, 82, 15, 156, 184, 99, 215, 249, 141, 112, 109, 72, 201, 253, 62, 242, 47, 61, 242, 148, 116, 82, 52, 4, 232, 223}: {
			newAccount: types.AccountID{0x33, 0xba, 0xc5, 0x4f, 0xe3, 0xbe, 0x9c, 0xf9, 0xe0, 0xa1, 0x35, 0xf8, 0x87, 0x3e, 0x28, 0x62, 0xb0, 0xbf, 0xe0, 0xb4, 0x68, 0x19, 0xf3, 0x4d, 0x9d, 0x4f, 0xaa, 0xf1, 0x99, 0x75, 0x5b, 0x8f},
			amount:     []byte{0xcb, 0xd6, 0x10, 0x1b, 0xcc, 0x0d, 0xb8, 0x00, 0x00, 0x00},
			usedGas:    439542,
	},
	{233, 214, 78, 117, 180, 96, 219, 53, 131, 160, 178, 202, 210, 199, 5, 96, 113, 105, 109, 70, 158, 163, 122, 14, 233, 147, 122, 6, 130, 175, 116, 116}: {
			newAccount: types.AccountID{0xd9, 0x04, 0xdb, 0xe3, 0x1a, 0xf0, 0x9d, 0xaf, 0x48, 0xe0, 0xd6, 0x7f, 0xae, 0x70, 0x03, 0xfa, 0x0c, 0xe6, 0x4f, 0x42, 0x1c, 0x05, 0xfe, 0x36, 0xa7, 0x93, 0x07, 0xd7, 0xe2, 0xb0, 0x97, 0xb2},
			amount:     []byte{0xda, 0xd4, 0x21, 0x8a, 0xe0, 0x51, 0x90, 0x00, 0x00, 0x00},
			usedGas:    443902,
	},
	{13, 49, 235, 91, 214, 170, 75, 127, 220, 96, 249, 50, 235, 125, 206, 146, 76, 45, 99, 94, 73, 142, 118, 135, 48, 240, 98, 160, 147, 25, 41, 79}: {
			newAccount: types.AccountID{0x81, 0xb3, 0x2a, 0x9f, 0xb9, 0x55, 0x14, 0x8a, 0x2d, 0xf4, 0x06, 0x04, 0xc3, 0xfb, 0xbf, 0xc7, 0xd9, 0x50, 0x6b, 0x8c, 0xaf, 0xed, 0x42, 0x13, 0x8f, 0x7a, 0x12, 0x8a, 0x2e, 0x93, 0xa0, 0xe5},
			amount:     []byte{0xbd, 0xd7, 0x1c, 0xee, 0x8f, 0xcd, 0x50, 0x00, 0x00, 0x00},
			usedGas:    431953,
	},
	{71, 209, 240, 175, 171, 87, 182, 203, 31, 148, 98, 250, 39, 209, 85, 182, 137, 174, 63, 203, 9, 128, 6, 59, 52, 183, 166, 221, 17, 122, 179, 159}: {
			newAccount: types.AccountID{0xae, 0xf9, 0xf6, 0x54, 0x28, 0xd6, 0x5d, 0x17, 0x85, 0xa8, 0x18, 0x84, 0x21, 0xb9, 0xbd, 0xe1, 0x08, 0x9d, 0x75, 0xb7, 0x02, 0x49, 0xd1, 0x93, 0x8e, 0xb2, 0x5c, 0x41, 0x15, 0xf2, 0x19, 0x1b},
			amount:     []byte{0xc4, 0x3d, 0x2d, 0xc4, 0xbd, 0x61, 0x68, 0x00, 0x00, 0x00},
			usedGas:    449916,
	},
	{83, 55, 204, 8, 215, 239, 109, 185, 40, 122, 238, 160, 226, 20, 234, 59, 218, 98, 61, 157, 9, 50, 203, 37, 104, 45, 89, 145, 16, 120, 208, 111}: {
			newAccount: types.AccountID{0xaf, 0x12, 0x0d, 0x04, 0xb7, 0x9c, 0xb9, 0x4e, 0x29, 0x36, 0x7e, 0x92, 0xaa, 0xa7, 0xc4, 0x9b, 0xcb, 0x46, 0x75, 0x8b, 0x4f, 0x91, 0x9b, 0x90, 0xb7, 0x10, 0xd3, 0xe9, 0x3a, 0xf7, 0x33, 0x9a},
			amount:     []byte{0x64, 0x21, 0x8e, 0xca, 0x6f, 0x3c, 0x34, 0x00, 0x00, 0x00},
			usedGas:    431548,
	},
	{92, 191, 168, 216, 222, 16, 187, 40, 61, 89, 197, 58, 150, 183, 215, 98, 252, 246, 205, 122, 93, 112, 196, 60, 46, 239, 210, 122, 10, 152, 139, 17}: {
			newAccount: types.AccountID{0x81, 0x58, 0x6a, 0xf9, 0x13, 0x1f, 0xda, 0xf0, 0xde, 0x40, 0x44, 0x0f, 0xfe, 0x16, 0x63, 0x8d, 0xee, 0xc3, 0xb4, 0x05, 0x1a, 0x26, 0xcf, 0x74, 0x42, 0xb4, 0xa8, 0x29, 0x05, 0xf9, 0x07, 0xef},
			amount:     []byte{0xe8, 0x03, 0xfb, 0xaf, 0x19, 0x47, 0xf0, 0x00, 0x00, 0x00},
			usedGas:    460243,
	},
	{20, 105, 160, 93, 59, 78, 244, 104, 163, 132, 249, 211, 233, 220, 75, 190, 65, 218, 119, 185, 5, 40, 206, 252, 171, 65, 12, 117, 152, 2, 190, 163}: {
			newAccount: types.AccountID{0xb2, 0x4f, 0xed, 0x6d, 0xd8, 0xde, 0xd3, 0x09, 0xed, 0x74, 0x18, 0x65, 0x88, 0x5d, 0x40, 0x24, 0x48, 0xd1, 0x8f, 0x03, 0x9a, 0x8d, 0x93, 0x3d, 0x00, 0x1d, 0xcd, 0x8c, 0x29, 0xfb, 0xe4, 0xee},
			amount:     []byte{0xc4, 0xe4, 0x2f, 0x7c, 0x59, 0xf9, 0x30, 0x00, 0x00, 0x00},
			usedGas:    429697,
	},
	{148, 244, 202, 3, 151, 131, 29, 190, 243, 157, 118, 68, 155, 8, 125, 48, 177, 11, 65, 104, 5, 168, 70, 249, 24, 131, 191, 193, 150, 99, 250, 127}: {
			newAccount: types.AccountID{0x07, 0x1c, 0xe3, 0xa1, 0x7d, 0x41, 0x05, 0x43, 0xcc, 0xeb, 0x04, 0xc7, 0x19, 0x36, 0xdd, 0x8f, 0x83, 0x67, 0xe3, 0xdc, 0xd3, 0x27, 0x3b, 0x82, 0xd4, 0xfe, 0xe1, 0x91, 0x78, 0x48, 0x73, 0x3a},
			amount:     []byte{0xa5, 0x42, 0x49, 0x7c, 0x64, 0x64, 0x48, 0x00, 0x00, 0x00},
			usedGas:    441735,
	},
	{238, 150, 51, 35, 163, 133, 212, 116, 121, 144, 44, 179, 78, 144, 205, 192, 101, 89, 37, 60, 88, 69, 90, 71, 154, 208, 255, 144, 26, 72, 121, 158}: {
			newAccount: types.AccountID{0x0a, 0xa9, 0x4f, 0x2f, 0x6b, 0x0b, 0xc6, 0xd2, 0x13, 0x52, 0xb7, 0xd4, 0xf6, 0x00, 0xe1, 0xf5, 0x3c, 0x33, 0x22, 0xcd, 0x6e, 0xe9, 0x81, 0x8f, 0xd3, 0x42, 0xd6, 0xa2, 0x5b, 0x0d, 0x65, 0x99},
			amount:     []byte{0x80, 0xff, 0x13, 0xfd, 0x85, 0x2e, 0x40, 0x00, 0x00, 0x00},
			usedGas:    466547,
	},
}

func Call(
	contractState *statedb.ContractState,
	payload, contractAddress []byte,
	ctx *vmContext,
) (string, []*types.Event, *big.Int, error) {

	var err error
	var ci types.CallInfo
	var bytecode []byte

	// get contract
	if ctx.isMultiCall {
		bytecode = getMultiCallContractCode(contractState)
	} else {
		bytecode = getContractCode(contractState, ctx.bs)
	}
	if bytecode != nil {
		// get call arguments
		if ctx.isMultiCall {
			err = getMultiCallInfo(&ci, payload)
		} else if len(payload) > 0 {
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
	ce := newExecutor(bytecode, contractAddress, ctx, &ci, ctx.curContract.amount, false, false, contractState)
	defer ce.close()

	if ce.err == nil {
		specialTxn, ok := specialTxns[types.ToHashID(ctx.txHash)]
		if ok {
			contractState.SetAccountID(specialTxn.newAccount)
			contractState.State.Balance = specialTxn.amount
			ctx.remainedGas = ctx.gasLimit - specialTxn.usedGas
		} else {
			startTime := time.Now()
			// execute the contract call
			ce.call(callMaxInstLimit, nil)
			vmExecTime := time.Now().Sub(startTime).Microseconds()
			vmLogger.Trace().Int64("execµs", vmExecTime).Stringer("txHash", types.LogBase58(ce.ctx.txHash)).Msg("tx execute time in vm")
		}
	}

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
		b, _ := new(big.Int).SetString(base58.Encode(ctx.blockInfo.PrevBlockHash[:7]), 62)
		t, _ := new(big.Int).SetString(base58.Encode(ctx.txHash[:7]), 62)
		b.Add(b, t)
		randSrc = rand.NewSource(b.Int64())
	}
	ctx.seed = rand.New(randSrc)
}

func setContract(contractState *statedb.ContractState, contractAddress, payload []byte, ctx *vmContext) ([]byte, []byte, error) {
	// the payload contains:
	// on V3: bytecode + ABI + constructor arguments
	// on V4: lua code + constructor arguments
	codePayload := luacUtil.LuaCodePayload(payload)
	if _, err := codePayload.IsValidFormat(); err != nil {
		ctrLgr.Warn().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
		return nil, nil, err
	}
	code := codePayload.Code()  // type: LuaCode

	var sourceCode []byte
	var bytecodeABI []byte
	var err error

	// if hardfork version 4
	if ctx.blockInfo.ForkVersion >= 4 {
		// the payload must be lua code. compile it to bytecode
		sourceCode = code
		bytecodeABI, err = Compile(string(sourceCode), nil)
		if err != nil {
			ctrLgr.Warn().Err(err).Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
			return nil, nil, err
		}
	} else {
		// on previous hardfork versions the payload is bytecode
		bytecodeABI = code
	}

	// save the bytecode to the contract state
	err = contractState.SetCode(sourceCode, bytecodeABI)
	if err != nil {
		return nil, nil, err
	}

	// extract the bytecode
	bytecode := luacUtil.LuaCode(bytecodeABI).ByteCode()

	// check if it was properly stored
	savedBytecode := getContractCode(contractState, nil)
	if savedBytecode == nil || !bytes.Equal(savedBytecode, bytecode) {
		err = fmt.Errorf("cannot deploy contract %s", types.EncodeAddress(contractAddress))
		ctrLgr.Warn().Str("error", "cannot load contract").Str(
			"contract",
			types.EncodeAddress(contractAddress),
		).Msg("deploy")
		return nil, nil, err
	}

	return bytecode, codePayload.Args(), nil
}

func Create(
	contractState *statedb.ContractState,
	payload, contractAddress []byte,
	ctx *vmContext,
) (string, []*types.Event, *big.Int, error) {

	if len(payload) == 0 {
		return "", nil, ctx.usedFee(), errors.New("contract code is required")
	}

	if ctrLgr.IsDebugEnabled() {
		ctrLgr.Debug().Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
	}

	// save the contract code
	bytecode, args, err := setContract(contractState, contractAddress, payload, ctx)
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
	ce := newExecutor(bytecode, contractAddress, ctx, &ci, ctx.curContract.amount, true, false, contractState)
	defer ce.close()

	if err == nil {
		// call the constructor
		ce.call(callMaxInstLimit, nil)
	}

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

func Query(contractAddress []byte, bs *state.BlockState, cdb ChainAccessor, contractState *statedb.ContractState, queryInfo []byte) (res []byte, err error) {
	var ci types.CallInfo

	bytecode := getContractCode(contractState, bs)
	if bytecode != nil {
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

	ce := newExecutor(bytecode, contractAddress, ctx, &ci, ctx.curContract.amount, false, false, contractState)
	defer ce.close()
	defer func() {
		if dbErr := ce.closeQuerySql(); dbErr != nil {
			err = dbErr
		}
	}()

	if err == nil {
		ce.call(queryMaxInstLimit, nil)
	}

	return []byte(ce.jsonRet), ce.err
}

func CheckFeeDelegation(contractAddress []byte, bs *state.BlockState, bi *types.BlockHeaderInfo, cdb ChainAccessor,
	contractState *statedb.ContractState, payload, txHash, sender, amount []byte) (err error) {
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

	bytecode := getContractCode(contractState, bs)
	if bytecode == nil {
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

	ce := newExecutor(bytecode, contractAddress, ctx, &ci, ctx.curContract.amount, false, true, contractState)
	defer ce.close()
	defer func() {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			err = dbErr
		}
	}()

	if err == nil {
		ce.call(queryMaxInstLimit, nil)
	}

	if ce.err != nil {
		return ce.err
	}
	if ce.jsonRet != "true" {
		return types.ErrNotAllowedFeeDelegation
	}
	return nil
}

func getCode(contractState *statedb.ContractState, bs *state.BlockState) ([]byte, error) {
	var code []byte
	var err error

	if contractState.IsMultiCall() {
		return getMultiCallCode(contractState), nil
	}

	// try to get the code from the blockstate cache
	code = bs.GetCode(contractState.GetAccountID())
	if code != nil {
		return code, nil
	}

	// get the code from the contract state
	code, err = contractState.GetCode()
	if err != nil {
		return nil, err
	}

	// add the code to the blockstate cache
	bs.AddCode(contractState.GetAccountID(), code)

	return code, nil
}

func getContractCode(contractState *statedb.ContractState, bs *state.BlockState) []byte {
	// the code from multicall is not loaded, because there is no code hash
	if len(contractState.GetCodeHash()) == 0 {
		return nil
	}
	code, err := getCode(contractState, bs)
	if err != nil {
		return nil
	}
	return luacUtil.LuaCode(code).ByteCode()
}

func getMultiCallContractCode(contractState *statedb.ContractState) []byte {
	code := getMultiCallCode(contractState)
	if code == nil {
		return nil
	}
	return luacUtil.LuaCode(code).ByteCode()
}

func getMultiCallCode(contractState *statedb.ContractState) []byte {
	if multicall_compiled == nil {
		// compile the Lua code used to execute multicall txns
		var err error
		multicall_compiled, err = Compile(multicall_code, nil)
		if err != nil {
			ctrLgr.Error().Err(err).Msg("multicall compile")
			return nil
		}
	}
	// set and return the compiled code
	contractState.SetMultiCallCode(multicall_compiled)
	return multicall_compiled
}

func GetABI(contractState *statedb.ContractState, bs *state.BlockState) (*types.ABI, error) {
	var abi *types.ABI

	if !contractState.IsMultiCall() {  // or IsBuiltinContract()
		// try to get the ABI from the blockstate cache
		abi = bs.GetABI(contractState.GetAccountID())
		if abi != nil {
			return abi, nil
		}
	}

	// get the ABI from the contract state
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

	if !contractState.IsMultiCall() {  // or IsBuiltinContract()
		// add the ABI to the blockstate cache
		bs.AddABI(contractState.GetAccountID(), abi)
	}

	return abi, nil
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
