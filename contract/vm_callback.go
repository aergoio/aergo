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
	"encoding/hex"
	"errors"
	"fmt"
	"index/suffixarray"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/minio/sha256-simd"
)

var (
	mulAergo, mulGaer, zeroBig *big.Int
	creatorMetaKey             = []byte("Creator")
)

const (
	maxEventCnt       = 50
	maxEventNameSize  = 64
	maxEventArgSize   = 4096
	luaCallCountDeduc = 1000
)

func init() {
	mulAergo = types.NewAmount(1, types.Aergo)
	mulGaer = types.NewAmount(1, types.Gaer)
	zeroBig = types.NewZeroAmount()
}

func addUpdateSize(s *vmContext, updateSize int64) error {
	if s.IsGasSystem() {
		return nil
	}
	if s.dbUpdateTotalSize+updateSize > dbUpdateMaxLimit {
		return errors.New("exceeded size of updates in the state database")
	}
	s.dbUpdateTotalSize += updateSize
	return nil
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
	val := []byte(C.GoString(value))
	if err := ctx.curContract.callState.ctrState.SetData(C.GoBytes(key, keyLen), val); err != nil {
		return C.CString(err.Error())
	}
	if err := addUpdateSize(ctx, int64(types.HashIDLength+len(val))); err != nil {
		C.luaL_setuncatchablerror(L)
		return C.CString(err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString("[Set]\n")
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("Key=%s Len=%v byte=%v\n",
			string(C.GoBytes(key, keyLen)), keyLen, C.GoBytes(key, keyLen)))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("Data=%s Len=%d byte=%v\n",
			string(val), len(val), val))
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
	if err := ctx.curContract.callState.ctrState.DeleteData(C.GoBytes(key, keyLen)); err != nil {
		return C.CString(err.Error())
	}
	if err := addUpdateSize(ctx, int64(32)); err != nil {
		C.luaL_setuncatchablerror(L)
		return C.CString(err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString("[Del]\n")
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("Key=%s Len=%v byte=%v\n",
			string(C.GoBytes(key, keyLen)), keyLen, C.GoBytes(key, keyLen)))
	}
	return nil
}

func getCallState(ctx *vmContext, aid types.AccountID) (*callState, error) {
	cs := ctx.callState[aid]
	if cs == nil {
		bs := ctx.bs

		prevState, err := bs.GetAccountState(aid)
		if err != nil {
			return nil, err
		}

		curState := types.Clone(*prevState).(types.State)
		cs = &callState{prevState: prevState, curState: &curState}
		ctx.callState[aid] = cs
	}
	return cs, nil
}

