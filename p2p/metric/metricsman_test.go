/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"testing"
	"time"
)

func TestMetricsManager_Stop(t *testing.T) {

	tests := []struct {
		name string
	}{
		{"T1"},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mm := NewMetricManager(1)
			go mm.Start()

			time.Sleep(time.Millisecond * 230 )
			mm.Stop()
		})
	}
}
