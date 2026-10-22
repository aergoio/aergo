/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/examples/component/message"
	"github.com/aergoio/aergo/v2/pkg/component"
)

type TestServer struct {
	*component.BaseComponent
}

func (ts *TestServer) BeforeStart() {
	// do nothing
}

func (cs *TestServer) AfterStart() {
	// do nothing
}

func (ts *TestServer) BeforeStop() {

	// add stop logics for this service
}

func (ts *TestServer) Statistics() *map[string]interface{} {
	return nil
}

func (ts *TestServer) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *message.HelloRsp:
		ts.Info().Msg(msg.Greeting)
	}
}
