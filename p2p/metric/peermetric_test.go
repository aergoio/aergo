/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestPeerMetric_OnRead(t *testing.T) {
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
			target := PeerMetric{mm: mm, PeerID: pid, seq: 1, InMetric: NewExponentMetric5(mm.interval), OutMetric: NewExponentMetric5(mm.interval), Since: time.Now()}

			if test.outSize > 0 {
				target.OnRead(0, test.inSize)
				target.InMetric.Calculate()
				assert.Equal(t, int64(test.inSize), target.TotalIn())
				assert.True(t, target.InMetric.LoadScore() > 0)
				assert.True(t, target.InMetric.LoadScore() <= target.TotalIn())
			}
		})
	}
}
