/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"context"
	"errors"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
)

var (
	// ErrBestBlock indicates that the best block is being changed in
	// chainservice soon.
	ErrBestBlock = errors.New("best block changed in chainservice")

	logger = log.NewLogger("consensus")
)

// FetchTXs requests to mempool and returns types.Tx array.
func FetchTXs(hs component.ICompSyncRequester, maxBlockBodySize uint32) []types.Transaction {
	//bf.RequestFuture(message.MemPoolSvc, &message.MemPoolGenerateSampleTxs{MaxCount: 3}, time.Second)
	result, err := hs.RequestFuture(message.MemPoolSvc,
		&message.MemPoolGet{MaxBlockBodySize: maxBlockBodySize}, time.Second,
		"consensus/util/info.FetchTXs").Result()
	if err != nil {
		logger.Info().Err(err).Msg("can't fetch transactions from mempool")
		return make([]types.Transaction, 0)
	}

	return result.(*message.MemPoolGetRsp).Txs
}

// TxOp is an interface used by GatherTXs for apply some transaction related operation.
type TxOp interface {
	Apply(bState *state.BlockState, tx types.Transaction) error
}

// TxOpFn is the type of arguments for CompositeTxDo.
type TxOpFn func(bState *state.BlockState, tx types.Transaction) error

// Apply applies f to tx.
func (f TxOpFn) Apply(bState *state.BlockState, tx types.Transaction) error {
	return f(bState, tx)
}

// NewCompTxOp returns a function which applies each function in fn.
func NewCompTxOp(fn ...TxOp) TxOp {
	return TxOpFn(func(bState *state.BlockState, tx types.Transaction) error {
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
	return TxOpFn(func(bState *state.BlockState, tx types.Transaction) error {
		if size += proto.Size(tx.GetTx()); uint32(size) > maxBlockBodySize {
			return errBlockSizeLimit
		}
		return nil
	})
}

// Lock acquires the chain lock in a blocking mode.
func Lock() {
	chain.InAddBlock <- struct{}{}
}

// LockNonblock acquires the chain lock in a non-blocking mode. It returns
// ErrBestBlock upon failure.
func LockNonblock() error {
	select {
	case chain.InAddBlock <- struct{}{}:
		return nil
	default:
		return ErrBestBlock
	}
}

// Unlock release the chain lock.
func Unlock() {
	<-chain.InAddBlock
}

// GatherTXs returns transactions from txIn. The selection is done by applying
// txDo.
func (g *BlockGenerator) GatherTXs() ([]types.Transaction, error) {
	var (
		bState = g.bState

		nCollected int
		nCand      int
	)

	if logger.IsDebugEnabled() {
		logger.Debug().Msg("start gathering tx")
	}

	if err := LockNonblock(); err != nil {
		return nil, ErrBestBlock
	}
	defer Unlock()

	txIn := g.fetchTXs(g.hs, g.maxBlockBodySize)
	nCand = len(txIn)

	txRes := make([]types.Transaction, 0, nCand)

	defer func() {
		logger.Info().
			Int("candidates", nCand).
			Int("collected", nCollected).
			Msg("transactions collected")
		contract.CloseDatabase()
	}()

	// block generation timeout check. this function works like BlockFactory#checkBpTimeout()
	checkBGTimeout := NewCompTxOp(
		TxOpFn(func(bState *state.BlockState, txIn types.Transaction) error {
			select {
			case <-g.ctx.Done():
				// TODO use function Cause() for precise control, later. cause can be used in go1.20 and later
				causeErr := g.ctx.Err()
				//causeErr := context.Cause(g.ctx)
				switch causeErr {
				case context.Canceled: // Only quitting of Aergo triggers Canceled error for now.
					return ErrQuit
				default:
					return ErrTimeout{Kind: "block"}
				}
			default:
				return nil
			}
		}),
	)

	if nCand > 0 {
		op := NewCompTxOp(checkBGTimeout, g.txOp)

		var preloadTx *types.Tx
		for i, tx := range txIn {
			// if not last tx, preload next tx
			if i != nCand-1 {
				preloadTx = txIn[i+1].GetTx()
				contract.RequestPreload(bState, g.bi, preloadTx, tx.GetTx(), contract.BlockFactory)
			}
			// process the transaction
			err := op.Apply(bState, tx)
			// mark the next preload tx to be executed
			contract.SetPreloadTx(preloadTx, contract.BlockFactory)

			//don't include tx that error is occurred
			if e, ok := err.(ErrTimeout); ok {
				if logger.IsDebugEnabled() {
					logger.Debug().Msg("stop gathering tx due to time limit")
				}
				err = e
				break
			} else if cause, ok := err.(*contract.VmTimeoutError); ok {
				if logger.IsDebugEnabled() {
					logger.Debug().Msg("stop gathering tx due to time limit")
				}
				// Mark the rejected TX by timeout. The marked TX will be
				// forced to be the first TX of the next block. By doing this,
				// the TX may have a chance to use the maximum block execution
				// time. If the TX is rejected by timeout even with this, it
				// may be evicted from the mempool after checking the actual
				// execution time.
				if g.tteEnabled() {
					g.setRejected(tx, cause, i == 0)
				}

				err = ErrTimeout{Kind: "contract"}

				break
			} else if err == errBlockSizeLimit {
				if logger.IsDebugEnabled() {
					logger.Debug().Msg("stop gathering tx due to size limit")
				}
				break
			} else if err != nil {
				if logger.IsDebugEnabled() {
					logger.Debug().Err(err).Int("idx", i).Str("hash", enc.ToString(tx.GetHash())).Msg("skip error tx")
				}
				//FIXME handling system error (panic?)
				// ex) gas error/nonce error skip, but other system error panic
				continue
			}

			txRes = append(txRes, tx)
		}

		nCollected = len(txRes)
	}

	// Warning: This line must be run even with 0 gathered TXs, since the
	// function below includes voting reward as well as BP reward.
	if err := chain.SendBlockReward(bState, chain.CoinbaseAccount); err != nil {
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
