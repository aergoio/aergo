/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
)

type PeerMetric struct {
	mm     MetricsManager
	PeerID types.PeerID
	seq    uint32

	Since    time.Time
	totalIn  int64
	totalOut int64

	InMetric  DataMetric
	OutMetric DataMetric
}

var _ p2pcommon.MsgIOListener = (*PeerMetric)(nil)

func (m *PeerMetric) OnRead(protocol p2pcommon.SubProtocol, read int) {
	atomic.AddInt64(&m.totalIn, int64(read))
	m.InMetric.AddBytes(read)
}

func (m *PeerMetric) OnWrite(protocol p2pcommon.SubProtocol, write int) {
	atomic.AddInt64(&m.totalOut, int64(write))
	m.OutMetric.AddBytes(write)
}

func (m *PeerMetric) TotalIn() int64 {
	return atomic.LoadInt64(&m.totalIn)
}

func (m *PeerMetric) TotalOut() int64 {
	return atomic.LoadInt64(&m.totalOut)
}

// Deprecated
func (m *PeerMetric) InputAdded(added int) {
	atomic.AddInt64(&m.totalIn, int64(added))
}

// Deprecated
func (m *PeerMetric) OutputAdded(added int) {
	atomic.AddInt64(&m.totalOut, int64(added))
}
