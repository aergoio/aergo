/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricsManager_Stop(t *testing.T) {

	tests := []struct {
		name string
	}{
		{"T1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mm := NewMetricManager(1)
			go mm.Start()

			time.Sleep(time.Millisecond * 230)
			mm.Stop()
		})
	}
}

func TestMetricsManager_Size(t *testing.T) {
	pid, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")

	tests := []struct {
		name string

		// inSize should be small or equal to out
		outSize int
		inSize  int

		remove bool
	}{
		{"Tzero", 0, 0, false},
		{"TzeroRM", 0, 0, true},
		{"TSameInOut", 999, 999, false},
		{"TSameInOutRM", 999, 999, true},
		{"TLeave", 2999, 999, false},
		{"TLeaveRM", 2999, 999, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mm := NewMetricManager(1)
			peerMetric := mm.NewMetric(pid, 1)

			assert.Equal(t, int64(0), mm.deadTotalIn)
			assert.Equal(t, int64(0), mm.deadTotalOut)
			if test.outSize > 0 {
				peerMetric.OnWrite(p2pcommon.PingRequest, test.outSize)
				peerMetric.OnRead(p2pcommon.PingResponse, test.inSize)

				assert.Equal(t, int64(0), mm.deadTotalIn)
				assert.Equal(t, int64(0), mm.deadTotalOut)
			}
			if test.remove {
				result := mm.Remove(pid, 1)
				assert.Equal(t, int64(test.inSize), result.totalIn)
				assert.Equal(t, int64(test.outSize), result.totalOut)
				assert.Equal(t, int64(test.inSize), mm.deadTotalIn)
				assert.Equal(t, int64(test.outSize), mm.deadTotalOut)
			}

			summary := mm.Summary()
			assert.Equal(t, int64(test.inSize), summary["in"].(int64))
			assert.Equal(t, int64(test.outSize), summary["out"].(int64))

		})
	}
}
