package contract

import "C"
import (
	"bytes"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/aergoio/aergo/fee"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/minio/sha256-simd"
)

type loadedReply struct {
	tx  *types.Tx
	ex  *executor
	err error
}

/* The preLoadWorker has 'current' and 'next' transactions to optimize the execution time by preloading the next transaction while executing the current transaction */
type preLoadReq struct {
	preLoadService int
	bs             *state.BlockState
	bi             *types.BlockHeaderInfo
	next           *types.Tx
	current        *types.Tx
}

type preLoadInfo struct {
	requestedTx *types.Tx
	replyCh     chan *loadedReply
}

var (
	loadReqCh     chan *preLoadReq
	preLoadInfos  [2]preLoadInfo
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
	loadReqCh = make(chan *preLoadReq, 10)
	preLoadInfos[BlockFactory].replyCh = make(chan *loadedReply, 4)
	preLoadInfos[ChainService].replyCh = make(chan *loadedReply, 4)
	addressRegexp, _ = regexp.Compile("^[a-zA-Z0-9]+$")

	go preLoadWorker()
}

func SetPreloadTx(tx *types.Tx, service int) {
	preLoadInfos[service].requestedTx = tx
}

func Execute(
	bs *state.BlockState,
	cdb ChainAccessor,
	tx *types.Tx,
	sender, receiver *state.V,
	bi *types.BlockHeaderInfo,
	preLoadService int,
	isFeeDelegation bool,
) (rv string, events []*types.Event, usedFee *big.Int, err error) {

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
			balance := new(big.Int).Sub(receiver.Balance(), usedFee)
			gasLimit = fee.MaxGasLimit(balance, bs.GasPrice)
			if gasLimit == 0 {
				err = newVmError(types.ErrNotEnoughGas)
				return
			}
		} else {
			gasLimit = txBody.GetGasLimit()
			if gasLimit == 0 {
				balance := new(big.Int).Sub(sender.Balance(), usedFee)
				gasLimit = fee.MaxGasLimit(balance, bs.GasPrice)
				if gasLimit == 0 {
					err = newVmError(types.ErrNotEnoughGas)
					return
				}
			} else {
				usedGas := fee.TxGas(len(txBody.GetPayload()))
				if gasLimit <= usedGas {
					err = newVmError(types.ErrNotEnoughGas)
					return
				}
				gasLimit -= usedGas
			}
		}
	}

	contractState, err := bs.OpenContractState(receiver.AccountID(), receiver.State())
	if err != nil {
		return
	}

	if receiver.IsRedeploy() {
		if err = checkRedeploy(sender, receiver, contractState); err != nil {
			return
		}
		bs.RemoveCache(receiver.AccountID())
	}

	var ex *executor

	// is there a request to preload an executor for this tx?
	if !receiver.IsDeploy() && preLoadInfos[preLoadService].requestedTx == tx {
		// get the reply channel
		replyCh := preLoadInfos[preLoadService].replyCh
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
		ctx := NewVmContext(bs, cdb, sender, receiver, contractState, sender.ID(),
			tx.GetHash(), bi, "", true, false, receiver.RP(),
			preLoadService, txBody.GetAmountBigInt(), gasLimit, isFeeDelegation)

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

func TxFee(payloadSize int, GasPrice *big.Int, version int32) *big.Int {
	if version < 2 {
		return fee.PayloadTxFee(payloadSize)
	}
	txGas := fee.TxGas(payloadSize)
	return new(big.Int).Mul(new(big.Int).SetUint64(txGas), GasPrice)
}

// send a request to preload an executor for the next tx
func RequestPreLoad(bs *state.BlockState, bi *types.BlockHeaderInfo, next, current *types.Tx, preLoadService int) {
	loadReqCh <- &preLoadReq{preLoadService, bs, bi, next, current}
}

// the preLoadWorker preloads executors for the next tx
func preLoadWorker() {
	for {
		var err error
		reqInfo := <-loadReqCh
		replyCh := preLoadInfos[reqInfo.preLoadService].replyCh

		// if there are more than 2 requests waiting for a reply, close the oldest one
		if len(replyCh) > 2 {
			select {
			case preload := <-replyCh:
				preload.ex.close()
			default:
			}
		}

		bs := reqInfo.bs
		tx := reqInfo.next  // next tx to be executed
		txBody := tx.GetBody()
		recipient := txBody.Recipient

		if (txBody.Type != types.TxType_NORMAL &&
			txBody.Type != types.TxType_TRANSFER &&
			txBody.Type != types.TxType_CALL &&
			txBody.Type != types.TxType_FEEDELEGATION) ||
			len(recipient) == 0 {
			continue
		}

		// if the tx currently being executed is a redeploy
		if reqInfo.current.GetBody().Type == types.TxType_REDEPLOY {
			// if the next tx is a call to the redeployed contract
			currentTxBody := reqInfo.current.GetBody()
			if bytes.Equal(recipient, currentTxBody.Recipient) {
				// do not preload an executor for a contract that is being redeployed
				replyCh <- &loadedReply{tx, nil, nil}
				continue
			}
		}

		// get the state of the recipient
		receiver, err := bs.GetAccountStateV(recipient)
		if err != nil {
			replyCh <- &loadedReply{tx, nil, err}
			continue
		}

		// when deploy and call in same block and not deployed yet
		if receiver.IsNew() || len(receiver.State().CodeHash) == 0 {
			// do not preload an executor for a contract that is not deployed yet
			replyCh <- &loadedReply{tx, nil, nil}
			continue
		}

		// open the contract state
		contractState, err := bs.OpenContractState(receiver.AccountID(), receiver.State())
		if err != nil {
			replyCh <- &loadedReply{tx, nil, err}
			continue
		}

		// create a new context
		ctx := NewVmContext(bs, nil, nil, receiver, contractState, txBody.GetAccount(),
			tx.GetHash(), reqInfo.bi, "", false, false, receiver.RP(),
			reqInfo.preLoadService, txBody.GetAmountBigInt(), txBody.GetGasLimit(),
			txBody.Type == types.TxType_FEEDELEGATION)

		// load a new executor
		ex, err := PreloadEx(bs, contractState, txBody.Payload, receiver.ID(), ctx)
		if ex == nil && ctx.traceFile != nil {
			ctx.traceFile.Close()
		}

		// send reply with executor
		replyCh <- &loadedReply{tx, ex, err}
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
	if len(receiver.State().CodeHash) == 0 || receiver.IsNew() {
		receiverAddr := types.EncodeAddress(receiver.ID())
		ctrLgr.Warn().Str("error", "not found contract").Str("contract", receiverAddr).Msg("redeploy")
		return newVmError(fmt.Errorf("not found contract %s", receiverAddr))
	}
	creator, err := contractState.GetData(creatorMetaKey)
	if err != nil {
		return err
	}
	if !bytes.Equal(creator, []byte(types.EncodeAddress(sender.ID()))) {
		return newVmError(types.ErrCreatorNotMatch)
	}
	return nil
}

func useGas(version int32) bool {
	return version >= 2 && PubNet
}

func SetBPTimeout(timeout <-chan struct{}) {
	bpTimeout = timeout
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
