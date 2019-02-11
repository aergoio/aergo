package contract

import "C"
import (
	"strconv"

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

var (
	loadReqCh    chan *preLoadReq
	preLoadInfos [2]preLoadInfo
	PubNet       bool
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

func Execute(bs *state.BlockState, tx *types.Tx, blockNo uint64, ts int64, prevBlockHash []byte,
	sender, receiver *state.V, preLoadService int) (string, error) {

	txBody := tx.GetBody()

	// Transfer balance
	if sender.AccountID() != receiver.AccountID() {
		if sender.Balance().Cmp(txBody.GetAmountBigInt()) < 0 {
			return "", types.ErrInsufficientBalance
		}
		sender.SubBalance(txBody.GetAmountBigInt())
		receiver.AddBalance(txBody.GetAmountBigInt())
	}

	if !receiver.IsCreate() && len(receiver.State().CodeHash) == 0 {
		return "", nil
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
				preload.ex.close()
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
		rv, err = PreCall(ex, bs, sender, contractState, blockNo, ts, receiver.RP(), prevBlockHash)
	} else {
		stateSet := NewContext(bs, sender, receiver, contractState, sender.ID(),
			tx.GetHash(), blockNo, ts, prevBlockHash, "", true,
			false, receiver.RP(), preLoadService, txBody.GetAmountBigInt())

		if receiver.IsCreate() {
			rv, err = Create(contractState, txBody.Payload, receiver.ID(), stateSet)
		} else {
			rv, err = Call(contractState, txBody.Payload, receiver.ID(), stateSet)
		}
	}
	if err != nil {
		if isSystemError(err) {
			return "", err
		}
		return "", newVmError(err)
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
			select {
			case preload := <-replyCh:
				preload.ex.close()
			default:
			}
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
		if receiver.IsNew() || len(receiver.State().CodeHash) == 0 {
			replyCh <- &loadedReply{tx, nil, nil}
			continue
		}
		contractState, err := bs.OpenContractState(receiver.AccountID(), receiver.State())
		if err != nil {
			replyCh <- &loadedReply{tx, nil, err}
			continue
		}
		stateSet := NewContext(bs, nil, receiver, contractState, txBody.GetAccount(),
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
