package dpos

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBlockFactory_context(t *testing.T) {
	quitC := make(chan interface{})
	factory := &BlockFactory{quit: quitC}
	factory.initContext()

	timeoutCtx, _ := context.WithTimeout(factory.ctx, time.Millisecond)
	cancelCtx, cancelFunc := context.WithTimeout(factory.ctx, time.Minute)
	quitCtx, _ := context.WithTimeout(factory.ctx, time.Minute)
	time.Sleep(time.Millisecond * time.Duration(2))

	// first channel is done by deadline
	select {
	case <-timeoutCtx.Done():
		err := timeoutCtx.Err()
		assert.Equal(t, context.DeadlineExceeded, err, "err by deadline")
		//assert.Equal(t, err, context.Cause(timeoutCtx), "cause and err is differ")
	default:
		assert.Fail(t, "deadline expected, but not")
	}

	// second channel is canceled by self cancel func
	cancelFunc()
	select {
	case <-cancelCtx.Done():
		err := cancelCtx.Err()
		assert.Equalf(t, context.Canceled, err, "err by cancel")
		//assert.Equal(t, err, context.Cause(cancelCtx), "cause and err is differ")
	default:
		assert.Fail(t, "cancel expected, but not")
	}

	// third channel is canceled by parent with cause
	close(quitC)
	<-factory.ctx.Done()
	assert.Equal(t, context.Canceled, factory.ctx.Err(), "factory err ErrQuit expected")
	//assert.Equal(t, chain.ErrQuit, context.Cause(factory.ctx), "factory cause ErrQuit expected")
	select {
	case <-quitCtx.Done():
		err := quitCtx.Err()
		assert.Equalf(t, context.Canceled, err, "err by quit")
		//assert.Equal(t, chain.ErrQuit, context.Cause(quitCtx), "cause ErrQuit expected")
	default:
		assert.Fail(t, "errQuit expected, but not")
	}
}
