package exec

import (
	"fmt"
	"strconv"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&setHardfork{})
}

type setHardfork struct{}

func (c *setHardfork) Command() string {
	return "hardfork"
}

func (c *setHardfork) Syntax() string {
	return fmt.Sprintf("%s", context.HardforkSymbol)
}

func (c *setHardfork) Usage() string {
	return fmt.Sprintf("hardfork <version>")
}

func (c *setHardfork) Describe() string {
	return "define the hardfork version"
}

func (c *setHardfork) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, err := c.parse(args)

	return err
}

func (c *setHardfork) parse(args string) (int32, error) {

	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 1 {
		return 0, fmt.Errorf("need 1 argument. usage: %s", c.Usage())
	}

	version, err := strconv.ParseInt(splitArgs[0].Text, 10, 64)
	if err != nil {
		return 0, err
	}

	return int32(version), nil
}

func (c *setHardfork) Run(args string) (string, uint64, []*types.Event, error) {
	version, _ := c.parse(args)

	context.Get().HardforkVersion = version

	return "hardfork set successfully", 0, nil, nil
}
