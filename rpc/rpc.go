/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package rpc

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/contract/enterprise"
	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	aergorpc "github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/opentracing/opentracing-go"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// RPC is actor for providing rpc service
type RPC struct {
	conf *config.Config

	*component.BaseComponent

	grpcServer    *grpc.Server
	grpcWebServer *grpcweb.WrappedGrpcServer
	actualServer  *AergoRPCService
	httpServer    *http.Server

	ca      types.ChainAccessor
	version string
	entConf *types.EnterpriseConfig
}

//var _ component.IComponent = (*RPCComponent)(nil)

// NewRPC create an rpc service
func NewRPC(cfg *config.Config, chainAccessor types.ChainAccessor, version string) *RPC {
	actualServer := &AergoRPCService{
		msgHelper:           message.GetHelper(),
		blockStream:         make(map[uint32]*ListBlockStream),
		blockMetadataStream: make(map[uint32]*ListBlockMetaStream),
		eventStream:         make(map[*EventStream]*EventStream),
	}

	tracer := opentracing.GlobalTracer()

	opts := make([]grpc.ServerOption, 0)

	if cfg.RPC.NetServiceTrace {
		opts = append(opts, grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)))
		opts = append(opts, grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)))
	}

	var entConf *types.EnterpriseConfig
	genesis := chainAccessor.GetGenesisInfo()
	if !genesis.ID.PublicNet {
		conf, err := chainAccessor.GetEnterpriseConfig("rpcpermissions")
		if err != nil {
			logger.Error().Err(err).Msg("could not get allowed client information")
		} else {
			entConf = conf
		}
	}

	if cfg.RPC.NSEnableTLS {
		certificate, err := tls.LoadX509KeyPair(cfg.RPC.NSCert, cfg.RPC.NSKey)
		if err != nil {
			logger.Error().Err(err).Msg("could not load server key pair")
		}
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(cfg.RPC.NSCACert)
		if err != nil {
			logger.Error().Err(err).Msg("could not read CA cert")
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			logger.Error().Bool("AppendCertsFromPEM", ok).Msg("failed to append server cert")
			err = fmt.Errorf("failed to append server cert")
		}
		if err == nil {
			creds := credentials.NewTLS(&tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				Certificates: []tls.Certificate{certificate},
				ClientCAs:    certPool,
			})
			opts = append(opts, grpc.Creds(creds))
			logger.Info().Str("cert", cfg.RPC.NSCert).Str("key", cfg.RPC.NSKey).Msg("grpc with TLS")
		}
	}

	grpcServer := network.GRPCSeverBuilder().
		MessageSize(int(types.GetMaxMessageSize(cfg.Blockchain.MaxBlockSize))).
		Opts(opts).
		GetInstance()

	grpcWebServer := grpcweb.WrapServer(
		grpcServer,
		grpcweb.WithWebsockets(true),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
			return true
		}))

	rpcsvc := &RPC{
		conf: cfg,

		grpcServer:    grpcServer,
		grpcWebServer: grpcWebServer,
		actualServer:  actualServer,
		ca:            chainAccessor,
		version:       version,
	}
	rpcsvc.BaseComponent = component.NewBaseComponent(message.RPCSvc, rpcsvc, logger)

	actualServer.actorHelper = rpcsvc
	actualServer.setClientAuth(entConf)

	rpcsvc.httpServer = &http.Server{
		Handler:        rpcsvc.grpcWebHandlerFunc(grpcWebServer, http.DefaultServeMux),
		ReadTimeout:    4 * time.Second,
		WriteTimeout:   4 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return rpcsvc
}

func (ns *RPC) SetHub(hub *component.ComponentHub) {
	ns.actualServer.hub = hub
	ns.BaseComponent.SetHub(hub)
}

func (ns *RPC) SetConsensusAccessor(ca consensus.ConsensusAccessor) {
	ns.actualServer.SetConsensusAccessor(ca)
}

// Start start rpc service.
func (ns *RPC) BeforeStart() {
	aergorpc.RegisterAergoRPCServiceServer(ns.grpcServer, ns.actualServer)
}

func (ns *RPC) AfterStart() {
	go ns.serve()
}

// Stop stops rpc service.
func (ns *RPC) BeforeStop() {
	ns.httpServer.Close()
	ns.grpcServer.Stop()
}

func (ns *RPC) Statistics() *map[string]interface{} {
	return &map[string]interface{}{
		"config":  ns.conf.RPC,
		"version": ns.version,
		"streams": ns.actualServer.Statistics(),
	}
}

