/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"github.com/aergoio/aergo/types"
	"sync/atomic"
	"time"
)

type PeerMetric struct {
	PeerID types.PeerID

	Since    time.Time
	totalIn  int64
	totalOut int64

	InMetric DataMetric
	OutMetric DataMetric
}

func (m *PeerMetric) TotalIn() int64 {
	return atomic.LoadInt64(&m.totalIn)
}

func (m *PeerMetric) TotalOut() int64 {
	return atomic.LoadInt64(&m.totalOut)
}

func (m *PeerMetric) InputAdded(added int) {
	atomic.AddInt64(&m.totalIn, int64(added))
}

func (m *PeerMetric) OutputAdded(added int) {
	atomic.AddInt64(&m.totalOut, int64(added))
}
