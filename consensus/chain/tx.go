/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"errors"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

var (
	// ErrBestBlock indicates that the best block is being changed in
	// chainservice soon.
	ErrBestBlock = errors.New("best block changed in chainservice")

	logger = log.NewLogger("consensus")
)

// FetchTXs requests to mempool and returns types.Tx array.
func FetchTXs(hs component.ICompSyncRequester, maxBlockBodySize uint32) []*types.Tx {
	//bf.RequestFuture(message.MemPoolSvc, &message.MemPoolGenerateSampleTxs{MaxCount: 3}, time.Second)
	result, err := hs.RequestFuture(message.MemPoolSvc,
		&message.MemPoolGet{MaxBlockBodySize: maxBlockBodySize}, time.Second,
		"consensus/util/info.FetchTXs").Result()
	if err != nil {
		logger.Info().Err(err).Msg("can't fetch transactions from mempool")
		return make([]*types.Tx, 0)
	}

	return result.(*message.MemPoolGetRsp).Txs
}

// TxOp is an interface used by GatherTXs for apply some transaction related operation.
type TxOp interface {
	Apply(bState *state.BlockState, tx *types.Tx) error
}

// TxOpFn is the type of arguments for CompositeTxDo.
type TxOpFn func(bState *state.BlockState, tx *types.Tx) error

// Apply applies f to tx.
func (f TxOpFn) Apply(bState *state.BlockState, tx *types.Tx) error {
	return f(bState, tx)
}

// NewCompTxOp returns a function which applies each function in fn.
func NewCompTxOp(fn ...TxOp) TxOp {
	return TxOpFn(func(bState *state.BlockState, tx *types.Tx) error {
		for _, f := range fn {
			var err error
			if err = f.Apply(bState, tx); err != nil {
				return err
			}
		}

		// If TxOp executes tx, it has a resulting BlockState. The final
		// BlockState must be sent to the chain service receiver.
		return nil
	})
}

func newBlockLimitOp(maxBlockBodySize uint32) TxOpFn {
	// Caution: the closure below captures the local variable 'size.' Generate
	// it whenever needed. Don't reuse it!
	size := 0
	return TxOpFn(func(bState *state.BlockState, tx *types.Tx) error {
		if size += proto.Size(tx); uint32(size) > maxBlockBodySize {
			return errBlockSizeLimit
		}
		return nil
	})
}

// LockChain aquires the chain lock in a non-blocking mode.
func LockChain() error {
	select {
	case chain.InAddBlock <- struct{}{}:
		return nil
	default:
		return ErrBestBlock
	}
}

// UnlockChain release the chain lock.
func UnlockChain() {
	<-chain.InAddBlock
}

// GatherTXs returns transactions from txIn. The selection is done by applying
// txDo.
func GatherTXs(hs component.ICompSyncRequester, bState *state.BlockState, txOp TxOp, maxBlockBodySize uint32) ([]*types.Tx, error) {
	var (
		nCollected int
		nCand      int
	)

	if logger.IsDebugEnabled() {
		logger.Debug().Msg("start gathering tx")
	}

	if err := LockChain(); err != nil {
		return nil, ErrBestBlock
	}
	defer UnlockChain()

	txIn := FetchTXs(hs, maxBlockBodySize)
	nCand = len(txIn)
	if nCand == 0 {
		return txIn, nil
	}
	txRes := make([]*types.Tx, 0, nCand)

	if logger.IsDebugEnabled() {
		defer func() {
			logger.Debug().
				Int("candidates", nCand).
				Int("collected", nCollected).
				Msg("transactions collected")
		}()
	}

	op := NewCompTxOp(txOp)

	var preLoadTx *types.Tx
	for i, tx := range txIn {
		if i != nCand-1 {
			preLoadTx = txIn[i+1]
			contract.PreLoadRequest(bState, preLoadTx, contract.BlockFactory)
		}

		err := op.Apply(bState, tx)
		contract.SetPreloadTx(preLoadTx, contract.BlockFactory)

		//don't include tx that error is occured
		if e, ok := err.(ErrTimeout); ok {
			if logger.IsDebugEnabled() {
				logger.Debug().Msg("stop gathering tx due to time limit")
			}
			err = e
			break
		} else if err == errBlockSizeLimit {
			if logger.IsDebugEnabled() {
				logger.Debug().Msg("stop gathering tx due to size limit")
			}
			break
		} else if err != nil {
			//FIXME handling system error (panic?)
			// ex) gas error/nonce error skip, but other system error panic
			logger.Debug().Err(err).Int("idx", i).Str("hash", enc.ToString(tx.GetHash())).Msg("skip error tx")
			continue
		}

		txRes = append(txRes, tx)
	}

	nCollected = len(txRes)

	if err := chain.SendRewardCoinbase(bState, chain.CoinbaseAccount); err != nil {
		return nil, err
	}

	if err := contract.SaveRecoveryPoint(bState); err != nil {
		return nil, err
	}

	if err := bState.Update(); err != nil {
		return nil, err
	}

	return txRes, nil
}
