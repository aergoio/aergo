/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package service

import (
	"fmt"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/examples/component/message"
	"github.com/aergoio/aergo/v2/pkg/component"
)

type ExampleService struct {
	*component.BaseComponent
	myname string
}

func NexExampleServie(myname string) *ExampleService {
	actor := &ExampleService{

		myname: myname,
	}
	actor.BaseComponent = component.NewBaseComponent(message.HelloService, actor, log.Default())

	return actor
}

func (es *ExampleService) BeforeStart() {
	// add init logics for this service
}

func (es *ExampleService) AfterStart() {
	// add init logics for this service
}

func (es *ExampleService) BeforeStop() {

	// add stop logics for this service
}

func (es *ExampleService) Statistics() *map[string]interface{} {
	return nil
}

func (es *ExampleService) Receive(context actor.Context) {

	switch msg := context.Message().(type) {
	case *message.HelloReq:
		context.Respond(
			&message.HelloRsp{
				Greeting: fmt.Sprintf("Hello %s, I'm %s", msg.Who, es.myname),
			})
	}

}
