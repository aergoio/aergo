/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"fmt"
	"time"
)

// Callable
type Callable interface {
	// DoCall run function. it should put result anything if call is over. It also stop  function if Cancel was called as soon as possible
	DoCall(done chan<- interface{})
	// Cancel should return instanly
	Cancel()
}

// InvokeWithTimer call DoCall method of m and return if m is finished or return error if timer fires.
func InvokeWithTimer(m Callable, timer *time.Timer) (interface{}, error) {
	done := make(chan interface{}, 1)
	go m.DoCall(done)
	select {
	case hsResult := <-done:
		return hsResult, nil
	case <-timer.C:
		m.Cancel()
		return nil, fmt.Errorf("timeout")
	}
}

