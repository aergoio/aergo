/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"time"

	"github.com/aergoio/aergo-actor/actor"
)

type IComponent interface {
	GetName() string
	Start()
	Stop()
	Status() Status
	SetHub(hub *ComponentHub)
	Hub() *ComponentHub

	Tell(message interface{})
	Request(message interface{}, sender *actor.PID)
	RequestFuture(message interface{}, timeout time.Duration, tip string) *actor.Future

	Receive(actor.Context)
}

type IActor interface {
	BeforeStart()
	BeforeStop()

	Receive(actor.Context)

	Statics() interface{}
}