func getCtrState(ctx *vmContext, aid types.AccountID) (*callState, error) {
	cs, err := getCallState(ctx, aid)
	if err != nil {
		return nil, err
	}
	if cs.ctrState == nil {
		cs.ctrState, err = ctx.bs.OpenContractState(aid, cs.curState)
	}
	return cs, err
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
	amount *C.char, gas uint64) (C.int, *C.char) {
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.LuaCallContract] contract state not found")
	}

	// get the contract address
	contractAddress := C.GoString(contractId)
	cid, err := getAddressNameResolved(contractAddress, ctx.bs)
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] invalid contractId: " + err.Error())
	}
	aid := types.ToAccountID(cid)

	// read the amount for the contract call
	amountBig, err := transformAmount(C.GoString(amount))
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] invalid amount: " + err.Error())
	}

	// get the contract state
	cs, err := getCtrState(ctx, aid)
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] getAccount error: " + err.Error())
	}

	// check if the contract exists
	callee := getContract(cs.ctrState, ctx.bs)
	if callee == nil {
		return -1, C.CString("[Contract.LuaCallContract] cannot find contract " + C.GoString(contractId))
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
	ctx.getRemainingGas(L)
	// create a new executor with the remaining gas on the child LState
	ce := newExecutor(callee, cid, ctx, &ci, amountBig, false, false, cs.ctrState)
	defer func() {
		// close the executor, closes also the child LState
		ce.close()
		// set the remaining gas on the parent LState
		ctx.setRemainingGas(L)
	}()

	if ce.err != nil {
		return -1, C.CString("[Contract.LuaCallContract] newExecutor error: " + ce.err.Error())
	}

	// send the amount to the contract
	senderState := prevContractInfo.callState.curState
	if amountBig.Cmp(zeroBig) > 0 {
		if ctx.isQuery == true || ctx.nestedView > 0 {
			return -1, C.CString("[Contract.LuaCallContract] send not permitted in query")
		}
		if r := sendBalance(L, senderState, cs.curState, amountBig); r != nil {
			return -1, r
		}
	}

	seq, err := setRecoveryPoint(aid, ctx, senderState, cs, amountBig, false, false)
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CALL Contract %v(%v) %v]\n",
			contractAddress, aid.String(), fnameStr))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("SendBalance: %s\n", amountBig.String()))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
			senderState.GetBalanceBigInt().String(), cs.curState.GetBalanceBigInt().String()))
	}
	if err != nil {
		return -1, C.CString("[System.LuaCallContract] database error: " + err.Error())
	}

	// set the current contract info
	ctx.curContract = newContractInfo(cs, prevContractInfo.contractId, cid,
		cs.curState.SqlRecoveryPoint, amountBig)
	defer func() {
		ctx.curContract = prevContractInfo
	}()

	// execute the contract call
	defer setInstCount(ctx, L, ce.L)
	ret := ce.call(minusCallCount(ctx, C.vm_instcount(L), luaCallCountDeduc), L)

	// check if the contract call failed
	if ce.err != nil {
		err := clearRecovery(L, ctx, seq, true)
		if err != nil {
			return -1, C.CString("[Contract.LuaCallContract] recovery err: " + err.Error())
		}
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
		}
		return -1, C.CString("[Contract.LuaCallContract] call err: " + ce.err.Error())
	}

	if seq == 1 {
		err := clearRecovery(L, ctx, seq, false)
		if err != nil {
			return -1, C.CString("[Contract.LuaCallContract] recovery err: " + err.Error())
		}
	}

	return ret, nil
}

func getOnlyContractState(ctx *vmContext, aid types.AccountID) (*state.ContractState, error) {
	cs := ctx.callState[aid]
	if cs == nil || cs.ctrState == nil {
		return ctx.bs.OpenContractStateAccount(aid)
	}
	return cs.ctrState, nil
}

