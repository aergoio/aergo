package exec

import (
	"fmt"
	"strings"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&help{})
}

type help struct{}

func (c *help) Command() string {
	return "help"
}

func (c *help) Syntax() string {
	return context.CommandSymbol
}

func (c *help) Usage() string {
	return "help [command]"
}

func (c *help) Describe() string {
	return "print usages and descriptions of commands"
}

func (c *help) Validate(args string) error {
	if args == "" {
		return nil
	}

	executor := GetExecutor(args)
	if executor == nil {
		return fmt.Errorf("command not found")
	}
	return nil
}

func (c *help) Run(args string) (string, uint64, []*types.Event, error) {
	var result strings.Builder

	if args == "" {
		// print whole usage guide
		result.WriteString(fmt.Sprintf("Aergo Brick, Toy for Developing Contracts, version %s\n", context.GitHash))

		result.WriteString(fmt.Sprintf("\n%-12s%s\n", "Command", "Usage"))
		result.WriteString(fmt.Sprintf("=====================================================\n"))

		for _, executor := range AllExecutors() {
			result.WriteString(fmt.Sprintf("%-12s%s\n", executor.Command(), executor.Usage()))
		}

		return result.String(), 0, nil, nil
	}

	// print details of a specific command
	executor := GetExecutor(args)
	if executor != nil {
		result.WriteString(fmt.Sprintf("%s\tusage: %s\tdescr: %s", executor.Command(), executor.Usage(), executor.Describe()))
	}

	return result.String(), 0, nil, nil
}
