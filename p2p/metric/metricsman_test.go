/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"bytes"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
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

func TestMetricsManager_Remove(t *testing.T) {
	pid, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")

	tests := []struct {
		name string

		// inSize should be small or equal to out
		outSize int
		inSize int

		remove bool
	}{
		{"Tzero", 0,0, false},
		{"TzeroRM", 0,0, true},
		{"TSameInOut", 999,999, false},
		{"TSameInOutRM", 999,999, true},
		{"TLeave", 2999,999, false},
		{"TLeaveRM", 2999,999, true},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			mm := NewMetricManager(1)
			rd := NewReader(buf)
			wt := NewWriter(buf)
			mm.Add(pid, rd, wt)

			assert.Equal(t, int64(0), mm.deadTotalIn)
			assert.Equal(t, int64(0), mm.deadTotalOut)
			if test.outSize > 0 {
				written, _ := wt.Write(make([]byte,test.outSize))
				assert.Equal(t, test.outSize, written)
				tempBuf := make([]byte, test.inSize)
				read,_ := rd.Read(tempBuf)
				assert.Equal(t, test.inSize, read)

				assert.Equal(t, int64(0), mm.deadTotalIn)
				assert.Equal(t, int64(0), mm.deadTotalOut)
			}
			if test.remove {
				result := mm.Remove(pid)
				assert.Equal(t, int64(test.inSize), result.totalIn)
				assert.Equal(t, int64(test.outSize), result.totalOut)
				assert.Equal(t, int64(test.inSize),  mm.deadTotalIn)
				assert.Equal(t, int64(test.outSize), mm.deadTotalOut)
			}

			summary := mm.Summary()
			assert.Equal(t, int64(test.inSize), summary["in"].(int64))
			assert.Equal(t, int64(test.outSize), summary["out"].(int64))

		})
	}
}
