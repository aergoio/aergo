package contract

import "C"
import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/aergoio/aergo/config"
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
	loadReqCh      chan *preLoadReq
	preLoadInfos   [2]preLoadInfo
	PubNet         bool
	TraceBlockNo   uint64
	HardforkConfig *config.HardforkConfig
	bpTimeout      <-chan struct{}
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

	usedFee = fee.PayloadTxFee(len(txBody.GetPayload()))

	// Transfer balance
	if sender.AccountID() != receiver.AccountID() {
		if sender.Balance().Cmp(txBody.GetAmountBigInt()) < 0 {
			err = types.ErrInsufficientBalance
			return
		}
		sender.SubBalance(txBody.GetAmountBigInt())
		receiver.AddBalance(txBody.GetAmountBigInt())
	}

	if !receiver.IsDeploy() && len(receiver.State().CodeHash) == 0 {
		return
	}

	gasLimit := txBody.GetGasLimit()
	if useGas(bi.Version) && gasLimit == 0 {
		balance := new(big.Int).Sub(sender.Balance(), new(big.Int).Add(txBody.GetAmountBigInt(), usedFee))
		n := balance.Div(balance, bs.GasPrice)
		if n.IsUint64() {
			gasLimit = n.Uint64()
		} else {
			gasLimit = math.MaxUint64
		}
		if gasLimit == 0 {
			err = newVmError(types.ErrNotEnoughGas)
			return
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
		bs.CodeMap.Remove(receiver.AccountID())
	}

	var ex *executor
	if !receiver.IsDeploy() && preLoadInfos[preLoadService].requestedTx == tx {
		replyCh := preLoadInfos[preLoadService].replyCh
		for {
			var preload *loadedReply
			if HardforkConfig.IsV2Fork(bi.No) {
				if preLoadService == BlockFactory {
					select {
					case preload = <-replyCh:
					case <-bpTimeout:
						err = &VmTimeoutError{}
						return
					default:
						continue
					}
				} else {
					select {
					case preload = <-replyCh:
					default:
						continue
					}
				}
			} else {
				preload = <-replyCh
			}
			if preload.tx != tx {
				preload.ex.close()
				continue
			}
			ex = preload.ex
			err = preload.err
			break
		}
		if err != nil {
			return
		}
	}

	var ctrFee *big.Int
	if ex != nil {
		rv, events, ctrFee, err = PreCall(ex, bs, sender, contractState, receiver.RP(), gasLimit)
	} else {
		ctx := newVmContext(bs, cdb, sender, receiver, contractState, sender.ID(),
			tx.GetHash(), bi, "", true, false, receiver.RP(),
			preLoadService, txBody.GetAmountBigInt(), gasLimit, isFeeDelegation)
		if ctx.traceFile != nil {
			defer ctx.traceFile.Close()
		}
		if receiver.IsDeploy() {
			rv, events, ctrFee, err = Create(contractState, txBody.Payload, receiver.ID(), ctx)
		} else {
			rv, events, ctrFee, err = Call(contractState, txBody.Payload, receiver.ID(), ctx)
		}
	}

	usedFee.Add(usedFee, ctrFee)

	if err != nil {
		if isSystemError(err) {
			return "", events, usedFee, err
		}
		return "", events, usedFee, newVmError(err)
	}
	if isFeeDelegation {
		if receiver.Balance().Cmp(usedFee) < 0 {
			return "", events, usedFee, newVmError(types.ErrInsufficientBalance)
		}
	} else {
		if sender.Balance().Cmp(usedFee) < 0 {
			return "", events, usedFee, newVmError(types.ErrInsufficientBalance)
		}
	}

	err = bs.StageContractState(contractState)
	if err != nil {
		return "", events, usedFee, err
	}

	return rv, events, usedFee, nil
}

func PreLoadRequest(bs *state.BlockState, bi *types.BlockHeaderInfo, next, current *types.Tx, preLoadService int) {
	loadReqCh <- &preLoadReq{preLoadService, bs, bi, next, current}
}

func preLoadWorker() {
	for {
		var err error
		reqInfo := <-loadReqCh
		replyCh := preLoadInfos[reqInfo.preLoadService].replyCh

		if len(replyCh) > 2 {
			select {
			case preload := <-replyCh:
				preload.ex.close()
			default:
			}
		}

		bs := reqInfo.bs
		tx := reqInfo.next
		txBody := tx.GetBody()
		recipient := txBody.Recipient

		if (txBody.Type != types.TxType_NORMAL &&
			txBody.Type != types.TxType_TRANSFER &&
			txBody.Type != types.TxType_CALL &&
			txBody.Type != types.TxType_FEEDELEGATION) ||
			len(recipient) == 0 {
			continue
		}

		if reqInfo.current.GetBody().Type == types.TxType_REDEPLOY {
			currentTxBody := reqInfo.current.GetBody()
			if bytes.Equal(recipient, currentTxBody.Recipient) {
				replyCh <- &loadedReply{tx, nil, nil}
				continue
			}
		}

		receiver, err := bs.GetAccountStateV(recipient)
		if err != nil {
			replyCh <- &loadedReply{tx, nil, err}
			continue
		}
		/* When deploy and call in same block and not deployed yet*/
		if receiver.IsNew() || len(receiver.State().CodeHash) == 0 {
			replyCh <- &loadedReply{tx, nil, nil}
			continue
		}
		contractState, err := bs.OpenContractState(receiver.AccountID(), receiver.State())
		if err != nil {
			replyCh <- &loadedReply{tx, nil, err}
			continue
		}
		ctx := newVmContext(bs, nil, nil, receiver, contractState, txBody.GetAccount(),
			tx.GetHash(), reqInfo.bi, "", false, false, receiver.RP(),
			reqInfo.preLoadService, txBody.GetAmountBigInt(), txBody.GetGasLimit(),
			txBody.Type == types.TxType_FEEDELEGATION)

		ex, err := PreloadEx(bs, contractState, receiver.AccountID(), txBody.Payload, receiver.ID(), ctx)
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
