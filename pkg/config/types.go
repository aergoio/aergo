/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package config

import (
	"github.com/spf13/viper"
)

const (
	homeDirPermission  = 0755 // TODO Validate
	confFilePermission = 0644 // TODO Validate
)

var _ IContext = (*BaseContext)(nil)

// BaseContext has IContext interface to get default information to
// find or generate a configuration file, and viper instance to
// store and perform actions
type BaseContext struct {
	IContext
	Vc *viper.Viper
}

// IContext provides base information to search or create default
// configuration file
type IContext interface {
	GetDefaultConfig() interface{}
	GetHomePath() string
	GetConfigFileName() string
	GetTemplate() string
}

/*
 How to write a toml conf file template
 =======================================

 string_type = "{{.STRUCT.FILED}}"
 bool/number_type = {{.STRUCT.FILED}}
 string_array_type = [{{range .STRUCT.FILED}}
 "{{.}}", {{end}}
 ]
 bool/number_array_type = [{{range .STRUCT.FILED}}
 {{.}}, {{end}}
 ]
 map = does not support
*/
