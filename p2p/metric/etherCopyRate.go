/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"math"
	"sync"
	"sync/atomic"
)

type etherCopyRate struct {
	mutex sync.Mutex
	uncounted int64
	// calculated 
	rate float64
	alpha float64
	init bool
}


func NewECRate5(tickInterval int) *etherCopyRate {
	ticIntF := float64(tickInterval)
	return NewECRate(1 - math.Exp(-ticIntF/60.0/5))
}

func NewECRate(alpha float64) *etherCopyRate {
	return &etherCopyRate{alpha: alpha}
}


func (a *etherCopyRate) APS() uint64 {
	return uint64(a.RateF())
}

func (a *etherCopyRate) LoadScore() uint64 {
	return uint64(a.RateF())
}

// APS returns the moving average rate of events per second.
func (a *etherCopyRate) RateF() float64 {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.rate * float64(1e9)
}

func (a *etherCopyRate) Calculate() {
	count := atomic.LoadInt64(&a.uncounted)
	atomic.AddInt64(&a.uncounted, -count)
	instantRate := float64(count) / float64(5e9)
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.init {
		a.rate += a.alpha * (instantRate - a.rate)
	} else {
		a.init = true
		a.rate = instantRate
	}
}

// Update adds n uncounted events.
func (a *etherCopyRate) AddBytes(n int64) {
	atomic.AddInt64(&a.uncounted, n)
}
