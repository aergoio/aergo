/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-actor/mailbox"
	"github.com/aergoio/aergo-lib/log"
)

var _ IComponent = (*BaseComponent)(nil)

type BaseComponent struct {
	*log.Logger
	IActor
	name            string
	pid             *actor.PID
	status          Status
	hub             *ComponentHub
	accQueuedMsg    uint64
	accProcessedMsg uint64
}

func NewBaseComponent(name string, actor IActor, logger *log.Logger) *BaseComponent {
	return &BaseComponent{
		Logger:          logger,
		IActor:          actor,
		name:            name,
		pid:             nil,
		status:          StoppedStatus,
		hub:             nil,
		accQueuedMsg:    0,
		accProcessedMsg: 0,
	}
}

func (base *BaseComponent) GetName() string {
	return base.name
}

func resumeDecider(_ interface{}) actor.Directive {
	return actor.ResumeDirective
}

func (base *BaseComponent) Start() {
	// call a init func, defined at an actor's implementation
	base.IActor.BeforeStart()

	skipResumeStrategy := actor.NewOneForOneStrategy(0, 0, resumeDecider)
	workerProps := actor.FromInstance(base).WithGuardian(skipResumeStrategy).WithMailbox(mailbox.Unbounded(base))

	var err error
	// create and spawn an actor using the name as an unique id
	base.pid, err = actor.SpawnNamed(workerProps, base.GetName())
	// if a same name of pid already exists, retry by attaching a sequential id
	// from actor.ProcessRegistry
	for ; err != nil; base.pid, err = actor.SpawnPrefix(workerProps, base.GetName()) {
		//TODO add log msg
	}

	// Wait for the messaging hub to be fully initilized. - Incomplete
	// initilization leads to a crash.
	hubInit.wait()
}

func (base *BaseComponent) Stop() {
	// call a cleanup func, defined at an actor's implementation
	base.IActor.BeforeStop()

	base.pid.Stop()
	base.pid = nil
}

func (base *BaseComponent) Tell(message interface{}) {
	if base.pid == nil {
		panic("PID is empty")
	}
	base.pid.Tell(message)
}

func (base *BaseComponent) TellTo(targetCompName string, message interface{}) {
	if base.hub == nil {
		panic("Component hub is not set")
	}
	base.hub.Tell(targetCompName, message)
}

func (base *BaseComponent) Request(message interface{}, sender *actor.PID) {
	if base.pid == nil {
		panic("PID is empty")
	}
	base.pid.Request(message, sender)
}

func (base *BaseComponent) RequestTo(targetCompName string, message interface{}) {
	if base.hub == nil {
		panic("Component hub is not set")
	}
	targetComp := base.hub.Get(targetCompName)
	targetComp.Request(message, base.pid)
}

func (base *BaseComponent) RequestFuture(message interface{}, timeout time.Duration, tip string) *actor.Future {
	if base.pid == nil {
		panic("PID is empty")
	}

	return base.pid.RequestFuturePrefix(message, tip, timeout)
}

func (base *BaseComponent) RequestToFuture(targetCompName string, message interface{}, timeout time.Duration) *actor.Future {
	if base.hub == nil {
		panic("Component hub is not set")
	}

	return base.hub.RequestFuture(targetCompName, message, timeout, base.name)
}

func (base *BaseComponent) SetHub(hub *ComponentHub) {
	base.hub = hub
}

func (base *BaseComponent) Hub() *ComponentHub {
	return base.hub
}

func (base *BaseComponent) Receive(context actor.Context) {
	base.accProcessedMsg++

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
		context.Respond(base.statics(msg))
	}

	base.IActor.Receive(context)
}

func (base *BaseComponent) Status() Status {
	return atomic.LoadUint32(&base.status)
}

func (base *BaseComponent) statics(req *CompStatReq) *CompStatRsp {
	thisMsgLatency := time.Now().Sub(req.SentTime)

	return &CompStatRsp{
		Status:            StatusToString(base.status),
		ProcessedMsg:      base.accProcessedMsg,
		QueuedMsg:         base.accQueuedMsg,
		MsgProcessLatency: thisMsgLatency.String(),
		Actor:             base.IActor.Statics(),
	}
}

// MessagePosted is called when a msg is inserted at a mailbox (or queue) of this component
// At this time, BaseComponent accumulates its counter to get a number of queued msgs
func (base *BaseComponent) MessagePosted(message interface{}) {
	base.accQueuedMsg++
}

// MessageReceived is called when msg is handled by the Receive func
// This does nothing, but needs to implement Mailbox Statics interface
func (base *BaseComponent) MessageReceived(message interface{}) {}

// MailboxStarted does nothing, but needs to implement Mailbox Statics interface
func (base *BaseComponent) MailboxStarted() {}

// MailboxEmpty does nothing, but needs to implement Mailbox Statics interface
func (base *BaseComponent) MailboxEmpty() {}
