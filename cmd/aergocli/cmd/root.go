/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const aergosystem = "aergo.system"

var (
	// Used for flags.
	home    string
	cfgFile string
	host    string
	port    int32

	privKey string
	pw      string
	dataDir string

	from   string
	to     string
	amount uint64

	address    string
	stateroot  string
	proof      bool
	compressed bool

	staking bool

	remote       bool
	importFormat string

	rootConfig CliConfig

	rootCmd = &cobra.Command{
		Use:               "aergocli",
		Short:             "Aergo light commandline interface",
		Long:              `Aergo is right`,
		PersistentPreRun:  connectAergo,
		PersistentPostRun: disconnectAergo,
	}
)

func init() {
	log.SetOutput(os.Stderr)
	cobra.OnInitialize(initConfig)
	rootCmd.SetOutput(os.Stdout)
	rootCmd.PersistentFlags().StringVar(&home, "home", ".", "aergo home path")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.toml", "config file (default is config.toml)")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "Host address to aergo server")
	rootCmd.PersistentFlags().Int32VarP(&port, "port", "p", 7845, "Port number to aergo server")
}

func initConfig() {
	cliCtx := NewCliContext(home, cfgFile)
	cliCtx.Vc.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	cliCtx.Vc.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	cliCtx.BindPFlags(rootCmd.PersistentFlags())

	rootConfig = cliCtx.GetDefaultConfig().(CliConfig)
	err := cliCtx.LoadOrCreateConfig(&rootConfig)
	if err != nil {
		log.Fatalf("Fail to load configuration file %v: %v", cliCtx.Vc.ConfigFileUsed(), err)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// GetServerAddress return ip address and port of server
func GetServerAddress() string {
	return fmt.Sprintf("%s:%d", rootConfig.Host, rootConfig.Port)
}

func connectAergo(cmd *cobra.Command, args []string) {
	serverAddr := GetServerAddress()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var ok bool
	client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient)
	if !ok {
		log.Fatal("internal error. wrong RPC client type")
	}
}

func disconnectAergo(cmd *cobra.Command, args []string) {
	if client != nil {
		client.Close()
	}
}
