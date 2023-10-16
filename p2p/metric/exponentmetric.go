/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"math"
	"sync/atomic"

	"github.com/aergoio/aergo/v2/p2p/p2putil"
)

// this struct calculate roughly approximate mean value.
type exponentMetric struct {
	unCalc        int64
	averageFactor int64

	subtotal int64
	average  int64
	inqueue  *p2putil.PressableQueue

	decayFactor float64
	loadScore   int64
}

func NewExponentMetric5(tickInterval int) DataMetric {
	return NewExponentMetric(tickInterval, 60*5)
}

// NewExponentMetric15
func NewExponentMetric15(tickInterval int) DataMetric {
	return NewExponentMetric(tickInterval, 15*60)
}

func NewExponentMetric(interval int, meanTime int) *exponentMetric {
	decayFactor := math.Exp(-float64(interval) / float64(meanTime))
	// rounded int value
	avr := (meanTime + interval>>1) / interval
	return &exponentMetric{averageFactor: int64(avr), inqueue: p2putil.NewPressableQueue(avr), decayFactor: decayFactor}
}

func (a *exponentMetric) APS() int64 {
	return atomic.LoadInt64(&a.average)
}

func (a *exponentMetric) LoadScore() int64 {
	return atomic.LoadInt64(&a.loadScore)
}

// Update adds n unCalc events.
func (a *exponentMetric) AddBytes(n int) {
	atomic.AddInt64(&a.unCalc, int64(n))
}

func (a *exponentMetric) Calculate() {
	count := atomic.LoadInt64(&a.unCalc)
	atomic.AddInt64(&a.unCalc, -count)
	a.subtotal += count
	out := a.inqueue.Press(count)
	if out != nil {
		a.subtotal -= out.(int64)
	}
	atomic.StoreInt64(&a.average, a.subtotal/int64(a.inqueue.Size()))

	atomic.StoreInt64(&a.loadScore, count+int64(float64(a.loadScore)*a.decayFactor))
}
