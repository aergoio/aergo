/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package rpc

import (
	"fmt"
	"net"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	aergorpc "github.com/aergoio/aergo/types"
	"google.golang.org/grpc"
)

// RPC is actor for providing rpc service
type RPC struct {
	conf *config.Config

	*component.BaseComponent

	hub *component.ComponentHub

	grpcServer   *grpc.Server
	actualServer aergorpc.AergoRPCServiceServer
}

//var _ component.IComponent = (*RPCComponent)(nil)

// NewRPC create an rpc service
func NewRPC(hub *component.ComponentHub, cfg *config.Config) *RPC {
	actualServer := &AergoRPCService{
		hub:       hub,
		msgHelper: message.GetHelper(),
	}
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 256),
	}

	grpcServer := grpc.NewServer(opts...)

	netsvc := &RPC{
		conf:          cfg,
		BaseComponent: component.NewBaseComponent("rpc", logger, cfg.EnableDebugMsg),
		hub:           hub,
		grpcServer:    grpcServer,
		actualServer:  actualServer,
	}
	actualServer.actorHelper = netsvc

	return netsvc
}

// Start start rpc service
func (ns *RPC) Start() {
	ns.BaseComponent.Start(ns)
	go ns.serve()
}

// Stop stops rpc service
func (ns *RPC) Stop() {
	ns.grpcServer.Stop()
	ns.BaseComponent.Stop()
}

func (ns *RPC) serve() {

	aergorpc.RegisterAergoRPCServiceServer(ns.grpcServer, ns.actualServer)

	ipAddr := net.ParseIP(ns.conf.RPC.NetServiceAddr)
	if ipAddr == nil {
		//TODO: warning?
		panic("wrong IP address format")
	}
	conn, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ipAddr, ns.conf.RPC.NetServicePort))
	if err != nil {
		panic(err)
	}

	err = ns.grpcServer.Serve(conn)
	if err != nil {
		panic(err)
	}

	return
}

const defaultTTL = time.Second * 4

// SendRequest implement interface method of ActorService
func (ns *RPC) SendRequest(actor string, msg interface{}) {
	ns.hub.Request(actor, msg, ns)
}

// FutureRequest implement interface method of ActorService
func (ns *RPC) FutureRequest(actor string, msg interface{}) *actor.Future {
	return ns.hub.RequestFuture(actor, msg, defaultTTL, "rpc.(*RPC).FutureRequest")
}

// CallRequest implement interface method of ActorService
func (ns *RPC) CallRequest(actor string, msg interface{}) (interface{}, error) {
	future := ns.hub.RequestFuture(actor, msg, defaultTTL, "rpc.(*RPC).CallRequest")

	return future.Result()
}
