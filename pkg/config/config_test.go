/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package config

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ========================================================
//  Dummy Config Struct for Testing
// ========================================================

const (
	defaultHomePath       = ".aergo"
	defaultConfigFileName = "config.toml"
)

type DummyContext struct {
	BaseContext
}

type DummyConf struct {
	DummyRootConf `mapstructure:",squash"`
	Sub           *DummySubConf `mapstruct:"sub"`
}
type DummyRootConf struct {
	One   int    `mapstruct:"one"`
	Unset int    `mapstruct:"unset"`
	Str   string `mapstruct:"str"`
	Bool  bool   `mapstruct:"bool"`
}

type DummySubConf struct {
	NumArray []int    `mapstruct:"numarray"`
	StrArray []string `mapstruct:"strarray"`
}

const dummyTemplate = `#dummy configration
one = {{.DummyRootConf.One}}
unset = {{.DummyRootConf.Unset}}
str = "{{.DummyRootConf.Str}}"
bool = {{.DummyRootConf.Bool}}

[sub]
numarray = [{{range .Sub.NumArray}}
{{.}}, {{end}}
]
strarray = [{{range .Sub.StrArray}}
"{{.}}", {{end}}
]
`

func NewDummyContext(homePath string, configFilePath string) *DummyContext {
	dummyCxt := &DummyContext{}
	dummyCxt.BaseContext = NewBaseContext(dummyCxt, homePath, configFilePath, "AG")

	return dummyCxt
}

func (ctx *DummyContext) GetHomePath() string {
	return defaultHomePath
}

func (ctx *DummyContext) GetConfigFileName() string {
	return defaultConfigFileName
}

func (ctx *DummyContext) GetDefaultConfig() interface{} {
	return &DummyConf{
		DummyRootConf: DummyRootConf{
			One:   1,
			Unset: 0,
			Str:   "text",
			Bool:  true,
		},
		Sub: &DummySubConf{
			NumArray: []int{1, 2},
			StrArray: []string{"a", "b"},
		},
	}
}

func (ctx *DummyContext) GetTemplate() string {
	return dummyTemplate
}

// ========================================================

/*
This is a conversion test between config, text, viper.
A test process is like this;
	1. load a default config instance
	2. convert 1 to a text
	3. read 2 to viper
	4. unmarshal 3 to an empty config instance
	5. 1 and 4 must be same each other
*/
func TestConvertConfig(t *testing.T) {
	dummyCxt := NewDummyContext("", "")
	// get a default configuration
	defaultConf := dummyCxt.GetDefaultConfig()

	// generate a text for a config file using a template
	generatedText, err := dummyCxt.fillTemplate(defaultConf)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	t.Log(generatedText)

	// load generatedText to viper
	var loadedConf DummyConf
	reader := strings.NewReader(generatedText) // convert string to io.Reader
	viperConf := viper.New()
	viperConf.SetConfigType("toml")
	err = viperConf.ReadConfig(reader)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	// unmarshal viper and fill a config instance
	viperConf.Unmarshal(&loadedConf)

	// compare original default config and loaded config from viper
	//assert.Equal(t, defaultConf.(*DummyConf), &loadedConf)
	assert.Equal(t, defaultConf, &loadedConf)
}

// create dummy configuration file to a temp dir
func createDummyConfFile() (string, error) {
	dummySample := `#dummy configration
	one = 1
	unset = 0
	str = "text"
	bool = true
	
	[sub]
	numarray = [
		1, 
		2,
	]
	strarray = [
		"a",
		"b",
	]
	`
	content := []byte(dummySample)
	tmpfile, err := ioutil.TempFile("", "dummyconf")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.Write(content); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}
	return tmpfile.Name(), nil
}

