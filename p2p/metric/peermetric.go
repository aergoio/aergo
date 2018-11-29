/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"github.com/libp2p/go-libp2p-peer"
	"time"
)

type PeerMetric struct {
	PeerID peer.ID

	Since time.Time
	SumIn int64
	SumOut int64

	InMetric DataMetric
	OutMetric DataMetric
}

func (m *PeerMetric) InputAdded(added int) {
	m.SumIn += int64(added)
}

func (m *PeerMetric) OutputAdded(added int) {
	m.SumOut += int64(added)
}