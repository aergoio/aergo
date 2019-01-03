/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */
package main

import (
	"fmt"
	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/p2p/pmap"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var (
	rootCmd = &cobra.Command{
		Use:   "polaris",
		Short: "Polaris node discovery service",
		Long:  "Polaris node discovery service for providing peer addresses to connect.",
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
	cfg.Consensus.EnableBp = false
}



func rootRun(cmd *cobra.Command, args []string) {

	svrlog = log.NewLogger("polaris")
	svrlog.Info().Msg("POLARIS STARTED")

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

	lntc := pmap.NewNTContainer(cfg)
	pmapSvc := pmap.NewPolarisService(cfg, lntc)
	rpcSvc := pmap.NewPolarisRPC(cfg)

	// Register services to Hub. Don't need to do nil-check since Register
	// function skips nil parameters.
	compMng.Register(lntc, pmapSvc, rpcSvc)

	//consensusSvc, err := impl.New(cfg.Consensus, compMng, chainSvc)
	//if err != nil {
	//	svrlog.Error().Err(err).Msg("Failed to start consensus service.")
	//	os.Exit(1)
	//}

	// All the services objects including Consensus must be created before the
	// actors are started.
	compMng.Start()


	common.HandleKillSig(func() {
		//consensus.Stop(consensusSvc)
		compMng.Stop()
	}, svrlog)

	// wait... TODO need to break out when system finished.
	for {
		time.Sleep(time.Minute)
	}
}

type RedirectService struct {
	*component.BaseComponent
}

func NewRedirectService(cfg *config.Config, svcPid string) *RedirectService {
	logger :=  log.NewLogger(svcPid)
	rs := &RedirectService{}
	rs.BaseComponent = component.NewBaseComponent(svcPid, rs, logger)

	return rs
}

func (rs *RedirectService) Receive(context actor.Context) {
	// ignore for now
}

func (rs *RedirectService)  BeforeStart() {}
func (rs *RedirectService) AfterStart() {}
func (rs *RedirectService) BeforeStop() {}


func (rs *RedirectService) Statistics() *map[string]interface{} {
	dummy := make(map[string]interface{})
	return &dummy
}