//export luaDelegateCallContract
func luaDelegateCallContract(L *LState, service C.int, contractId *C.char,
	fname *C.char, args *C.char, gas uint64) (C.int, *C.char) {
	contractIdStr := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] contract state not found")
	}

	// get the contract address
	cid, err := getAddressNameResolved(contractIdStr, ctx.bs)
	if err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] invalid contractId: " + err.Error())
	}
	aid := types.ToAccountID(cid)

	// get the contract state
	contractState, err := getOnlyContractState(ctx, aid)
	if err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract]getContractState error" + err.Error())
	}

	// check if the contract exists
	contract := getContract(contractState, ctx.bs)
	if contract == nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] cannot find contract " + contractIdStr)
	}

	// read the arguments for the contract call
	var ci types.CallInfo
	ci.Name = fnameStr
	err = getCallInfo(&ci.Args, []byte(argsStr), cid)
	if err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] invalid arguments: " + err.Error())
	}

	// get the remaining gas from the parent LState
	ctx.getRemainingGas(L)
	// create a new executor with the remaining gas on the child LState
	ce := newExecutor(contract, cid, ctx, &ci, zeroBig, false, false, contractState)
	defer func() {
		// close the executor, closes also the child LState
		ce.close()
		// set the remaining gas on the parent LState
		ctx.setRemainingGas(L)
	}()

	if ce.err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] newExecutor error: " + ce.err.Error())
	}

	seq, err := setRecoveryPoint(aid, ctx, nil, ctx.curContract.callState, zeroBig, false, false)
	if err != nil {
		return -1, C.CString("[System.LuaDelegateCallContract] database error: " + err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[DELEGATECALL Contract %v %v]\n", contractIdStr, fnameStr))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
	}

	// execute the contract call
	defer setInstCount(ctx, L, ce.L)
	ret := ce.call(minusCallCount(ctx, C.vm_instcount(L), luaCallCountDeduc), L)

	// check if the contract call failed
	if ce.err != nil {
		err := clearRecovery(L, ctx, seq, true)
		if err != nil {
			return -1, C.CString("[Contract.LuaDelegateCallContract] recovery error: " + err.Error())
		}
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
		}
		return -1, C.CString("[Contract.LuaDelegateCallContract] call error: " + ce.err.Error())
	}

	if seq == 1 {
		err := clearRecovery(L, ctx, seq, false)
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
func luaSendAmount(L *LState, service C.int, contractId *C.char, amount *C.char) *C.char {

	ctx := contexts[service]
	if ctx == nil {
		return C.CString("[Contract.LuaSendAmount] contract state not found")
	}

	// read the amount to be sent
	amountBig, err := transformAmount(C.GoString(amount))
	if err != nil {
		return C.CString("[Contract.LuaSendAmount] invalid amount: " + err.Error())
	}

	// cannot send amount in query
	if (ctx.isQuery == true || ctx.nestedView > 0) && amountBig.Cmp(zeroBig) > 0 {
		return C.CString("[Contract.LuaSendAmount] send not permitted in query")
	}

	// get the receiver account
	cid, err := getAddressNameResolved(C.GoString(contractId), ctx.bs)
	if err != nil {
		return C.CString("[Contract.LuaSendAmount] invalid contractId: " + err.Error())
	}

	// get the receiver state
	aid := types.ToAccountID(cid)
	cs, err := getCallState(ctx, aid)
	if err != nil {
		return C.CString("[Contract.LuaSendAmount] getAccount error: " + err.Error())
	}

	// get the sender state
	senderState := ctx.curContract.callState.curState

	// check if the receiver is a contract
	if len(cs.curState.GetCodeHash()) > 0 {

		// get the contract state
		if cs.ctrState == nil {
			cs.ctrState, err = ctx.bs.OpenContractState(aid, cs.curState)
			if err != nil {
				return C.CString("[Contract.LuaSendAmount] getContractState error: " + err.Error())
			}
		}

		// set the function to be called
		var ci types.CallInfo
		ci.Name = "default"

		// get the contract code
		code := getContract(cs.ctrState, ctx.bs)
		if code == nil {
			return C.CString("[Contract.LuaSendAmount] cannot find contract:" + C.GoString(contractId))
		}

		// get the remaining gas from the parent LState
		ctx.getRemainingGas(L)
		// create a new executor with the remaining gas on the child LState
		ce := newExecutor(code, cid, ctx, &ci, amountBig, false, false, cs.ctrState)
		defer func() {
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
			if r := sendBalance(L, senderState, cs.curState, amountBig); r != nil {
				return r
			}
		}

		// create a recovery point
		seq, err := setRecoveryPoint(aid, ctx, senderState, cs, amountBig, false, false)
		if err != nil {
			return C.CString("[System.LuaSendAmount] database error: " + err.Error())
		}

		// log some info
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(
				fmt.Sprintf("[Send Call default] %s(%s) : %s\n", types.EncodeAddress(cid), aid.String(), amountBig.String()))
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
				senderState.GetBalanceBigInt().String(), cs.curState.GetBalanceBigInt().String()))
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
		}

		// set the current contract info
		prevContractInfo := ctx.curContract
		ctx.curContract = newContractInfo(cs, prevContractInfo.contractId, cid,
			cs.curState.SqlRecoveryPoint, amountBig)
		defer func() {
			ctx.curContract = prevContractInfo
		}()

		// execute the contract call
		defer setInstCount(ctx, L, ce.L)
		ce.call(minusCallCount(ctx, C.vm_instcount(L), luaCallCountDeduc), L)

		// check if the contract call failed
		if ce.err != nil {
			// recover to the previous state
			err := clearRecovery(L, ctx, seq, true)
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
			err := clearRecovery(L, ctx, seq, false)
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
	if r := sendBalance(L, senderState, cs.curState, amountBig); r != nil {
		return r
	}

	// update the recovery point
	if ctx.lastRecoveryEntry != nil {
		_, _ = setRecoveryPoint(aid, ctx, senderState, cs, amountBig, true, false)
	}

	// log some info
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[Send] %s(%s) : %s\n",
			types.EncodeAddress(cid), aid.String(), amountBig.String()))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
			senderState.GetBalanceBigInt().String(), cs.curState.GetBalanceBigInt().String()))
	}

	return nil
}

