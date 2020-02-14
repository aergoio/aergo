package cmd

import "github.com/aergoio/aergo-lib/config"

const (
	EnvironmentPrefix             = "AG"
	defaultAergoHomePath          = ".aergo"
	defaultAergoCliConfigFileName = "cliconfig.toml"
)

type CliContext struct {
	config.BaseContext
}

func NewCliContext(homePath string, configFilePath string) *CliContext {
	cliCtx := &CliContext{}
	cliCtx.BaseContext = config.NewBaseContext(cliCtx, homePath, configFilePath, EnvironmentPrefix)

	return cliCtx
}

// CliConfig is configs for aergo cli.
type CliConfig struct {
	Host         string     `mapstructure:"host" description:"Target server host. default is localhost"`
	Port         int        `mapstructure:"port" description:"Target server port. default is 7845"`
	TLS          *TLSConfig `mapstructure:"tls"`
	KeyStorePath string     `mapstructure:"keystore" description:"Path to keystore directory"`
}

type TLSConfig struct {
	ServerName string `mapstructure:"servername" description:"Target server name for TLS"`
	CACert     string `mapstructure:"cacert" description:"CA(Certificate Authority) certification file path for TLS"`
	ClientCert string `mapstructure:"clientcert" description:"Client PEM certification file path for TLS"`
	ClientKey  string `mapstructure:"clientkey" description:"Client key file path for TLS"`
}

// GetDefaultConfig return cliconfig with default value. It ALWAYS returns NEW object.
func (ctx *CliContext) GetDefaultConfig() interface{} {
	return CliConfig{
		Host:         "localhost",
		Port:         7845,
		KeyStorePath: defaultAergoHomePath,
		TLS:          ctx.GetDefaultTLSConfig(),
	}
}

// GetDefaultTLSConfig return tlsconfig with default value. It ALWAYS returns NEW object.
func (ctx *CliContext) GetDefaultTLSConfig() *TLSConfig {
	return &TLSConfig{ServerName: "localhost"}
}

func (ctx *CliContext) GetHomePath() string {
	return defaultAergoHomePath
}

func (ctx *CliContext) GetConfigFileName() string {
	return defaultAergoCliConfigFileName
}

func (ctx *CliContext) GetTemplate() string {
	return configTemplate
}

const configTemplate = `# aergo cli TOML Configuration File (https://github.com/toml-lang/toml)
host = "{{.Host}}"
port = "{{.Port}}"
keystore = "{{.KeyStorePath}}"

[tls]
servername = "{{.TLS.ServerName}}"
cacert = "{{.TLS.CACert}}"
clientcert = "{{.TLS.ClientCert}}"
clientkey = "{{.TLS.ClientKey}}"
`
