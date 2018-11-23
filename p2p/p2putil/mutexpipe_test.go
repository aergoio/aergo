/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/stretchr/testify/assert"
)

func TestMutexPipe(t *testing.T) {
	const arrSize = 30
	var mos [arrSize]TestItem
	for i := 0; i < arrSize; i++ {
		mos[i] = &testItem{i}
	}

	logger := log.NewLogger("test")
	tests := []struct {
		name      string
		cap       int
		finishIdx int

		expectMinOut uint64
		expectConsec uint64
		//expectMinOut uint64
		//expectOut    uint64
	}{
		{"tStall", 10, 0, 2, 18},
		{"tmidStall", 10, 5, 7, 13},
		{"tfast", 10, 1000, arrSize - 10, 0},
		//{"tStall", 10, 1, 2, 2},
		//{"tmidStall", 10, 5, 6, 6},
		//{"tfast", 10, 1000, arrSize - 10, 30},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doneC := make(chan int, 1)
			statListener := NewStatLister()
			listener := NewMultiListener(statListener, &logListener{logger}, &orderCheckListener{t: t, outId: -1, dropId: -1})
			//c := newMutexPipe(tt.cap, listener)
			c := newMutexPipe(tt.cap, listener)
			c.Open()
			failCount := 0
			go consumeStall(c, tt.finishIdx, arrSize, doneC)
			for _, mo := range mos {
				if !c.Put(mo) {
					failCount++
				}
				time.Sleep(time.Millisecond)

			}
			consumedCount := <-doneC
			lock := &sync.Mutex{}
			lock.Lock()
			actStat := statListener
			lock.Unlock()

			fmt.Printf("In %d , out %d , failed count %d\n", actStat.incnt, actStat.outcnt, failCount)
			assert.Equal(t, uint64(arrSize), actStat.incnt)
			assert.True(t, actStat.incnt == arrSize)
			if tt.expectConsec == 0 {
				assert.Equal(t, uint64(consumedCount), actStat.outcnt)
				assert.Equal(t, actStat.incnt, actStat.outcnt)
			} else {
				assert.Equal(t, uint64(consumedCount+1), actStat.outcnt)
				assert.Equal(t, actStat.incnt, actStat.outcnt+actStat.consecdrop+uint64(tt.cap))
			}

			//c.Close()
		})
	}
}

func TestMutexPipe_nonBlockWriteChan2(t *testing.T) {
	const arrSize = 30
	var mos [arrSize]TestItem
	for i := 0; i < arrSize; i++ {
		mos[i] = &testItem{i}
	}
	logger := log.NewLogger("test")

	tests := []struct {
		name     string
		cap      int
		stallIdx int

		expectMinOut uint64
		expectConsec uint64
	}{
		{"tfast", 10, 1000, arrSize - 10, 0},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doneC := make(chan int, 1)
			statListener := NewStatLister()
			listener := NewMultiListener(statListener, &logListener{logger}, &orderCheckListener{t: t, outId: -1, dropId: -1})
			c := newMutexPipe(tt.cap, listener)
			c.Open()

			go consumeStall2(c, tt.stallIdx, arrSize, doneC)
			for _, mo := range mos {
				for !c.Put(mo) {
					time.Sleep(time.Millisecond<<3)
				}
			}
			consumeCount := <-doneC
			lock := &sync.Mutex{}
			lock.Lock()
			actStat := statListener
			lock.Unlock()

			fmt.Printf("In %d , out %d , consecutive drop %d\n", actStat.incnt, actStat.outcnt, actStat.consecdrop)
			assert.True(t, actStat.incnt == arrSize)
			assert.Equal(t, uint64(consumeCount), actStat.outcnt)

			c.Close()
		})
	}
}

