/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"container/list"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/aergoio/aergo/pkg/log"
)

var _ log.ILogger = (*BaseComponent)(nil)

type BaseComponent struct {
	*log.Logger
	name            string
	pid             *actor.PID
	status          Status
	hub             *ComponentHub
	enableDebugMsg  bool
	msgCount        map[string]*list.List
	lastHandleMsgTs int64
}

func NewBaseComponent(name string, logger *log.Logger, enableDebugMsg bool) *BaseComponent {
	return &BaseComponent{
		Logger:          logger,
		name:            name,
		pid:             nil,
		status:          StoppedStatus,
		hub:             nil,
		enableDebugMsg:  enableDebugMsg,
		msgCount:        make(map[string]*list.List),
		lastHandleMsgTs: time.Now().UnixNano(),
	}
}

func (base *BaseComponent) GetName() string {
	return base.name
}

func (base *BaseComponent) Start(inheritant IComponent) {
	workerProps := actor.FromInstance(inheritant)
	if base.enableDebugMsg {
		workerProps = workerProps.WithMailbox(newMailBoxLogger(base.hub.Metrics(base.name)))
	}
	base.pid = actor.Spawn(workerProps)
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
		base.Fatal("PID is empty")
	}
}

func (base *BaseComponent) RequestFuture(message interface{}, timeout time.Duration) *actor.Future {

	if base.pid == nil {
		base.Fatal("PID is empty")
	}

	return base.pid.RequestFuture(message, timeout)
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

	switch context.Message().(type) {

	case *actor.Started:
		//base.Info("Started, initialize actor here")
		base.status = StartedStatus

	case *actor.Stopping:
		//base.Info("Stopping, actor is about shut down")
		base.status = StoppingStatus

	case *actor.Stopped:
		//base.Info("Stopped, actor and it's children are stopped")
		base.status = StoppedStatus

	case *actor.Restarting:
		//base.Info("Restarting, actor is about restart")
		base.status = RestartingStatus
	}
}
