/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"container/list"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
)

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

func resumeDecider(_ interface{}) actor.Directive {
	return actor.ResumeDirective
}

func (base *BaseComponent) Start(inheritant IComponent) {
	skipResumeStrategy := actor.NewOneForOneStrategy(0, 0, resumeDecider)
	workerProps := actor.FromInstance(inheritant).WithGuardian(skipResumeStrategy)
	if base.enableDebugMsg {
		workerProps = workerProps.WithMailbox(newMailBoxLogger(base.hub.Metrics(base.name)))
	}
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
