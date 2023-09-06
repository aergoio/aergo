package exec

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/types"
	"github.com/rs/zerolog"
)

var logger = log.NewLogger("brick")

var storedCmdLine = ""

type Executor interface {
	Command() string
	Syntax() string
	Usage() string
	Describe() string
	Validate(args string) error
	Run(args string) (string, uint64, []*types.Event, error) // message, gas used, events, error
}

var execImpls = make(map[string]Executor)

func registerExec(executor Executor) {
	execImpls[executor.Command()] = executor
	Index(context.CommandSymbol, executor.Command())
}

func GetExecutor(cmd string) Executor {
	return execImpls[cmd]
}

func AllExecutors() []Executor {

	var result []Executor

	for _, executor := range execImpls {
		result = append(result, executor)
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(result[i].Command(), result[j].Command()) < 0
	})

	return result
}

func Broker(cmdStr string) {

	cmd, args := context.ParseFirstWord(cmdStr)
	if len(cmd) == 0 || context.Comment == cmd {
		return
	}

	Execute(cmd, args)
}

func Execute(cmd, args string) {
	executor := GetExecutor(cmd)

	if executor == nil {
		letBatchKnowErr = fmt.Errorf("command not found: %s", cmd)
		batchErrorCount++
		logger.Error().Str("cmd", cmd).Msg("command not found")
		return
	}

	if err := executor.Validate(args); err != nil {
		letBatchKnowErr = err
		batchErrorCount++
		logger.Error().Err(err).Str("cmd", cmd).Msg("validation fail")
		return
	}

	result, gasUsed, events, err := executor.Run(args)
	if err != nil {
		letBatchKnowErr = err
		batchErrorCount++
		logger.Error().Err(err).Str("cmd", cmd).Msg("execution fail")
		return
	}

	eventArray := zerolog.Arr()

	for _, event := range events {
		eventArgs := zerolog.Arr()
		var parsedArgs []interface{}
		json.Unmarshal([]byte(event.GetJsonArgs()), &parsedArgs)

		for _, v := range parsedArgs {
			eventArgs.Interface(v)
		}

		eventArray.RawJSON([]byte(fmt.Sprintf(`{"name":"%s", "args":%s}`, event.GetEventName(), event.GetJsonArgs())))
	}

	logEntry := logger.Info().Str("cmd", cmd)

	if gasUsed > 0 {
		logEntry.Uint64("gas", gasUsed)
	}
	if events != nil {
		logEntry.Array("events", eventArray)
	}
	logEntry.Msg(result)
}
