/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/examples/component/message"
	"github.com/aergoio/aergo/examples/component/server"
	"github.com/aergoio/aergo/examples/component/service"
	"github.com/aergoio/aergo/pkg/component"
)

func main() {

	compHub := component.NewComponentHub()

	testServer := &server.TestServer{
		BaseComponent: component.NewBaseComponent("TestServer", log.Default()),
	}
	helloService := service.NexExampleServie("Den")

	compHub.Register(testServer)
	compHub.Register(helloService)
	compHub.Start()

	// request and go through
	compHub.Request(message.HelloService, &message.HelloReq{Who: "Roger"}, testServer)

	// request and wait
	rawResponse, err := compHub.RequestFuture(message.HelloService,
		&component.CompStatReq{SentTime: time.Now()},
		time.Second, "examples/component.main").Result()
	if err != nil {
		fmt.Println(err)
	} else {
		response := rawResponse.(*component.CompStatRsp)
		fmt.Printf("RequestFuture Test Result: %v\n", response)
	}

	// collect all component's statuses
	statics := compHub.Profile(time.Second)
	if data, err := json.MarshalIndent(statics, "", "\t"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("All Component's Statics: %s\n", data)
	}

	time.Sleep(1 * time.Second)

	compHub.Stop()
}
