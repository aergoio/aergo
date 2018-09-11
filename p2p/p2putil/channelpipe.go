package p2putil

// ChannelPipe serve non blocking limited size channel.
// It preserve input ordering, and not block caller unless it is Closed()
// it has own goroutine, it must be called Open before using it, and Close for dispose resource.
type ChannelPipe interface {
	// In returns channel for input. it should be used after Open() method is called.
	// It cause block if item is pushed after Close() is called.
	In() chan<- interface{}
	Out() <-chan interface{}
	// Done should be called after get item from out channel
	Done() chan<- interface{}

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
	OnIn(element interface{})
	OnDrop(element interface{})
	OnOut(element interface{})
}

// NewDefaultChannelPipe create pipe to output channel out
func NewDefaultChannelPipe(bufsize int, listener PipeEventListener) ChannelPipe {
	return newDefaultChannelPipe(bufsize, listener)
}

// newDefaultChannelPipe create pipe to output channel out
func newDefaultChannelPipe(bufsize int, listener PipeEventListener) *channelPipe {
	if listener == nil {
		listener = &StatListener{}
	}
	c := &channelPipe{
		in:   make(chan interface{}),
		out:  make(chan interface{}, 1),
		done: make(chan interface{}),

		queue: NewPressableQueue(bufsize),
		stop:  make(chan interface{}),

		listener: listener,
	}

	return c
}

func (c *channelPipe) In() chan<- interface{} {
	return c.in
}

func (c *channelPipe) Out() <-chan interface{} {
	return c.out
}

func (c *channelPipe) Done() chan<- interface{} {
	return c.done
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
