package rpc

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
)

func Test_txPutter_Commit(t *testing.T) {
	sampleSize := 1000
	sampleTXs := make([]*types.Tx, sampleSize)
	idMap := make(map[types.TxID]time.Duration)
	for i := 0; i < sampleSize; i++ {
		tx := types.NewTx()
		tx.Body.Nonce = uint64(i + 1)
		tx.Hash = tx.CalculateTxHash()
		sampleTXs[i] = tx
		id, _ := types.ParseToTxID(tx.Hash)
		idMap[id] = 0
	}
	slows := make([]*types.Tx, 10)
	for i := 0; i < 10; i++ {
		tx := types.NewTx()
		tx.Body.Nonce = uint64(i + 1 + sampleSize)
		tx.Hash = tx.CalculateTxHash()
		slows[i] = tx
		id, _ := types.ParseToTxID(tx.Hash)
		idMap[id] = time.Second>>4 + time.Second>>5
	}

	tests := []struct {
		name    string
		txs     []*types.Tx
		wantErr bool
	}{
		{"TFast1000", sampleTXs, false},
		{"TSlow10", slows, true},
		{"TSlow1", concat(sampleTXs[:10], slows[:1], sampleTXs[10:100]), false},
		{"TSlow2", concat(sampleTXs[:10], slows[:1], sampleTXs[10:20], slows[1:2], sampleTXs[20:]), false},
		{"TSlow3", concat(sampleTXs[:10], slows[:1], sampleTXs[10:20], slows[1:2], sampleTXs[20:30], slows[2:3]), true},
		{"TSlow13", concat(sampleTXs[:10], slows[:3], sampleTXs[10:]), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := component.NewComponentHub()
			stub := &MPStub{succ: idMap, added: make(map[types.TxID]bool)}
			stub.BaseComponent = component.NewBaseComponent(message.MemPoolSvc, stub, logger)
			hub.Register(stub)
			hub.Start()
			defer hub.Stop()

			ctx := context.TODO()
			ctx, cancelFunc := context.WithTimeout(ctx, time.Second*3)
			defer cancelFunc()
			m := newPutter(ctx, tt.txs, hub, defaultActorTimeout)
			m.actorTimeout = time.Second >> 4

			err := m.Commit()
			if (err != nil) != tt.wantErr {
				t.Errorf("Commit() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(stub.added) != len(tt.txs) {
					t.Errorf("Commit() added = %v, want %v", len(stub.added), len(tt.txs))
				}
				for _, tx := range tt.txs {
					if _, exists := stub.added[types.ToTxID(tx.GetHash())]; !exists {
						t.Errorf("Commit() tx hash %v was not put, want put", types.ToTxID(tx.GetHash()))
					}
				}
				for i, r := range m.rs {
					if r == nil ||
						// TX_ALREADY_EXISTS will be sent to when retrying
						(r.Error != types.CommitStatus_TX_OK && r.Error != types.CommitStatus_TX_ALREADY_EXISTS) {
						t.Errorf("idx %d, err %v (%s), want all success", i, r.Error, r.Detail)
					}
				}
			}
		})
	}
}

func BenchmarkPutter(b *testing.B) {
	sampleSize := 1000
	sampleTXs := make([]*types.Tx, sampleSize)
	idMap := make(map[types.TxID]time.Duration)
	for i := 0; i < sampleSize; i++ {
		tx := types.NewTx()
		tx.Body.Nonce = uint64(i + 1)
		tx.Hash = tx.CalculateTxHash()
		sampleTXs[i] = tx
		id, _ := types.ParseToTxID(tx.Hash)
		idMap[id] = 0
	}
	slows := make([]*types.Tx, 10)
	for i := 0; i < 10; i++ {
		tx := types.NewTx()
		tx.Body.Nonce = uint64(i + 1 + sampleSize)
		tx.Hash = tx.CalculateTxHash()
		slows[i] = tx
		id, _ := types.ParseToTxID(tx.Hash)
		idMap[id] = time.Second>>4 + time.Second>>5
	}

	hub := component.NewComponentHub()
	stub := &MPStub{succ: idMap, added: make(map[types.TxID]bool)}
	stub.BaseComponent = component.NewBaseComponent(message.MemPoolSvc, stub, logger)
	hub.Register(stub)
	hub.Start()
	defer hub.Stop()

	const (
		queued = iota
		single
	)
	benchmarks := []struct {
		name      string
		putType   int
		queueSize int
	}{
		{"BBuf1", queued, 1},
		{"BBuf10", queued, 10},
		{"BBuf100", queued, 100},
		{"BSingle", single, 1},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {

				ctx := context.TODO()
				m := newPutter(ctx, sampleTXs, hub, defaultActorTimeout)
				m.q = p2putil.NewPressableQueue(bm.queueSize)
				m.actorTimeout = time.Second >> 4
				if bm.putType == single {
					m2 := &singlePutter{m}
					_ = m2.Commit()
				} else {
					_ = m.Commit()
				}

			}
		})
	}
}

type singlePutter struct {
	*txPutter
}

func (m *singlePutter) Commit() error {
	toRetry := 0
	for i, tx := range m.Txs {
		hash := tx.Hash
		var r types.CommitResult
		r.Hash = hash
		m.rs[i] = &r
		calculated := tx.CalculateTxHash()
		if !bytes.Equal(hash, calculated) {
			r.Error = types.CommitStatus_TX_INVALID_HASH
		} else {
			future := m.hub.RequestFuture(message.MemPoolSvc,
				&message.MemPoolPut{Tx: tx},
				m.actorTimeout, "rpc.(*AergoRPCService).CommitTX")
			var result interface{}
			var err error
			for result, err = future.Result(); err != nil; {
				if err == actor.ErrTimeout && toRetry < m.maxRetry {
					toRetry++
					m.logger.Debug().Int("idx", i).Int("retryCnt", toRetry).Msg("Retrying timeout job")
					future = m.hub.RequestFuture(message.MemPoolSvc,
						&message.MemPoolPut{Tx: tx},
						m.actorTimeout, "rpc.(*AergoRPCService).CommitTX")
				} else {
					m.logger.Debug().Err(err).Int("idx", i).Int("retryCnt", toRetry).Msg("Exiting commit")
					return err
				}
			}
			m.logger.Debug().Int("idx", i).Msg("job was finished in time")
			m.writeResult(result, i)
		}
	}
	return nil
}

func concat(a ...[]*types.Tx) []*types.Tx {
	ret := make([]*types.Tx, 0, len(a))
	for _, element := range a {
		ret = append(ret, element...)
	}
	return ret
}

type MPStub struct {
	*component.BaseComponent
	succ  map[types.TxID]time.Duration
	added map[types.TxID]bool
}

func (a *MPStub) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *message.MemPoolPut:
		id := types.ToTxID(msg.Tx.Hash)
		if _, exist := a.added[id]; !exist {
			a.added[id] = true
			if dt, exist := a.succ[id]; exist {
				if dt > 0 {
					time.Sleep(dt)
				}
			}
			context.Respond(&message.MemPoolPutRsp{Err: nil})
		} else {
			context.Respond(&message.MemPoolPutRsp{Err: types.ErrTxAlreadyInMempool})
		}
	}
}
func (a *MPStub) BeforeStart() {}
func (a *MPStub) AfterStart()  {}
func (a *MPStub) BeforeStop()  {}

func (a *MPStub) Statistics() *map[string]interface{} {
	stat := make(map[string]interface{})
	return &stat
}
