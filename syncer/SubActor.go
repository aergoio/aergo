package syncer

import (
	"fmt"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-actor/router"
	"github.com/aergoio/aergo/pkg/component"
)

type SubActor struct {
	actor actor.Actor
	*actor.PID

	hub *component.ComponentHub //for communicate with other service

	name  string
	count int
}

func newSubActor(actor actor.Actor, name string, cntWorker int, hub *component.ComponentHub) *SubActor {
	return &SubActor{
		actor: actor,
		name:  name,
		count: cntWorker,
		hub:   hub,
	}
}

func (subactor *SubActor) start() {
	if subactor == nil {
		panic("subactor is nil")
	}

	subactor.PID = actor.Spawn(router.NewRoundRobinPool(subactor.count).WithInstance(subactor.actor))

	msg := fmt.Sprintf("%s[%d] :pid[%s] started", subactor.name, subactor.count, subactor.PID.String())
	logger.Info().Msg(msg)
}

func (subactor *SubActor) stop() {
	subactor.GracefulStop()
	msg := fmt.Sprintf("%s stoped", subactor.name)
	logger.Info().Msg(msg)
}

func (subactor *SubActor) Receive(context actor.Context) {
}
