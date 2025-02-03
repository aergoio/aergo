package contract

/*
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include "db_msg.h"
#include "db_module.h"

#define ERR_BF_TIMEOUT "contract timeout"

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
	"math/big"
	"math/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
	"runtime"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/contract/msg"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
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

////////////////////////////////////////////////////////////////////////////////
// VM API
////////////////////////////////////////////////////////////////////////////////

func (ctx *vmContext) handleSetVariable(args []string) (result string, err error) {
	if len(args) != 2 {
		return "", errors.New("[System.SetVariable] invalid number of arguments")
	}
	key, value := []byte(args[0]), []byte(args[1])
	if ctx.isQuery || ctx.nestedView > 0 {
		return "", errors.New("[System.SetVariable] set not permitted in query")
	}
	if err := ctx.curContract.callState.ctrState.SetData(key, value); err != nil {
		return "", err
	}
	if err := ctx.addUpdateSize(int64(types.HashIDLength + len(value))); err != nil {
		err = errors.New("uncatchable: " + err.Error())
		return "", err
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString("[Set]\n")
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("Key=%s Len=%v byte=%v\n",
			string(key), len(key), key))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("Data=%s Len=%d byte=%v\n",
			string(value), len(value), value))
	}
	return "", nil
}

func (ctx *vmContext) handleGetVariable(args []string) (result string, err error) {
	if len(args) != 2 {
		return "", errors.New("[System.GetVariable] invalid number of arguments")
	}
	key := []byte(args[0])
	blkno := args[1]
	if len(blkno) > 0 {
		bigNo, _ := new(big.Int).SetString(strings.TrimSpace(blkno), 10)
		if bigNo == nil || bigNo.Sign() < 0 {
			return "", errors.New("[System.GetVariable] invalid blockheight value :" + blkno)
		}
		blkNo := bigNo.Uint64()

		chainBlockHeight := ctx.blockInfo.No
		if chainBlockHeight == 0 {
			bestBlock, err := ctx.cdb.GetBestBlock()
			if err != nil {
				return "", errors.New("[System.GetVariable] get best block error")
			}
			chainBlockHeight = bestBlock.GetHeader().GetBlockNo()
		}
		if blkNo < chainBlockHeight {
			blk, err := ctx.cdb.GetBlockByNo(blkNo)
			if err != nil {
				return "", err
			}
			accountId := types.ToAccountID(ctx.curContract.contractId)
			contractProof, err := ctx.bs.GetAccountAndProof(accountId[:], blk.GetHeader().GetBlocksRootHash(), false)
			if err != nil {
				return "", errors.New("[System.GetVariable] failed to get snapshot state for account")
			} else if contractProof.Inclusion {
				trieKey := common.Hasher(key)
				varProof, err := ctx.bs.GetVarAndProof(trieKey, contractProof.GetState().GetStorageRoot(), false)
				if err != nil {
					return "", errors.New("[System.GetVariable] failed to get snapshot state variable in contract")
				}
				if varProof.Inclusion {
					if len(varProof.GetValue()) == 0 {
						return "", nil
					}
					return string(varProof.GetValue()), nil
				}
			}
			return "", nil
		}
	}

	data, err := ctx.curContract.callState.ctrState.GetData(key)
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", nil
	}
	return string(data), nil
}

func (ctx *vmContext) handleDelVariable(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[System.DelVariable] invalid number of arguments")
	}
	key := []byte(args[0])
	if ctx.isQuery || ctx.nestedView > 0 {
		return "", errors.New("[System.DelVariable] delete not permitted in query")
	}
	if err := ctx.curContract.callState.ctrState.DeleteData(key); err != nil {
		return "", err
	}
	if err := ctx.addUpdateSize(int64(32)); err != nil {
		err = errors.New("uncatchable: " + err.Error())
		return "", err
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString("[Del]\n")
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("Key=%s Len=%v byte=%v\n",
			string(key), len(key), key))
	}
	return "", nil
}


/*
func (ctx *vmContext) setInstCount(parent *LState, child *LState) {
	if !ctx.IsGasSystem() {
		C.vm_setinstcount(parent, C.vm_instcount(child))
	}
}

func (ctx *vmContext) setInstMinusCount(L *LState, deduc C.int) {
	if !ctx.IsGasSystem() {
		C.vm_setinstcount(L, ctx.minusCallCount(C.vm_instcount(L), deduc))
	}
}

func (ctx *vmContext) minusCallCount(curCount, deduc C.int) C.int {
	if ctx.IsGasSystem() {
		return 0
	}
	remain := curCount - deduc
	if remain <= 0 {
		remain = 1
	}
	return remain
}
*/


