package contract

import (
	"encoding/hex"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/minio/sha256-simd"
	"strconv"
)

func Execute(bs *state.BlockState, tx *types.Tx, blockNo uint64, ts int64,
	sender, receiver *state.V) (string, error) {

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

	bcCtx := NewContext(bs, sender.State(), contractState, types.EncodeAddress(txBody.GetAccount()),
		hex.EncodeToString(tx.GetHash()), blockNo, ts, "", 0,
		types.EncodeAddress(receiver.ID()), 0, nil, sqlTx.GetHandle())

	var rv string
	if receiver.IsNew() {
		rv, err = Create(contractState, txBody.Payload, receiver.ID(), bcCtx)
	} else {
		rv, err = Call(contractState, txBody.Payload, receiver.ID(), bcCtx)
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

func CreateContractID(account []byte, nonce uint64) []byte {
	h := sha256.New()
	h.Write(account)
	h.Write([]byte(strconv.FormatUint(nonce, 10)))
	recipientHash := h.Sum(nil)                   // byte array with length 32
	return append([]byte{0x0C}, recipientHash...) // prepend 0x0C to make it same length as account addresses
}
