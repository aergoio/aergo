/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

var NotSupportError = fmt.Errorf("not supported cmd")

type PolarisRPC struct {
	*component.BaseComponent

	conf   *config.Config
	logger *log.Logger

	actorHelper p2pcommon.ActorService

	grpcServer   *grpc.Server
	actualServer *PRPCServer
}

func NewPolarisRPC(cfg *config.Config) *PolarisRPC {
	logger := log.NewLogger("prpc")
	actualServer := &PRPCServer{
		logger: logger,
	}
	tracer := opentracing.GlobalTracer()
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 8),
	}

	if cfg.RPC.NetServiceTrace {
		opts = append(opts, grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)))
		opts = append(opts, grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)))
	}
	grpcServer := grpc.NewServer(opts...)

	rpc := &PolarisRPC{
		conf:         cfg,
		logger:       logger,
		grpcServer:   grpcServer,
		actualServer: actualServer,
	}
	rpc.BaseComponent = component.NewBaseComponent(message.PolarisRPCSvc, rpc, logger)
	actualServer.actorHelper = rpc

	return rpc
}

func (rpc *PolarisRPC) SetHub(hub *component.ComponentHub) {
	rpc.actualServer.hub = hub
	rpc.BaseComponent.SetHub(hub)
}

// Start start rpc service.
func (rpc *PolarisRPC) BeforeStart() {
	types.RegisterPolarisRPCServiceServer(rpc.grpcServer, rpc.actualServer)
}

func (rpc *PolarisRPC) AfterStart() {
	go rpc.serve()
}

// Stop stops rpc service.
func (rpc *PolarisRPC) BeforeStop() {
	rpc.grpcServer.Stop()
}

// Statistics show statistic information of p2p module. NOTE: It it not implemented yet
func (rpc *PolarisRPC) Statistics() *map[string]interface{} {
	return nil
}

func (rpc *PolarisRPC) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	default:
		// maybe BaseComponent have processed this message.
		rpc.Debug().Str("msgType", reflect.TypeOf(msg).String()).Msg("msg not directly handled by rpc received")
	}
}

// Serve TCP multiplexer
func (rpc *PolarisRPC) serve() {
	ipAddr := net.ParseIP(rpc.conf.RPC.NetServiceAddr)
	if ipAddr == nil {
		panic("Wrong IP address format in RPC.NetServiceAddr")
	}

	addr := fmt.Sprintf("%s:%d", ipAddr, rpc.conf.RPC.NetServicePort)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	rpc.Info().Msg(fmt.Sprintf("Starting RPC server listening on %s, with TLS: %v", addr, rpc.conf.RPC.NSEnableTLS))
	if rpc.conf.RPC.NSEnableTLS {
		rpc.Warn().Msg("TLS is enabled in configuration, but currently not supported")
	}

	// Server both servers
	go rpc.serveGRPC(l, rpc.grpcServer)

	return
}

// Serve GRPC server over TCP
func (rpc *PolarisRPC) serveGRPC(l net.Listener, server *grpc.Server) {
	if err := server.Serve(l); err != nil {
		switch err {
		case cmux.ErrListenerClosed:
			// Server killed, usually by ctrl-c signal
		default:
			panic(err)
		}
	}
}

const defaultTTL = time.Second * 4

// TellRequest implement interface method of ActorService
func (rpc *PolarisRPC) TellRequest(actor string, msg interface{}) {
	rpc.TellTo(actor, msg)
}

// SendRequest implement interface method of ActorService
func (rpc *PolarisRPC) SendRequest(actor string, msg interface{}) {
	rpc.RequestTo(actor, msg)
}

// FutureRequest implement interface method of ActorService
func (rpc *PolarisRPC) FutureRequest(actor string, msg interface{}, timeout time.Duration) *actor.Future {
	return rpc.RequestToFuture(actor, msg, timeout)
}

// FutureRequestDefaultTimeout implement interface method of ActorService
func (rpc *PolarisRPC) FutureRequestDefaultTimeout(actor string, msg interface{}) *actor.Future {
	return rpc.RequestToFuture(actor, msg, defaultTTL)
}

// CallRequest implement interface method of ActorService
func (rpc *PolarisRPC) CallRequest(actor string, msg interface{}, timeout time.Duration) (interface{}, error) {
	future := rpc.RequestToFuture(actor, msg, timeout)
	return future.Result()
}

