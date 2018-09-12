package p2putil

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/stretchr/testify/assert"
)

func TestChannelpipe(t *testing.T) {
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
		{"tStall", 10, 0, 2, 18},
		{"tmidStall", 10, 5, 7, 13},
		{"tfast", 10, 1000, arrSize - 10, 0},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doneC := make(chan int, 1)
			statListener := NewStatLister()
			listener := NewMultiListener(statListener, &logListener{logger}, &orderCheckListener{t: t, outId: -1, dropId: -1})
			c := newDefaultChannelPipe(tt.cap, listener)
			c.Open()
			go consumeStall(c, tt.stallIdx, arrSize, doneC)
			for _, mo := range mos {
				c.In() <- mo
				time.Sleep(time.Millisecond)
			}
			consumeCount := <-doneC
			lock := &sync.Mutex{}
			lock.Lock()
			actStat := statListener
			lock.Unlock()

			fmt.Printf("In %d , out %d , consecutive drop %d\n", actStat.incnt, actStat.outcnt, actStat.consecdrop)
			assert.True(t, actStat.incnt == arrSize)
			if tt.expectConsec == 0 {
				assert.Equal(t, uint64(consumeCount), actStat.outcnt)
				assert.Equal(t, actStat.incnt, actStat.outcnt)
			} else {
				assert.Equal(t, uint64(consumeCount+1), actStat.outcnt)
				assert.Equal(t, actStat.incnt, actStat.outcnt+actStat.consecdrop+uint64(tt.cap))
			}

			c.Close()
		})
	}
}

func Test_nonBlockWriteChan2(t *testing.T) {
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
			c := newDefaultChannelPipe(tt.cap, listener)
			c.Open()

			go consumeStall2(c, tt.stallIdx, arrSize, doneC)
			for _, mo := range mos {
				c.In() <- mo
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

func consumeStall(wc *channelPipe, idx int, maxCnt int, doneChannel chan<- int) {
	arrs := make([]int, 0, maxCnt)
	cnt := 0
LOOP:
	for cnt < maxCnt {
		select {
		case <-time.NewTimer(time.Millisecond * 200).C:
			fmt.Printf("Internal expiretime is out \n")
			break LOOP
		case mo := <-wc.Out():
			cnt++
			arrs = append(arrs, mo.(TestItem).ID())
			wc.Done() <- mo
			// fmt.Printf("Consuming mo %s \n", mo.GetRequestID())
			if cnt >= idx {
				fmt.Printf("Finishing consume after index %d \n", cnt)
				break LOOP
			}
		}
	}
	fmt.Println("Consumed ", arrs)
	doneChannel <- cnt
}

func consumeStall2(wc *channelPipe, idx int, maxCnt int, doneChannel chan<- int) {
	arrs := make([]int, 0, maxCnt)
	cnt := 0
LOOP:
	for cnt < maxCnt {
		select {
		case <-time.NewTimer(time.Millisecond * 200).C:
			fmt.Printf("Internal expiretime is out \n")
			break LOOP
		case mo := <-wc.Out():
			cnt++
			if cnt%4 == 3 {
				time.Sleep(time.Millisecond)
			}
			arrs = append(arrs, mo.(TestItem).ID())
			wc.Done() <- mo
			// fmt.Printf("Consuming mo %s \n", mo.GetRequestID())
		}
	}
	fmt.Println("Consumed ", arrs)
	doneChannel <- cnt
}

type logListener struct {
	logger *log.Logger
}

func (l *logListener) OnIn(element interface{}) {
	l.logger.Info().Int("id", element.(TestItem).ID()).Msg("In")
}

func (l *logListener) OnDrop(element interface{}) {
	l.logger.Info().Int("id", element.(TestItem).ID()).Msg("Drop")
}

func (l *logListener) OnOut(element interface{}) {
	l.logger.Info().Int("id", element.(TestItem).ID()).Msg("Out")
}

type orderCheckListener struct {
	t      *testing.T
	outId  int
	dropId int
}

func (l *orderCheckListener) OnIn(element interface{}) {
}

func (l *orderCheckListener) OnDrop(element interface{}) {
	id := element.(TestItem).ID()
	assert.Truef(l.t, id > l.dropId, "dropId expected higer thant %d, but %d", l.dropId, id)
	l.dropId = id
}

func (l *orderCheckListener) OnOut(element interface{}) {
	id := element.(TestItem).ID()
	assert.Truef(l.t, id > l.outId, "outId expected higer thant %d, but %d", l.outId, id)
	l.outId = id
}

func TestLongterm(t *testing.T) {
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
			c := newDefaultChannelPipe(tt.cap, listener)
			c.Open()

			go consumeForLongterm(c, tt.testTime+time.Minute, doneC, finish)
			expire := time.Now().Add(tt.testTime)

			i := 0
			for time.Now().Before(expire) {
				c.In() <- mos[i%arrSize]
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

func TestMultiLoads(t *testing.T) {
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
			for j := 0; j < threadsize; j++ {
				go func(tid int) {
					wg.Add(1)
					i := 0
					for time.Now().Before(expire) {
						c.In() <- mos[i%arrSize]
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
			// last one is in channel and not consumed
			assert.Equal(t, uint64(consumeCount+1), actStat.outcnt)
			// in should equal to sum of out, drop, and remained in queue
			assert.Equal(t, actStat.incnt, actStat.outcnt+actStat.dropcnt+uint64(rqueue.Size()))

			c.Close()
		})
	}
}

type testItem2 struct {
	testItem
	routineId int
}

func consumeForLongterm(wc *channelPipe, ttl time.Duration, doneChannel chan<- int, finishChannel <-chan interface{}) {
	finishTime := time.NewTimer(ttl)
	cnt := 0
LOOP:
	for {
		select {
		case <-finishChannel:
			fmt.Printf("Finish loop by signal\n")
			break LOOP
		case <-finishTime.C:
			fmt.Printf("Finish loop by time expire \n")
			break LOOP
		case mo := <-wc.Out():
			cnt++
			time.Sleep(time.Millisecond >> 2)
			wc.Done() <- mo
		}
	}
	fmt.Printf("Consumed %d items\n", cnt)
	doneChannel <- cnt
}