func TestMutexPipe_Longterm(t *testing.T) {
	// skip unit test in normal time..
	t.SkipNow()
	const arrSize = 30
	var mos [arrSize]TestItem
	for i := 0; i < arrSize; i++ {
		mos[i] = &testItem{i}
	}

	logger := log.NewLogger("test")
	tests := []struct {
		name     string
		cap      int
		testTime time.Duration
	}{
		{"tlong", 20, time.Second * 10},
		{"tlong", 20, time.Second * 11},
		{"tlong", 20, time.Second * 12},
		{"tlong", 20, time.Second * 13},
		{"tlong", 20, time.Second * 14},
		{"tlong", 20, time.Second * 15},
		{"tlong", 20, time.Second * 16},
		{"tlong", 20, time.Second * 17},
		{"tlong", 20, time.Second * 18},
		{"tlong", 20, time.Second * 19},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doneC := make(chan int, 1)
			finish := make(chan interface{})
			statListener := NewStatLister()
			listener := NewMultiListener(statListener, &logListener{logger})
			c := newMutexPipe(tt.cap, listener)
			c.Open()

			go consumeForLongterm(c, tt.testTime+time.Minute, doneC, finish)
			expire := time.Now().Add(tt.testTime)

			i := 0
			for time.Now().Before(expire) {
				c.Put(mos[i%arrSize])
				time.Sleep(time.Millisecond * 5)
				i++

			}
			finish <- struct{}{}
			consumeCount := <-doneC
			lock := &sync.Mutex{}
			lock.Lock()
			actStat := statListener
			rqueue := c.queue
			lock.Unlock()

			fmt.Printf("In %d , out %d , drop %d, consecutive drop %d\n", actStat.incnt, actStat.outcnt, actStat.dropcnt, actStat.consecdrop)
			assert.Equal(t, actStat.incnt, uint64(i))
			// last one is in channel and not consumed
			assert.Equal(t, uint64(consumeCount+1), actStat.outcnt)
			// in should equal to sum of out, drop, and remained in queue
			assert.Equal(t, actStat.incnt, actStat.outcnt+actStat.dropcnt+uint64(rqueue.Size()))

			c.Close()
		})
	}
}

func TestMutexPipe_MultiLoads(t *testing.T) {
	// skip unit test in normal time..
	// t.SkipNow()
	const threadsize = 100
	const arrSize = 30
	var mos [threadsize][arrSize]TestItem
	for j := 0; j < threadsize; j++ {
		for i := 0; i < arrSize; i++ {
			mos[j][i] = &testItem2{testItem{i}, j}
		}
	}

	// logger := log.NewLogger("test")
	tests := []struct {
		name     string
		cap      int
		testTime time.Duration
	}{
		{"tlong", 20, time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doneC := make(chan int, 1)
			finish := make(chan interface{})
			statListener := NewStatLister()
			listener := NewMultiListener(statListener)
			c := newDefaultChannelPipe(tt.cap, listener)
			c.Open()

			go consumeForLongterm(c, tt.testTime+time.Minute, doneC, finish)
			wg := sync.WaitGroup{}
			expire := time.Now().Add(tt.testTime)
			wg.Add(threadsize)
			for j := 0; j < threadsize; j++ {
				go func(tid int) {
					i := 0
					for time.Now().Before(expire) {
						c.Put(mos[i%arrSize])
						i++
					}
					wg.Done()
				}(j)
			}
			wg.Wait()
			finish <- struct{}{}
			consumeCount := <-doneC
			lock := &sync.Mutex{}
			lock.Lock()
			actStat := statListener
			rqueue := c.queue
			lock.Unlock()

			fmt.Printf("In %d , out %d , drop %d, consecutive drop %d\n", actStat.incnt, actStat.outcnt, actStat.dropcnt, actStat.consecdrop)
			// There are two cases, one is last one is in channel and not consumed, and another is consumed all items.
			assert.True(t, actStat.outcnt-uint64(consumeCount) <= 1 )
			// in should equal to sum of out, drop, and remained in queue
			assert.Equal(t, actStat.incnt, actStat.outcnt+actStat.dropcnt+uint64(rqueue.Size()))

			c.Close()
		})
	}
}
