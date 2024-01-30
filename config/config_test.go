/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package config

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetParamConfPath(t *testing.T) {
	// create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	// generate a random conf file path and set to the conf
	generatedConfFilePath := path.Join(tmpDir, "aergo.toml")

	// create a default config
	serverCxt := NewServerContext("", generatedConfFilePath)
	defaultConf := serverCxt.GetDefaultConfig()
	var loadedConf Config

	// create a config
	err = serverCxt.LoadOrCreateConfig(defaultConf)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	// load a saved config file
	if err := serverCxt.LoadOrCreateConfig(&loadedConf); err != nil {
		assert.Fail(t, err.Error())
	}

	// compare each other
	assert.Equal(t, defaultConf.(*Config), &loadedConf)
}
