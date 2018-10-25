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
	"time"

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
	rest "github.com/aergoio/aergo/rest"
	"github.com/aergoio/aergo/rpc"
	"github.com/aergoio/aergo/syncer"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
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
	svrlog         *log.Logger

	cfg *config.Config
)

func init() {
	cobra.OnInitialize(initConfig)
	fs := rootCmd.PersistentFlags()
	fs.StringVar(&homePath, "home", "", "path of aergo home")
	fs.StringVar(&configFilePath, "config", "", "path of configuration file")
	fs.BoolVar(&enableTestmode, "testmode", false, "enable unsafe test mode (skips certain validations)")
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
}

func rootRun(cmd *cobra.Command, args []string) {

	svrlog = log.NewLogger("asvr")
	svrlog.Info().Msg("AERGO SVR STARTED")

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

	p2p.InitNodeInfo(cfg.P2P, svrlog)

	compMng := component.NewComponentHub()

	consensusSvc, err := impl.New(cfg, compMng)
	if err != nil {
		svrlog.Error().Err(err).Msg("failed to start consensus service. server shutdown")
		os.Exit(1)
	}

	mpoolSvc := mempool.NewMemPoolService(cfg)
	compMng.Register(mpoolSvc)
	chainSvc := chain.NewChainService(cfg, consensusSvc, mpoolSvc)
	compMng.Register(chainSvc)
	consensusSvc.SetChainAccessor(chainSvc)
	accountsvc := account.NewAccountService(cfg)
	compMng.Register(accountsvc)
	rpcSvc := rpc.NewRPC(compMng, cfg, chainSvc)
	compMng.Register(rpcSvc)
	syncSvc := syncer.NewSyncer(cfg, chainSvc)
	compMng.Register(syncSvc)
	p2pSvc := p2p.NewP2P(compMng, cfg, chainSvc)
	compMng.Register(p2pSvc)

	if cfg.EnableRest {
		svrlog.Info().Msg("Start Rest server")
		restsvc := rest.NewRestService(cfg, chainSvc)
		compMng.Register(restsvc)
		//restsvc.Start()
	} else {
		svrlog.Info().Msg("Do not Start Rest server")
	}

	compMng.Start()

	if cfg.Consensus.EnableBp {
		// Warning!!!: The consensus service must start after all the other
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
