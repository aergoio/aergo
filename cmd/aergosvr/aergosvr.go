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

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/account"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/consensus/impl"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/mempool"
	"github.com/aergoio/aergo/v2/p2p"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/pkg/component"
	polarisclient "github.com/aergoio/aergo/v2/polaris/client"
	"github.com/aergoio/aergo/v2/rpc"
	"github.com/aergoio/aergo/v2/syncer"
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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if homePath != "" && configFilePath != "" && verbose {
				fmt.Println("ignore home path if given config file has valid configuration")
			}
		},
		Run: rootRun,
	}
	homePath       string
	configFilePath string
	enableTestmode bool
	useTestnet     bool

	verbose bool

	svrlog *log.Logger

	cfg *config.Config
)

func init() {
	cobra.OnInitialize(initConfig)

	localFlags := rootCmd.Flags()
	localFlags.SortFlags = false
	localFlags.BoolVar(&useTestnet, "testnet", false, "use Aergo TestNet; this only affects if there's no genesis block")
	localFlags.BoolVar(&enableTestmode, "testmode", false, "enable unsafe test mode (skips certain validations); can NOT use with --testnet")

	fs := rootCmd.PersistentFlags()
	fs.StringVar(&homePath, "home", "", "path of aergo home")
	fs.StringVar(&configFilePath, "config", "", "path of configuration file")
	fs.BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
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
	if cfg.EnableTestmode && cfg.UseTestnet {
		fmt.Println("Turn off test mode for Aergo Public Chains")
		os.Exit(1)
	}
}

func rootRun(cmd *cobra.Command, args []string) {
	svrlog = log.NewLogger("asvr")
	svrlog.Info().Str("revision", gitRevision).Str("branch", gitBranch).Msg("AERGO SVR STARTED")

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

	p2pkey.InitNodeInfo(&cfg.BaseConfig, cfg.P2P, githash, svrlog)

	compMng := component.NewComponentHub()

	chainSvc := chain.NewChainService(cfg)

	mpoolSvc := mempool.NewMemPoolService(cfg, chainSvc)
	rpcSvc := rpc.NewRPC(cfg, chainSvc, githash)
	admSvc := rpc.NewAdminService(cfg.RPC, compMng)
	syncSvc := syncer.NewSyncer(cfg, chainSvc, nil)
	p2pSvc := p2p.NewP2P(cfg, chainSvc)
	pmapSvc := polarisclient.NewPolarisConnectSvc(cfg.P2P, p2pSvc)

	var accountSvc component.IComponent
	if cfg.Personal {
		accountSvc = account.NewAccountService(cfg, chainSvc.SDB())
	}

	// Register services to Hub. Don't need to do nil-check since Register
	// function skips nil parameters.
	var verifyOnly = cfg.Blockchain.VerifyOnly || cfg.Blockchain.VerifyBlock != 0
	if !verifyOnly {
		compMng.Register(chainSvc, mpoolSvc, rpcSvc, syncSvc, p2pSvc, accountSvc, pmapSvc)
	} else {
		compMng.Register(chainSvc, mpoolSvc, rpcSvc)
	}

	consensusSvc, err := impl.New(cfg, compMng, chainSvc, p2pSvc, rpcSvc)
	if err != nil {
		svrlog.Error().Err(err).Msg("Failed to start consensus service.")
		os.Exit(1)
	}

	dmp := NewDumper(cfg, compMng)

	// All the services objects including Consensus must be created before the
	// actors are started.
	compMng.Start()

	if cfg.Consensus.EnableBp {
		// Warning: The consensus service must start after all the other
		// services.
		consensus.Start(consensusSvc)
	}

	if cfg.EnableDump {
		dmp.Start()
	}

	if len(cfg.RPC.NetServicePath) > 0 {
		admSvc.Start()
	}

	var interrupt = common.HandleKillSig(func() {
		consensus.Stop(consensusSvc)
		compMng.Stop()
	}, svrlog)

	// Wait main routine to stop
	<-interrupt.C
}
