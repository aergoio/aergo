package exec

import (
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/contract/vm_dummy"
	"github.com/aergoio/aergo/v2/types"
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

func (c *injectAccount) parse(args string) (string, *big.Int, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 2 {
		return "", nil, fmt.Errorf("need 2 arguments. usage: %s", c.Usage())
	}

	amount, success := new(big.Int).SetString(splitArgs[1].Text, 10)
	if success == false {
		return "", nil, fmt.Errorf("fail to parse number %s", splitArgs[1].Text)
	}

	return splitArgs[0].Text, amount, nil
}

func (c *injectAccount) Run(args string) (string, uint64, []*types.Event, error) {
	accountName, amount, _ := c.parse(args)

	err := context.Get().ConnectBlock(
		vm_dummy.NewLuaTxAccountBig(accountName, amount),
	)

	if err != nil {
		return "", 0, nil, err
	}

	Index(context.AccountSymbol, accountName)

	return "inject an account successfully", 0, nil, nil
}
