/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"github.com/AsynkronIT/protoactor-go/actor"
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
		ts.Info("%v\n", msg.Greeting)
	}
}
