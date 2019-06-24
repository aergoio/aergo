package p2putil

// ChannelPipe serve non blocking limited size channel.
// It preserve input ordering, and not block caller unless it is Closed()
// Tt must be called Open before using it, and Close for dispose resource.
type ChannelPipe interface {
	// Put item to pipe. it should be used after Open() method is called.
	// It always returns true and guaranty that item is queued.
	Put(item interface{}) bool
	Out() <-chan interface{}
	// Done should be called after get item from out channel
	Done()

	Open()
	Close()
}

type channelPipe struct {
	in   chan interface{}
	out  chan interface{}
	done chan interface{}

	queue *PressableQueue
	stop  chan interface{}

	listener PipeEventListener
}

// PipeEventListener listen event of ChannelPipe
type PipeEventListener interface {
	// OnIn is called when item is queued
	OnIn(element interface{})
	// OnDrop is called when queued item is dropped and not out to channel receiver
	OnDrop(element interface{})
	// OnOut is called when queued item went to out channel (and will be sent to receiver)
	OnOut(element interface{})
}

// NewDefaultChannelPipe create pipe to output channel out
func NewDefaultChannelPipe(bufSize int, listener PipeEventListener) ChannelPipe {
	return newDefaultChannelPipe(bufSize, listener)
}

// newDefaultChannelPipe create pipe to output channel out
func newDefaultChannelPipe(bufSize int, listener PipeEventListener) *channelPipe {
	if listener == nil {
		listener = &StatListener{}
	}
	c := &channelPipe{
		in:   make(chan interface{}),
		out:  make(chan interface{}, 1),
		done: make(chan interface{}),

		queue: NewPressableQueue(bufSize),
		stop:  make(chan interface{}),

		listener: listener,
	}

	return c
}

func (c *channelPipe) Put(item interface{}) bool {
	c.in <- item
	return true
}

func (c *channelPipe) Out() <-chan interface{} {
	return c.out
}

func (c *channelPipe) Done() {
	c.done <- struct{}{}
}

func (c *channelPipe) Open() {
	go c.run()
}
func (c *channelPipe) Close() {
	c.stop <- struct{}{}
}

func (c *channelPipe) run() {
LOOP:
	for {
		select {
		case mo := <-c.in:
			c.listener.OnIn(mo)
			if len(c.out) == 0 {
				if c.queue.Empty() {
					c.pushToOut(mo)
					continue LOOP
				} else {
					c.pushToOut(c.queue.Poll())
					c.queue.Offer(mo)
				}
			} else {
				if dropped := c.queue.Press(mo); dropped != nil {
					c.listener.OnDrop(dropped)
				}
			}
		case <-c.done:
			// Next done will come, if out is not empty
			if len(c.out) == 0 {
				first := c.queue.Poll()
				if first != nil {
					c.pushToOut(first)
				}
			}
		case <-c.stop:
			break LOOP
		}
	}
}

func (c *channelPipe) pushToOut(e interface{}) {
	c.out <- e
	c.listener.OnOut(e)
}
