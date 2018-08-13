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
	Request(message interface{}, sender IComponent)
	RequestFuture(message interface{}, timeout time.Duration, tip string) *actor.Future
	Pid() *actor.PID

	Receive(actor.Context)

	SetHub(hub *ComponentHub)
	Hub() *ComponentHub
}
