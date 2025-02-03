/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package contract

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"strconv"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	luacUtil "github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/contract/msg"
	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
	"github.com/aergoio/aergo/v2/blacklist"
	jsoniter "github.com/json-iterator/go"
)

const (
	callMaxInstLimit     = 5000000
	queryMaxInstLimit    = callMaxInstLimit * 10
	dbUpdateMaxLimit     = fee.StateDbMaxUpdateSize
	maxCallDepthOld      = 5
	maxCallDepth         = 20
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
	CurrentForkVersion int32
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
	nestedView        int32 // indicates whether the parent called the contract in view (read-only) mode
	isFeeDelegation   bool
	isMultiCall       bool
	service           int
	callState         map[types.AccountID]*callState
	lastRecoveryEntry *recoveryEntry
	dbUpdateTotalSize int64
	seed              *rand.Rand
	events            []*types.Event
	eventCount        int32
	callDepth         int32
	callStack         []*executor
	traceFile         *os.File
	gasLimit          uint64
	remainingGas      uint64
	execCtx           context.Context
	deadline          time.Time
}

type executor struct {
	vmInstance *VmInstance
	code       []byte
	err        error
	ci         *types.CallInfo
	fname      string
	ctx        *vmContext
	contractGasLimit uint64
	usedGas    uint64
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

func MaxPossibleCallDepth() int {
	return maxCallDepth
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
		service:         executionMode,
		gasLimit:        gasLimit,
		remainingGas:    gasLimit,
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

	// use the deadline from the execution context
	if deadline, ok := execCtx.Deadline(); ok {
		ctx.deadline = deadline
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

////////////////////////////////////////////////////////////////////////////////
// GAS
////////////////////////////////////////////////////////////////////////////////

func (ctx *vmContext) IsGasSystem() bool {
	return fee.GasEnabled(ctx.blockInfo.ForkVersion) && !ctx.isQuery
}

// check if the gas limit set by the parent VM instance is valid
func (ctx *vmContext) parseGasLimit(gas string) (uint64, error) {
	// it must be a valid uint64 value
	if len(gas) != 8 {
		return 0, errors.New("uncatchable: invalid gas limit")
	}
	gasLimit := binary.LittleEndian.Uint64([]byte(gas))
	// gas limit must be less than or equal to the remaining gas
	if gasLimit > ctx.remainingGas {
		return 0, errors.New("uncatchable: gas limit exceeds the remaining gas")
	}
	return gasLimit, nil
}

// get the total gas used by all contracts in the current transaction
func (ctx *vmContext) usedGas() uint64 {
	if fee.IsZeroFee() || !ctx.IsGasSystem() {
		return 0
	}
	return ctx.gasLimit - ctx.remainingGas
}

// get the contracts execution fee
func (ctx *vmContext) usedFee() *big.Int {
	return fee.TxExecuteFee(ctx.blockInfo.ForkVersion, ctx.bs.GasPrice, ctx.usedGas(), ctx.dbUpdateTotalSize)
}

////////////////////////////////////////////////////////////////////////////////


// TODO: is this used on private chains? if not, remove it
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
	bytecode []byte,
	contractId []byte,
	ctx *vmContext,
	ci *types.CallInfo,
	amount *big.Int,
	isCreate bool,
	isFeeDelegation bool,
	ctrState *statedb.ContractState,
) *executor {

	if ctx.blockInfo.ForkVersion != CurrentForkVersion {
		// force the VM Pool to regenerate the VM instances
		// using the new hardfork version
		CurrentForkVersion = ctx.blockInfo.ForkVersion
		FlushVmInstances()
	}

	// create a new executor and add it to the call stack
	ce := &executor{
		ctx:  ctx,
		code: bytecode,
	}
	ctx.callStack = append(ctx.callStack, ce)
	ctx.callDepth++
	if ctx.callDepth > MaxCallDepth(ctx.blockInfo.ForkVersion) {
		ce.err = fmt.Errorf("exceeded the maximum call depth(%d)", MaxCallDepth(ctx.blockInfo.ForkVersion))
		return ce
	}

	if blacklist.Check(types.EncodeAddress(contractId)) {
		ce.err = fmt.Errorf("contract not available")
		ctrLgr.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("blocked contract")
		return ce
	}

	// get a connection to an unused VM instance
	ce.vmInstance = GetVmInstance()
	if ce.vmInstance == nil {
		ce.err = ErrVmStart
		ctrLgr.Error().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("new AergoLua executor")
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
			// the constructor function does not need to be declared with abi.register()
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
	} else if isFeeDelegation {
		_, err := resolveFunction(ctrState, ctx.bs, checkFeeDelegationFn, false)
		if err != nil {
			ce.preErr = err
			ctrLgr.Debug().Err(ce.err).Str("contract", types.EncodeAddress(contractId)).Msg("not found function")
			return ce
		}
		ce.isView = true
		ce.fname = checkFeeDelegationFn
		ce.isAutoload = true
	} else {
		f, err := resolveFunction(ctrState, ctx.bs, ci.Name, false)
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
	}
	ce.ci = ci

	return ce
}

func checkPayable(callee *types.Function, amount *big.Int) error {
	if amount.Cmp(big.NewInt(0)) <= 0 || callee.Payable {
		return nil
	}
	return fmt.Errorf("'%s' is not payable", callee.Name)
}

func (ce *executor) getEvents() []*types.Event {
	if ce == nil || ce.ctx == nil {
		return nil
	}
	return ce.ctx.events
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

func (ce *executor) close() {
	if ce != nil {
		FreeVmInstance(ce.vmInstance)
		if ce.ctx != nil {
			ce.ctx.callDepth--
			ce.ctx.callStack = ce.ctx.callStack[:len(ce.ctx.callStack)-1]
			if ce.ctx.traceFile != nil {
				ce.ctx.traceFile.Close()
				ce.ctx.traceFile = nil
			}
		}
	}
}




func getMultiCallInfo(ci *types.CallInfo, payload []byte) error {
	payload = append([]byte{'['}, payload...)
	payload = append(payload, ']')
	ci.Name = "execute"
	return getCallInfo(&ci.Args, payload, []byte("multicall"))
}

// ci is a pointer to a CallInfo struct: { Name string, Args []interface{} }
// args is a JSON array of arguments
func getCallInfo(ci interface{}, args []byte, contractAddress []byte) error {
	d := json.NewDecoder(bytes.NewReader(args))
	d.UseNumber()
	d.DisallowUnknownFields()
	err := d.Decode(ci)
	if err != nil {
		ctrLgr.Debug().AnErr("error", err).Str("contract", types.EncodeAddress(contractAddress)).Msg("invalid calling information")
	}
	return err
}

// return only the arguments as a single string containing the JSON array
func getCallInfoArgs(ci *types.CallInfo) (string, error) {
	args, err := json.Marshal(ci.Args)
	if err != nil {
		return "", err
	}
	return string(args), nil
}




////////////////////////////////////////////////////////////////////////////////
// Called Externally
////////////////////////////////////////////////////////////////////////////////

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
	defer ce.close()  // close the executor and the VM instance

	// set the gas limit from the transaction
	ce.contractGasLimit = ctx.gasLimit

	// execute the contract call
	ce.call(true)

	err = ctx.updateUsedGas(ce.usedGas)
	if err != nil {
		return "", nil, ctx.usedFee(), err
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

	// create a new executor for the constructor
	ce := newExecutor(bytecode, contractAddress, ctx, &ci, ctx.curContract.amount, true, false, contractState)
	defer ce.close()  // close the executor and the VM instance

	// set the gas limit from the transaction
	ce.contractGasLimit = ctx.gasLimit

	if ce.err == nil {
		// call the constructor
		ce.call(true)
	}

	err = ctx.updateUsedGas(ce.usedGas)
	if err != nil {
		return "", nil, ctx.usedFee(), err
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
			ctx.service = index
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
	defer ce.close()  // close the executor and the VM instance
	defer func() {
		if dbErr := ce.closeQuerySql(); dbErr != nil {
			err = dbErr
		}
	}()

	// set the gas limit from the transaction
	ce.contractGasLimit = ctx.gasLimit

	if ce.err == nil {
		ce.call(true)
	}

	err = ctx.updateUsedGas(ce.usedGas)
	if err != nil {
		return nil, err
	}

	return []byte(ce.jsonRet), ce.err
}



//! this is complicated, a query before the actual execution
//  and queried many times, even by mempool

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
	defer ce.close()  // close the executor and the VM instance
	defer func() {
		if dbErr := ce.rollbackToSavepoint(); dbErr != nil {
			err = dbErr
		}
	}()

	// set the gas limit from the transaction
	ce.contractGasLimit = ctx.gasLimit

	if ce.err == nil {
		ce.call(true)
	}

	err = ctx.updateUsedGas(ce.usedGas)
	if err != nil {
		return err
	}

	if ce.err != nil {
		return ce.err
	}
	if ce.jsonRet != "true" {
		return types.ErrNotAllowedFeeDelegation
	}
	return nil
}

func (ctx *vmContext) updateUsedGas(usedGas uint64) error {
	if usedGas > ctx.remainingGas {
		ctx.remainingGas = 0
		return errors.New("run out of gas")
	}
	// deduct the used gas
	ctx.remainingGas -= usedGas
	return nil
}



////////////////////////////////////////////////////////////////////////////////
// Contract Code
////////////////////////////////////////////////////////////////////////////////

// only called by a deploy transaction
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
		bytecodeABI, err = Compile(string(sourceCode), false)
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
		ctrLgr.Warn().Str("error", "cannot load contract").Str("contract", types.EncodeAddress(contractAddress)).Msg("deploy")
		return nil, nil, err
	}

	return bytecode, codePayload.Args(), nil
}

func getContract(contractState *statedb.ContractState, bs *state.BlockState) []byte {
	code, err := getCode(contractState, bs)
	if err != nil {
		return nil
	}
	return luacUtil.LuaCode(code).ByteCode()
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
		multicall_compiled, err = Compile(multicall_code, false)
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

// send the source code to a VM instance, to be compiled
func Compile(code string, hasParent bool) (luacUtil.LuaCode, error) {

	// get a connection to an unused VM instance
	vmInstance := GetVmInstance()
	if vmInstance == nil {
		err := ErrVmStart
		ctrLgr.Error().Err(err).Msg("get vm instance for compilation")
		return nil, err
	}
	defer FreeVmInstance(vmInstance)

	// build the message
	message := msg.SerializeMessage("compile", code, strconv.FormatBool(hasParent))

	/*/ encrypt the message
	message, err = msg.Encrypt(message, secretKey)
	if err != nil {
		return nil, err
	}
	*/

	// send the execution request to the VM instance
	err := msg.SendMessage(vmInstance.conn, message)
	if err != nil {
		return nil, fmt.Errorf("compile: send message: %v", err)
	}

	// timeout of 250 ms
	deadline := time.Now().Add(250 * time.Millisecond)
	response, err := msg.WaitForMessage(vmInstance.conn, deadline)
	if err != nil {
		return nil, fmt.Errorf("compile: wait for message: %v", err)
	}

	/*/ decrypt the message
	response, err = msg.Decrypt(response, secretKey)
	if err != nil {
		return nil, err
	}
	*/

	results, err := msg.DeserializeMessage(response)
	if len(results) != 2 {
		return nil, fmt.Errorf("compile: invalid number of results: %v", results)
	}
	bytecodeAbi := results[0]
	errMsg := results[1]

	if len(errMsg) > 0 {
		return nil, fmt.Errorf("compile: %s", errMsg)
	}
	return luacUtil.LuaCode(bytecodeAbi), nil
}
