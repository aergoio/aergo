/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"sync/atomic"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
)

var _ IComponent = (*BaseComponent)(nil)

type ActorSpan struct {
	span opentracing.Span
}

// BaseComponent provides a basic implementations for IComponent interface
type BaseComponent struct {
	*log.Logger
	IActor
	name            string
	pid             *actor.PID
	status          Status
	hub             *ComponentHub
	accProcessedMsg uint64
	inbounds        []actor.InboundMiddleware
	outbounds       []actor.OutboundMiddleware
}

// NewBaseComponent is a helper to create BaseComponent
// This func requires this component's name, implementation of IActor, and
// logger to record internal log msg
// Setting a logger with a same name with the component is recommended
func NewBaseComponent(name string, actor IActor, logger *log.Logger) *BaseComponent {
	return &BaseComponent{
		Logger:          logger,
		IActor:          actor,
		name:            name,
		pid:             nil,
		status:          StoppedStatus,
		hub:             nil,
		accProcessedMsg: 0,
	}
}

// GetName returns a name of this component
func (base *BaseComponent) GetName() string {
	return base.name
}

// resumeDecider advices a behavior when panic is occured during receving a msg
// A component, which its strategy is this, will throw away a current failing msg
// and just keep going to process a next msg
func resumeDecider(_ interface{}) actor.Directive {
	return actor.ResumeDirective
}

// Start inits internal modules and spawns actor process
// let this component
func (base *BaseComponent) Start() {
	// call a init func, defined at an actor's implementation
	base.IActor.BeforeStart()

	skipResumeStrategy := actor.NewOneForOneStrategy(0, 0, resumeDecider)

	inbound := func(next actor.ActorFunc) actor.ActorFunc {
		fn := func(c actor.Context) {
			parentSpanId := c.MessageHeader().Get("opentracing-span")
			parentSpan := base.hub.RestoreSpan(parentSpanId)
			var span opentracing.Span

			if nil == parentSpan {
				span = opentracing.StartSpan(base.name)
			} else {
				span = opentracing.StartSpan(
					base.name,
					opentracing.ChildOf((*parentSpan).Context()))
			}
			spanId := base.hub.SaveSpan(span)
			defer base.hub.DestroySpan(spanId)

			next(c)
		}
		return fn
	}
	outbound := func(next actor.SenderFunc) actor.SenderFunc {
		fn := func(c actor.Context, target *actor.PID, envelope *actor.MessageEnvelope) {
			if nil == envelope.Header {
				envelope.Header = make(map[string]string)
			}
			parentSpanId := c.MessageHeader().Get("opentracing-span")
			parentSpan := base.hub.RestoreSpan(parentSpanId)
			var span opentracing.Span

			if nil == parentSpan {
				span = opentracing.StartSpan(base.name)
			} else {
				span = opentracing.StartSpan(
					base.name,
					opentracing.ChildOf((*parentSpan).Context()))
			}
			spanId := base.hub.SaveSpan(span)
			defer base.hub.DestroySpan(spanId)

			envelope.Header.Set("opentracing-span", spanId)

			next(c, target, envelope)
		}
		return fn
	}

	workerProps := actor.FromInstance(base).
		WithGuardian(skipResumeStrategy).
		WithMiddleware(inbound).
		WithOutboundMiddleware(outbound)

	var err error
	// create and spawn an actor using the name as an unique id
	base.pid, err = actor.SpawnNamed(workerProps, base.GetName())
	// if a same name of pid already exists, retry by attaching a sequential id
	// from actor.ProcessRegistry
	for ; err != nil; base.pid, err = actor.SpawnPrefix(workerProps, base.GetName()) {
		base.Warn().Err(err).Msg("actor name is duplicate")
	}

	// Wait for the messaging hub to be fully initialized. - Incomplete
	// initialization leads to a crash.
	hubInit.wait()

	base.IActor.AfterStart()
}

// Stop lets this component stop and terminate
func (base *BaseComponent) Stop() {
	// call a cleanup func, defined at an actor's implementation
	base.IActor.BeforeStop()

	base.pid.Stop()
	base.pid = nil
}

// Tell passes a given message to this component and forgets
func (base *BaseComponent) Tell(message interface{}) {
	if base.pid == nil {
		logger.Warn().Msg("PID is empty")
		return // do nothing
	}
	base.pid.Tell(message)
}