func (ctx *vmContext) handleCall(args []string) (result string, err error) {
	if len(args) != 5 {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Call] invalid number of arguments")
		}
		return "", errors.New("[Contract.LuaCallContract] invalid number of arguments")
	}
	contractAddress, fname, fargs, amount, gas := args[0], args[1], args[2], args[3], args[4]
	// gas => remaining gas
	// but it can also be the gas limit set by the caller contract

	// get the contract address
	cid, err := getAddressNameResolved(contractAddress, ctx.bs)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Call] invalid contractId: " + err.Error())
		}
		return "", errors.New("[Contract.LuaCallContract] invalid contractId: " + err.Error())
	}
	aid := types.ToAccountID(cid)

	// read the amount for the contract call
	amountBig, err := transformAmount(amount, ctx.blockInfo.ForkVersion)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Call] invalid amount: " + err.Error())
		}
		return "", errors.New("[Contract.LuaCallContract] invalid amount: " + err.Error())
	}

	// get the contract state
	cs, err := getContractState(ctx, cid)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Call] getAccount error: " + err.Error())
		}
		return "", errors.New("[Contract.LuaCallContract] getAccount error: " + err.Error())
	}

	// check if the contract exists
	bytecode := getContractCode(cs.ctrState, ctx.bs)
	if bytecode == nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Call] cannot find contract " + contractAddress)
		}
		return "", errors.New("[Contract.LuaCallContract] cannot find contract " + contractAddress)
	}

	// read the arguments for the contract call
	var ci types.CallInfo
	ci.Name = fname
	err = getCallInfo(&ci.Args, []byte(fargs), cid)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Call] invalid arguments: " + err.Error())
		}
		return "", errors.New("[Contract.LuaCallContract] invalid arguments: " + err.Error())
	}

	// get the remaining gas or gas limit from the parent contract
	gasLimit, err := ctx.parseGasLimit(gas)
	if err != nil {
		return "", err
	}

	// create a new executor
	ce := newExecutor(bytecode, cid, ctx, &ci, amountBig, false, false, cs.ctrState)
	defer ce.close()  // close the executor and the VM instance
	if ce.err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Call] newExecutor error: " + ce.err.Error())
		}
		return "", errors.New("[Contract.LuaCallContract] newExecutor error: " + ce.err.Error())
	}

	// set the remaining gas or gas limit from the parent contract
	ce.contractGasLimit = gasLimit

	// send the amount to the contract
	senderState := ctx.curContract.callState.accState
	receiverState := cs.accState
	if amountBig.Cmp(zeroBig) > 0 {
		if ctx.isQuery == true || ctx.nestedView > 0 {
			if CurrentForkVersion >= 5 {
				return "", errors.New("[Contract.Call] send not permitted in query")
			}
			return "", errors.New("[Contract.LuaCallContract] send not permitted in query")
		}
		if r := sendBalance(senderState, receiverState, amountBig); r != nil {
			if CurrentForkVersion >= 5 {
				return "", errors.New("[Contract.Call] " + r.Error())
			}
			return "", errors.New("[Contract.LuaCallContract] " + r.Error())
		}
	}

	seq, err := setRecoveryPoint(aid, ctx, senderState, cs, amountBig, false, false)
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[CALL Contract %v(%v) %v]\n",
			contractAddress, aid.String(), fname))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("SendBalance: %s\n", amountBig.String()))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
			senderState.Balance().String(), receiverState.Balance().String()))
	}
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Call] database error: " + err.Error())
		}
		return "", errors.New("[Contract.LuaCallContract] database error: " + err.Error())
	}

	// set the current contract info
	prevContract := ctx.curContract
	ctx.curContract = newContractInfo(cs, prevContract.contractId, cid, receiverState.RP(), amountBig)
	defer func() {
		ctx.curContract = prevContract
	}()

	// execute the contract call
	ce.call(false)

	// the result contains the used gas in the first 8 bytes
	result = ce.jsonRet

	// check if the contract call failed
	if ce.err != nil {
		// revert the contract to the previous state
		err := clearRecovery(ctx, seq, true)
		if err != nil {
			if CurrentForkVersion >= 5 {
				return result, errors.New("[Contract.Call] recovery err: " + err.Error())
			}
			return "", errors.New("[Contract.LuaCallContract] recovery err: " + err.Error())
		}
		// log some info
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
		}
		// in case of timeout, return the original error message
		switch ceErr := ce.err.(type) {
		case *VmTimeoutError:
			return result, errors.New(ceErr.Error())
		default:
			if CurrentForkVersion >= 5 {
				return result, errors.New("[Contract.Call] call err: " + ceErr.Error())
			}
			return "", errors.New("[Contract.LuaCallContract] call err: " + ceErr.Error())
		}
	}

	// release the recovery point
	if seq == 1 {
		err := clearRecovery(ctx, seq, false)
		if err != nil {
			if CurrentForkVersion >= 5 {
				return result, errors.New("[Contract.Call] recovery err: " + err.Error())
			}
			return "", errors.New("[Contract.LuaCallContract] recovery err: " + err.Error())
		}
	}

	// return the result
	return result, nil
}

func (ctx *vmContext) handleDelegateCall(args []string) (result string, err error) {
	if len(args) != 4 {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.DelegateCall] invalid number of arguments")
		}
		return "", errors.New("[Contract.LuaDelegateCallContract] invalid number of arguments")
	}
	contractAddress, fname, fargs, gas := args[0], args[1], args[2], args[3]

	var isMultiCall bool
	var cid []byte

	// get the contract address
	if contractAddress == "multicall" {
		isMultiCall = true
		fargs = fname
		fname = "execute"
		cid = ctx.curContract.contractId
	} else {
		cid, err = getAddressNameResolved(contractAddress, ctx.bs)
		if err != nil {
			if CurrentForkVersion >= 5 {
				return "", errors.New("[Contract.DelegateCall] invalid contractId: " + err.Error())
			}
			return "", errors.New("[Contract.LuaDelegateCallContract] invalid contractId: " + err.Error())
		}
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
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.DelegateCall] getContractState error: " + err.Error())
		}
		return "", errors.New("[Contract.LuaDelegateCallContract] getContractState error: " + err.Error())
	}

	// get the contract code
	var bytecode []byte
	if isMultiCall {
		bytecode = getMultiCallContractCode(contractState)
	} else {
		bytecode = getContractCode(contractState, ctx.bs)
	}
	if bytecode == nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.DelegateCall] cannot find contract " + contractAddress)
		}
		return "", errors.New("[Contract.LuaDelegateCallContract] cannot find contract " + contractAddress)
	}

	// read the arguments for the contract call
	var ci types.CallInfo
	if isMultiCall {
		err = getMultiCallInfo(&ci, []byte(fargs))
	} else {
		ci.Name = fname
		err = getCallInfo(&ci.Args, []byte(fargs), cid)
	}
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.DelegateCall] invalid arguments: " + err.Error())
		}
		return "", errors.New("[Contract.LuaDelegateCallContract] invalid arguments: " + err.Error())
	}

	// get the remaining gas or gas limit from the parent contract
	gasLimit, err := ctx.parseGasLimit(gas)
	if err != nil {
		return "", err
	}

	// create a new executor
	ce := newExecutor(bytecode, cid, ctx, &ci, zeroBig, false, false, contractState)
	defer ce.close()  // close the executor and the VM instance
	if ce.err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.DelegateCall] newExecutor error: " + ce.err.Error())
		}
		return "", errors.New("[Contract.LuaDelegateCallContract] newExecutor error: " + ce.err.Error())
	}

	// set the remaining gas or gas limit from the parent contract
	ce.contractGasLimit = gasLimit

	seq, err := setRecoveryPoint(aid, ctx, nil, ctx.curContract.callState, zeroBig, false, false)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.DelegateCall] database error: " + err.Error())
		}
		return "", errors.New("[Contract.LuaDelegateCallContract] database error: " + err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[DELEGATECALL Contract %v %v]\n", contractAddress, fname))
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
	}

	// execute the contract call
	ce.call(false)

	// the result contains the used gas in the first 8 bytes
	result = ce.jsonRet

	// check if the contract call failed
	if ce.err != nil {
		// revert the contract to the previous state
		err := clearRecovery(ctx, seq, true)
		if err != nil {
			if CurrentForkVersion >= 5 {
				return result, errors.New("[Contract.DelegateCall] recovery error: " + err.Error())
			}
			return "", errors.New("[Contract.LuaDelegateCallContract] recovery error: " + err.Error())
		}
		// log some info
		if ctx.traceFile != nil {
			_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
		}
		// in case of timeout, return the original error message
		switch ceErr := ce.err.(type) {
		case *VmTimeoutError:
			return result, errors.New(ceErr.Error())
		default:
			if CurrentForkVersion >= 5 {
				return result, errors.New("[Contract.DelegateCall] call error: " + ce.err.Error())
			}
			return "", errors.New("[Contract.LuaDelegateCallContract] call error: " + ce.err.Error())
		}
	}

	// release the recovery point
	if seq == 1 {
		err := clearRecovery(ctx, seq, false)
		if err != nil {
			if CurrentForkVersion >= 5 {
				return result, errors.New("[Contract.DelegateCall] recovery error: " + err.Error())
			}
			return "", errors.New("[Contract.LuaDelegateCallContract] recovery error: " + err.Error())
		}
	}

	// return the result
	return result, nil
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

