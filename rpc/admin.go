package rpc

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/aergoio/aergo/types"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

const adminSvc = "AdminRPCService"

type AdminService struct {
	*component.ComponentHub
	*log.Logger
	run func()
}

func NewAdminService(conf *config.RPCConfig, hub *component.ComponentHub) *AdminService {
	as := &AdminService{
		ComponentHub: hub,
		Logger:       log.NewLogger("admin"),
	}
	as.run = func() {
		socketFile := conf.NetServicePath
		// Remove the previous socket file.
		os.Remove(socketFile)
		l, err := net.Listen("unix", socketFile)
		if err != nil {
			panic(err)
		}
		grpcServer := newGRPCServer(conf.NetServiceTrace)
		types.RegisterAdminRPCServiceServer(grpcServer, as)

		grpcServer.Serve(l)
		as.Info().Msg(fmt.Sprintf("Starting Admin RPC server listening on %s", l.Addr()))
	}

	return as
}

func newGRPCServer(trace bool) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 8),
	}
	if trace {
		tracer := opentracing.GlobalTracer()
		opts = append(opts, grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)))
		opts = append(opts, grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)))
	}
	return grpc.NewServer(opts...)
}

func (as *AdminService) Start() {
	go as.run()
}

// MempoolTxStat returns the TX-relasted statistics of the current mempool.
func (as *AdminService) MempoolTxStat(ctx context.Context, in *types.Empty) (*types.SingleBytes, error) {
	return &types.SingleBytes{Value: []byte("mempool tx stat")}, nil
}

// MempoolTx returns the TX-relasted statistics of the current mempool.
func (as *AdminService) MempoolTx(ctx context.Context, in *types.AccountList) (*types.SingleBytes, error) {
	return &types.SingleBytes{Value: []byte("mempool tx")}, nil
}
