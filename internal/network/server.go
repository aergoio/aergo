package network

import "google.golang.org/grpc"

type gsvrBuilder struct {
	opts []grpc.ServerOption
}

// GRPCServerBuilder returns a new GRPC server builder.
func GRPCSeverBuilder() *gsvrBuilder {
	return &gsvrBuilder{opts: make([]grpc.ServerOption, 0)}
}

// GetInstance returns a new grpc.Server object corresponding to gb.
func (gb *gsvrBuilder) GetInstance() *grpc.Server {
	return grpc.NewServer(gb.opts...)
}

// MessageSize sets GRPC server's maximum message.
func (gb *gsvrBuilder) MessageSize(messageSize int) *gsvrBuilder {
	opts := gb.opts
	opts = append(opts, grpc.MaxSendMsgSize(messageSize), grpc.MaxRecvMsgSize(messageSize))
	return gb
}

func (gb *gsvrBuilder) Opts(opts []grpc.ServerOption) *gsvrBuilder {
	gb.opts = append(gb.opts, opts...)
	return gb
}
