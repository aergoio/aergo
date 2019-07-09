package contract

import "C"
import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	"github.com/aergoio/aergo/fee"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/minio/sha256-simd"
)

type loadedReply struct {
	tx  *types.Tx
	ex  *Executor
	err error
}

type preLoadReq struct {
	preLoadService int
	bs             *state.BlockState
	next           *types.Tx
	current        *types.Tx
}

type preLoadInfo struct {
	requestedTx *types.Tx
	replyCh     chan *loadedReply
}

var (
	loadReqCh    chan *preLoadReq
	preLoadInfos [2]preLoadInfo
	PubNet       bool
	TraceBlockNo uint64
)

const BlockFactory = 0
const ChainService = 1

func init() {
	loadReqCh = make(chan *preLoadReq, 10)
	preLoadInfos[BlockFactory].replyCh = make(chan *loadedReply, 4)
	preLoadInfos[ChainService].replyCh = make(chan *loadedReply, 4)

	go preLoadWorker()
}

func SetPreloadTx(tx *types.Tx, service int) {
	preLoadInfos[service].requestedTx = tx
}

func Execute(bs *state.BlockState, cdb ChainAccessor, tx *types.Tx, blockNo uint64, ts int64, prevBlockHash []byte,
	sender, receiver *state.V, preLoadService int) (rv string, events []*types.Event, usedFee *big.Int, err error) {

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

	var ex *Executor
	if !receiver.IsDeploy() && preLoadInfos[preLoadService].requestedTx == tx {
		replyCh := preLoadInfos[preLoadService].replyCh
		for {
			preload := <-replyCh
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

	var cFee *big.Int
	if ex != nil {
		rv, events, cFee, err = PreCall(ex, bs, sender, contractState, blockNo, ts, receiver.RP(), prevBlockHash)
	} else {
		stateSet := NewContext(bs, cdb, sender, receiver, contractState, sender.ID(),
			tx.GetHash(), blockNo, ts, prevBlockHash, "", true,
			false, receiver.RP(), preLoadService, txBody.GetAmountBigInt())
		if stateSet.traceFile != nil {
			defer stateSet.traceFile.Close()
		}
		if receiver.IsDeploy() {
			rv, events, cFee, err = Create(contractState, txBody.Payload, receiver.ID(), stateSet)
		} else {
			rv, events, cFee, err = Call(contractState, txBody.Payload, receiver.ID(), stateSet)
		}
	}

	usedFee.Add(usedFee, cFee)

	if err != nil {
		if isSystemError(err) {
			return "", events, usedFee, err
		}
		return "", events, usedFee, newVmError(err)
	}

	err = bs.StageContractState(contractState)
	if err != nil {
		return "", events, usedFee, err
	}

	return rv, events, usedFee, nil
}

func PreLoadRequest(bs *state.BlockState, next, current *types.Tx, preLoadService int) {
	loadReqCh <- &preLoadReq{preLoadService, bs, next, current}
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

		if txBody.Type != types.TxType_NORMAL || len(recipient) == 0 {
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
		stateSet := NewContext(bs, nil, nil, receiver, contractState, txBody.GetAccount(),
			tx.GetHash(), 0, 0, nil, "", false,
			false, receiver.RP(), reqInfo.preLoadService, txBody.GetAmountBigInt())

		ex, err := PreloadEx(bs, contractState, receiver.AccountID(), txBody.Payload, receiver.ID(), stateSet)
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
		logger.Warn().Str("error", "not found contract").Str("contract", receiverAddr).Msg("redeploy")
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
