/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/examples/component/message"
	"github.com/aergoio/aergo/pkg/component"
)

type TestServer struct {
	*component.BaseComponent
}

var _ component.IComponent = (*TestServer)(nil)

func (ts *TestServer) Start() {
	ts.BaseComponent.Start(ts)
}

func (ts *TestServer) Receive(context actor.Context) {
	ts.BaseComponent.Receive(context)

	switch msg := context.Message().(type) {
	case *message.HelloRsp:
		ts.Info().Msg(msg.Greeting)

	case *component.CompStatReq:
		context.Respond(
			&component.CompStatRsp{
				"component": ts.BaseComponent.Statics(msg),
			})
	}
}
