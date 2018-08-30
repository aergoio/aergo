/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"time"

	"github.com/aergoio/aergo-actor/actor"
)

// IComponent provides a common interface for easy management
// and providing a communication channel between components
// BaseComponent struct provides general implementation of this
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

// IActor describes functions that each components have to implement
// A BeforeStart func is called before a IComponent.Start func
// So developers can write component specific initalization codes in here
// A BeforeStop func is called before a IComponent.Stop func
// In a Receive func, component's actions is described
// For each type of message, developer can define a behavior
// If there is component specific statics or debug info are exists,
// than developers can get those by defining it in Statics func
type IActor interface {
	BeforeStart()
	AfterStart()
	BeforeStop()

	Receive(actor.Context)

	Statics() *map[string]interface{}
}
