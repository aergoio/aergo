package contract

import (
	"encoding/hex"
	"errors"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/minio/sha256-simd"
	"strconv"
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

	contractState, err := bs.OpenContractState(receiver.State())
	if err != nil {
		return "", err
	}
	sqlTx, err := BeginTx(receiver.AccountID(), receiver.RP())
	if err != nil {
		return "", err
	}
	err = sqlTx.Savepoint()
	if err != nil {
		return "", err
	}

	var rv string
	var ex *Executor
	if preLoadInfos[preLoadService].requestedTx == tx {
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
		rv, err = PreCall(ex, bs, sender.State(), contractState, blockNo, ts, sqlTx.GetHandle())
	} else {
		bcCtx := NewContext(bs, sender.State(), contractState, types.EncodeAddress(txBody.GetAccount()),
			hex.EncodeToString(tx.GetHash()), blockNo, ts, "", 0,
			types.EncodeAddress(receiver.ID()), 0, nil, sqlTx.GetHandle(), preLoadService)

		if receiver.IsNew() {
			rv, err = Create(contractState, txBody.Payload, receiver.ID(), bcCtx)
		} else {
			rv, err = Call(contractState, txBody.Payload, receiver.ID(), bcCtx)
		}
	}
	if err != nil {
		if rErr := sqlTx.RollbackToSavepoint(); rErr != nil {
			return "", rErr
		}
		if err == types.ErrInsufficientBalance {
			return "", err
		}
		return "", VmError(err)
	}

	err = bs.CommitContractState(contractState)
	if err != nil {
		_ = sqlTx.RollbackToSavepoint()
		return "", err
	}

	return rv, sqlTx.Release()
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
		contractState, err := bs.OpenContractState(receiver.State())
		if err != nil {
			replyCh <- &loadedReply{tx, nil, err}
		}
		bcCtx := NewContext(bs, nil, contractState, types.EncodeAddress(txBody.GetAccount()),
			hex.EncodeToString(tx.GetHash()), 0, 0, "", 0,
			types.EncodeAddress(receiver.ID()), 0, nil, nil, reqInfo.preLoadService)

		ex, err := PreloadEx(contractState, txBody.Payload, receiver.ID(), bcCtx)
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
