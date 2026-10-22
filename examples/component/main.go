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
	"github.com/aergoio/aergo/v2/examples/component/message"
	"github.com/aergoio/aergo/v2/examples/component/server"
	"github.com/aergoio/aergo/v2/examples/component/service"
	"github.com/aergoio/aergo/v2/pkg/component"
)

func main() {

	compHub := component.NewComponentHub()

	testServer := &server.TestServer{}
	testServer.BaseComponent = component.NewBaseComponent("TestServer", testServer, log.Default())

	helloService := service.NexExampleServie("Den")

	compHub.Register(testServer)
	compHub.Register(helloService)
	compHub.Start()

	// request and go through
	testServer.RequestTo(message.HelloService, &message.HelloReq{Who: "Roger"})

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
	statics, _ := compHub.Statistics(time.Second, "")
	if data, err := json.MarshalIndent(statics, "", "\t"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("All Component's Statistics: %s\n", data)
	}

	time.Sleep(1 * time.Second)

	compHub.Stop()
}
