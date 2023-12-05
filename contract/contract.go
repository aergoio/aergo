package contract

import "C"
import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
	"github.com/minio/sha256-simd"
)

/* The preloadWorker optimizes the execution time by preloading the next transaction while executing the current transaction */
type preloadRequest struct {
	preloadService int
	bs             *state.BlockState
	bi             *types.BlockHeaderInfo
	next           *types.Tx // the tx to preload the executor for, the next to be executed
	current        *types.Tx // the tx currently being executed
}

type preloadReply struct {
	tx  *types.Tx
	ex  *executor
	err error
}

type preloader struct {
	requestedTx *types.Tx // the next preload tx to be executed
	replyCh     chan *preloadReply
}

var (
	loadReqCh     chan *preloadRequest
	preloaders    [2]preloader
	PubNet        bool
	TraceBlockNo  uint64
	bpTimeout     <-chan struct{}
	maxSQLDBSize  uint64
	addressRegexp *regexp.Regexp
)

const (
	BlockFactory = iota
	ChainService
	MaxVmService
)

func init() {
	loadReqCh = make(chan *preloadRequest, 10)
	preloaders[BlockFactory].replyCh = make(chan *preloadReply, 4)
	preloaders[ChainService].replyCh = make(chan *preloadReply, 4)
	addressRegexp, _ = regexp.Compile("^[a-zA-Z0-9]+$")

	go preloadWorker()
}

// mark the next preload tx to be executed
func SetPreloadTx(tx *types.Tx, service int) {
	preloaders[service].requestedTx = tx
}

