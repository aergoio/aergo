/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/aergoio/aergo/pkg/log"
)

type IComponent interface {
	GetName() string

	Start()
	Stop()
	Request(message interface{}, sender IComponent)
	RequestFuture(message interface{}, timeout time.Duration) *actor.Future
	Pid() *actor.PID

	Receive(actor.Context)

	SetHub(hub *ComponentHub)
	Hub() *ComponentHub

	log.ILogger
}
