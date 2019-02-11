/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"math"
	"time"

	"github.com/aergoio/aergo-lib/log"
)

var durations []time.Duration

var maxTrial = 15

func init() {
	// It will get [20s 36s 1m6s 2m1s 3m40s 6m42s 12m12s 22m14s 40m30s 1h13m48s 2h14m29s 4h5m2s 7h26m29s 13h33m32s 24h42m21s]
	durations = generateExpDuration(20, 0.6, maxTrial)
}

type reconnectJob struct {
	meta   p2pcommon.PeerMeta
	trial  int
	rm     ReconnectManager
	pm     PeerManager
	logger *log.Logger

	cancel chan struct{}
}

func newReconnectRunner(meta p2pcommon.PeerMeta, rm ReconnectManager, pm PeerManager, logger *log.Logger) *reconnectJob {
	return &reconnectJob{meta: meta, trial: 0, rm: rm, pm: pm, cancel: make(chan struct{}, 1), logger: logger}
}
func (rj *reconnectJob) runJob() {
	timer := time.NewTimer(getNextInterval(rj.trial))
RETRYLOOP:
	for {
		// wait for duration
		select {
		case <-timer.C:
			_, found := rj.pm.GetPeer(rj.meta.ID)
			if found {
				break RETRYLOOP
			}
			rj.logger.Debug().Str("peer_meta", rj.meta.String()).Int("trial", rj.trial).Msg("Trying to connect")
			rj.pm.AddNewPeer(rj.meta)
			rj.trial++
			timer.Reset(getNextInterval(rj.trial))
		case <-rj.cancel:
			break RETRYLOOP
		}
	}
	rj.rm.jobFinished(rj.meta.ID)
}

func getNextInterval(trial int) time.Duration {
	if trial < maxTrial {
		return durations[trial]
	}
	return durations[maxTrial-1]
}

func generateExpDuration(initSecs int, inc float64, count int) []time.Duration {
	arr := make([]time.Duration, 0, count)
	num := float64(0)
	for i := 0; i < count; i++ {
		x := math.Exp(num) * float64(initSecs)
		arr = append(arr, time.Second*time.Duration(math.Round(x)))
		num += inc
	}
	return arr
}
