/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"testing"
	"time"
)

func TestClosedChannel(t *testing.T) {
	ca := make(chan int, 10)
	cb := make(chan int)

	go func(size int) {
		for i:=0; i<size; i++ {
			ca <- i
		}
	}(100)
	close(cb)
	LOOP:
	for {
		select {
		case v := <-ca:
			t.Log("got val ", v)
			time.Sleep(time.Millisecond<<2)
		case <-cb:
			t.Log("closed")
			break LOOP
		}
	}

	t.Logf("wait closed channel again")
	<-cb
	t.Logf("finished")
}
