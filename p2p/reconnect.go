/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"math"
	"time"
)

var durations []time.Duration

const maxTrial = 15

func init() {
	// It will get [20s 36s 1m6s 2m1s 3m40s 6m42s 12m12s 22m14s 40m30s 1h13m48s 2h14m29s 4h5m2s 7h26m29s 13h33m32s 24h42m21s]
	durations = generateExpDuration(20, 0.6, maxTrial)
}

type reconnectRunner struct {
	meta     PeerMeta
	maxTrial int
	pm       PeerManager

	cancel chan struct{}
}

func newReconnectRunner(meta PeerMeta, pm PeerManager) *reconnectRunner {
	return &reconnectRunner{meta: meta, maxTrial: maxTrial, pm: pm, cancel: make(chan struct{}, 1)}
}
func (rr *reconnectRunner) runReconnect() {
	for _, duration := range durations {
		// wait for duration
		select {
		case <-time.NewTimer(duration).C:
			_, found := rr.pm.GetPeer(rr.meta.ID)
			if found {
				// found means that peer is registered in other goroutine. so close and cancel it.
			}
			rr.pm.AddNewPeer(rr.meta)
		case <-rr.cancel:
			return
		}
	}
}

func generateExpDuration(initSecs int, inc float64, count int) []time.Duration {
	durations := make([]time.Duration, 0, count)
	num := float64(0)
	for i := 0; i < count; i++ {
		x := math.Exp(num) * float64(initSecs)
		durations = append(durations, time.Second*time.Duration(math.Round(x)))
		num += inc
	}
	return durations
}