func TestLoad(t *testing.T) {
	// write dummy configuration file
	confFilePath, err := createDummyConfFile()
	if err != nil {
		assert.Fail(t, err.Error())
	}

	// create a default configuration
	toOverwriteConf := &DummyConf{
		DummyRootConf: DummyRootConf{
			One:   11,
			Unset: 12,
			Str:   "modified_text",
			Bool:  false,
		},
		Sub: &DummySubConf{
			NumArray: []int{},
			StrArray: []string{"y"},
		},
	}

	// load a config file over the default configuration
	dummyCxt := NewDummyContext("", confFilePath)
	defaultConf := dummyCxt.GetDefaultConfig()
	if err := dummyCxt.LoadOrCreateConfig(toOverwriteConf); err != nil {
		assert.Fail(t, err.Error())
	}

	// check field changes
	assert.Equal(t, defaultConf, toOverwriteConf)
}

func TestExpand(t *testing.T) {
	myText := "$HOME/mypath"
	dummyCxt := NewDummyContext("/myhome", "")

	assert.Equal(t, dummyCxt.ExpandPathEnv(myText), "/myhome/mypath")
}

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) SetupTest() {
	os.Unsetenv("AG_HOME")
}

func TestLoadOrCreate(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) TestSetNothing() {
	// remove os default user home path
	homeBackup := os.Getenv("HOME")
	userProfileBackup := os.Getenv("USERPROFILE")
	os.Unsetenv("HOME")
	os.Unsetenv("USERPROFILE")

	// create an empty viper config
	dummyCxt := NewDummyContext("", "")
	viperConf := dummyCxt.Vc

	// calculate and fill home & config path
	dummyCxt.retrievePath(viperConf)

	// path must be a default, if there is no env and parameters
	suite.Equal(viperConf.Get("home"), defaultHomePath)
	confPath := path.Join(defaultHomePath, defaultConfigFileName)
	suite.Equal(viperConf.ConfigFileUsed(), confPath)

	// restore environment vars
	os.Setenv("HOME", homeBackup)
	os.Setenv("USERPROFILE", userProfileBackup)
}

func (suite *ConfigTestSuite) TestSetOsHome() {
	var homePath, confPath string

	// generate home and conf path based on a os home path
	if os.Getenv("HOME") != "" { // for unix
		homePath = path.Join(os.Getenv("HOME"), defaultHomePath)
		confPath = path.Join(homePath, defaultConfigFileName)
	} else if os.Getenv("USERPROFILE") != "" { // for windows
		homePath = path.Join(os.Getenv("USERPROFILE"), defaultHomePath)
		confPath = path.Join(homePath, defaultConfigFileName)
	}

	// create an empty config
	dummyCxt := NewDummyContext("", "")
	viperConf := dummyCxt.Vc

	// calculate home and config path
	dummyCxt.retrievePath(viperConf)

	// compare result
	suite.Equal(viperConf.Get("home"), homePath)
	suite.Equal(viperConf.ConfigFileUsed(), confPath)
}

func (suite *ConfigTestSuite) TestSetEnvHome() {
	// create a temporal directory
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		suite.Fail(err.Error())
	}
	// set the temporal directory as a aergo home
	os.Setenv("AG_HOME", tmpDir)

	// create a default config
	dummyCxt := NewDummyContext("", "")
	defaultConf := dummyCxt.GetDefaultConfig()

	// create a config file
	err = dummyCxt.LoadOrCreateConfig(defaultConf)
	if err != nil {
		suite.Fail(err.Error())
	}

	// check an existence of the config file that is created by LoadOrCreateConfig func
	// configuration file must be located at ag_home, which is set at env
	confFilePath := path.Join(tmpDir, defaultConfigFileName)
	if _, err := os.Open(confFilePath); os.IsNotExist(err) {
		suite.Fail(err.Error())
	}
}

