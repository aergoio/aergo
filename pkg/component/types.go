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
	Status() Status

	Receive(actor.Context)

	SetHub(hub *ComponentHub)
	Hub() *ComponentHub
}

type Statics struct {
	Status            string `json:"status"`
	ProcessedMsg      uint64 `json:"acc_processed_msg"`
	QueuedMsg         uint64 `json:"acc_queued_msg"`
	MsgProcessLatency string `json:"msg_latency"`
	Error             string `json:"error"`
}