func (ctx *vmContext) handleSend(args []string) (result string, err error) {
	if len(args) != 3 {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Send] invalid number of arguments")
		}
		return "", errors.New("[Contract.LuaSendAmount] invalid number of arguments")
	}
	contractAddress, amount, gas := args[0], args[1], args[2]

	// read the amount to be sent
	amountBig, err := transformAmount(amount, ctx.blockInfo.ForkVersion)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Send] invalid amount: " + err.Error())
		}
		return "", errors.New("[Contract.LuaSendAmount] invalid amount: " + err.Error())
	}

	// cannot send amount in query
	if (ctx.isQuery == true || ctx.nestedView > 0) && amountBig.Cmp(zeroBig) > 0 {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Send] send not permitted in query")
		}
		return "", errors.New("[Contract.LuaSendAmount] send not permitted in query")
	}

	// get the receiver account
	cid, err := getAddressNameResolved(contractAddress, ctx.bs)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Send] invalid contractId: " + err.Error())
		}
		return "", errors.New("[Contract.LuaSendAmount] invalid contractId: " + err.Error())
	}

	// get the receiver state
	aid := types.ToAccountID(cid)
	cs, err := getCallState(ctx, cid)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Send] getAccount error: " + err.Error())
		}
		return "", errors.New("[Contract.LuaSendAmount] getAccount error: " + err.Error())
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
				if CurrentForkVersion >= 5 {
					return "", errors.New("[Contract.Send] getContractState error: " + err.Error())
				}
				return "", errors.New("[Contract.LuaSendAmount] getContractState error: " + err.Error())
			}
		}

		// set the function to be called
		var ci types.CallInfo
		ci.Name = "default"

		// get the contract code
		bytecode := getContractCode(cs.ctrState, ctx.bs)
		if bytecode == nil {
			if CurrentForkVersion >= 5 {
				return "", errors.New("[Contract.Send] cannot find contract:" + contractAddress)
			}
			return "", errors.New("[Contract.LuaSendAmount] cannot find contract:" + contractAddress)
		}

		// get the remaining gas or gas limit from the parent contract
		gasLimit, err := ctx.parseGasLimit(gas)
		if err != nil {
			return "", err
		}

		// create a new executor
		ce := newExecutor(bytecode, cid, ctx, &ci, amountBig, false, false, cs.ctrState)
		defer ce.close()  // close the executor and the VM instance
		if ce.err != nil {
			if CurrentForkVersion >= 5 {
				return "", errors.New("[Contract.Send] newExecutor error: " + ce.err.Error())
			}
			return "", errors.New("[Contract.LuaSendAmount] newExecutor error: " + ce.err.Error())
		}

		// set the remaining gas or gas limit from the parent contract
		ce.contractGasLimit = gasLimit

		// send the amount to the contract
		if amountBig.Cmp(zeroBig) > 0 {
			if r := sendBalance(senderState, receiverState, amountBig); r != nil {
				if CurrentForkVersion >= 5 {
					return "", errors.New("[Contract.Send] " + r.Error())
				}
				return "", errors.New("[Contract.LuaSendAmount] " + r.Error())
			}
		}

		// create a recovery point
		seq, err := setRecoveryPoint(aid, ctx, senderState, cs, amountBig, false, false)
		if err != nil {
			if CurrentForkVersion >= 5 {
				return "", errors.New("[Contract.Send] database error: " + err.Error())
			}
			return "", errors.New("[Contract.LuaSendAmount] database error: " + err.Error())
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
		prevContract := ctx.curContract
		ctx.curContract = newContractInfo(cs, prevContract.contractId, cid, receiverState.RP(), amountBig)
		defer func() {
			ctx.curContract = prevContract
		}()

		// execute the contract call
		ce.call(false)

		// the result contains the used gas in the first 8 bytes
		result = ce.jsonRet

		// check if the contract call failed
		if ce.err != nil {
			// revert the contract to the previous state
			err := clearRecovery(ctx, seq, true)
			if err != nil {
				if CurrentForkVersion >= 5 {
					return result, errors.New("[Contract.Send] recovery err: " + err.Error())
				}
				return "", errors.New("[Contract.LuaSendAmount] recovery err: " + err.Error())
			}
			// log some info
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
			}
			// in case of timeout, return the original error message
			switch ceErr := ce.err.(type) {
			case *VmTimeoutError:
				return result, errors.New(ceErr.Error())
			default:
				if CurrentForkVersion >= 5 {
					return result, errors.New("[Contract.Send] call err: " + ce.err.Error())
				}
				return "", errors.New("[Contract.LuaSendAmount] call err: " + ce.err.Error())
			}
		}

		// release the recovery point
		if seq == 1 {
			err := clearRecovery(ctx, seq, false)
			if err != nil {
				if CurrentForkVersion >= 5 {
					return result, errors.New("[Contract.Send] recovery err: " + err.Error())
				}
				return "", errors.New("[Contract.LuaSendAmount] recovery err: " + err.Error())
			}
		}

		// the transfer and contract call succeeded
		return result, nil
	}

	// the receiver is not a contract, just send the amount

	// if amount is zero, do nothing
	if amountBig.Cmp(zeroBig) == 0 {
		return "", nil
	}

	// send the amount to the receiver
	if r := sendBalance(senderState, receiverState, amountBig); r != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.Send] " + r.Error())
		}
		return "", errors.New("[Contract.LuaSendAmount] " + r.Error())
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
			senderState.Balance().String(), receiverState.Balance().String()))
	}

	return "", nil
}