func (ns *RPC) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *types.Block:
		server := ns.actualServer
		server.BroadcastToListBlockStream(msg)
		meta := msg.GetMetadata()
		server.BroadcastToListBlockMetadataStream(meta)
	case []*types.Event:
		server := ns.actualServer
		for _, e := range msg {
			if bytes.Equal(e.GetContractAddress(), types.AddressPadding([]byte(types.AergoEnterprise))) {
				eventName := strings.Split(e.GetEventName(), " ")
				conf := strings.ToUpper(eventName[1])
				switch eventName[0] {
				case "Enable":
					if conf == enterprise.RPCPermissions {
						value := false
						if e.JsonArgs == "true" {
							value = true
						}
						server.setClientAuthOn(value)
					} else if conf == enterprise.AccountWhite {
						value := false
						if e.JsonArgs == "true" {
							value = true
						}
						msg := &message.MemPoolEnableWhitelist{On: value}
						ns.TellTo(message.MemPoolSvc, msg)
					} else if conf == enterprise.P2PBlack || conf == enterprise.P2PWhite {
						value := false
						if e.JsonArgs == "true" {
							value = true
						}
						msg := message.P2PWhiteListConfEnableEvent{Name: conf, On: value}
						ns.TellTo(message.P2PSvc, msg)
					}
				case "Set":
					if conf == enterprise.RPCPermissions {
						values := make([]string, 1024)
						if err := json.Unmarshal([]byte(e.JsonArgs), &values); err != nil {
							return
						}
						server.setClientAuthMap(values)
					} else if conf == enterprise.AccountWhite {
						values := make([]string, 1024)
						if err := json.Unmarshal([]byte(e.JsonArgs), &values); err != nil {
							return
						}
						msg := &message.MemPoolSetWhitelist{
							Accounts: values,
						}
						ns.TellTo(message.MemPoolSvc, msg)
					} else if conf == enterprise.P2PBlack || conf == enterprise.P2PWhite {
						values := make([]string, 1024)
						if err := json.Unmarshal([]byte(e.JsonArgs), &values); err != nil {
							return
						}
						msg := message.P2PWhiteListConfSetEvent{
							Values: values,
						}
						ns.TellTo(message.P2PSvc, msg)
					}
				default:
					logger.Warn().Str("Enterprise event", eventName[0]).Str("conf", conf).Msg("unknown message in RPCPERMISSION")
				}
			}
		}
		server.BroadcastToEventStream(msg)
	case *message.GetServerInfo:
		context.Respond(ns.CollectServerInfo(msg.Categories))
	case *actor.Started, *actor.Stopping, *actor.Stopped, *component.CompStatReq: // donothing
		// Ignore actor lfiecycle messages
	default:
		ns.Warn().Msgf("unknown msg received in rpc %s", reflect.TypeOf(msg).String())
	}
}

// Create HTTP handler that redirects matching grpc-web requests to the grpc-web wrapper.
func (ns *RPC) grpcWebHandlerFunc(grpcWebServer *grpcweb.WrappedGrpcServer, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if grpcWebServer.IsAcceptableGrpcCorsRequest(r) || grpcWebServer.IsGrpcWebRequest(r) || grpcWebServer.IsGrpcWebSocketRequest(r) {
			grpcWebServer.ServeHTTP(w, r)
		} else {
			ns.Info().Msg("Request handled by other hanlder. is this correct?")
			otherHandler.ServeHTTP(w, r)
		}
	})
}

// Serve GRPC server over TCP
func (ns *RPC) serveGRPC(l net.Listener, server *grpc.Server) {
	if err := server.Serve(l); err != nil {
		switch err {
		case cmux.ErrListenerClosed:
			// Server killed, usually by ctrl-c signal
		default:
			panic(err)
		}
	}
}

