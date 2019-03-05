/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aergoio/aergo/p2p/pmap"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/account"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/impl"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/mempool"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/rpc"
	"github.com/aergoio/aergo/syncer"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/spf13/cobra"
)

var (
	gitRevision, gitBranch string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var (
	rootCmd = &cobra.Command{
		Use:   "aergosvr",
		Short: "Aergo Server",
		Long:  "Aergo Server Full-node implementation",
		Run:   rootRun,
	}
	homePath       string
	configFilePath string
	enableTestmode bool
	useTestnet     bool
	svrlog         *log.Logger

	cfg *config.Config
)

func init() {
	cobra.OnInitialize(initConfig)
	fs := rootCmd.PersistentFlags()
	fs.StringVar(&homePath, "home", "", "path of aergo home")
	fs.StringVar(&configFilePath, "config", "", "path of configuration file")
	fs.BoolVar(&enableTestmode, "testmode", false, "enable unsafe test mode (skips certain validations)")
	fs.BoolVar(&useTestnet, "testnet", false, "use Aergo TestNet")
}

func initConfig() {
	serverCtx := config.NewServerContext(homePath, configFilePath)
	cfg = serverCtx.GetDefaultConfig().(*config.Config)
	err := serverCtx.LoadOrCreateConfig(cfg)
	if err != nil {
		fmt.Printf("Fail to load configuration file %v: %v", serverCtx.Vc.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}
	if enableTestmode {
		cfg.EnableTestmode = true
	}
	if useTestnet {
		cfg.UseTestnet = true
	}
}

func configureZipkin() {
	protocol := cfg.Monitor.ServerProtocol
	endpoint := cfg.Monitor.ServerEndpoint
	var collector zipkin.Collector
	var err error
	if "http" == protocol || "https" == protocol {
		zipkinURL := fmt.Sprintf("%s://%s/api/v1/spans", protocol, endpoint)
		collector, err = zipkin.NewHTTPCollector(zipkinURL)
		if err != nil {
			panic("Error connecting to zipkin server at " + zipkinURL + ". Error: " + err.Error())
		}
	} else if "kafka" == protocol {
		endpoints := strings.Split(endpoint, ",")
		collector, err = zipkin.NewKafkaCollector(endpoints)
		if err != nil {
			panic("Error connecting to kafka endpoints at " + endpoint + ". Error: " + err.Error())
		}
	}

	if nil != collector {
		myEndpoint := cfg.RPC.NetServiceAddr + ":" + strconv.Itoa(cfg.RPC.NetServicePort)
		tracer, err := zipkin.NewTracer(zipkin.NewRecorder(collector, false, myEndpoint, "aergosvr"))
		if err != nil {
			panic("Error starting new zipkin tracer. Error: " + err.Error())
		}
		opentracing.InitGlobalTracer(tracer)
	}
}

func rootRun(cmd *cobra.Command, args []string) {

	svrlog = log.NewLogger("asvr")
	svrlog.Info().Str("revision", gitRevision).Str("branch", gitBranch).Msg("AERGO SVR STARTED")

	configureZipkin()

	if cfg.EnableProfile {
		svrlog.Info().Msgf("Enable Profiling on localhost: %d", cfg.ProfilePort)
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.ProfilePort), nil)
			svrlog.Info().Err(err).Msg("Run Profile Server")
		}()
	}

	if cfg.EnableTestmode {
		svrlog.Warn().Msgf("Running with unsafe test mode. Turn off test mode for production use!")
	}

	p2p.InitNodeInfo(&cfg.BaseConfig, cfg.P2P, svrlog)

	compMng := component.NewComponentHub()

	chainSvc := chain.NewChainService(cfg)

	mpoolSvc := mempool.NewMemPoolService(cfg, chainSvc)
	rpcSvc := rpc.NewRPC(cfg, chainSvc)
	syncSvc := syncer.NewSyncer(cfg, chainSvc, nil)
	p2pSvc := p2p.NewP2P(cfg, chainSvc)
	pmapSvc := pmap.NewPolarisConnectSvc(cfg.P2P, p2pSvc)

	var accountSvc component.IComponent
	if cfg.Personal {
		accountSvc = account.NewAccountService(cfg, chainSvc.SDB())
	}

	// Register services to Hub. Don't need to do nil-check since Register
	// function skips nil parameters.
	compMng.Register(chainSvc, mpoolSvc, rpcSvc, syncSvc, p2pSvc, accountSvc, pmapSvc)

	consensusSvc, err := impl.New(cfg, compMng, chainSvc)
	if err != nil {
		svrlog.Error().Err(err).Msg("Failed to start consensus service.")
		os.Exit(1)
	}

	// All the services objects including Consensus must be created before the
	// actors are started.
	compMng.Start()

	if cfg.Consensus.EnableBp {
		// Warning: The consensus service must start after all the other
		// services.
		consensus.Start(consensusSvc)
	}

	common.HandleKillSig(func() {
		consensus.Stop(consensusSvc)
		compMng.Stop()
	}, svrlog)

	// wait... TODO need to break out when system finished.
	for {
		time.Sleep(time.Minute)
	}
}