func (ctx *vmContext) handlePrint(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[Contract.Print] invalid number of arguments")
	}
	ctrLgr.Info().Str("Contract SystemPrint", types.EncodeAddress(ctx.curContract.contractId)).Msg(args[0])
	return "", nil
}

func (ctx *vmContext) handleSetRecoveryPoint() (result string, err error) {
	if ctx.isQuery || ctx.nestedView > 0 {
		return "", nil
	}
	curContract := ctx.curContract
	// if it is the multicall code, ignore
	if curContract.callState.ctrState.IsMultiCall() {
		return "", nil
	}
	aid := types.ToAccountID(curContract.contractId)
	seq, err := setRecoveryPoint(aid, ctx, nil, curContract.callState, zeroBig, false, false)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.SetRecoveryPoint] database error: " + err.Error())
		}
		return "", errors.New("[Contract.pcall] database error: " + err.Error())
	}
	if ctx.traceFile != nil {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[pcall] snapshot set %d\n", seq))
	}
	return strconv.Itoa(seq), nil
}

func clearRecovery(ctx *vmContext, start int, revert bool) error {
	item := ctx.lastRecoveryEntry
	for {
		if revert {
			if item.revertState(ctx) != nil {
				return errors.New("database error")
			}
		}
		if item.seq == start {
			if revert || item.prev == nil {
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

func (ctx *vmContext) handleClearRecovery(args []string) (result string, err error) {
	if len(args) != 2 {
		return "", errors.New("[Contract.ClearRecovery] invalid number of arguments")
	}
	start, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.New("[Contract.ClearRecovery] invalid start")
	}
	revert, err := strconv.ParseBool(args[1])
	if err != nil {
		return "", errors.New("[Contract.ClearRecovery] invalid revert")
	}
	err = clearRecovery(ctx, start, revert)
	if err != nil {
		return "", err
	}
	if ctx.traceFile != nil && revert == true {
		_, _ = ctx.traceFile.WriteString(fmt.Sprintf("pcall recovery snapshot: %d\n", start))
	}
	return "", nil
}

func (ctx *vmContext) handleGetBalance(args []string) (result string, err error) {
	if len(args) != 1 {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.GetBalance] invalid number of arguments")
		}
		return "", errors.New("[Contract.LuaGetBalance] invalid number of arguments")
	}
	contractAddress := args[0]
	if contractAddress == "" {
		return ctx.curContract.callState.ctrState.GetBalanceBigInt().String(), nil
	}
	cid, err := getAddressNameResolved(contractAddress, ctx.bs)
	if err != nil {
		if CurrentForkVersion >= 5 {
			return "", errors.New("[Contract.GetBalance] invalid contractId: " + err.Error())
		}
		return "", errors.New("[Contract.LuaGetBalance] invalid contractId: " + err.Error())
	}
	aid := types.ToAccountID(cid)
	cs := ctx.callState[aid]
	if cs == nil {
		as, err := ctx.bs.GetAccountState(aid)
		if err != nil {
			if CurrentForkVersion >= 5 {
				return "", errors.New("[Contract.GetBalance] getAccount error: " + err.Error())
			}
			return "", errors.New("[Contract.LuaGetBalance] getAccount error: " + err.Error())
		}
		return as.GetBalanceBigInt().String(), nil
	}
	return cs.accState.Balance().String(), nil
}



func (ctx *vmContext) getContractId() string {
	return types.EncodeAddress(ctx.curContract.contractId)
}

func (ctx *vmContext) getSender() string {
	return types.EncodeAddress(ctx.curContract.sender)
}

func (ctx *vmContext) getAmount() string {
	return ctx.curContract.amount.String()
}

func (ctx *vmContext) getTxHash() string {
	return base58.Encode(ctx.txHash)
}

func (ctx *vmContext) getOrigin() string {
	return types.EncodeAddress(ctx.origin)
}

func (ctx *vmContext) getIsFeeDelegation() bool {
	return ctx.isFeeDelegation
}

func (ctx *vmContext) getBlockNo() uint64 {
	return ctx.blockInfo.No
}

func (ctx *vmContext) getPrevBlockHash() string {
	return base58.Encode(ctx.blockInfo.PrevBlockHash)
}

func (ctx *vmContext) getTimestamp() uint64 {
	return uint64(ctx.blockInfo.Ts / 1e9)
}



func (ctx *vmContext) handleGetContractId() (result string, err error) {
	//setInstMinusCount(ctx, L, 1000)
	return types.EncodeAddress(ctx.curContract.contractId), nil
}

func (ctx *vmContext) handleGetSender() (result string, err error) {
	//setInstMinusCount(ctx, L, 1000)
	return types.EncodeAddress(ctx.curContract.sender), nil
}

func (ctx *vmContext) handleGetAmount() (result string, err error) {
	return ctx.curContract.amount.String(), nil
}

func (ctx *vmContext) handleGetTxHash() (result string, err error) {
	return base58.Encode(ctx.txHash), nil
}

func (ctx *vmContext) handleGetOrigin() (result string, err error) {
	//setInstMinusCount(ctx, L, 1000)
	return types.EncodeAddress(ctx.origin), nil
}

func (ctx *vmContext) handleIsFeeDelegation() (result string, err error) {
	if ctx.isFeeDelegation {
		return "1", nil
	}
	return "0", nil
}

func (ctx *vmContext) handleGetBlockNo() (result string, err error) {
	return strconv.Itoa(int(ctx.blockInfo.No)), nil
}

func (ctx *vmContext) handleGetPrevBlockHash() (result string, err error) {
	return base58.Encode(ctx.blockInfo.PrevBlockHash), nil
}

func (ctx *vmContext) handleGetTimeStamp() (result string, err error) {
	return strconv.FormatInt(ctx.blockInfo.Ts / 1e9, 10), nil
}


//export checkDbExecContext
func checkDbExecContext(service C.int) bool {
	// check if service is valid
	if service < 0 || service >= C.int(len(contexts)) {
		return false
	}
	if PubNet {
		return false
	}
	return true
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

	// make sure that this go routine does not migrate to another thread
	runtime.LockOSThread()

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

func (ctx *vmContext) handleCryptoSha256(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("[Contract.CryptoSha256] invalid number of arguments")
	}
	data := []byte(args[0])
	if checkHexString(string(data)) {
		dataStr := data[2:]
		var err error
		data, err = hex.Decode(string(dataStr))
		if err != nil {
			return "", fmt.Errorf("[Contract.CryptoSha256] hex decoding error: %v", err)
		}
	}
	h := sha256.New()
	h.Write(data)
	resultHash := h.Sum(nil)
	return "0x" + hex.Encode(resultHash), nil
}

func decodeHex(hexStr string) ([]byte, error) {
	if checkHexString(hexStr) {
		hexStr = hexStr[2:]
	}
	return hex.Decode(hexStr)
}

func (ctx *vmContext) handleECVerify(args []string) (result string, err error) {
	if len(args) != 3 {
		return "", errors.New("[Contract.EcVerify] invalid number of arguments")
	}
	msg, sig, addr := args[0], args[1], args[2]
	bMsg, err := decodeHex(msg)
	if err != nil {
		return "", errors.New("[Contract.EcVerify] invalid message format: " + err.Error())
	}
	bSig, err := decodeHex(sig)
	if err != nil {
		return "", errors.New("[Contract.EcVerify] invalid signature format: " + err.Error())
	}

	var pubKey *btcec.PublicKey
	var verifyResult bool
	isAergo := len(addr) == types.EncodedAddressLength

	/*Aergo Address*/
	if isAergo {
		bAddress, err := types.DecodeAddress(addr)
		if err != nil {
			return "", errors.New("[Contract.EcVerify] invalid aergo address: " + err.Error())
		}
		pubKey, err = btcec.ParsePubKey(bAddress)
		if err != nil {
			return "", errors.New("[Contract.EcVerify] error parsing pubKey: " + err.Error())
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
			return "", errors.New("[Contract.EcVerify] error recoverCompact: " + err.Error())
		}
		if pubKey != nil {
			verifyResult = pubKey.IsEqual(pub)
		} else {
			bAddress, err := decodeHex(addr)
			if err != nil {
				return "", errors.New("[Contract.EcVerify] invalid Ethereum address: " + err.Error())
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
			return "", errors.New("[Contract.EcVerify] error parsing signature: " + err.Error())
		}
		if pubKey == nil {
			return "", errors.New("[Contract.EcVerify] error recovering pubKey")
		}
		verifyResult = sign.Verify(bMsg, pubKey)
	}
	if verifyResult {
		return "1", nil
	}
	return "0", nil
}

func luaCryptoToBytes(data []byte) ([]byte, bool) {
	var d []byte
	isHex := checkHexString(string(data))
	if isHex {
		var err error
		d, err = hex.Decode(string(data[2:]))
		if err != nil {
			isHex = false
		}
	}
	if !isHex {
		d = data
	}
	return d, isHex
}

func cryptoBytesToRlpObject(data []byte) rlpObject {
	// read the first byte to determine the type of the RLP object
	rlpType := data[0]
	data = data[1:]
	// convert the remaining bytes to the appropriate type
	if rlpType == C.RLP_TSTRING {
		return rlpString(data)
	}
	// if the type is not a list, return nil
	if rlpType != C.RLP_TLIST {
		return nil
	}
	// the type is a list. deserialize it
	items, err := msg.DeserializeMessage(data)
	if err != nil {
		return nil
	}
	// convert the items to rlpList
	list := make(rlpList, len(items))
	for i, item := range items {
		list[i] = rlpString(item)
	}
	return list
}

func (ctx *vmContext) handleCryptoVerifyEthStorageProof(args []string) (result string, err error) {
	if len(args) != 4 {
		return "", errors.New("[Contract.CryptoVerifyEthStorageProof] invalid number of arguments")
	}
	key := []byte(args[0])
	value := cryptoBytesToRlpObject([]byte(args[1]))
	hash := []byte(args[2])
	proof, err := msg.DeserializeMessage([]byte(args[3]))
	if err != nil {
		return "", errors.New("[Contract.CryptoVerifyEthStorageProof] error deserializing proof: " + err.Error())
	}
	proofBytes := make([][]byte, len(proof))
	for i, p := range proof {
		proofBytes[i] = []byte(p)
	}

	if verifyEthStorageProof(key, value, hash, proofBytes) {
		return "1", nil
	}
	return "0", nil
}

func (ctx *vmContext) handleCryptoKeccak256(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[Contract.CryptoKeccak256] invalid number of arguments")
	}
	data, isHex := luaCryptoToBytes([]byte(args[0]))
	h := keccak256(data)
	if isHex {
		hexb := "0x" + hex.Encode(h)
		return hexb, nil
	} else {
		return string(h), nil
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

func (ctx *vmContext) handleDeploy(args []string) (result string, err error) {
	if len(args) != 4 {
		return "", errors.New("[Contract.Deploy] invalid number of arguments")
	}
	codeOrAddress, fargs, amount, gas := args[0], args[1], args[2], args[3]

	if ctx.isQuery || ctx.nestedView > 0 {
		return "", errors.New("[Contract.Deploy] deploy not permitted in query")
	}
	bs := ctx.bs

	// contract code
	var codeABI []byte
	var sourceCode []byte

	// check if contract name or address is given
	cid, err := getAddressNameResolved(codeOrAddress, bs)
	if err == nil {
		// check if contract exists
		contractState, err := getOnlyContractState(ctx, cid)
		if err != nil {
			return "", errors.New("[Contract.Deploy] " + err.Error())
		}
		// read the contract code
		codeABI, err = contractState.GetCode()
		if err != nil {
			return "", errors.New("[Contract.Deploy] " + err.Error())
		} else if len(codeABI) == 0 {
			return "", errors.New("[Contract.Deploy] not found code")
		}
		if ctx.blockInfo.ForkVersion >= 4 {
			sourceCode = contractState.GetSourceCode()
		}
	}

	//! maybe not needed on hardfork 5, if using Lua for new contracts
	// but it could at least check the code for validity

	// compile contract code if not found
	if len(codeABI) == 0 {
		codeABI, err = Compile(codeOrAddress, true)
		if err != nil {
			// check if string contains timeout error
			if strings.Contains(err.Error(), C.ERR_BF_TIMEOUT) {
				return "", err  //errors.New(C.ERR_BF_TIMEOUT)
			} else if err == ErrVmStart {
				return "", errors.New("[Contract.Deploy] get luaState error")
			}
			return "", errors.New("[Contract.Deploy] compile error: " + err.Error())
		}
		if ctx.blockInfo.ForkVersion >= 4 {
			sourceCode = []byte(codeOrAddress)
		}
	}

	err = ctx.addUpdateSize(int64(len(codeABI) + len(sourceCode)))
	if err != nil {
		return "", errors.New("[Contract.Deploy] " + err.Error())
	}

	// create account for the contract
	creator := ctx.curContract.callState.accState
	newContract, err := state.CreateAccountState(CreateContractID(ctx.curContract.contractId, creator.Nonce()), bs.StateDB)
	if err != nil {
		return "", errors.New("[Contract.Deploy] " + err.Error())
	}
	contractState, err := statedb.OpenContractState(newContract.ID(), newContract.State(), bs.StateDB)
	if err != nil {
		return "", errors.New("[Contract.Deploy] " + err.Error())
	}

	cs := &callState{isCallback: true, isDeploy: true, ctrState: contractState, accState: newContract}
	ctx.callState[newContract.AccountID()] = cs

	// read the amount transferred to the contract
	amountBig, err := transformAmount(amount, ctx.blockInfo.ForkVersion)
	if err != nil {
		return "", errors.New("[Contract.Deploy] value not proper format: " + err.Error())
	}

	// read the arguments for the constructor call
	var ci types.CallInfo
	err = getCallInfo(&ci.Args, []byte(fargs), newContract.ID())
	if err != nil {
		return "", errors.New("[Contract.Deploy] invalid args: " + err.Error())
	}

	// send the amount to the contract
	senderState := ctx.curContract.callState.accState
	receiverState := cs.accState
	if amountBig.Cmp(zeroBig) > 0 {
		if rv := sendBalance(senderState, receiverState, amountBig); rv != nil {
			return "", errors.New("[Contract.Deploy] " + rv.Error())
		}
	}

	// create a recovery point
	seq, err := setRecoveryPoint(newContract.AccountID(), ctx, senderState, cs, amountBig, false, true)
	if err != nil {
		return "", errors.New("[System.DeployContract] DB err: " + err.Error())
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
	prevContract := ctx.curContract
	ctx.curContract = newContractInfo(cs, prevContract.contractId, newContract.ID(), receiverState.RP(), amountBig)
	defer func() {
		ctx.curContract = prevContract
	}()

	bytecode := util.LuaCode(codeABI).ByteCode()

	// save the contract code
	err = contractState.SetCode(sourceCode, codeABI)
	if err != nil {
		return "", errors.New("[Contract.Deploy] " + err.Error())
	}

	// save the contract creator
	err = contractState.SetData(dbkey.CreatorMeta(), []byte(types.EncodeAddress(prevContract.contractId)))
	if err != nil {
		return "", errors.New("[Contract.Deploy] " + err.Error())
	}

	// get the remaining gas or gas limit from the parent contract
	gasLimit, err := ctx.parseGasLimit(gas)
	if err != nil {
		return "", err
	}

	// create a new executor
	ce := newExecutor(bytecode, newContract.ID(), ctx, &ci, amountBig, true, false, contractState)
	defer ce.close()  // close the executor and the VM instance
	if ce.err != nil {
		return "", errors.New("[Contract.Deploy] newExecutor Error: " + ce.err.Error())
	}

	// set the remaining gas or gas limit from the parent contract
	ce.contractGasLimit = gasLimit

	// increment the nonce of the creator
	senderState.SetNonce(senderState.Nonce() + 1)

	addr := types.EncodeAddress(newContract.ID())

	if ce != nil {
		// run the constructor
		ce.call(false)

		// the result contains the used gas in the first 8 bytes
		result = ce.jsonRet

		// check if the execution was successful
		if ce.err != nil {
			// revert the contract to the previous state
			err := clearRecovery(ctx, seq, true)
			if err != nil {
				return result, errors.New("[Contract.Deploy] recovery error: " + err.Error())
			}
			// log some info
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("recovery snapshot: %d\n", seq))
			}
			// in case of timeout, return the original error message
			switch ceErr := ce.err.(type) {
			case *VmTimeoutError:
				return result, errors.New(ceErr.Error())
			default:
				return result, errors.New("[Contract.Deploy] call err: " + ce.err.Error())
			}
		}
	}

	// release the recovery point
	if seq == 1 {
		err := clearRecovery(ctx, seq, false)
		if err != nil {
			return result, errors.New("[Contract.Deploy] recovery error: " + err.Error())
		}
	}

	// the result already contains a JSON array
	// insert the contract address before the other returned values
	// the first 8 bytes contain the used gas
	result = result[:8] + `["` + addr + `",` + result[9:]

	return result, nil
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

func (ctx *vmContext) handleRandomInt(args []string) (result string, err error) {
	if len(args) != 2 {
		return "", errors.New("[Contract.RandomInt] invalid number of arguments")
	}
	min, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.New("[Contract.RandomInt] invalid min")
	}
	max, err := strconv.Atoi(args[1])
	if err != nil {
		return "", errors.New("[Contract.RandomInt] invalid max")
	}
	if ctx.seed == nil {
		setRandomSeed(ctx)
	}
	return strconv.Itoa(ctx.seed.Intn(max+1-min) + min), nil
}

func (ctx *vmContext) handleEvent(args []string) (result string, err error) {
	if len(args) != 2 {
		return "", errors.New("[Contract.Event] invalid number of arguments")
	}
	eventName, eventArgs := args[0], args[1]
	if ctx.isQuery || ctx.nestedView > 0 {
		return "", errors.New("[Contract.Event] event not permitted in query")
	}
	if ctx.eventCount >= maxEventCnt(ctx) {
		return "", errors.New(fmt.Sprintf("[Contract.Event] exceeded the maximum number of events(%d)", maxEventCnt(ctx)))
	}
	if len(eventName) > maxEventNameSize {
		return "", errors.New(fmt.Sprintf("[Contract.Event] exceeded the maximum length of event name(%d)", maxEventNameSize))
	}
	if len(eventArgs) > maxEventArgSize {
		return "", errors.New(fmt.Sprintf("[Contract.Event] exceeded the maximum length of event args(%d)", maxEventArgSize))
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
	return "", nil
}

func (ctx *vmContext) handleToPubkey(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[Contract.ToPubkey] invalid number of arguments")
	}
	address := args[0]
	// check the length of address
	if len(address) != types.EncodedAddressLength {
		return "", errors.New("[Contract.ToPubkey] invalid address length")
	}
	// decode the address in string format to bytes (public key)
	pubkey, err := types.DecodeAddress(address)
	if err != nil {
		return "", errors.New("[Contract.ToPubkey] invalid address")
	}
	// return the public key in hex format
	return "0x" + hex.Encode(pubkey), nil
}

func (ctx *vmContext) handleToAddress(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[Contract.ToAddress] invalid number of arguments")
	}
	pubkey := args[0]
	// decode the pubkey in hex format to bytes
	pubkeyBytes, err := decodeHex(pubkey)
	if err != nil {
		return "", errors.New("[Contract.ToAddress] invalid public key")
	}
	// check the length of pubkey
	if len(pubkeyBytes) != types.AddressLength {
		return "", errors.New("[Contract.ToAddress] invalid public key length")
		// or convert the pubkey to compact format - SerializeCompressed()
	}
	// encode the pubkey in bytes to an address in string format
	address := types.EncodeAddress(pubkeyBytes)
	// return the address
	return address, nil
}

func (ctx *vmContext) handleIsContract(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[Contract.IsContract] invalid number of arguments")
	}
	contractAddress := args[0]

	cid, err := getAddressNameResolved(contractAddress, ctx.bs)
	if err != nil {
		return "", errors.New("[Contract.IsContract] invalid contractId: " + err.Error())
	}

	cs, err := getCallState(ctx, cid)
	if err != nil {
		return "", errors.New("[Contract.IsContract] getAccount error: " + err.Error())
	}

	return strconv.Itoa(len(cs.accState.CodeHash())), nil
}

