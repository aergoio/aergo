/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package service

import (
	"fmt"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/examples/component/message"
	"github.com/aergoio/aergo/pkg/component"
)

type ExampleService struct {
	*component.BaseComponent
	myname string
}

var _ component.IComponent = (*ExampleService)(nil)

func NexExampleServie(myname string) *ExampleService {
	return &ExampleService{
		BaseComponent: component.NewBaseComponent(message.HelloService, log.Default()),
		myname:        myname,
	}
}

func (es *ExampleService) Start() {
	es.BaseComponent.Start(es)
	//TODO add init logics for this service
}

func (es *ExampleService) Stop() {
	es.BaseComponent.Stop()
	//TODO add stop logics for this service
}

func (es *ExampleService) Receive(context actor.Context) {
	es.BaseComponent.Receive(context)

	switch msg := context.Message().(type) {
	case *message.HelloReq:
		context.Respond(
			&message.HelloRsp{
				Greeting: fmt.Sprintf("Hello %s, I'm %s", msg.Who, es.myname),
			})
	case *component.CompStatReq:
		context.Respond(
			&component.CompStatRsp{
				"component": es.BaseComponent.Statics(msg),
				"hello": map[string]interface{}{
					"name": es.myname,
				},
			})
	}

}
