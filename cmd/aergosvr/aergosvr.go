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

	"github.com/aergoio/aergo/account"
	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/factory"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/mempool"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
	rest "github.com/aergoio/aergo/rest"
	"github.com/aergoio/aergo/rpc"
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
	svrlog         *log.Logger

	cfg *config.Config
)

func init() {
	cobra.OnInitialize(initConfig)
	fs := rootCmd.PersistentFlags()
	fs.StringVar(&homePath, "home", "", "path of aergo home")
	fs.StringVar(&configFilePath, "config", "", "path of configuration file")
}

func initConfig() {
	serverCtx := config.NewServerContext(homePath, configFilePath)
	cfg = serverCtx.GetDefaultConfig().(*config.Config)
	err := serverCtx.LoadOrCreateConfig(cfg)
	if err != nil {
		fmt.Printf("Fail to load configuration file %v: %v", serverCtx.Vc.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}
}

func rootRun(cmd *cobra.Command, args []string) {
	moduleLogLevels, err := log.ParseLevels(cfg.BaseConfig.LogLevel)
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	log.SetModuleLevels(moduleLogLevels)

	svrlog = log.NewLogger(log.ASVR)
	svrlog.Info("AERGO SVR STARTED")

	if cfg.EnableProfile {
		svrlog.Info("Enable Profiling on localhost:", cfg.ProfilePort)
		go func() {
			svrlog.Info(http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.ProfilePort), nil))
		}()
	}

	compMng := component.NewComponentHub()
	chainsvc := blockchain.NewChainService(cfg)
	compMng.Register(chainsvc)
	mpoolsvc := mempool.NewMemPoolService(cfg)
	compMng.Register(mpoolsvc)
	accountsvc := account.NewAccountService(cfg)
	compMng.Register(accountsvc)
	rpcsvc := rpc.NewRPC(compMng, cfg)
	compMng.Register(rpcsvc)
	p2psvc := p2p.NewP2P(compMng, cfg, chainsvc)
	compMng.Register(p2psvc)

	if cfg.EnableRest {
		svrlog.Info("Start Rest server")
		restsvc := rest.NewRestService(cfg, chainsvc)
		compMng.Register(restsvc)
		//restsvc.Start()
	} else {
		svrlog.Info("Do not Start Rest server")
	}

	compMng.Start()
	common.HandleKillSig(compMng.Stop, svrlog)

	if cfg.Consensus.EnableBp {
		c, err := factory.New(cfg, compMng)
		if err == nil {
			// ???
			consensus.Start(c)
		} else {
			svrlog.Errorf("failed to start consensus service: %s. server shutdown", err.Error())
			os.Exit(1)
		}
		chainsvc.SendChainInfo(c)
	} else {
		chainsvc.SendChainInfo(nil)
	}
	// wait... TODO need to break out when system finished.
	for {
		time.Sleep(time.Minute)
	}
}