// CallRequest implement interface method of ActorService
func (rpc *PolarisRPC) CallRequestDefaultTimeout(actor string, msg interface{}) (interface{}, error) {
	future := rpc.RequestToFuture(actor, msg, defaultTTL)
	return future.Result()
}
func (rpc *PolarisRPC) GetChainAccessor() types.ChainAccessor {
	return nil
}

type PRPCServer struct {
	logger *log.Logger

	hub         *component.ComponentHub
	actorHelper p2pcommon.ActorService
}

func (rs *PRPCServer) NodeState(ctx context.Context, in *types.NodeReq) (*types.SingleBytes, error) {
	timeout := int64(binary.LittleEndian.Uint64(in.Timeout))
	component := string(in.Component)

	rs.logger.Debug().Str("comp", component).Int64("timeout", timeout).Msg("nodestate")

	statics, err := rs.hub.Statistics(time.Duration(timeout)*time.Second, component)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(statics, "", "\t")
	if err != nil {
		return nil, err
	}
	return &types.SingleBytes{Value: data}, nil
}

func (rs *PRPCServer) Metric(ctx context.Context, in *types.MetricsRequest) (*types.Metrics, error) {
	// TODO: change types.Metrics to support polaris metric informations or refactor to make more generalized structure (i.e change aergo-protobuf first)
	result := &types.Metrics{}

	return result, nil
}

func (rs *PRPCServer) CurrentList(ctx context.Context, p1 *types.Paginations) (*types.PolarisPeerList, error) {
	result, err := rs.actorHelper.CallRequestDefaultTimeout(PolarisSvc,
		&CurrentListMsg{p1.Ref, p1.Size})
	if err != nil {
		return nil, err
	}
	list, ok := result.(*types.PolarisPeerList)
	if !ok {
		return nil, fmt.Errorf("unkown error")
	}
	return list, nil
}

func (rs *PRPCServer) WhiteList(ctx context.Context, p1 *types.Paginations) (*types.PolarisPeerList, error) {
	result, err := rs.actorHelper.CallRequestDefaultTimeout(PolarisSvc,
		&WhiteListMsg{p1.Ref, p1.Size})
	if err == nil {
		list := result.(*types.PolarisPeerList)
		return list, nil
	} else {
		return nil, err
	}
}

func (rs *PRPCServer) BlackList(ctx context.Context, p1 *types.Paginations) (*types.PolarisPeerList, error) {
	result, err := rs.actorHelper.CallRequestDefaultTimeout(PolarisSvc,
		&BlackListMsg{p1.Ref, p1.Size})
	if err == nil {
		list := result.(*types.PolarisPeerList)
		return list, nil
	} else {
		return nil, err
	}
}

func (rs *PRPCServer) ListBLEntries(ctx context.Context, entInfo *types.Empty) (*types.BLConfEntries, error) {
	result, err := rs.actorHelper.CallRequestDefaultTimeout(PolarisSvc, ListEntriesMsg{})
	if err != nil {
		return nil, err
	}
	list, ok := result.(*types.BLConfEntries)
	if !ok {
		return nil, fmt.Errorf("unkown error")
	}
	return list, nil
}

func (rs *PRPCServer) AddBLEntry(ctx context.Context, entInfo *types.AddEntryParams) (*types.SingleString, error) {
	ret := &types.SingleString{}

	if len(entInfo.PeerID) == 0 && len(entInfo.Cidr) == 0 && len(entInfo.Address) == 0 {
		ret.Value = "at least one flags is required"
		return ret, types.RPCErrInvalidArgument
	} else if len(entInfo.Cidr) > 0 && len(entInfo.Address) > 0 {
		ret.Value = "either address or cidr is allowed, not both "
		return ret, types.RPCErrInvalidArgument
	}

	result, err := rs.actorHelper.CallRequestDefaultTimeout(PolarisSvc, entInfo)
	if err != nil {
		ret.Value = err.Error()
		return ret, types.RPCErrInternalError
	}
	if result != nil {
		ret.Value = result.(error).Error()
		return ret, types.RPCErrInvalidArgument
	}
	return ret, nil
}

func (rs *PRPCServer) RemoveBLEntry(ctx context.Context, msg *types.RmEntryParams) (*types.SingleString, error) {
	ret := &types.SingleString{}
	result, err := rs.actorHelper.CallRequestDefaultTimeout(PolarisSvc, msg)
	if err != nil {
		ret.Value = err.Error()
		return ret, types.RPCErrInternalError
	}
	removed := result.(bool)
	if !removed {
		ret.Value = "index out of range"
		return ret, types.RPCErrInvalidArgument
	}
	return ret, nil

}
