/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricReader_Read(t *testing.T) {
	tests := []struct {
		name string
		listeners []ListenAdded
	}{
		{"TNoListener",nil},
		{"TSingle",[]ListenAdded{testCountingListener}},
		{"TMulti",[]ListenAdded{testCountingListener, logListener}},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpSize=0
			tenBytes := []byte("0123456789")
			buf := bytes.NewBuffer(tenBytes)
			r := &MetricReader{rd:buf}
			for _, l := range test.listeners {
				r.AddListener(l)
			}
			outBuf := make([]byte, 100)
			n, err := r.Read(outBuf)
			assert.Equal(t, len(tenBytes), n)
			assert.Nil(t, err)

			assert.True(t, bytes.Equal(tenBytes, outBuf[:len(tenBytes)]))
			if test.listeners != nil {
				assert.Equal(t, int64(len(tenBytes)), tmpSize)
			}
		})
	}
}


func TestMetricWriter_Write(t *testing.T) {
	tests := []struct {
		name string
		listeners []ListenAdded
	}{
		{"TNoListener",nil},
		{"TSingle",[]ListenAdded{testCountingListener}},
		{"TMulti",[]ListenAdded{testCountingListener, logListener}},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpSize=0
			tenBytes := []byte("0123456789")
			buf := &bytes.Buffer{}
			r := &MetricWriter{wt:buf}
			for _, l := range test.listeners {
				r.AddListener(l)
			}
			n, err := r.Write(tenBytes)
			assert.Equal(t, len(tenBytes), n)
			assert.Nil(t, err)

			assert.True(t, bytes.Equal(tenBytes, buf.Bytes()[:len(tenBytes)]))
			if test.listeners != nil {
				assert.Equal(t, int64(len(tenBytes)), tmpSize)
			}
		})
	}
}


var (
	tmpSize int64
)
func testCountingListener(added int) {
	tmpSize += int64(added)
}
func logListener(added int ) {
	fmt.Println(added," bytes added")
}