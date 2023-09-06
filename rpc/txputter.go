package rpc

import (
	"bytes"
	"context"
	"reflect"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const actorBufSize = 10

type txPutter struct {
	hub          component.ICompSyncRequester
	logger       *log.Logger
	actorTimeout time.Duration
	maxRetry     int

	ctx context.Context
	Txs []*types.Tx

	txSize int
	offset int
	q      *p2putil.PressableQueue

	rs      []*types.CommitResult
	futures []*actor.Future
}

func newPutter(ctx context.Context, txs []*types.Tx, hub component.ICompSyncRequester, timeout time.Duration) *txPutter {
	m := &txPutter{ctx: ctx, Txs: txs, hub: hub, actorTimeout: timeout}
	m.logger = log.NewLogger("txputter")
	txSize := len(m.Txs)
	m.txSize = txSize
	m.q = p2putil.NewPressableQueue(actorBufSize)
	m.maxRetry = actorBufSize << 1
	m.rs = make([]*types.CommitResult, txSize)
	m.futures = make([]*actor.Future, txSize)

	return m
}

func (m *txPutter) Commit() error {
	//phase.1 send tx message to mempool with size of workers
	for true {
		idx := m.putToNextTx()
		if idx < 0 || m.q.Full() { // nothing to put or job queue is full
			break
		}
	}

	toRetry := 0
	for !m.q.Empty() {
		select {
		case <-m.ctx.Done():
			return m.ctx.Err()
		default:
		}

		i := m.q.Poll().(int)
		future := m.futures[i]
		result, err := future.Result()
		if err != nil { // error by actors
			if err == actor.ErrTimeout && toRetry < m.maxRetry {
				toRetry++
				m.logger.Debug().Int("idx", i).Int("retryCnt", toRetry).Msg("Retrying timeout job")
				m.rePutTx(i) // retry
			} else {
				m.logger.Debug().Err(err).Int("idx", i).Int("retryCnt", toRetry).Msg("Exiting commit")
				return err
			}
		} else { //
			m.writeResult(result, i)
			m.putToNextTx()
		}
	}
	m.logger.Debug().Int("txSize", m.txSize).Int("retryCnt", toRetry).Msg("putting txs complete")
	return nil
}

func (m *txPutter) writeResult(result interface{}, i int) {
	var err error
	rsp, ok := result.(*message.MemPoolPutRsp)
	if !ok {
		err = status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	} else {
		err = rsp.Err
	}
	m.rs[i].Error = convertError(err)
	if err != nil {
		m.rs[i].Detail = err.Error()
	}
}

// put valid tx, skipping invalid tx
func (m *txPutter) putToNextTx() int {
	for ; m.offset < m.txSize; m.offset++ {
		tx := m.Txs[m.offset]
		hash := tx.Hash
		var r types.CommitResult
		r.Hash = hash
		m.rs[m.offset] = &r
		calculated := tx.CalculateTxHash()
		if !bytes.Equal(hash, calculated) {
			m.logger.Trace().Object("calculated", types.LogBase58{Bytes: &calculated}).Object("in", types.LogBase58{Bytes: &hash}).Msg("tx hash mismatch")
			r.Error = types.CommitStatus_TX_INVALID_HASH
		} else {
			f := m.hub.RequestFuture(message.MemPoolSvc,
				&message.MemPoolPut{Tx: tx},
				m.actorTimeout, "rpc.(*AergoRPCService).CommitTX")
			m.futures[m.offset] = f
			m.q.Offer(m.offset)
			point := m.offset
			m.offset++
			m.logger.Trace().Object("tx", types.LogTxHash{Tx: tx}).Msg("putting tx to mempool")
			return point
		}
	}
	return -1
}

func (m *txPutter) rePutTx(i int) {
	tx := m.Txs[i]
	f := m.hub.RequestFuture(message.MemPoolSvc,
		&message.MemPoolPut{Tx: tx},
		m.actorTimeout, "rpc.(*AergoRPCService).CommitTX")
	m.futures[i] = f
	m.q.Offer(i)
	m.logger.Trace().Object("tx", types.LogTxHash{Tx: tx}).Msg("putting tx to mempool")
}
