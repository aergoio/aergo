/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
)

func TestClosedChannel(t *testing.T) {
	ca := make(chan int, 10)
	cb := make(chan int)

	go func(size int) {
		for i := 0; i < size; i++ {
			ca <- i
		}
	}(100)
	close(cb)
LOOP:
	for {
		select {
		case v := <-ca:
			t.Log("got val ", v)
			time.Sleep(time.Millisecond << 2)
		case <-cb:
			t.Log("closed")
			break LOOP
		}
	}

	t.Logf("wait closed channel again")
	<-cb
	t.Logf("finished")
}

func BenchmarkMapIteration(b *testing.B) {
	b.SkipNow()
	sizes := []int{100, 1000, 10000, 100000}
	ms := make([]map[types.TxID]*p2pcommon.PeerMeta, len(sizes))
	for i, s := range sizes {
		ms[i] = make(map[types.TxID]*p2pcommon.PeerMeta)
		for j := 0; j < s; j++ {
			ms[i][types.ToTxID([]byte(types.RandomPeerID()))] = &p2pcommon.PeerMeta{}
		}
	}

	benchmarks := []struct {
		name string
		m    map[types.TxID]*p2pcommon.PeerMeta
	}{
		{"B100", ms[0]},
		{"B1000", ms[1]},
		{"B10000", ms[2]},
		{"B100000", ms[3]},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var sum uint64
				for k, p := range bm.m {
					sum++
					_ = k.String()
					_ = p.ProducerIDs
				}
			}
		})
	}
}