// TellTo tells (sends and forgets) a message to a target component
// Internally this component will try to find the target component
// using a hub set
func (base *BaseComponent) TellTo(targetCompName string, message interface{}) {
	if base.hub == nil {
		logger.Warn().Str("from", base.GetName()).Str("to", targetCompName).Interface("msg", message).Msg("Hub is not set")
		return // do nothing
	}
	base.hub.Tell(targetCompName, message)
}

// Request passes a given message to this component.
// And a message sender will expect to get a response in form of
// an actor request
func (base *BaseComponent) Request(message interface{}, sender *actor.PID) {
	if base.pid == nil {
		logger.Warn().Str("to", base.GetName()).Str("from", sender.GetId()).Interface("msg", message).Msg("PID is empty")
		return // do nothing
	}
	base.pid.Request(message, sender)
}

// RequestTo passes a given message to a target component
// And a message sender, this component, will expect to get a response
// from the target component in form of an actor request
func (base *BaseComponent) RequestTo(targetCompName string, message interface{}) {
	if base.hub == nil {
		logger.Warn().Str("to", targetCompName).Str("from", base.GetName()).Interface("msg", message).Msg("Hub is not set")
		return // do nothing
	}
	targetComp := base.hub.Get(targetCompName)
	targetComp.Request(message, base.pid)
}

// RequestFuture is similar with Request; passes a given message to this component.
// And this returns a future, that represent an asynchronous result
func (base *BaseComponent) RequestFuture(message interface{}, timeout time.Duration, tip string) *actor.Future {
	if base.pid == nil {
		logger.Warn().Str("to", base.GetName()).Str("from", tip).Interface("msg", message).Msg("PID is empty")
		retFuture := actor.NewFuturePrefix("NilPID", timeout)
		retFuture.PID().Tell(ErrHubUnregistered)

		return retFuture
	}

	return base.pid.RequestFuturePrefix(message, tip, timeout)
}

// RequestToFuture is similar with RequestTo; passes a given message to this component.
// And this returns a future, that represent an asynchronous result
func (base *BaseComponent) RequestToFuture(targetCompName string, message interface{}, timeout time.Duration) *actor.Future {
	if base.hub == nil {
		logger.Warn().Str("from", base.GetName()).Str("to", targetCompName).Interface("msg", message).Msg("Hub is not set")
		retFuture := actor.NewFuturePrefix("NilHub", timeout)
		retFuture.PID().Tell(ErrHubUnregistered)

		return retFuture
	}

	return base.hub.RequestFuture(targetCompName, message, timeout, base.name)
}

// SetHub assigns a component hub to be used internally
func (base *BaseComponent) SetHub(hub *ComponentHub) {
	base.hub = hub
}

// Hub returns a component hub set
func (base *BaseComponent) Hub() *ComponentHub {
	return base.hub
}

// MsgQueueLen gives a number of queued msgs in this component's mailbox
func (base *BaseComponent) MsgQueueLen() int32 {
	return base.pid.MsgNum()
}

// Receive in the BaseComponent handles system messages and invokes actor's
// receive function; implementation to handle incomming messages
func (base *BaseComponent) Receive(context actor.Context) {
	switch msg := context.Message().(type) {

	case *actor.Started:
		atomic.SwapUint32(&base.status, StartedStatus)

	case *actor.Stopping:
		atomic.SwapUint32(&base.status, StoppingStatus)

	case *actor.Stopped:
		atomic.SwapUint32(&base.status, StoppedStatus)

	case *actor.Restarting:
		atomic.SwapUint32(&base.status, RestartingStatus)

	case *CompStatReq:
		base.accProcessedMsg++
		context.Respond(base.statics(msg))

	default:
		base.accProcessedMsg++
	}

	base.IActor.Receive(context)
}

// Status returns status of this component; started, stopped, stopping, restarting
// This func is thread-safe
func (base *BaseComponent) Status() Status {
	return atomic.LoadUint32(&base.status)
}

func (base *BaseComponent) statics(req *CompStatReq) *CompStatRsp {
	thisMsgLatency := time.Now().Sub(req.SentTime)

	return &CompStatRsp{
		Status:            StatusToString(base.status),
		AccProcessedMsg:   base.accProcessedMsg,
		MsgQueueLen:       uint64(base.pid.MsgNum()),
		MsgProcessLatency: thisMsgLatency.String(),
		Actor:             base.IActor.Statistics(),
	}
}
