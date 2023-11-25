package exec

import (
	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&resetChain{})
}

type resetChain struct{}

func (c *resetChain) Command() string {
	return "reset"
}

func (c *resetChain) Syntax() string {
	return ""
}

func (c *resetChain) Usage() string {
	return "reset"
}

func (c *resetChain) Describe() string {
	return "reset to a new dummy chain"
}

func (c *resetChain) Validate(args string) error {
	return nil
}

func (c *resetChain) Run(args string) (string, uint64, []*types.Event, error) {
	context.Reset()
	resetContractInfoInterface()
	return "reset a dummy chain successfully", 0, nil, nil
}
