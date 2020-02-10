/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const aergosystem = "aergo.system"

var (
	// Used for test.
	test bool

	// Used for flags.
	home    string
	cfgFile string
	host    string
	port    int32

	crtFile   string
	cacrtFile string
	svrName   string
	keyFile   string
	certPeer  string
	privKey   string
	pw        string
	dataDir   string

	from   string
	to     string
	amount string
	unit   string
	name   string

	address    string
	stateroot  string
	proof      bool
	compressed bool

	staking bool

	remote           bool
	importFormat     string
	keystoreFilePath string
	exportAsWif      bool

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
	rootCmd.PersistentFlags().StringVar(&home, "home", "", "aergo cli home path")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is cliconfig.toml)")
	rootCmd.PersistentFlags().StringVar(&svrName, "tlsservername", "", "aergosvr name for TLS ")
	rootCmd.PersistentFlags().StringVar(&cacrtFile, "tlscacert", "", "aergosvr CA certification file for TLS ")
	rootCmd.PersistentFlags().StringVar(&crtFile, "tlscert", "", "client certification file for TLS ")
	rootCmd.PersistentFlags().StringVar(&keyFile, "tlskey", "", "client key file for TLS ")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "Host address to aergo server")
	rootCmd.PersistentFlags().Int32VarP(&port, "port", "p", 7845, "Port number to aergo server")
}

func initConfig() {
	cliCtx := NewCliContext(home, cfgFile)
	cliCtx.Vc.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	cliCtx.Vc.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	cliCtx.Vc.BindPFlag("tls.servername", rootCmd.PersistentFlags().Lookup("tlsservername"))
	cliCtx.Vc.BindPFlag("tls.cacert", rootCmd.PersistentFlags().Lookup("tlscacert"))
	cliCtx.Vc.BindPFlag("tls.clientcert", rootCmd.PersistentFlags().Lookup("tlscert"))
	cliCtx.Vc.BindPFlag("tls.clientkey", rootCmd.PersistentFlags().Lookup("tlskey"))

	cliCtx.BindPFlags(rootCmd.PersistentFlags())

	rootConfig = cliCtx.GetDefaultConfig().(CliConfig)
	err := cliCtx.LoadOrCreateConfig(&rootConfig)
	if err != nil {
		log.Fatalf("Fail to load configuration file %v: %v", cliCtx.Vc.ConfigFileUsed(), err)
	}
}

func Execute() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// GetServerAddress return ip address and port of server
func GetServerAddress() string {
	return fmt.Sprintf("%s:%d", rootConfig.Host, rootConfig.Port)
}

func connectAergo(cmd *cobra.Command, args []string) {
	if test {
		return
	}
	serverAddr := GetServerAddress()
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 256)),
	}
	if rootConfig.TLS.ClientCert != "" || rootConfig.TLS.ClientKey != "" {
		certificate, err := tls.LoadX509KeyPair(rootConfig.TLS.ClientCert, rootConfig.TLS.ClientKey)
		if err != nil {
			log.Fatal("wrong tls setting : ", err)
		}
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(rootConfig.TLS.CACert)
		if err != nil {
			log.Fatal("could not read server certification file : ", err)
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			log.Fatal("failed to append server certification to CA certs")
		}
		creds := credentials.NewTLS(&tls.Config{
			ServerName:   rootConfig.TLS.ServerName,
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	var ok bool
	client, ok = util.GetClient(serverAddr, opts).(*util.ConnClient)
	if !ok {
		log.Fatal("internal error. wrong RPC client type")
	}
}

func disconnectAergo(cmd *cobra.Command, args []string) {
	if test {
		return
	}

	if client != nil {
		client.Close()
	}
}
