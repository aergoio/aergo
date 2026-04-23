package config

import (
	"os"
	"strconv"

	"github.com/aergoio/aergo-lib/log"
)

// This config is special config for NHIS site, and used only in aergosvr-2.3.x.
// And, these configs are set by environment variables only.

const (
	TURN_OFF_CONTROL_COMPACTION = "TURN_OFF_CONTROL_COMPACTION"
	COMPACTION_CHAIN_DB_PORT    = "COMPACTION_CHAIN_DB_PORT"
	COMPACTION_STATE_DB_PORT    = "COMPACTION_STATE_DB_PORT"

	DEFAULT_CHAIN_DB_PORT = 17092
	DEFAULT_STATE_DB_PORT = 17091
)

// DBConfig defines configurations for db modnitoring
type DBConfig struct {
	ControlCompaction bool `mapstructure:"controlcompaction" description:"enable control compaction"`
	StateDBPort       int  `mapstructure:"statedbport" description:"StateDB control port"`
	ChainDBPort       int  `mapstructure:"chaindbport" description:"ChainDB control port"`
}

var (
	logger = log.NewLogger("db_compaction")

	defaultDBConfig *DBConfig
)

func init() {
	var turnOffCompaction bool = false
	var chainDBPort int = DEFAULT_CHAIN_DB_PORT
	var stateDBPort int = DEFAULT_STATE_DB_PORT

	chainDBPortVar := os.Getenv(COMPACTION_CHAIN_DB_PORT)
	if chainDBPortVar != "" {
		chainDBPort, _ = strconv.Atoi(chainDBPortVar)
	}
	stateDBPortVar := os.Getenv(COMPACTION_STATE_DB_PORT)
	if stateDBPortVar != "" {
		stateDBPort, _ = strconv.Atoi(stateDBPortVar)
	}

	turnOffCompactionVar := os.Getenv(TURN_OFF_CONTROL_COMPACTION)
	if turnOffCompactionVar == "true" || turnOffCompactionVar == "1" {
		logger.Info().Msg("turn off control compaction. the control port will not be opened.")
		turnOffCompaction = true
	} else {
		logger.Debug().Int("chaindbPort", chainDBPort).Int("stateDBPort", stateDBPort).Msg("turn on control compaction. the control port will be opened.")
	}
	defaultDBConfig = &DBConfig{
		ControlCompaction: !turnOffCompaction,
		StateDBPort:       stateDBPort,
		ChainDBPort:       chainDBPort,
	}
}

func GetDBConfig() *DBConfig {
	return defaultDBConfig
}