// Serve HTTP server over TCP
func (ns *RPC) serveHTTP(l net.Listener, server *http.Server) {
	if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

// Serve TCP multiplexer
func (ns *RPC) serve() {
	ipAddr := net.ParseIP(ns.conf.RPC.NetServiceAddr)
	if ipAddr == nil {
		panic("Wrong IP address format in RPC.NetServiceAddr")
	}

	addr := fmt.Sprintf("%s:%d", ipAddr, ns.conf.RPC.NetServicePort)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	// Setup TCP multiplexer
	tcpm := cmux.New(l)

	//grpcL := tcpm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	matchers := []cmux.Matcher{cmux.HTTP2()}
	if ns.conf.RPC.NSEnableTLS {
		matchers = append(matchers, cmux.TLS())
	} else {
		//http1 only work without TLS
		httpL := tcpm.Match(cmux.HTTP1Fast())
		go ns.serveHTTP(httpL, ns.httpServer)
	}
	grpcL := tcpm.Match(matchers...)
	go ns.serveGRPC(grpcL, ns.grpcServer)

	ns.Info().Msg(fmt.Sprintf("Starting RPC server listening on %s, with TLS: %v", addr, ns.conf.RPC.NSEnableTLS))

	// Serve TCP multiplexer
	if err := tcpm.Serve(); !strings.Contains(err.Error(), "use of closed network connection") {
		ns.Fatal().Msg(fmt.Sprintf("%v", err))
	}

	return
}

func (ns *RPC) CollectServerInfo(categories []string) *types.ServerInfo {
	// 3 items are needed
	statusInfo := make(map[string]string)
	rsp, err := ns.CallRequestDefaultTimeout(message.P2PSvc, &message.GetSelf{})
	statusInfo["version"] = ns.version
	if err != nil {
		ns.Logger.Error().Err(err).Msg("p2p actor error")
		statusInfo["id"] = p2pkey.NodeSID()
	} else {
		meta := rsp.(p2pcommon.PeerMeta)
		statusInfo["id"] = meta.ID.String()
		statusInfo["addr"] = meta.PrimaryAddress()
		statusInfo["port"] = strconv.Itoa(int(meta.PrimaryPort()))
	}
	configInfo := make(map[string]*types.ConfigItem)
	types.AddCategory(configInfo, "base").AddBool("personal", ns.conf.BaseConfig.Personal)
	types.AddCategory(configInfo, "account").AddInt("unlocktimeout", int(ns.conf.Account.UnlockTimeout))
	return &types.ServerInfo{Status: statusInfo, Config: configInfo}
}

const defaultTTL = time.Second * 4

// TellRequest implement interface method of ActorService
func (ns *RPC) TellRequest(actor string, msg interface{}) {
	ns.TellTo(actor, msg)
}

// SendRequest implement interface method of ActorService
func (ns *RPC) SendRequest(actor string, msg interface{}) {
	ns.RequestTo(actor, msg)
}

// FutureRequest implement interface method of ActorService
func (ns *RPC) FutureRequest(actor string, msg interface{}, timeout time.Duration) *actor.Future {
	return ns.RequestToFuture(actor, msg, timeout)
}

// FutureRequestDefaultTimeout implement interface method of ActorService
func (ns *RPC) FutureRequestDefaultTimeout(actor string, msg interface{}) *actor.Future {
	return ns.RequestToFuture(actor, msg, defaultTTL)
}

// CallRequest implement interface method of ActorService
func (ns *RPC) CallRequest(actor string, msg interface{}, timeout time.Duration) (interface{}, error) {
	future := ns.RequestToFuture(actor, msg, timeout)
	return future.Result()
}

// CallRequest implement interface method of ActorService
func (ns *RPC) CallRequestDefaultTimeout(actor string, msg interface{}) (interface{}, error) {
	future := ns.RequestToFuture(actor, msg, defaultTTL)
	return future.Result()
}

// GetChainAccessor implment interface method of ActorService
func (ns *RPC) GetChainAccessor() types.ChainAccessor {
	return ns.ca
}

func convertError(err error) types.CommitStatus {
	switch err {
	case nil:
		return types.CommitStatus_TX_OK
	case types.ErrTxNonceTooLow:
		return types.CommitStatus_TX_NONCE_TOO_LOW
	case types.ErrTxAlreadyInMempool:
		return types.CommitStatus_TX_ALREADY_EXISTS
	case types.ErrTxHasInvalidHash:
		return types.CommitStatus_TX_INVALID_HASH
	case types.ErrTxFormatInvalid:
		return types.CommitStatus_TX_INVALID_FORMAT
	case types.ErrInsufficientBalance:
		return types.CommitStatus_TX_INSUFFICIENT_BALANCE
	case types.ErrSameNonceAlreadyInMempool:
		return types.CommitStatus_TX_HAS_SAME_NONCE
	default:
		//logger.Info().Str("hash", err.Error()).Msg("RPC encountered unconvertable error")
		return types.CommitStatus_TX_INTERNAL_ERROR
	}
}