func (ctx *vmContext) handleNameResolve(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[Contract.NameResolve] invalid number of arguments")
	}
	account := args[0]  // account name or address
	var addr []byte
	if len(account) == types.EncodedAddressLength {
		// also checks if valid address
		addr, err = types.DecodeAddress(account)
	} else {
		addr, err = name.Resolve(ctx.bs, []byte(account), false)
	}
	if err != nil {
		return "", errors.New("[Contract.NameResolve] " + err.Error())
	}
	return types.EncodeAddress(addr), nil
}

func (ctx *vmContext) handleGovernance(args []string) (result string, err error) {
	if len(args) != 2 {
		return "", errors.New("[Contract.Governance] invalid number of arguments")
	}
	gType, arg := args[0], args[1]

	if ctx.isQuery || ctx.nestedView > 0 {
		return "", errors.New("[Contract.Governance] governance not permitted in query")
	}

	var amountBig *big.Int
	var payload []byte

	switch gType {
	case "S", "U":
		var err error
		amountBig, err = transformAmount(arg, ctx.blockInfo.ForkVersion)
		if err != nil {
			return "", errors.New("[Contract.Governance] invalid amount: " + err.Error())
		}
		if gType == "S" {
			payload = []byte(fmt.Sprintf(`{"Name":"%s"}`, types.Opstake.Cmd()))
		} else {
			payload = []byte(fmt.Sprintf(`{"Name":"%s"}`, types.Opunstake.Cmd()))
		}
	case "V":
		amountBig = zeroBig
		payload = []byte(fmt.Sprintf(`{"Name":"%s","Args":%s}`, types.OpvoteBP.Cmd(), arg))
	case "D":
		amountBig = zeroBig
		payload = []byte(fmt.Sprintf(`{"Name":"%s","Args":%s}`, types.OpvoteDAO.Cmd(), arg))
	}

	cid := []byte(types.AergoSystem)
	aid := types.ToAccountID(cid)
	scsState, err := getContractState(ctx, cid)
	if err != nil {
		return "", errors.New("[Contract.Governance] getAccount error: " + err.Error())
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
		return "", errors.New("[Contract.Governance] error: " + err.Error())
	}

	// create a recovery point
	seq, err := setRecoveryPoint(aid, ctx, senderState, scsState, zeroBig, false, false)
	if err != nil {
		return "", errors.New("[Contract.Governance] database error: " + err.Error())
	}

	// execute the system transaction
	events, err := system.ExecuteSystemTx(scsState.ctrState, &txBody, senderState, receiverState, ctx.blockInfo)
	if err != nil {
		// revert the contract to the previous state
		rErr := clearRecovery(ctx, seq, true)
		if rErr != nil {
			return "", errors.New("[Contract.Governance] recovery error: " + rErr.Error())
		}
		return "", errors.New("[Contract.Governance] error: " + err.Error())
	}

	// release the recovery point
	if seq == 1 {
		err := clearRecovery(ctx, seq, false)
		if err != nil {
			return "", errors.New("[Contract.Governance] recovery error: " + err.Error())
		}
	}

	// add the events to the context
	ctx.eventCount += int32(len(events))
	ctx.events = append(ctx.events, events...)

	if ctx.lastRecoveryEntry != nil {
		if gType == "S" {
			seq, _ = setRecoveryPoint(aid, ctx, senderState, scsState, amountBig, true, false)
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[GOVERNANCE]aid(%s)\n", aid.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("staking : %s\n", amountBig.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
					senderState.Balance().String(), receiverState.Balance().String()))
			}
		} else if gType == "U" {
			seq, _ = setRecoveryPoint(aid, ctx, receiverState, ctx.curContract.callState, amountBig, true, false)
			if ctx.traceFile != nil {
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("[GOVERNANCE]aid(%s)\n", aid.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("snapshot set %d\n", seq))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("unstaking : %s\n", amountBig.String()))
				_, _ = ctx.traceFile.WriteString(fmt.Sprintf("After sender: %s receiver: %s\n",
					senderState.Balance().String(), receiverState.Balance().String()))
			}
		}
	}

	return "", nil
}


