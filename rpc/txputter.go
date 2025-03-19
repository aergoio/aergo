package rpc

import (
	"bytes"
	"context"
	"reflect"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
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

// Commit sends transactions to the mempool and processes the results.
// It implements a producer-consumer pattern where it first fills a queue with
// transaction requests, then processes the responses as they come in.
func (m *txPutter) Commit() error {
	// phase 1: Fill the queue with transaction requests up to the buffer size
	// this acts as the producer part of the pattern
	for true {
		idx := m.putNextTx()
		if idx < 0 || m.q.Full() { // nothing to put or job queue is full
			break
		}
	}

	// phase 2: process responses from the mempool (consumer part)
	// keep track of retries for transactions that time out
	toRetry := 0
	for !m.q.Empty() {
		// check if the context has been canceled
		select {
		case <-m.ctx.Done():
			// exit with the context error
			return m.ctx.Err()
		default:
		}

		// get the next transaction index from the queue
		i := m.q.Poll().(int)
		future := m.futures[i]
		// wait for and retrieve the result of the transaction submission
		result, err := future.Result()

		if err != nil { // error occurred during processing
			// if it's a timeout and we haven't exceeded max retries
			if err == actor.ErrTimeout && toRetry < m.maxRetry {
				toRetry++
				m.logger.Debug().Int("idx", i).Int("retryCnt", toRetry).Msg("Retrying timeout job")
				// re-send the transaction to the mempool, now without verification
				m.rePutTx(i)
			} else {
				// for other errors or if max retries exceeded, exit with error
				m.logger.Debug().Err(err).Int("idx", i).Int("retryCnt", toRetry).Msg("Exiting commit")
				return err
			}
		} else { // on success
			// record the commit result
			m.writeResult(result, i)
			// try to add another transaction to keep the queue full
			m.putNextTx()
		}
	}

	// all transactions have been processed
	m.logger.Debug().Int("txSize", m.txSize).Int("retryCnt", toRetry).Msg("putting txs complete")
	return nil
}

// writeResult processes the result of a transaction submission and updates the commit result
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

// send valid txs to the mempool, skipping invalid txs
func (m *txPutter) putNextTx() int {
	// start at the current offset and iterates over the transactions (m.Txs)
	for ; m.offset < m.txSize; m.offset++ {
		// get the next transaction from the list
		tx := m.Txs[m.offset]
		// get the hash of the transaction
		hash := tx.Hash
		// create a commit result object
		var r types.CommitResult
		r.Hash = hash
		m.rs[m.offset] = &r
		// re-calculate the hash of the transaction
		calculated := tx.CalculateTxHash()
		// if the calculated hash is not equal to the original hash, set the error to TX_INVALID_HASH
		if !bytes.Equal(hash, calculated) {
			m.logger.Trace().Stringer("calculated", types.LogBase58(calculated)).Stringer("in", types.LogBase58(hash)).Msg("tx hash mismatch")
			r.Error = types.CommitStatus_TX_INVALID_HASH
		} else {
			// send the transaction to the mempool
			f := m.hub.RequestFuture(message.MemPoolSvc,
				&message.MemPoolPut{Tx: tx},
				m.actorTimeout, "rpc.(*AergoRPCService).CommitTX")
			// add the future to the list of futures
			m.futures[m.offset] = f
			// add the offset to the queue
			m.q.Offer(m.offset)
			// set the point to the current offset
			point := m.offset
			// increment the offset
			m.offset++
			// log the operation
			m.logger.Trace().Object("tx", types.LogTxHash{Tx: tx}).Msg("putting tx to mempool")
			// return the point
			return point
		}
	}
	// return -1 if no valid txs were found
	return -1
}

// re-send the transaction to the mempool, without verification
func (m *txPutter) rePutTx(i int) {
	// get the transaction from the list
	tx := m.Txs[i]
	// send the transaction to the mempool
	f := m.hub.RequestFuture(message.MemPoolSvc,
		&message.MemPoolPut{Tx: tx},
		m.actorTimeout, "rpc.(*AergoRPCService).CommitTX")
	// add the future to the list of futures
	m.futures[i] = f
	// add the offset to the queue
	m.q.Offer(i)
	// log the operation
	m.logger.Trace().Object("tx", types.LogTxHash{Tx: tx}).Msg("re-putting tx to mempool")
}
