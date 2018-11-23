/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"sync"
	"sync/atomic"
)

type mutexPipe struct {
	mutex *sync.Mutex

	out  chan interface{}

	queue *PressableQueue
	stop  int32

	listener PipeEventListener
}

// newMutexPipe create pipe to output channel out
func newMutexPipe(bufsize int, listener PipeEventListener) *mutexPipe {
	if listener == nil {
		listener = &StatListener{}
	}
	c := &mutexPipe{
		mutex: new(sync.Mutex),
		out:  make(chan interface{}, 1),

		queue: NewPressableQueue(bufsize),

		listener: listener,
	}

	return c
}

func (c *mutexPipe) Put(mo interface{}) bool {
	// stop is set after this pipe is closed
	if atomic.LoadInt32(&c.stop) != 0 {
		return false
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.listener.OnIn(mo)
	if len(c.out) == 0 {
		if c.queue.Empty() {
			c.pushToOut(mo)
		} else {
			c.pushToOut(c.queue.Poll())
			c.queue.Offer(mo) // this offer always return true
		}
	} else {
		if dropped := c.queue.Press(mo); dropped != nil {
			c.listener.OnDrop(dropped)
		}
	}
	return true
}

func (c *mutexPipe) Out() <-chan interface{} {
	return c.out
}

func (c *mutexPipe) Done() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.out) == 0 && !c.queue.Empty() {
		c.pushToOut(c.queue.Poll())
	}
}

func (c *mutexPipe) Open() {
	atomic.StoreInt32(&c.stop,0)
}
func (c *mutexPipe) Close() {
	atomic.StoreInt32(&c.stop,1)
}

func (c *mutexPipe) pushToOut(e interface{}) {
	c.out <- e
	c.listener.OnOut(e)
}
