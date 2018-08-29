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
)

var (
	// Used for flags.
	home    string
	cfgFile string
	keyFile string
	host    string
	port    int32

	rootConfig CliConfig

	rootCmd = &cobra.Command{
		Use:   "aergocli",
		Short: "Argo light commandline interface",
		Long:  `Argo is right`,
	}
)

func init() {
	log.SetOutput(os.Stderr)
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&home, "home", "", "aergo home path")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $AG_HOME/.aergo/config.toml)")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "", "private key file (readonly mode if missing)")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "Host address to aergo server")
	rootCmd.PersistentFlags().Int32VarP(&port, "port", "p", 7845, "Port number to aergo server")
}

func initConfig() {
	cliCtx := NewCliContext(home, cfgFile)
	//cliCtx.Vc.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key"))
	//cliCtx.Vc.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	//cliCtx.Vc.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	cliCtx.BindPFlags(rootCmd.PersistentFlags())

	rootConfig = cliCtx.GetDefaultConfig().(CliConfig)
	err := cliCtx.LoadOrCreateConfig(&rootConfig)
	if err != nil {
		log.Fatalf("Fail to load configuration file %v: %v", cliCtx.Vc.ConfigFileUsed(), err)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// GetServerAddress return ip address and port of server
func GetServerAddress() string {
	return fmt.Sprintf("%s:%d", rootConfig.Host, rootConfig.Port)
}
