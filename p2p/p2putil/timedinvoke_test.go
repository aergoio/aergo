/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

var logger *log.Logger

func init() {
	logger = log.NewLogger("test")
}

func Test_InvokeWithTimerSimple(t *testing.T) {
	tests := []struct {
		name string
		ttl time.Duration
		iteration int
		expectErr bool
		expected interface{}
	}{
		{"TFast", time.Second, 2, false, 2},
		{"TTimeout", time.Millisecond*200, 1000, true,nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := &samplConstCaller{iteration:test.iteration}
			actual, err := InvokeWithTimer(m, time.NewTimer(test.ttl))
			assert.Equal(t, test.expectErr, err != nil)
			if !test.expectErr {
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

type samplConstCaller struct {
	iteration int
	cancel int32
}

func (c *samplConstCaller) DoCall(done chan<- interface{}) {
	i := 0
	for ; i < c.iteration ; i++ {
		fmt.Printf("%d th iteration \n",i)
		time.Sleep(time.Millisecond*50)
		if atomic.LoadInt32(&c.cancel) != 0 {
			// actually no put to done channel is needed.
			return
		}
	}
	done <- i
}

func (c *samplConstCaller) Cancel() {
	atomic.StoreInt32(&c.cancel,1)
}


func Test_InvokeWithTimerRetainedTime(t *testing.T) {
	// some test will be resulted by process time. so run it manually and make skip it in CI test
	t.SkipNow()
	tests := []struct {
		name string
		ttl time.Duration
		iteration int
		retainCount int
		expectErr bool
		expected interface{}
	}{

		{"TTimeout", time.Millisecond*100, 4, 1, true,nil},
		{"TSuccRetain", time.Millisecond*100, 4, 5, false,4},
		{"TTooLongIfRetain", time.Millisecond*20, 1000, 100, true,nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := &ratainableCaller{iteration:test.iteration, retainCount:test.retainCount, retainDuration:test.ttl, timer:time.NewTimer(test.ttl)}
			actual, err := InvokeWithTimer(m, m.timer)
			assert.Equal(t, test.expectErr, err != nil)
			if !test.expectErr {
				assert.Equal(t, test.expected, actual)
			}
			logger.Info().Str("name",test.name).Msg("Test finished")

		})
	}
	time.Sleep(time.Millisecond*200)
}


type ratainableCaller struct {
	retainCount int
	retainDuration time.Duration
	timer *time.Timer

	iteration int
	cancel int32
}

func (c *ratainableCaller) DoCall(done chan<- interface{}) {
	retainCnt := c.retainCount
	i := 0
	for ; i < c.iteration ; i++ {
		fmt.Printf("%d th iteration \n",i)
		time.Sleep(time.Millisecond*50)
		if retainCnt > 0 {
			// do not retain if timer was already expired. and return soon because cancel will be set.
			if !c.timer.Stop() {
				logger.Info().Msg("Timer already expired.")
				logger.Info().Msg("make place holder.")
				time.Sleep(time.Millisecond*100)
				logger.Info().Msg("out")
				return
			}
			c.timer.Reset(c.retainDuration)
			retainCnt--
			logger.Info().Msg("Retained Timer")
		}
		if atomic.LoadInt32(&c.cancel) != 0 {
			// actually no put to done channel is needed.
			return
		}
	}
	done <- i
}

func (c *ratainableCaller) Cancel() {
	atomic.StoreInt32(&c.cancel,1)
}