// Execute executes a normal transaction which is possibly executing smart contract.
func Execute(execCtx context.Context, bs *state.BlockState, cdb ChainAccessor, tx *types.Tx, sender, receiver *state.AccountState, bi *types.BlockHeaderInfo, preloadService int, isFeeDelegation bool) (rv string, events []*types.Event, usedFee *big.Int, err error) {

	var (
		txBody     = tx.GetBody()
		txType     = txBody.GetType()
		txPayload  = txBody.GetPayload()
		txAmount   = txBody.GetAmountBigInt()
		txGasLimit = txBody.GetGasLimit()
	)

	// compute the base fee
	usedFee = fee.TxBaseFee(bi.ForkVersion, bs.GasPrice, len(txPayload))

	// transfer the amount from the sender to the receiver
	if err = state.SendBalance(sender, receiver, txBody.GetAmountBigInt()); err != nil {
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
	contractState, err := statedb.OpenContractState(receiver.ID(), receiver.State(), bs.LuaStateDB)
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

	var ex *executor

	// is there a request to preload an executor for this tx?
	if !receiver.IsDeploy() && preloaders[preloadService].requestedTx == tx {
		// get the reply channel
		replyCh := preloaders[preloadService].replyCh
		// wait for the reply
		for {
			preload := <-replyCh
			if preload.tx != tx {
				preload.ex.close()
				continue
			}
			// get the executor and error from the reply
			ex = preload.ex
			err = preload.err
			break
		}
		if err != nil {
			return
		}
	}

	var ctrFee *big.Int

	// is there a preloaded executor?
	if ex != nil {
		// execute the transaction
		rv, events, ctrFee, err = PreCall(ex, bs, sender, contractState, receiver.RP(), gasLimit)
	} else {
		// create a new context
		ctx := NewVmContext(execCtx, bs, cdb, sender, receiver, contractState, sender.ID(), tx.GetHash(), bi, "", true, false, receiver.RP(), preloadService, txBody.GetAmountBigInt(), gasLimit, isFeeDelegation)

		// execute the transaction
		if receiver.IsDeploy() {
			rv, events, ctrFee, err = Create(contractState, txBody.Payload, receiver.ID(), ctx)
		} else {
			rv, events, ctrFee, err = Call(contractState, txBody.Payload, receiver.ID(), ctx)
		}

		// close the trace file
		if ctx.traceFile != nil {
			defer ctx.traceFile.Close()
		}
	}

	// check if the execution fee is negative
	if ctrFee != nil && ctrFee.Sign() < 0 {
		return "", events, usedFee, ErrVmStart
	}
	// add the execution fee to the total fee
	usedFee.Add(usedFee, ctrFee)

	// check if the execution failed
	if err != nil {
		if isSystemError(err) {
			return "", events, usedFee, err
		}
		return "", events, usedFee, newVmError(err)
	}

	// check for sufficient balance for fee
	if isFeeDelegation {
		if receiver.Balance().Cmp(usedFee) < 0 {
			return "", events, usedFee, newVmError(types.ErrInsufficientBalance)
		}
	} else {
		if sender.Balance().Cmp(usedFee) < 0 {
			return "", events, usedFee, newVmError(types.ErrInsufficientBalance)
		}
	}

	// save the contract state
	err = statedb.StageContractState(contractState, bs.LuaStateDB)
	if err != nil {
		return "", events, usedFee, err
	}

	// return the result
	return rv, events, usedFee, nil
}

// send a request to preload an executor for the next tx
func RequestPreload(bs *state.BlockState, bi *types.BlockHeaderInfo, next, current *types.Tx, preloadService int) {
	loadReqCh <- &preloadRequest{preloadService, bs, bi, next, current}
}

// the preloadWorker preloads an executor for the next tx
func preloadWorker() {
	// infinite loop
	for {
		var err error

		// wait for a preload request
		request := <-loadReqCh
		// get the reply channel for this request
		replyCh := preloaders[request.preloadService].replyCh

		// if there are more than 2 requests waiting for a reply, close the oldest one
		if len(replyCh) > 2 {
			select {
			case preload := <-replyCh:
				preload.ex.close()
			default:
			}
		}

		bs := request.bs
		tx := request.next // the tx to preload the executor for
		txBody := tx.GetBody()
		recipient := txBody.Recipient

		// only preload an executor for a normal, transfer, call or fee delegation tx
		if (txBody.Type != types.TxType_NORMAL &&
			txBody.Type != types.TxType_TRANSFER &&
			txBody.Type != types.TxType_CALL &&
			txBody.Type != types.TxType_FEEDELEGATION) ||
			len(recipient) == 0 {
			continue
		}

		// if the tx currently being executed is a redeploy
		if request.current.GetBody().Type == types.TxType_REDEPLOY {
			// if the next tx is a call to the redeployed contract
			currentTxBody := request.current.GetBody()
			if bytes.Equal(recipient, currentTxBody.Recipient) {
				// do not preload an executor for a contract that is being redeployed
				replyCh <- &preloadReply{tx, nil, nil}
				continue
			}
		}

		// get the state of the recipient
		receiver, err := state.GetAccountState(recipient, bs.LuaStateDB, bs.EvmStateDB)
		if err != nil {
			replyCh <- &preloadReply{tx, nil, err}
			continue
		}

		// when deploy and call in same block and not deployed yet
		if receiver.IsNew() || !receiver.IsContract() {
			// do not preload an executor for a contract that is not deployed yet
			replyCh <- &preloadReply{tx, nil, nil}
			continue
		}

		// open the contract state
		contractState, err := statedb.OpenContractState(receiver.ID(), receiver.State(), bs.LuaStateDB)
		if err != nil {
			replyCh <- &preloadReply{tx, nil, err}
			continue
		}

		// create a new context
		// FIXME need valid context
		ctx := NewVmContext(context.Background(), bs, nil, nil, receiver, contractState, txBody.GetAccount(), tx.GetHash(), request.bi, "", false, false, receiver.RP(), request.preloadService, txBody.GetAmountBigInt(), txBody.GetGasLimit(), txBody.Type == types.TxType_FEEDELEGATION)

		// load a new executor
		ex, err := PreloadExecutor(bs, contractState, txBody.Payload, receiver.ID(), ctx)
		if ex == nil && ctx.traceFile != nil {
			ctx.traceFile.Close()
		}

		// send reply with executor
		replyCh <- &preloadReply{tx, ex, err}
	}
}

// check if the tx is valid and if the code should be executed
func checkExecution(txType types.TxType, amount *big.Int, payloadSize int, version int32, isDeploy, isContract bool) (do_execute bool, err error) {

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

func CreateContractID(account []byte, nonce uint64) []byte {
	h := sha256.New()
	h.Write(account)
	h.Write([]byte(strconv.FormatUint(nonce, 10)))
	recipientHash := h.Sum(nil)                   // byte array with length 32
	return append([]byte{0x0C}, recipientHash...) // prepend 0x0C to make it same length as account addresses
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
