package rpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

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

const requestTimeout = 10 * time.Second

// MempoolTxStat returns the TX-relasted statistics of the current mempool.
func (as *AdminService) MempoolTxStat(ctx context.Context, in *types.Empty) (*types.SingleBytes, error) {
	r, err := as.RequestFuture(message.MemPoolSvc, &message.MemPoolTxStat{}, requestTimeout, "rpc/MempoolTxStat").Result()
	return &types.SingleBytes{Value: r.(*message.MemPoolTxStatRsp).Data}, err
}

// MempoolTx returns the TX-relasted statistics of the current mempool.
func (as *AdminService) MempoolTx(ctx context.Context, in *types.AccountList) (*types.SingleBytes, error) {
	m := &message.MemPoolTx{Accounts: make([]types.Address, len(in.Accounts))}
	for i, acc := range in.Accounts {
		m.Accounts[i] = acc.Address
	}

	var data []byte
	r, err := as.RequestFuture(message.MemPoolSvc, m, requestTimeout, "rpc/MempoolTxStat").Result()
	if r != nil {
		data = r.(*message.MemPoolTxRsp).Data
	}
	return &types.SingleBytes{Value: data}, err
}
