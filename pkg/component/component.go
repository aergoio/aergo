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

type BaseComponent struct {
	*log.Logger
	name            string
	pid             *actor.PID
	status          Status
	hub             *ComponentHub
	accQueuedMsg    uint64
	accProcessedMsg uint64
}

func NewBaseComponent(name string, logger *log.Logger) *BaseComponent {
	return &BaseComponent{
		Logger:          logger,
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

func (base *BaseComponent) Start(inheritant IComponent) {
	skipResumeStrategy := actor.NewOneForOneStrategy(0, 0, resumeDecider)
	workerProps := actor.FromInstance(inheritant).WithGuardian(skipResumeStrategy).WithMailbox(mailbox.Unbounded(base))

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
	base.pid.Stop()
	base.pid = nil
}

func (base *BaseComponent) Request(message interface{}, sender IComponent) {

	if base.pid != nil {
		base.pid.Request(message, sender.Pid())
	} else {
		log.Default().Fatal().Msg("PID is empty")
	}
}

func (base *BaseComponent) RequestFuture(message interface{}, timeout time.Duration, tip string) *actor.Future {

	if base.pid == nil {
		log.Default().Fatal().Msg("PID is empty")
	}

	return base.pid.RequestFuturePrefix(message, tip, timeout)
}

func (base *BaseComponent) Pid() *actor.PID {
	return base.pid
}

func (base *BaseComponent) SetHub(hub *ComponentHub) {
	base.hub = hub
}

func (base *BaseComponent) Hub() *ComponentHub {
	return base.hub
}

func (base *BaseComponent) Receive(context actor.Context) {
	base.accProcessedMsg++

	switch context.Message().(type) {

	case *actor.Started:
		//base.Info("Started, initialize actor here")
		atomic.SwapUint32(&base.status, StartedStatus)

	case *actor.Stopping:
		//base.Info("Stopping, actor is about shut down")
		atomic.SwapUint32(&base.status, StoppingStatus)

	case *actor.Stopped:
		//base.Info("Stopped, actor and it's children are stopped")
		atomic.SwapUint32(&base.status, StoppedStatus)

	case *actor.Restarting:
		//base.Info("Restarting, actor is about restart")
		atomic.SwapUint32(&base.status, RestartingStatus)
	}
}

func (base *BaseComponent) Status() Status {
	return atomic.LoadUint32(&base.status)
}

func (base *BaseComponent) Statics(req *CompStatReq) *Statics {
	thisMsgLatency := time.Now().Sub(req.SentTime)

	return &Statics{
		Status:            StatusToString(base.status),
		ProcessedMsg:      base.accProcessedMsg,
		QueuedMsg:         base.accQueuedMsg,
		MsgProcessLatency: thisMsgLatency.String(),
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
