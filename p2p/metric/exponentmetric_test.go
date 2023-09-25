/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"math"
	"math/rand"
	"testing"

	"github.com/aergoio/aergo-lib/log"
)

var logger = log.NewLogger("metric")

func TestExponentMetric_Calculate(t *testing.T) {
	lenIn := 200
	constIn := []int64{779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779, 779}
	varIn := make([]int64, lenIn)
	for i := 0; i < lenIn; i++ {
		varIn[i] = rand.Int63() & 0x0fffffff
	}

	tests := []struct {
		name     string
		interval int
		meantime int
		inBytes  []int64

		//expectLF []int64
	}{
		{"TConst", 15, 30, constIn},
		{"TVarin", 5, 60, varIn},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := NewExponentMetric(test.interval, test.meantime)

			//errorRatio := 1.0
			var total int64 = 0
			for i, in := range test.inBytes {
				m.AddBytes(int(in))
				total += in
				realbps := total / int64(i+1)
				m.Calculate()
				diffRatio := math.Abs(float64(realbps-m.APS()) / float64(realbps))
				logger.Debug().Msgf("%03d: in %10d, Total %11d, expectBps %9d, aps %9d, diffR %.4f loadScore %10d \n", i, in, total, realbps, m.APS(), diffRatio, m.LoadScore())
				//assert.Equal(t, realbps, m.APS())
				// assert.True(t, diffRatio <= errorRatio)
				//errorRatio = diffRatio
				//assert.Equal(t, test.expectLF[i], m.LoadScore())
			}
		})
	}
}