////////////////////////////////////////////////////////////////////////////////

func (ctx *vmContext) handleDbExec(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.Exec] invalid number of arguments")
	}
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_db_exec(&cReq, (*C.char)(unsafe.Pointer(&[]byte(args[0])[0])), C.int(len(args[0])))
	return processResult(&cReq)
}

func (ctx *vmContext) handleDbQuery(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.Query] invalid number of arguments")
	}
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_db_query(&cReq, (*C.char)(unsafe.Pointer(&[]byte(args[0])[0])), C.int(len(args[0])))
	return processResult(&cReq)
}

func (ctx *vmContext) handleDbPrepare(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.Prepare] invalid number of arguments")
	}
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_db_prepare(&cReq, (*C.char)(unsafe.Pointer(&[]byte(args[0])[0])), C.int(len(args[0])))
	return processResult(&cReq)
}

//stmtExec
func (ctx *vmContext) handleStmtExec(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.StmtExec] invalid number of arguments")
	}
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_stmt_exec(&cReq, (*C.char)(unsafe.Pointer(&[]byte(args[0])[0])), C.int(len(args[0])))
	return processResult(&cReq)
}

//stmtQuery
func (ctx *vmContext) handleStmtQuery(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.StmtQuery] invalid number of arguments")
	}
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_stmt_query(&cReq, (*C.char)(unsafe.Pointer(&[]byte(args[0])[0])), C.int(len(args[0])))
	return processResult(&cReq)
}

