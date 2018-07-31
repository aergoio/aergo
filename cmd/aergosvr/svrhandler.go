/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package main

// BaseDispatcher is basic implementation of RequestDispatcher
// type BaseDispatcher struct {
// 	netsrv.RequestDispatcher

// 	listeners []netsrv.Listener
// }

// var dispatcher *BaseDispatcher

// func createDispatcher(cfg *config.Config) *BaseDispatcher {
// 	dispatcher = &BaseDispatcher{}

// 	httpListener := CreateHTTPListener(cfg, dispatcher)
// 	rpcListener := CreateRPCListener(cfg, dispatcher)

// 	dispatcher.listeners = []netsrv.Listener{httpListener, rpcListener}
// 	for _, listener := range dispatcher.listeners {

// 		go listener.Start()
// 	}

// 	return dispatcher
// }

// func (d *BaseDispatcher) Stop() {
// 	for _, listener := range d.listeners {
// 		listener.Stop()
// 	}
// }
