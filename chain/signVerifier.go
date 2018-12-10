package chain

import (
	"errors"
	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
)

type SignVerifier struct {
	workerCnt int
	workCh    chan verifyWork
	doneCh    chan VerifyResult
}

type verifyWork struct {
	idx int
	tx  *types.Tx
}

type VerifyResult struct {
	work *verifyWork
	err  error
}

var (
	ErrTxFormatInvalid = errors.New("tx invalid format")

	//logger = log.NewLogger("signverifier")
)

func NewSignVerifier(workerCnt int) *SignVerifier {
	sv := &SignVerifier{
		workerCnt: workerCnt,
		workCh:    make(chan verifyWork, workerCnt),
		doneCh:    make(chan VerifyResult, workerCnt),
	}

	for i := 0; i < workerCnt; i++ {
		go sv.verifyTxLoop(i)
	}

	return sv
}

func (sv *SignVerifier) Stop() {
	close(sv.workCh)
	close(sv.doneCh)
}

func (sv *SignVerifier) verifyTxLoop(workerNo int) {
	logger.Debug().Int("worker", workerNo).Msg("verify worker run")

	for txWork := range sv.workCh {
		//logger.Debug().Int("worker", workerNo).Int("idx", txWork.idx).Msg("get work to verify tx")
		err := verifyTx(txWork.tx)

		if err != nil {
			logger.Error().Int("worker", workerNo).Str("hash", enc.ToString(txWork.tx.GetHash())).
				Err(err).Msg("error verify tx")
		}

		sv.doneCh <- VerifyResult{work: &txWork, err: err}
	}

	logger.Debug().Int("worker", workerNo).Msg("verify worker stop")

}

func verifyTx(tx *types.Tx) error {
	account := tx.GetBody().GetAccount()
	if account == nil {
		return ErrTxFormatInvalid
	}

	err := key.VerifyTx(tx)

	if err != nil {
		return err
	}

	return nil
}

func (sv *SignVerifier) VerifyTxs(txlist *types.TxList) (bool, []error) {
	txs := txlist.GetTxs()
	txLen := len(txs)

	if txLen == 0 {
		return false, nil
	}

	errors := make([]error, txLen, txLen)

	//logger.Debug().Int("txlen", txLen).Msg("verify tx start")

	go func() {
		for i, tx := range txs {
			//logger.Debug().Int("idx", i).Msg("push tx start")
			sv.workCh <- verifyWork{idx: i, tx: tx}
		}
	}()

	var doneCnt = 0
	failed := false

LOOP:
	for {
		select {
		case result := <-sv.doneCh:
			doneCnt++
			//logger.Debug().Int("donecnt", doneCnt).Msg("verify tx done")

			if result.work.idx < 0 || result.work.idx >= txLen {
				logger.Error().Int("idx", result.work.idx).Msg("Invalid Verify Result Index")
				continue
			}

			errors[result.work.idx] = result.err

			if result.err != nil {
				logger.Error().Err(result.err).Int("txno", result.work.idx).
					Msg("verifing tx failed")
				failed = true
			}

			if doneCnt == txLen {
				break LOOP
			}
		}
	}

	logger.Debug().Msg("verify tx done")
	return failed, errors
}

func (bv *SignVerifier) verifyTxsInplace(txlist *types.TxList) (bool, []error) {
	txs := txlist.GetTxs()
	txLen := len(txs)
	errors := make([]error, txLen, txLen)
	failed := false

	logger.Debug().Int("txlen", txLen).Msg("verify tx inplace start")

	for i, tx := range txs {
		errors[i] = verifyTx(tx)
		failed = true
	}

	logger.Debug().Msg("verify tx done")
	return failed, errors
}
