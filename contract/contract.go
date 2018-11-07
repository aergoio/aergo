package contract

import "C"
import (
	"errors"
	"strconv"

	"github.com/aergoio/aergo/internal/enc"
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
	tx             *types.Tx
}

type preLoadInfo struct {
	requestedTx *types.Tx
	replyCh     chan *loadedReply
}

var loadReqCh chan *preLoadReq
var preLoadInfos [2]preLoadInfo

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

func Execute(bs *state.BlockState, tx *types.Tx, blockNo uint64, ts int64,
	sender, receiver *state.V, preLoadService int) (string, error) {

	txBody := tx.GetBody()

	// Transfer balance
	if sender.AccountID() != receiver.AccountID() {
		if sender.Balance() < txBody.Amount {
			return "", types.ErrInsufficientBalance
		}
		sender.SubBalance(txBody.Amount)
		receiver.AddBalance(txBody.Amount)
	}

	if txBody.Payload == nil {
		return "", nil
	}

	if !receiver.IsNew() && len(receiver.State().CodeHash) == 0 {
		return "", errors.New("account is not a contract")
	}

	contractState, err := bs.OpenContractState(receiver.AccountID(), receiver.State())
	if err != nil {
		return "", err
	}

	var rv string
	var ex *Executor
	if !receiver.IsCreate() && preLoadInfos[preLoadService].requestedTx == tx {
		replyCh := preLoadInfos[preLoadService].replyCh
		for {
			preload := <-replyCh
			if preload.tx != tx {
				preload.ex.close(true)
				continue
			}
			ex = preload.ex
			err = preload.err
			break
		}
		if err != nil {
			return "", err
		}
	}
	if ex != nil {
		rv, err = PreCall(ex, bs, sender.State(), contractState, blockNo, ts, receiver.RP())
	} else {
		bcCtx := NewContext(bs, sender.State(), contractState, types.EncodeAddress(txBody.GetAccount()),
			enc.ToString(tx.GetHash()), blockNo, ts, "", 0,
			types.EncodeAddress(receiver.ID()), 0, nil, receiver.RP(),
			preLoadService, txBody.GetAmount())

		if receiver.IsCreate() {
			rv, err = Create(contractState, txBody.Payload, receiver.ID(), bcCtx)
		} else {
			rv, err = Call(contractState, txBody.Payload, receiver.ID(), bcCtx)
		}
	}
	if err != nil {
		if err == types.ErrInsufficientBalance {
			return "", err
		} else if _, ok := err.(DbSystemError); ok {
			return "", err
		}
		return "", VmError(err)
	}

	err = bs.StageContractState(contractState)
	if err != nil {
		return "", err
	}

	return rv, nil
}

func PreLoadRequest(bs *state.BlockState, tx *types.Tx, preLoadService int) {
	loadReqCh <- &preLoadReq{preLoadService, bs, tx}
}

func preLoadWorker() {
	for {
		var err error
		reqInfo := <-loadReqCh
		replyCh := preLoadInfos[reqInfo.preLoadService].replyCh

		if len(replyCh) > 2 {
			preload := <-replyCh
			preload.ex.close(true)
		}

		bs := reqInfo.bs
		tx := reqInfo.tx
		txBody := tx.GetBody()
		recipient := txBody.Recipient

		if txBody.Type != types.TxType_NORMAL || len(recipient) == 0 ||
			txBody.Payload == nil {
			continue
		}
		receiver, err := bs.GetAccountStateV(recipient)
		if err != nil {
			replyCh <- &loadedReply{tx, nil, err}
			continue
		}
		/* When deploy and call in same block and not deployed yet*/
		if receiver.IsNew() {
			replyCh <- &loadedReply{tx, nil, nil}
			continue
		}
		if len(receiver.State().CodeHash) == 0 {
			replyCh <- &loadedReply{tx, nil, errors.New("account is not a contract")}
			continue
		}
		contractState, err := bs.OpenContractState(receiver.AccountID(), receiver.State())
		if err != nil {
			replyCh <- &loadedReply{tx, nil, err}
			continue
		}
		txHash := enc.ToString(tx.GetHash())
		sender := types.EncodeAddress(txBody.GetAccount())
		contractId := types.EncodeAddress(receiver.ID())

		bcCtx := &LBlockchainCtx{
			sender:     C.CString(sender),
			txHash:     C.CString(txHash),
			contractId: C.CString(contractId),
			service:    C.int(reqInfo.preLoadService),
			amount:     C.ulonglong(txBody.GetAmount()),
			isQuery:    C.int(0),
			confirmed:  C.int(0),
			node:       C.CString(""),
		}

		ex, err := PreloadEx(contractState, txBody.Payload, receiver.ID(), bcCtx)
		if err != nil {
			bcCtx.Del()
		}
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
