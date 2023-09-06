package exec

import (
	"fmt"
	"strconv"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&setTimestamp{})
}

type setTimestamp struct{}

func (c *setTimestamp) Command() string {
	return "timestamp"
}

func (c *setTimestamp) Syntax() string {
	return fmt.Sprintf("%s", context.TimestampSymbol)
}

func (c *setTimestamp) Usage() string {
	return fmt.Sprintf("timestamp <value_or_increment>")
}

func (c *setTimestamp) Describe() string {
	return "define or increment the current timestamp"
}

func (c *setTimestamp) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, err := c.parse(args)

	return err
}

func (c *setTimestamp) parse(args string) (string, int64, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 1 {
		return "", 0, fmt.Errorf("need 1 argument. usage: %s", c.Usage())
	}

	operation := "set"
	typed := splitArgs[0].Text

	if typed[0:1] == "+" {
		typed = typed[1:]
		operation = "add"
	}

	value, err := strconv.ParseInt(typed, 10, 64)
	if err != nil {
		return "", 0, err
	}

	return operation, value, nil
}

func (c *setTimestamp) Run(args string) (string, uint64, []*types.Event, error) {
	operation, value, _ := c.parse(args)

	context.Get().SetTimestamp(operation == "add", value)

	return "timestamp set successfully", 0, nil, nil
}
