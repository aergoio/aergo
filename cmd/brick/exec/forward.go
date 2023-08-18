package exec

import (
	"fmt"
	"strconv"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&forward{})
}

type forward struct{}

func (c *forward) Command() string {
	return "forward"
}

func (c *forward) Syntax() string {
	return context.AmountSymbol
}

func (c *forward) Usage() string {
	return "forward [height_to_skip]"
}

func (c *forward) Describe() string {
	return "fast forward blocks n times (default = 1)"
}

func (c *forward) Validate(args string) error {
	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, err := c.parse(args)

	return err
}

func (c *forward) parse(args string) (int, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) == 0 {
		height, _ := strconv.Atoi("1")
		return height, nil
	} else if len(splitArgs) > 1 {
		return 0, fmt.Errorf("need 1 or 0 arguments. usage: %s", c.Usage())
	}

	amount, err := strconv.Atoi(splitArgs[0].Text)
	if err != nil {
		return 0, fmt.Errorf("fail to parse number %s: %s", splitArgs[0].Text, err.Error())
	}

	return amount, nil
}

func (c *forward) Run(args string) (string, uint64, []*types.Event, error) {
	amount, _ := c.parse(args)

	for i := 0; i < amount; i++ {
		if err := context.Get().ConnectBlock(); err != nil {
			return "", 0, nil, err
		}
	}

	return "fast forward blocks successfully", 0, nil, nil
}
