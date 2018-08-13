/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package main

import (
	"fmt"
	"time"

	"github.com/aergoio/aergo/examples/component/message"
	"github.com/aergoio/aergo/examples/component/server"
	"github.com/aergoio/aergo/examples/component/service"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
)

func main() {

	compHub := component.NewComponentHub()

	testServer := &server.TestServer{
		BaseComponent: component.NewBaseComponent("TestServer", log.Default(), true),
	}
	helloService := service.NexExampleServie("Den")

	compHub.Register(testServer)
	compHub.Register(helloService)
	compHub.Start()

	// request and go through
	compHub.Request(message.HelloService, &message.HelloReq{Who: "Roger"}, testServer)

	// request and wait
	rawResponse, err := compHub.RequestFuture(message.HelloService, &component.StatusReq{}, time.Second, "examples/component.main").Result()
	if err != nil {
		fmt.Println(err)
	} else {
		response := rawResponse.(*component.StatusRsp)
		fmt.Printf("Hello Component Status is : %d\n", response.Status)
	}

	// collect all component's statuses
	statusMap := compHub.Status()
	fmt.Println(statusMap)

	compHub.Stop()

	time.Sleep(1 * time.Second)
}
