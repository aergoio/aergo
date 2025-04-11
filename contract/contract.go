package contract

import "C"
import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
)

var (
	PubNet        bool
	TraceBlockNo  uint64
	bpTimeout     <-chan struct{}
	maxSQLDBSize  uint64
	addressRegexp *regexp.Regexp
)

// These constants indicate the situation in which the contract is executed. BlockFactory refers to execution by a block producer to create a block, from which BlockFactory class calls, and ChainService refers to execution to verify and apply blocks created by other producers. Depending on the mode, the operations or policies may be different, and currently timeout policy is different.
// TODO These values are also used to select slots for internal parallel processing. They have multiple roles or meanings, so it make harder to understand the source code.
const (
	BlockFactory = iota
	ChainService
	MaxVmService
)

func init() {
	addressRegexp, _ = regexp.Compile("^[a-zA-Z0-9]+$")
}

// Execute executes a normal transaction which is possibly executing smart contract.
func Execute(
	execCtx context.Context,
	bs *state.BlockState,
	cdb ChainAccessor,
	tx *types.Tx,
	sender, receiver *state.AccountState,
	bi *types.BlockHeaderInfo,
	executionMode int,
	isFeeDelegation bool,
) (rv string, events []*types.Event, internalOps string, usedFee *big.Int, err error) {

	var (
		txBody     = tx.GetBody()
		txType     = txBody.GetType()
		txPayload  = txBody.GetPayload()
		txAmount   = txBody.GetAmountBigInt()
		txGasLimit = txBody.GetGasLimit()
		isMultiCall= (txType == types.TxType_MULTICALL)
	)

	// compute the base fee
	usedFee = fee.TxBaseFee(bi.ForkVersion, bs.GasPrice, len(txPayload))

	// transfer the amount from the sender to the receiver
	if err = state.SendBalance(sender, receiver, txAmount); err != nil {
		return
	}

	// check if the tx is valid and if the code should be executed
	var do_execute bool
	if do_execute, err = checkExecution(txType, txAmount, len(txPayload), bi.ForkVersion, receiver.IsDeploy(), receiver.IsContract()); do_execute != true {
		return
	}

	// compute gas limit
	var gasLimit uint64
	if gasLimit, err = fee.GasLimit(bi.ForkVersion, isFeeDelegation, txGasLimit, len(txPayload), bs.GasPrice, usedFee, sender.Balance(), receiver.Balance()); err != nil {
		err = newVmError(types.ErrNotEnoughGas)
		return
	}

	// open the contract state
	var contractState *statedb.ContractState
	if isMultiCall {
		contractState = statedb.GetMultiCallState(sender.ID(), sender.State())
	} else {
		contractState, err = statedb.OpenContractState(receiver.ID(), receiver.State(), bs.StateDB)
	}
	if err != nil {
		return
	}

	// check if this is a contract redeploy
	if receiver.IsRedeploy() {
		// check if the redeploy is valid
		if err = checkRedeploy(sender, receiver, contractState); err != nil {
			return
		}
		// remove the contract from the cache
		bs.RemoveCache(receiver.AccountID())
	}
	var ctrFee *big.Int

	// create a new context
	ctx := NewVmContext(execCtx, bs, cdb, sender, receiver, contractState, sender.ID(), tx.GetHash(), bi, "", true, false, receiver.RP(), executionMode, txAmount, gasLimit, isFeeDelegation, isMultiCall)

	// execute the transaction
	if receiver.IsDeploy() {
		rv, events, internalOps, ctrFee, err = Create(contractState, txPayload, receiver.ID(), ctx)
	} else {
		rv, events, internalOps, ctrFee, err = Call(contractState, txPayload, receiver.ID(), ctx)
	}

	// close the trace file
	if ctx.traceFile != nil {
		defer ctx.traceFile.Close()
	}

	// check if the execution fee is negative
	if ctrFee != nil && ctrFee.Sign() < 0 {
		return "", events, internalOps, usedFee, ErrVmStart
	}
	// add the execution fee to the total fee
	usedFee.Add(usedFee, ctrFee)

	// check if the execution failed
	if err != nil {
		if isSystemError(err) {
			return "", events, internalOps, usedFee, err
		}
		return "", events, internalOps, usedFee, newVmError(err)
	}

	// check for sufficient balance for fee
	if isFeeDelegation {
		if receiver.Balance().Cmp(usedFee) < 0 {
			return "", events, internalOps, usedFee, newVmError(types.ErrInsufficientBalance)
		}
	} else {
		if sender.Balance().Cmp(usedFee) < 0 {
			return "", events, internalOps, usedFee, newVmError(types.ErrInsufficientBalance)
		}
	}

	if !isMultiCall {
		// save the contract state
		err = statedb.StageContractState(contractState, bs.StateDB)
		if err != nil {
			return "", events, internalOps, usedFee, err
		}
	}

	// return the result
	return rv, events, internalOps, usedFee, nil
}

