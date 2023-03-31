/** @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"math"
	"time"
)

const firstReconnectCoolTime = time.Minute >> 1

var (
	durations []time.Duration
	maxTrial  = 15
)

func init() {
	// It will get [20s 36s 1m6s 2m1s 3m40s 6m42s 12m12s 22m14s 40m30s 1h13m48s 2h14m29s 4h5m2s 7h26m29s 13h33m32s 24h42m21s]
	// 20 sec for dpos
	durations = generateExpDuration(20, 0.6, maxTrial)
	// 3 sec for raft
	//durations = generateExpDuration(3, 0.6, 20)
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
