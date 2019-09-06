/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const aergosystem = "aergo.system"

var (
	client *PolarisClient
)

var (
	// Used for test.
	test bool

	// Used for flags.
	home    string
	cfgFile string
	host    string
	port    int32

	privKey string
	pw      string

	rootConfig CliConfig

	rootCmd = &cobra.Command{
		Use:               "colaris",
		Short:             "colaris to Polaris",
		Long:              `light commandline interface`,
		PersistentPreRun:  connectPolaris,
		PersistentPostRun: disconnectPolaris,
	}
)

// flags for blacklist
var (
	addAddr string
	addCidr string
	addPid  string // base58 encoded string
	rmIdx   int
)

func init() {
	log.SetOutput(os.Stderr)
	cobra.OnInitialize(initConfig)
	rootCmd.SetOutput(os.Stdout)
	rootCmd.PersistentFlags().StringVar(&home, "home", "", "aergo cli home path")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is cliconfig.toml)")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "Host address to aergo server")
	rootCmd.PersistentFlags().Int32VarP(&port, "port", "p", 8915, "Port number to polaris server")
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

func connectPolaris(cmd *cobra.Command, args []string) {
	if test {
		return
	}

	serverAddr := GetServerAddress()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	var ok bool
	client, ok = GetClient(serverAddr, opts).(*PolarisClient)
	if !ok {
		log.Fatal("internal error. wrong RPC client type")
	}
}

func disconnectPolaris(cmd *cobra.Command, args []string) {
	if test {
		return
	}

	if client != nil {
		client.Close()
	}
}
