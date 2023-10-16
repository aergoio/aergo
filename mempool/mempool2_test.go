package mempool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/types"
)

func TestCheckExpired(t *testing.T) {
	for i := -20; i < 10; i++ {
		now := time.Now()
		tTime := now.Add(time.Hour * time.Duration(i))
		eTime := now.Add(-1 * evictPeriod)
		b1 := time.Since(tTime) < evictPeriod
		b2 := tTime.After(eTime)
		if b1 != b2 {
			t.Errorf("CheckResult is differ for time %v :  %v and %v ", tTime, b1, b2)
		}
	}
}

func BenchmarkTimerCheck(b *testing.B) {
	pool := make(map[types.AccountID]*txList)
	st := types.NewState()
	totSize := 0
	for i := 0; i < 8000; i++ {
		acc := types.AccountID(types.ToHashID([]byte(types.RandomPeerID())))
		txCnt := i / 2
		totSize += txCnt
		l := newTxList(types.HashID(acc).Bytes(), st, nil)
		l.list = make([]types.Transaction, txCnt)
		for j := 0; j < txCnt; j++ {
			tx := types.NewTx()
			tx.Body.Nonce = uint64(j + 1)
			l.list[j] = types.NewTransaction(tx)
		}
		pool[acc] = l
	}
	cache := sync.Map{}
	ttl := 4 * time.Millisecond
	benchmarks := []struct {
		name string
		fn   func(pool map[types.AccountID]*txList, cache sync.Map, ttl time.Duration) (int, time.Duration)
	}{
		{"BNoTO", noTimeout},
		{"BCtx", timeoutWithCtx},
		{"BTimer", timeoutWithTimer},
		{"BTimer2", timeoutWithTimer2},
		{"BNowFn", timeoutWithNow},
		{"BNowFn64", timeoutWithNow64},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tot, et := bm.fn(pool, cache, ttl)
				if tot < totSize {
					b.Logf("Not all tx processed proc %d, total %d, in ttl %v", tot, totSize, et)
				}
			}
		})
	}
}

func noTimeout(pool map[types.AccountID]*txList, cache sync.Map, ttl time.Duration) (int, time.Duration) {
	start := time.Now()
	//expireT := start.Add(ttl)
	total := 0
	for _, list := range pool {
		if time.Since(list.GetLastModifiedTime()) < evictPeriod {
			continue
		}
		txs := list.GetAll()
		total += len(txs)

		//for _, tx := range txs {
		//	cache.Delete(types.ToTxID(tx.GetHash()))
		//}
	}
	elapsed := time.Now().Sub(start)
	return total, elapsed
}

func timeoutWithNow(pool map[types.AccountID]*txList, cache sync.Map, ttl time.Duration) (int, time.Duration) {
	start := time.Now()
	expireT := start.Add(ttl)
	total := 0
	for _, list := range pool {
		if !time.Now().Before(expireT) {
			break
		}
		if time.Since(list.GetLastModifiedTime()) < evictPeriod {
			continue
		}
		txs := list.GetAll()
		total += len(txs)

		//for _, tx := range txs {
		//	cache.Delete(types.ToTxID(tx.GetHash()))
		//}
	}
	elapsed := time.Now().Sub(start)
	return total, elapsed
}

func timeoutWithNow64(pool map[types.AccountID]*txList, cache sync.Map, ttl time.Duration) (int, time.Duration) {
	start := time.Now()
	expireT := start.Add(ttl)
	total := 0
	cnt := 0
	for _, list := range pool {
		cnt++
		if (cnt&0x03f) == 0 && !time.Now().Before(expireT) {
			break
		}
		if time.Since(list.GetLastModifiedTime()) < evictPeriod {
			continue
		}
		txs := list.GetAll()
		total += len(txs)

		//for _, tx := range txs {
		//	cache.Delete(types.ToTxID(tx.GetHash()))
		//}
	}
	elapsed := time.Now().Sub(start)
	return total, elapsed
}

func timeoutWithTimer(pool map[types.AccountID]*txList, cache sync.Map, ttl time.Duration) (int, time.Duration) {
	start := time.Now()
	expireT := time.NewTimer(ttl)
	total := 0

L:
	for _, list := range pool {
		select {
		case <-expireT.C:
			break L
		default:
		}

		if time.Since(list.GetLastModifiedTime()) < evictPeriod {
			continue
		}
		txs := list.GetAll()
		total += len(txs)

		//for _, tx := range txs {
		//	cache.Delete(types.ToTxID(tx.GetHash()))
		//}
	}
	elapsed := time.Now().Sub(start)
	return total, elapsed
}
func timeoutWithTimer2(pool map[types.AccountID]*txList, cache sync.Map, ttl time.Duration) (int, time.Duration) {
	start := time.Now()
	evictTime := start.Add(-evictPeriod)
	expireT := time.NewTimer(ttl)
	total := 0

L:
	for _, list := range pool {
		select {
		case <-expireT.C:
			break L
		default:
		}

		if list.GetLastModifiedTime().After(evictTime) {
			continue
		}
		txs := list.GetAll()
		total += len(txs)

		//for _, tx := range txs {
		//	cache.Delete(types.ToTxID(tx.GetHash()))
		//}
	}
	elapsed := time.Now().Sub(start)
	return total, elapsed
}

func timeoutWithCtx(pool map[types.AccountID]*txList, cache sync.Map, ttl time.Duration) (int, time.Duration) {
	start := time.Now()
	expireT, _ := context.WithTimeout(context.TODO(), ttl)
	total := 0
L:
	for _, list := range pool {
		select {
		case <-expireT.Done():
			break L
		default:
		}

		if time.Since(list.GetLastModifiedTime()) < evictPeriod {
			continue
		}
		txs := list.GetAll()
		total += len(txs)

		//for _, tx := range txs {
		//	cache.Delete(types.ToTxID(tx.GetHash()))
		//}
	}
	elapsed := time.Now().Sub(start)
	return total, elapsed
}