func sendBalance(L *LState, sender *types.State, receiver *types.State, amount *big.Int) *C.char {
	if sender == receiver {
		return nil
	}
	if sender.GetBalanceBigInt().Cmp(amount) < 0 {
		return C.CString("[Contract.sendBalance] insufficient balance: " +
			sender.GetBalanceBigInt().String() + " : " + amount.String())
	} else {
		sender.Balance = new(big.Int).Sub(sender.GetBalanceBigInt(), amount).Bytes()
	}
	receiver.Balance = new(big.Int).Add(receiver.GetBalanceBigInt(), amount).Bytes()

	return nil
}

//export luaPrint
func luaPrint(L *LState, service C.int, args *C.char) {
	ctx := contexts[service]
	setInstMinusCount(ctx, L, 1000)
	ctrLgr.Info().Str("Contract SystemPrint", types.EncodeAddress(ctx.curContract.contractId)).Msg(C.GoString(args))
}

func setRecoveryPoint(aid types.AccountID, ctx *vmContext, senderState *types.State,
	cs *callState, amount *big.Int, isSend, isDeploy bool) (int, error) {
	var seq int
	prev := ctx.lastRecoveryEntry
	if prev != nil {
		seq = prev.seq + 1
	} else {
		seq = 1
	}
	re := &recoveryEntry{
		seq,
		amount,
		senderState,
		senderState.GetNonce(),
		cs,
		isSend,
		isDeploy,
		nil,
		-1,
		prev,
	}
	ctx.lastRecoveryEntry = re
	if isSend {
		return seq, nil
	}
	re.stateRevision = cs.ctrState.Snapshot()
	tx := cs.tx
	if tx != nil {
		saveName := fmt.Sprintf("%s_%p", aid.String(), &re)
		err := tx.subSavepoint(saveName)
		if err != nil {
			return seq, err
		}
		re.sqlSaveName = &saveName
	}
	return seq, nil
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
	seq, err := setRecoveryPoint(types.ToAccountID(curContract.contractId), ctx, nil,
		curContract.callState, zeroBig, false, false)
	if err != nil {
		return -1, C.CString("[Contract.pcall] database error: " + err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[Pcall] snapshot set %d\n", seq))
	}
	return C.int(seq), nil
}

func clearRecovery(L *LState, ctx *vmContext, start int, error bool) error {
	item := ctx.lastRecoveryEntry
	for {
		if error {
			if item.recovery(ctx.bs) != nil {
				return errors.New("database error")
			}
		}
		if item.seq == start {
			if error || item.prev == nil {
				ctx.lastRecoveryEntry = item.prev
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
func luaClearRecovery(L *LState, service C.int, start int, error bool) *C.char {
	ctx := contexts[service]
	if ctx == nil {
		return C.CString("[Contract.pcall] contract state not found")
	}
	err := clearRecovery(L, ctx, start, error)
	if err != nil {
		return C.CString(err.Error())
	}
	if ctx.traceFile != nil && error == true {
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
	return C.CString(cs.curState.GetBalanceBigInt().String()), nil
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
	return C.CString(enc.ToString(ctx.txHash))
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
	return C.CString(enc.ToString(ctx.blockInfo.PrevBlockHash))
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
		data, err = hex.DecodeString(string(dataStr))
		if err != nil {
			return nil, C.CString("[Contract.LuaCryptoSha256] hex decoding error: " + err.Error())
		}
	}
	h := sha256.New()
	h.Write(data)
	resultHash := h.Sum(nil)

	return C.CString("0x" + hex.EncodeToString(resultHash)), nil
}

func decodeHex(hexStr string) ([]byte, error) {
	if checkHexString(hexStr) {
		hexStr = hexStr[2:]
	}
	return hex.DecodeString(hexStr)
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
		pubKey, err = btcec.ParsePubKey(bAddress, btcec.S256())
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
		pub, _, err := btcec.RecoverCompact(btcec.S256(), bSig, bMsg)
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
		sign, err := btcec.ParseSignature(bSig, btcec.S256())
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
		d, err = hex.DecodeString(string(b[2:]))
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
		hexb := []byte("0x" + hex.EncodeToString(h))
		return C.CBytes(hexb), len(hexb)
	} else {
		return C.CBytes(h), len(h)
	}
}

func transformAmount(amountStr string) (*big.Int, error) {
	var ret *big.Int
	var prev int
	if len(amountStr) == 0 {
		return zeroBig, nil
	}
	index := suffixarray.New([]byte(amountStr))
	r := regexp.MustCompile("(?i)aergo|gaer|aer")

	res := index.FindAllIndex(r, -1)
	for _, pair := range res {
		amountBig, _ := new(big.Int).SetString(strings.TrimSpace(amountStr[prev:pair[0]]), 10)
		if amountBig == nil {
			return nil, errors.New("converting error for BigNum: " + amountStr[prev:])
		}
		cmp := amountBig.Cmp(zeroBig)
		if cmp < 0 {
			return nil, errors.New("negative amount not allowed")
		} else if cmp == 0 {
			prev = pair[1]
			continue
		}
		switch pair[1] - pair[0] {
		case 3:
		case 4:
			amountBig = new(big.Int).Mul(amountBig, mulGaer)
		case 5:
			amountBig = new(big.Int).Mul(amountBig, mulAergo)
		}
		if ret != nil {
			ret = new(big.Int).Add(ret, amountBig)
		} else {
			ret = amountBig
		}
		prev = pair[1]
	}

	if prev >= len(amountStr) {
		if ret != nil {
			return ret, nil
		} else {
			return zeroBig, nil
		}
	}
	num := strings.TrimSpace(amountStr[prev:])
	if len(num) == 0 {
		if ret != nil {
			return ret, nil
		} else {
			return zeroBig, nil
		}
	}

	amountBig, _ := new(big.Int).SetString(num, 10)

	if amountBig == nil {
		return nil, errors.New("converting error for Integer: " + amountStr[prev:])
	}
	if amountBig.Cmp(zeroBig) < 0 {
		return nil, errors.New("negative amount not allowed")
	}
	if ret != nil {
		ret = new(big.Int).Add(ret, amountBig)
	} else {
		ret = amountBig
	}
	return ret, nil
}

//export luaDeployContract
func luaDeployContract(
	L *LState,
	service C.int,
	contract *C.char,
	args *C.char,
	amount *C.char,
) (C.int, *C.char) {

	argsStr := C.GoString(args)
	contractStr := C.GoString(contract)

	ctx := contexts[service]
	if ctx == nil {
		return -1, C.CString("[Contract.LuaDeployContract]not found contract state")
	}
	if ctx.isQuery == true || ctx.nestedView > 0 {
		return -1, C.CString("[Contract.LuaDeployContract]send not permitted in query")
	}
	bs := ctx.bs

	// contract code
	var code []byte

	// check if contract name or address is given
	cid, err := getAddressNameResolved(contractStr, bs)
	if err == nil {
		aid := types.ToAccountID(cid)
		// check if contract exists
		contractState, err := getOnlyContractState(ctx, aid)
		if err != nil {
			return -1, C.CString("[Contract.LuaDeployContract]" + err.Error())
		}
		// read the contract code
		code, err = contractState.GetCode()
		if err != nil {
			return -1, C.CString("[Contract.LuaDeployContract]" + err.Error())
		} else if len(code) == 0 {
			return -1, C.CString("[Contract.LuaDeployContract]: not found code")
		}
	}

	// compile contract code if not found
	if len(code) == 0 {
		if ctx.blockInfo.ForkVersion >= 2 {
			code, err = Compile(contractStr, L)
		} else {
			code, err = Compile(contractStr, nil)
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
	}

	err = addUpdateSize(ctx, int64(len(code)))
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	// create account for the contract
	prevContractInfo := ctx.curContract
	creator := prevContractInfo.callState.curState
	newContract, err := bs.CreateAccountStateV(CreateContractID(prevContractInfo.contractId, creator.GetNonce()))
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}
	contractState, err := bs.OpenContractState(newContract.AccountID(), newContract.State())
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	cs := &callState{ctrState: contractState, prevState: &types.State{}, curState: newContract.State()}
	ctx.callState[newContract.AccountID()] = cs

	// read the amount transferred to the contract
	amountBig, err := transformAmount(C.GoString(amount))
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
	senderState := prevContractInfo.callState.curState
	if amountBig.Cmp(zeroBig) > 0 {
		if rv := sendBalance(L, senderState, cs.curState, amountBig); rv != nil {
			return -1, rv
		}
	}

	// create a recovery point
	seq, err := setRecoveryPoint(newContract.AccountID(), ctx, senderState, cs, amountBig, false, true)
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
			senderState.GetBalanceBigInt().String(), cs.curState.GetBalanceBigInt().String()))
	}

	// set the contract info
	ctx.curContract = newContractInfo(cs, prevContractInfo.contractId, newContract.ID(),
		cs.curState.SqlRecoveryPoint, amountBig)
	defer func() {
		ctx.curContract = prevContractInfo
	}()

	runCode := util.LuaCode(code).ByteCode()

	// save the contract code
	err = contractState.SetCode(code)
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	// save the contract creator
	err = contractState.SetData(creatorMetaKey, []byte(types.EncodeAddress(prevContractInfo.contractId)))
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	// get the remaining gas from the parent LState
	ctx.getRemainingGas(L)
	// create a new executor with the remaining gas on the child LState
	ce := newExecutor(runCode, newContract.ID(), ctx, &ci, amountBig, true, false, contractState)
	defer func() {
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
	senderState.Nonce += 1

	addr := C.CString(types.EncodeAddress(newContract.ID()))
	ret := C.int(1)

	if ce != nil {
		// run the constructor
		defer setInstCount(ce.ctx, L, ce.L)
		ret += ce.call(minusCallCount(ctx, C.vm_instcount(L), luaCallCountDeduc), L)

		// check if the execution was successful
		if ce.err != nil {
			// rollback the recovery point
			err := clearRecovery(L, ctx, seq, true)
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
		err := clearRecovery(L, ctx, seq, false)
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
func luaEvent(L *LState, service C.int, eventName *C.char, args *C.char) *C.char {
	ctx := contexts[service]
	if ctx.isQuery == true || ctx.nestedView > 0 {
		return C.CString("[Contract.Event] event not permitted in query")
	}
	if ctx.eventCount >= maxEventCnt {
		return C.CString(fmt.Sprintf("[Contract.Event] exceeded the maximum number of events(%d)", maxEventCnt))
	}
	if len(C.GoString(eventName)) > maxEventNameSize {
		return C.CString(fmt.Sprintf("[Contract.Event] exceeded the maximum length of event name(%d)", maxEventNameSize))
	}
	if len(C.GoString(args)) > maxEventArgSize {
		return C.CString(fmt.Sprintf("[Contract.Event] exceeded the maximum length of event args(%d)", maxEventArgSize))
	}
	ctx.events = append(
		ctx.events,
		&types.Event{
			ContractAddress: ctx.curContract.contractId,
			EventIdx:        ctx.eventCount,
			EventName:       C.GoString(eventName),
			JsonArgs:        C.GoString(args),
		},
	)
	ctx.eventCount++
	return nil
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

	aid := types.ToAccountID(cid)
	cs, err := getCallState(ctx, aid)
	if err != nil {
		return -1, C.CString("[Contract.LuaIsContract] getAccount error: " + err.Error())
	}

	return C.int(len(cs.curState.GetCodeHash())), nil
}

//export luaGovernance
func luaGovernance(L *LState, service C.int, gType C.char, arg *C.char) *C.char {

	ctx := contexts[service]
	if ctx == nil {
		return C.CString("[Contract.LuaGovernance] contract state not found")
	}

	if ctx.isQuery == true || ctx.nestedView > 0 {
		return C.CString("[Contract.LuaGovernance] governance not permitted in query")
	}

	var amountBig *big.Int
	var payload []byte

	switch gType {
	case 'S', 'U':
		var err error
		amountBig, err = transformAmount(C.GoString(arg))
		if err != nil {
			return C.CString("[Contract.LuaGovernance] invalid amount: " + err.Error())
		}
		if gType == 'S' {
			payload = []byte(fmt.Sprintf(`{"Name":"%s"}`, types.Opstake.Cmd()))
		} else {
			payload = []byte(fmt.Sprintf(`{"Name":"%s"}`, types.Opunstake.Cmd()))
		}
	case 'V':
		amountBig = zeroBig
		payload = []byte(fmt.Sprintf(`{"Name":"%s","Args":%s}`, types.OpvoteBP.Cmd(), C.GoString(arg)))
	case 'D':
		amountBig = zeroBig
		payload = []byte(fmt.Sprintf(`{"Name":"%s","Args":%s}`, types.OpvoteDAO.Cmd(), C.GoString(arg)))
	}

	aid := types.ToAccountID([]byte(types.AergoSystem))
	scsState, err := getCtrState(ctx, aid)
	if err != nil {
		return C.CString("[Contract.LuaGovernance] getAccount error: " + err.Error())
	}

	curContract := ctx.curContract

	senderState := curContract.callState.curState
	sender := ctx.bs.InitAccountStateV(curContract.contractId,
		curContract.callState.prevState, curContract.callState.curState)
	receiver := ctx.bs.InitAccountStateV([]byte(types.AergoSystem),
		scsState.prevState, scsState.curState)

	if sender.AccountID().String() == "A9zXKkooeGYAZC5ReCcgeg4ddsvMHAy2ivUafXhrnzpj" {
		sender.ClearAid()
	}

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

	seq, err := setRecoveryPoint(aid, ctx, senderState, scsState, zeroBig, false, false)
	if err != nil {
		return C.CString("[Contract.LuaGovernance] database error: " + err.Error())
	}

	events, err := system.ExecuteSystemTx(scsState.ctrState, &txBody, sender, receiver, ctx.blockInfo)
	if err != nil {
		rErr := clearRecovery(L, ctx, seq, true)
		if rErr != nil {
			return C.CString("[Contract.LuaGovernance] recovery error: " + rErr.Error())
		}
		return C.CString("[Contract.LuaGovernance] error: " + err.Error())
	}

	if seq == 1 {
		err := clearRecovery(L, ctx, seq, false)
		if err != nil {
			return C.CString("[Contract.LuaGovernance] recovery error: " + err.Error())
		}
	}

	ctx.eventCount += int32(len(events))
	ctx.events = append(ctx.events, events...)

	if ctx.lastRecoveryEntry != nil {
		if gType == 'S' {
			seq, _ = setRecoveryPoint(aid, ctx, senderState, scsState, amountBig, true, false)
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[GOVERNANCE]aid(%s)\n", aid.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("staking : %s\n", amountBig.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
					senderState.GetBalanceBigInt().String(), scsState.curState.GetBalanceBigInt().String()))
			}
		} else if gType == 'U' {
			seq, _ = setRecoveryPoint(aid, ctx, scsState.curState, ctx.curContract.callState, amountBig, true, false)
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[GOVERNANCE]aid(%s)\n", aid.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("unstaking : %s\n", amountBig.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
					senderState.GetBalanceBigInt().String(), scsState.curState.GetBalanceBigInt().String()))
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

	select {
	case <-bpTimeout:
		return 1
	default:
		return 0
	}

	// Temporarily disable timeout check to prevent contract timeout raised from chain service
	// if service < BlockFactory {
	// 	service = service + MaxVmService
	// }
	// if service != BlockFactory {
	// 	return 0
	// }
	// select {
	// case <-bpTimeout:
	// 	return 1
	// default:
	// 	return 0
	// }
	//return 0
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

// set the remaining gas on the given LState
func (ctx *vmContext) setRemainingGas(L *LState) {
	if ctx.IsGasSystem() {
		C.lua_gasset(L, C.ulonglong(ctx.remainedGas))
	}
}

//export luaGetStaking
func luaGetStaking(service C.int, addr *C.char) (*C.char, C.lua_Integer, *C.char) {

	var (
		ctx          *vmContext
		scs, namescs *state.ContractState
		err          error
		staking      *types.Staking
	)

	ctx = contexts[service]
	scs, err = ctx.bs.GetSystemAccountState()
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}

	namescs, err = ctx.bs.GetNameAccountState()
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}

	staking, err = system.GetStaking(scs, name.GetAddress(namescs, types.ToAddress(C.GoString(addr))))
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}

	return C.CString(staking.GetAmountBigInt().String()), C.lua_Integer(staking.When), nil
}