//stmtColumnInfo
func (ctx *vmContext) handleStmtColumnInfo(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.StmtColumnInfo] invalid number of arguments")
	}
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_stmt_column_info(&cReq, (*C.char)(unsafe.Pointer(&[]byte(args[0])[0])), C.int(len(args[0])))
	return processResult(&cReq)
}

//rsNext
func (ctx *vmContext) handleRsNext(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.RsNext] invalid number of arguments")
	}
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_rs_next(&cReq, (*C.char)(unsafe.Pointer(&[]byte(args[0])[0])), C.int(len(args[0])))
	return processResult(&cReq)
}

//rsGet
func (ctx *vmContext) handleRsGet(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.RsGet] invalid number of arguments")
	}
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_rs_get(&cReq, (*C.char)(unsafe.Pointer(&[]byte(args[0])[0])), C.int(len(args[0])))
	return processResult(&cReq)
}

/*
//rsColumnInfo
func (ctx *vmContext) handleRsColumnInfo(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.RsColumnInfo] invalid number of arguments")
	}
	col_id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.New("[DB.RsColumnInfo] invalid column id")
	}

	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_rs_column_info(&cReq, C.int(col_id))
	return processResult(&cReq)
}
*/

/*
//rsClose
func (ctx *vmContext) handleRsClose(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.RsClose] invalid number of arguments")
	}
	query_id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.New("[DB.RsClose] invalid query id")
	}

	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_rs_close(&cReq, C.int(query_id))
	return processResult(&cReq)
}
*/

