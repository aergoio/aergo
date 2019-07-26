package chain

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

type ErrDebug struct {
	cond  StopCond
	value int
}

type StopCond int

// stop before swap chain
const (
	DEBUG_CHAIN_STOP StopCond = 0 + iota
	DEBUG_CHAIN_RANDOM_STOP
	DEBUG_CHAIN_BP_SLEEP
	DEBUG_CHAIN_OTHER_SLEEP
	DEBUG_SYNCER_CRASH
	DEBUG_RAFT_SNAP_FREQ // change snap frequency after first snapshot
)

const (
	DEBUG_CHAIN_STOP_INF = DEBUG_RAFT_SNAP_FREQ
)

var (
	EnvNameStaticCrash     = "DEBUG_CHAIN_CRASH"       // 1 ~ 4
	EnvNameRandomCrashTime = "DEBUG_RANDOM_CRASH_TIME" // 1 ~ 600000(=10min) ms
	EnvNameChainBPSleep    = "DEBUG_CHAIN_BP_SLEEP"    // bp node sleeps before connecting block for each block (ms). used
	EnvNameChainOtherSleep = "DEBUG_CHAIN_OTHER_SLEEP" // non bp node sleeps before connecting block for each block (ms).
	EnvNameSyncCrash       = "DEBUG_SYNCER_CRASH"      // case 1
	EnvNameRaftSnapFreq    = "DEBUG_RAFT_SNAP_FREQ"    // case 1
)

var stopConds = [...]string{
	EnvNameStaticCrash,
	EnvNameRandomCrashTime,
	EnvNameChainBPSleep,
	EnvNameChainOtherSleep,
	EnvNameSyncCrash,
	EnvNameRaftSnapFreq,
}

type DebugHandler func(value int) error

func (c StopCond) String() string { return stopConds[c] }

func (ec *ErrDebug) Error() string {
	return fmt.Sprintf("stopped by debugger cond[%s]=%d", ec.cond.String(), ec.value)
}

type Debugger struct {
	sync.RWMutex
	condMap map[StopCond]int
	isEnv   map[StopCond]bool
}

func newDebugger() *Debugger {
	dbg := &Debugger{condMap: make(map[StopCond]int), isEnv: make(map[StopCond]bool)}

	checkEnv := func(condName StopCond) {
		envName := stopConds[condName]

		envStr := os.Getenv(envName)
		if len(envStr) > 0 {
			val, err := strconv.Atoi(envStr)
			if err != nil {
				logger.Error().Err(err).Msgf("%s environment varialble must be integer", envName)
				return
			}
			logger.Debug().Int("value", val).Msgf("env variable[%s] is set", envName)

			dbg.Set(condName, val, true)
		}
	}

	checkEnv(DEBUG_CHAIN_STOP)
	checkEnv(DEBUG_CHAIN_RANDOM_STOP)
	checkEnv(DEBUG_CHAIN_BP_SLEEP)
	checkEnv(DEBUG_CHAIN_OTHER_SLEEP)
	checkEnv(DEBUG_SYNCER_CRASH)
	checkEnv(DEBUG_RAFT_SNAP_FREQ)

	return dbg
}

func (debug *Debugger) Set(cond StopCond, value int, env bool) {
	if debug == nil {
		return
	}

	debug.Lock()
	defer debug.Unlock()

	logger.Debug().Int("cond", int(cond)).Str("name", stopConds[cond]).Int("val", value).Msg("set debug condition")

	debug.condMap[cond] = value
	debug.isEnv[cond] = env
}

func (debug *Debugger) Unset(cond StopCond) {
	if debug == nil {
		return
	}

	debug.Lock()
	defer debug.Unlock()

	logger.Debug().Str("cond", cond.String()).Msg("deubugger condition is unset")
	delete(debug.condMap, cond)
}

func (debug *Debugger) clear() {
	if debug == nil {
		return
	}

	debug.Lock()
	defer debug.Unlock()

	debug.condMap = make(map[StopCond]int)
	debug.isEnv = make(map[StopCond]bool)
}

func (debug *Debugger) Check(cond StopCond, value int, handler DebugHandler) error {
	if debug == nil {
		return nil
	}

	debug.Lock()
	defer debug.Unlock()

	if setVal, ok := debug.condMap[cond]; ok {
		logger.Debug().Str("cond", stopConds[cond]).Int("val", setVal).Msg("check debug condition")

		switch cond {
		case DEBUG_CHAIN_STOP:
			if setVal == value {
				if debug.isEnv[cond] {
					logger.Fatal().Str("cond", stopConds[cond]).Msg("shutdown by DEBUG_CHAIN_CRASH")
				} else {
					return &ErrDebug{cond: cond, value: value}
				}
			}

		case DEBUG_CHAIN_RANDOM_STOP:
			go crashRandom(setVal)
			handleCrashRandom(setVal)

		case DEBUG_CHAIN_OTHER_SLEEP, DEBUG_CHAIN_BP_SLEEP:
			handleChainSleep(setVal)

		case DEBUG_SYNCER_CRASH:
			if setVal == value {
				return handleSyncerCrash(setVal, cond)
			}
		case DEBUG_RAFT_SNAP_FREQ:
			handler(setVal)
		}
	}

	return nil
}

func handleChainSleep(sleepMils int) {
	logger.Debug().Int("sleep(ms)", sleepMils).Msg("before chain sleep")

	time.Sleep(time.Millisecond * time.Duration(sleepMils))

	logger.Debug().Msg("after chain sleep")
}

func handleCrashRandom(waitMils int) {
	logger.Debug().Int("after(ms)", waitMils).Msg("before random crash")

	go crashRandom(waitMils)
}

func handleSyncerCrash(val int, cond StopCond) error {
	if val >= 1 {
		logger.Fatal().Int("val", val).Msg("sync crash by DEBUG_SYNC_CRASH")
		return nil
	} else {
		return &ErrDebug{cond: cond, value: val}
	}
}

func crashRandom(waitMils int) {
	if waitMils <= 0 {
		return
	}

	time.Sleep(time.Millisecond * time.Duration(waitMils))

	logger.Debug().Msg("shutdown by DEBUG_RANDOM_CRASH_TIME")

	os.Exit(100)
}
