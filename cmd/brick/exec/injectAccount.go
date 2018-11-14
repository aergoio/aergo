package exec

import (
	"fmt"
	"strconv"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/aergoio/aergo/contract"
)

func init() {
	registerExec(&injectAccount{})
}

type injectAccount struct{}

func (c *injectAccount) Command() string {
	return "inject"
}

func (c *injectAccount) Syntax() string {
	return fmt.Sprintf("%s %s", context.AccountSymbol, context.AmountSymbol)
}

func (c *injectAccount) Usage() string {
	return fmt.Sprintf("inject <account_name> <amount>")
}

func (c *injectAccount) Describe() string {
	return "create an account with a given amount of balance"
}

func (c *injectAccount) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, err := c.parse(args)

	return err
}

func (c *injectAccount) parse(args string) (string, uint64, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 2 {
		return "", 0, fmt.Errorf("need 2 arguments. usage: %s", c.Usage())
	}

	amount, err := strconv.ParseUint(splitArgs[1].Text, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("fail to parse number %s: %s", splitArgs[1].Text, err.Error())
	}

	return splitArgs[0].Text, amount, nil
}

func (c *injectAccount) Run(args string) (string, error) {
	accountName, amount, _ := c.parse(args)

	err := context.Get().ConnectBlock(
		contract.NewLuaTxAccount(accountName, amount),
	)

	if err != nil {
		return "", err
	}

	Index(context.AccountSymbol, accountName)

	return "inject an account successfully", nil
}
