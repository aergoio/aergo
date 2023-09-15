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
	"github.com/aergoio/aergo/v2/types"
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

func Execute(txCtx context.Context, bs *state.BlockState, cdb ChainAccessor, tx *types.Tx, sender, receiver *state.V, bi *types.BlockHeaderInfo, preloadService int, isFeeDelegation bool) (rv string, events []*types.Event, usedFee *big.Int, err error) {

	txBody := tx.GetBody()

	// compute the base fee
	usedFee = TxFee(len(txBody.GetPayload()), bs.GasPrice, bi.ForkVersion)

	// check if sender and receiver are not the same
	if sender.AccountID() != receiver.AccountID() {
		// check if sender has enough balance
		if sender.Balance().Cmp(txBody.GetAmountBigInt()) < 0 {
			err = types.ErrInsufficientBalance
			return
		}
		// transfer the amount from the sender to the receiver
		sender.SubBalance(txBody.GetAmountBigInt())
		receiver.AddBalance(txBody.GetAmountBigInt())
	}

	// check if the receiver is a not contract
	if !receiver.IsDeploy() && len(receiver.State().CodeHash) == 0 {
		// Before the chain version 3, any tx with no code hash is
		// unconditionally executed as a simple Aergo transfer. Since this
		// causes confusion, emit error for call-type tx with a wrong address
		// from the chain version 3 by not returning error but fall-through for
		// correct gas estimation.
		if !(bi.ForkVersion >= 3 && txBody.Type == types.TxType_CALL) {
			// Here, the condition for fee delegation TX essentially being
			// call-type, is not necessary, because it is rejected from the
			// mempool without code hash.
			return
		}
	}

	var gasLimit uint64
	if useGas(bi.ForkVersion) {
		if isFeeDelegation {
			// check if the contract has enough balance for fee
			balance := new(big.Int).Sub(receiver.Balance(), usedFee)
			gasLimit = fee.MaxGasLimit(balance, bs.GasPrice)
			if gasLimit == 0 {
				err = newVmError(types.ErrNotEnoughGas)
				return
			}
		} else {
			// read the gas limit from the tx
			gasLimit = txBody.GetGasLimit()
			if gasLimit == 0 {
				// no gas limit specified, the limit is the sender's balance
				balance := new(big.Int).Sub(sender.Balance(), usedFee)
				gasLimit = fee.MaxGasLimit(balance, bs.GasPrice)
				if gasLimit == 0 {
					err = newVmError(types.ErrNotEnoughGas)
					return
				}
			} else {
				// check if the sender has enough balance for gas
				usedGas := fee.TxGas(len(txBody.GetPayload()))
				if gasLimit <= usedGas {
					err = newVmError(types.ErrNotEnoughGas)
					return
				}
				// subtract the used gas from the gas limit
				gasLimit -= usedGas
			}
		}
	}

	// open the contract state
	contractState, err := bs.OpenContractState(receiver.AccountID(), receiver.State())
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
		ctx := NewVmContext(txCtx, bs, cdb, sender, receiver, contractState, sender.ID(), tx.GetHash(), bi, "", true, false, receiver.RP(), preloadService, txBody.GetAmountBigInt(), gasLimit, isFeeDelegation)

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
	err = bs.StageContractState(contractState)
	if err != nil {
		return "", events, usedFee, err
	}

	// return the result
	return rv, events, usedFee, nil
}

// compute the base fee for a transaction
func TxFee(payloadSize int, GasPrice *big.Int, version int32) *big.Int {
	if version < 2 {
		return fee.PayloadTxFee(payloadSize)
	}
	// get the amount of gas needed for the payload
	txGas := fee.TxGas(payloadSize)
	// multiply the amount of gas with the gas price
	return new(big.Int).Mul(new(big.Int).SetUint64(txGas), GasPrice)
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
		receiver, err := bs.GetAccountStateV(recipient)
		if err != nil {
			replyCh <- &preloadReply{tx, nil, err}
			continue
		}

		// when deploy and call in same block and not deployed yet
		if receiver.IsNew() || len(receiver.State().CodeHash) == 0 {
			// do not preload an executor for a contract that is not deployed yet
			replyCh <- &preloadReply{tx, nil, nil}
			continue
		}

		// open the contract state
		contractState, err := bs.OpenContractState(receiver.AccountID(), receiver.State())
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

func CreateContractID(account []byte, nonce uint64) []byte {
	h := sha256.New()
	h.Write(account)
	h.Write([]byte(strconv.FormatUint(nonce, 10)))
	recipientHash := h.Sum(nil)                   // byte array with length 32
	return append([]byte{0x0C}, recipientHash...) // prepend 0x0C to make it same length as account addresses
}

func checkRedeploy(sender, receiver *state.V, contractState *state.ContractState) error {
	// check if the contract exists
	if len(receiver.State().CodeHash) == 0 || receiver.IsNew() {
		receiverAddr := types.EncodeAddress(receiver.ID())
		ctrLgr.Warn().Str("error", "not found contract").Str("contract", receiverAddr).Msg("redeploy")
		return newVmError(fmt.Errorf("not found contract %s", receiverAddr))
	}
	// get the contract creator
	creator, err := contractState.GetData(creatorMetaKey)
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

func useGas(version int32) bool {
	return version >= 2 && PubNet
}

func GasUsed(txFee, gasPrice *big.Int, txType types.TxType, version int32) uint64 {
	if fee.IsZeroFee() || txType == types.TxType_GOVERNANCE || version < 2 {
		return 0
	}
	return new(big.Int).Div(txFee, gasPrice).Uint64()
}

func SetStateSQLMaxDBSize(size uint64) {
	if size > stateSQLMaxDBSize {
		maxSQLDBSize = stateSQLMaxDBSize
	} else if size < stateSQLMinDBSize {
		maxSQLDBSize = stateSQLMinDBSize
	} else {
		maxSQLDBSize = size
	}
	sqlLgr.Info().Uint64("size", maxSQLDBSize).Msg("set max database size(MB)")
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