//lastInsertRowid
func (ctx *vmContext) handleLastInsertRowid(args []string) (result string, err error) {
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_last_insert_rowid(&cReq)
	return processResult(&cReq)
}

//dbOpenWithSnapshot
func (ctx *vmContext) handleDbOpenWithSnapshot(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[DB.DbOpenWithSnapshot] invalid number of arguments")
	}
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_db_open_with_snapshot(&cReq, (*C.char)(unsafe.Pointer(&[]byte(args[0])[0])), C.int(len(args[0])))
	return processResult(&cReq)
}

//dbGetSnapshot
func (ctx *vmContext) handleDbGetSnapshot(args []string) (result string, err error) {
	var cReq C.request
	cReq.service = C.int(ctx.service)
	C.handle_db_get_snapshot(&cReq)
	return processResult(&cReq)
}

func processResult(cReq *C.request) (result string, err error) {
	if cReq.result.ptr != nil {
		result = C.GoStringN(cReq.result.ptr, cReq.result.len)
		C.free(unsafe.Pointer(cReq.result.ptr))
	}
	if cReq.error != nil {
		errstr := C.GoString(cReq.error)
		C.free(unsafe.Pointer(cReq.error))
		err = errors.New(errstr)
	}
	return result, err
}



////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////


//export isPublic
func isPublic() C.int {
	if PubNet {
		return C.int(1)
	} else {
		return C.int(0)
	}
}





// this is only used at server side, by db_module.c

//export luaIsView
func luaIsView(service C.int) C.bool {
	ctx := contexts[service]
	return C.bool(ctx.nestedView > 0)
}




// checks whether the block creation timeout occurred
//
func checkTimeout(service int) bool {

	// only check timeout for the block factory
	if service != BlockFactory {
		return false
	}

	ctx := contexts[service]
	select {
	case <-ctx.execCtx.Done():
		return true
	default:
		return false
	}

}

//export LuaGetDbHandleSnap
func LuaGetDbHandleSnap(service C.int, snapshot *C.char) *C.char {
	ctx := contexts[service]

	curContract := ctx.curContract
	callState := curContract.callState

	if ctx.isQuery != true {
		return C.CString("[Contract.SetDbSnap] not permitted in transaction")
	}

	if callState.tx != nil {
		return C.CString("[Contract.SetDbSnap] transaction already started")
	}

	rp, err := strconv.ParseUint(C.GoString(snapshot), 10, 64)
	if err != nil {
		return C.CString("[Contract.SetDbSnap] snapshot is not valid: " + C.GoString(snapshot))
	}

	aid := types.ToAccountID(curContract.contractId)
	tx, err := beginReadOnly(aid.String(), rp)
	if err != nil {
		return C.CString("[Contract.SetDbSnap] Error Begin SQL Transaction")
	}

	callState.tx = tx
	return nil
}

//export LuaGetDbSnapshot
func LuaGetDbSnapshot(service C.int) *C.char {
	ctx := contexts[service]
	return C.CString(strconv.FormatUint(ctx.curContract.rp, 10))
}




func (ctx *vmContext) handleGetStaking(args []string) (result string, err error) {
	if len(args) != 1 {
		return "", errors.New("[Contract.GetStaking] invalid number of arguments")
	}
	addr := args[0]

	systemcs, err := statedb.GetSystemAccountState(ctx.bs.StateDB)
	if err != nil {
		return "", err
	}

	namecs, err := statedb.GetNameAccountState(ctx.bs.StateDB)
	if err != nil {
		return "", err
	}

	staking, err := system.GetStaking(systemcs, name.GetAddress(namecs, types.ToAddress(addr)))
	if err != nil {
		return "", err
	}

	// returns a string with the amount and when
	result = staking.GetAmountBigInt().String() + "," + strconv.FormatUint(staking.When, 10)
	return result, nil
}


////////////////////////////////////////////////////////////////////////////////

func sendBalance(sender *state.AccountState, receiver *state.AccountState, amount *big.Int) error {
	if err := state.SendBalance(sender, receiver, amount); err != nil {
		return errors.New("insufficient balance: " + sender.Balance().String() +
		                  " amount to transfer: " + amount.String())
	}
	return nil
}
