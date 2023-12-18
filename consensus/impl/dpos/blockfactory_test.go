package dpos

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlockFactory_context(t *testing.T) {
	quitC := make(chan interface{})
	factory := &BlockFactory{quit: quitC}
	factory.initContext()

	timeoutCtx, _ := context.WithTimeout(factory.ctx, time.Millisecond)
	contextDeadline, _ := timeoutCtx.Deadline()
	cancelCtx, cancelFunc := context.WithTimeout(factory.ctx, time.Minute)
	quitCtx, _ := context.WithTimeout(factory.ctx, time.Minute)

	// first channel is done by deadline
	select {
	case <-timeoutCtx.Done():
		err := timeoutCtx.Err()
		assert.Equal(t, context.DeadlineExceeded, err, "unexpected error type")
		// check if timeout occured not before deadline
		assert.True(t, time.Now().After(contextDeadline))
	case <-time.NewTimer(5 * time.Second).C:
		assert.Fail(t, "timeout did not occur within expected time frame")
	}

	// second channel is canceled by self cancel func
	cancelFunc()
	select {
	case <-cancelCtx.Done():
		err := cancelCtx.Err()
		assert.Equalf(t, context.Canceled, err, "unexpected error type")
		//assert.Equal(t, err, context.Cause(cancelCtx), "cause and err is differ")
	default:
		assert.Fail(t, "cancel expected, but not")
	}

	// third channel is canceled by parent with cause
	close(quitC)
	<-factory.ctx.Done()
	assert.Equal(t, context.Canceled, factory.ctx.Err(), "unexpected error type")
	//assert.Equal(t, chain.ErrQuit, context.Cause(factory.ctx), "factory cause ErrQuit expected")
	select {
	case <-quitCtx.Done():
		err := quitCtx.Err()
		assert.Equalf(t, context.Canceled, err, "unexpected error type")
		//assert.Equal(t, chain.ErrQuit, context.Cause(quitCtx), "cause ErrQuit expected")
	default:
		assert.Fail(t, "errQuit expected, but not")
	}
}