// check if the tx is valid and if the code should be executed
func checkExecution(txType types.TxType, amount *big.Int, payloadSize int, version int32, isDeploy, isContract bool) (do_execute bool, err error) {

	if txType == types.TxType_MULTICALL {
		return true, nil
	}

	// transactions with type NORMAL should not call smart contracts
	// transactions with type TRANSFER can only call smart contracts when:
	//  * the amount is greater than 0
	//  * the payload is empty (only transfer to "default" function)
	if version >= 4 && isContract {
		if txType == types.TxType_NORMAL || (txType == types.TxType_TRANSFER && (payloadSize > 0 || types.IsZeroAmount(amount))) {
			// emit an error
			return false, newVmError(types.ErrTxNotAllowedRecipient)
		}
	}

	// check if the receiver is a not contract
	if !isDeploy && !isContract {
		// before the hardfork version 3, all transactions in which the recipient
		// is not a contract were processed as a simple Aergo transfer, including
		// type CALL and FEEDELEGATION.
		// starting from hardfork version 3, transactions expected to CALL a
		// contract but without a valid recipient will emit an error.
		// FEEDELEGATION txns with invalid recipient are rejected on mempool.
		if version >= 3 && txType == types.TxType_CALL {
			// continue and emit an error for correct gas estimation
			// it will fail because there is no code to execute
		} else {
			// no code to execute, just return
			return false, nil
		}
	}

	return true, nil
}

func checkRedeploy(sender, receiver *state.AccountState, contractState *statedb.ContractState) error {
	// check if the contract exists
	if !receiver.IsContract() || receiver.IsNew() {
		receiverAddr := types.EncodeAddress(receiver.ID())
		ctrLgr.Warn().Str("error", "not found contract").Str("contract", receiverAddr).Msg("redeploy")
		return newVmError(fmt.Errorf("not found contract %s", receiverAddr))
	}
	// get the contract creator
	creator, err := contractState.GetData(dbkey.CreatorMeta())
	if err != nil {
		return err
	}
	// check if the sender is the creator
	if !bytes.Equal(creator, []byte(types.EncodeAddress(sender.ID()))) {
		return newVmError(types.ErrCreatorNotMatch)
	}
	// no problem found
	return nil
}

func CreateContractID(account []byte, nonce uint64) []byte {
	h := sha256.New()
	h.Write(account)
	h.Write([]byte(strconv.FormatUint(nonce, 10)))
	recipientHash := h.Sum(nil)                   // byte array with length 32
	return append([]byte{0x0C}, recipientHash...) // prepend 0x0C to make it same length as account addresses
}

func SetStateSQLMaxDBSize(size uint64) {
	if size > stateSQLMaxDBSize {
		maxSQLDBSize = stateSQLMaxDBSize
	} else if size < stateSQLMinDBSize {
		maxSQLDBSize = stateSQLMinDBSize
	} else {
		maxSQLDBSize = size
	}
	//sqlLgr.Info().Uint64("size", maxSQLDBSize).Msg("set max database size(MB)")
}

func StrHash(d string) []byte {
	// using real address
	if len(d) == types.EncodedAddressLength && addressRegexp.MatchString(d) {
		return types.ToAddress(d)
	} else {
		// using alias
		h := sha256.New()
		h.Write([]byte(d))
		b := h.Sum(nil)
		b = append([]byte{0x0C}, b...)
		return b
	}
}