func (suite *ConfigTestSuite) TestSetEnvHomeAndConfPath() {
	// create a temporal directory for home dir
	tmpHomeDir, err := ioutil.TempDir("", "test")
	if err != nil {
		suite.Fail(err.Error())
	}
	// set the temporal directory as a home
	os.Setenv("AG_HOME", tmpHomeDir)

	// create another a temporal directory for conf path
	tmpConfDir, err := ioutil.TempDir("", "test")
	if err != nil {
		suite.Fail(err.Error())
	}
	confFilePath := path.Join(tmpConfDir, defaultConfigFileName)

	// create a default config
	dummyCxt := NewDummyContext("", confFilePath)
	defaultConf := dummyCxt.GetDefaultConfig()

	// create a config
	err = dummyCxt.LoadOrCreateConfig(defaultConf)
	if err != nil {
		suite.Fail(err.Error())
	}

	// check an existence of the config file that is created by LoadOrCreateConfig func
	// when user specifies the confFilePath, even if he set the ag_home,
	// the config file at the confFilePath must be loaded
	if _, err := os.Open(confFilePath); os.IsNotExist(err) {
		suite.Fail(err.Error())
	}

	// the conf file must not be created at home folder
	homeConfFilePath := path.Join(tmpHomeDir, defaultConfigFileName)
	if _, err := os.Open(homeConfFilePath); os.IsNotExist(err) == false {
		suite.Failf("file %s must not exist", homeConfFilePath)
	}
}

func (suite *ConfigTestSuite) TestSetParamHome() {
	// create a temporal directory
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		suite.Fail(err.Error())
	}
	// generate a random home path and set to the conf
	generatedHomePath := path.Join(tmpDir, "aergoHome")

	// create a default config
	dummyCxt := NewDummyContext(generatedHomePath, "")
	defaultConf := dummyCxt.GetDefaultConfig()

	// create a config
	err = dummyCxt.LoadOrCreateConfig(defaultConf)
	if err != nil {
		suite.Fail(err.Error())
	}

	// check an existence of the config file that is created by LoadOrCreateConfig func
	// the config file must exist in the home path
	confFilePath := path.Join(generatedHomePath, defaultConfigFileName)
	if _, err := os.Open(confFilePath); os.IsNotExist(err) {
		suite.Fail(err.Error())
	}
}

func (suite *ConfigTestSuite) TestSetParamConfPath() {
	// create a temporal directory
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		suite.Fail(err.Error())
	}
	// generate a random conf file path and set to the conf
	generatedConfFilePath := path.Join(tmpDir, "aergo.toml")

	// create a default config
	dummyCxt := NewDummyContext("", generatedConfFilePath)
	defaultConf := dummyCxt.GetDefaultConfig()

	// create a config
	err = dummyCxt.LoadOrCreateConfig(defaultConf)
	if err != nil {
		suite.Fail(err.Error())
	}

	// check an existence of the config file that is created by LoadOrCreateConfig func
	// when an user specify the exact config file path, just check it
	// if it does not exist, then create it
	if _, err := os.Open(generatedConfFilePath); os.IsNotExist(err) {
		suite.Fail(err.Error())
	}
}

func (suite *ConfigTestSuite) TestSetParamHomeAndConfPath() {
	// create a temporal directory
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		suite.Fail(err.Error())
	}
	// generate a random home path and set to the conf
	generatedHomePath := path.Join(tmpDir, "aergoHome")

	// generate a random conf file path and set to the conf
	generatedConfFilePath := path.Join(tmpDir, "aergo.toml")

	// create an default config
	dummyCxt := NewDummyContext(generatedHomePath, generatedConfFilePath)
	defaultConf := dummyCxt.GetDefaultConfig()

	// create a config
	err = dummyCxt.LoadOrCreateConfig(defaultConf)
	if err != nil {
		suite.Fail(err.Error())
	}

	// check an existence of the config file that is created by LoadOrCreateConfig func
	if _, err := os.Open(generatedConfFilePath); os.IsNotExist(err) {
		suite.Fail(err.Error())
	}

	// because a conf path is set, the conf file must be located at that path,
	// not at the home path
	confFilePath := path.Join(generatedHomePath, defaultConfigFileName)
	if _, err := os.Open(confFilePath); os.IsNotExist(err) == false {
		suite.Failf("file %s must not exist", confFilePath)
	}
}
