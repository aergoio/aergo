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
	Host string `mapstructure:"host" description:"Target server host. default is localhost"`
	Port int    `mapstructure:"port" description:"Target server port. default is 7845"`
}

// GetDefaultConfig return cliconfig with default value. It ALWAYS returns NEW object.
func (ctx *CliContext) GetDefaultConfig() interface{} {
	return CliConfig{
		Host: "localhost",
		Port: 7845,
	}
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
`
