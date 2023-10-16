/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"github.com/aergoio/aergo-lib/config"
	"github.com/aergoio/aergo/v2/polaris/common"
)

const (
	EnvironmentPrefix            = "AG"
	defaultColarisHomePath       = ".polaris"
	defaultColarisConfigFileName = "colaris.toml"
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
	Port int    `mapstructure:"port" description:"Target server port. default is 8915"`
}

// GetDefaultConfig return cliconfig with default value. It ALWAYS returns NEW object.
func (ctx *CliContext) GetDefaultConfig() interface{} {
	return CliConfig{
		Host: "localhost",
		Port: common.DefaultRPCPort,
	}
}

func (ctx *CliContext) GetHomePath() string {
	return defaultColarisHomePath
}

func (ctx *CliContext) GetConfigFileName() string {
	return defaultColarisConfigFileName
}

func (ctx *CliContext) GetTemplate() string {
	return configTemplate
}

const configTemplate = `# aergo cli TOML Configuration File (https://github.com/toml-lang/toml)
host = "{{.Host}}"
port = "{{.Port}}"
`
