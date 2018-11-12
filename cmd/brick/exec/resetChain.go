package exec

import (
	"github.com/aergoio/aergo/cmd/brick/context"
)

func init() {
	registerExec(&resetChain{})
}

type resetChain struct{}

func (c *resetChain) Command() string {
	return "reset"
}

func (c *resetChain) Syntax() string {
	return "reset"
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

func (c *resetChain) Run(args string) (string, error) {

	context.Reset()

	return "reset a dummy chain successfully", nil
}
