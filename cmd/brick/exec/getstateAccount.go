package exec

import (
	"fmt"
	"strings"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/aergoio/aergo/contract"
)

func init() {
	registerExec(&getStateAccount{})
}

type getStateAccount struct{}

func (c *getStateAccount) Command() string {
	return "getstate"
}

func (c *getStateAccount) Syntax() string {
	return fmt.Sprintf("getstate %s", context.AccountSymbol)
}

func (c *getStateAccount) Usage() string {
	return fmt.Sprintf("getstate <account_name>")
}

func (c *getStateAccount) Describe() string {
	return "create an account with a given amount of balance"
}

func (c *getStateAccount) Validate(args string) error {
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, err := c.parse(args)

	return err
}

func (c *getStateAccount) parse(args string) (string, error) {
	splitArgs := strings.Fields(args)
	if len(splitArgs) < 1 {
		return "", fmt.Errorf("need an arguments. usage: %s", c.Usage())
	}

	return splitArgs[0], nil
}

func (c *getStateAccount) Run(args string) (string, error) {
	accountName, _ := c.parse(args)

	state, err := context.Get().GetAccountState(accountName)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s=%d", contract.StrToAddress(accountName), state.GetBalance()), nil
}
