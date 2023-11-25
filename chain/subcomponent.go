package chain

import (
	"fmt"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-actor/router"
	"github.com/aergoio/aergo/v2/pkg/component"
)

// SubComponent handles message with Receive(), and requests to other actor services with IComponentRequester
// To use SubComponent, only need to implement Actor interface
type SubComponent struct {
	actor.Actor
	component.IComponentRequester // use baseComponent to request to other actors

	name  string
	pid   *actor.PID
	count int
}

func NewSubComponent(subactor actor.Actor, requester *component.BaseComponent, name string, cntWorker int) *SubComponent {
	return &SubComponent{
		Actor:               subactor,
		IComponentRequester: requester,
		name:                name,
		count:               cntWorker}
}

// spawn new subactor
func (sub *SubComponent) Start() {
	sub.pid = actor.Spawn(router.NewRoundRobinPool(sub.count).WithInstance(sub.Actor))

	msg := fmt.Sprintf("%s[%d] started", sub.name, sub.count)
	logger.Info().Msg(msg)
}

// stop subactor
func (sub *SubComponent) Stop() {
	sub.pid.GracefulStop()
	msg := fmt.Sprintf("%s stoped", sub.name)
	logger.Info().Msg(msg)
}

// send message to this subcomponent and reply to actor with pid respondTo
func (sub *SubComponent) Request(message interface{}, respondTo *actor.PID) {
	sub.pid.Request(message, respondTo)
}
