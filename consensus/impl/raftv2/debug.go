package raftv2

import (
	"os"
	"strconv"
	"time"
)

var (
	DEBUG_PROPOSE_SLEEP = "DEBUG_PROPOSE_SLEEP"
)

func checkEnv(envName string) int {
	envStr := os.Getenv(envName)
	if len(envStr) > 0 {
		val, err := strconv.Atoi(envStr)
		if err != nil {
			logger.Error().Err(err).Msgf("%s environment varialble must be integer", envName)
			return 0
		}
		logger.Debug().Int("value", val).Msgf("env variable[%s] is set", envName)

		return val
	}
	return 0
}

func debugRaftProposeSleep() {
	val := checkEnv(DEBUG_PROPOSE_SLEEP)

	if val > 0 {
		logger.Debug().Int("sleep", val).Msg("sleep raft propose")
		time.Sleep(time.Second * time.Duration(val))
	}
}
