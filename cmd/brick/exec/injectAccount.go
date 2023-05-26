package exec

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/types"
)

func init() {
	registerExec(&injectAccount{})
}

type injectAccount struct{}

func (c *injectAccount) Command() string {
	return "inject"
}

func (c *injectAccount) Syntax() string {
	return fmt.Sprintf("%s %s %s", context.VersionSymbol, context.AccountSymbol, context.AmountSymbol)
}

func (c *injectAccount) Usage() string {
	return fmt.Sprintf("inject <version> <account_name> <amount>")
}

func (c *injectAccount) Describe() string {
	return "create an account with a given amount of balance"
}

func (c *injectAccount) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, _, err := c.parse(args)

	return err
}

func (c *injectAccount) parse(args string) (int32, string, *big.Int, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 3 {
		return 0, "", nil, fmt.Errorf("need 3 arguments. usage: %s", c.Usage())
	}

	version, err := strconv.ParseInt(splitArgs[0].Text, 10, 32)
	if err != nil {
		return 0, "", nil, fmt.Errorf("fail to parse version %s", splitArgs[0].Text)
	}

	amount, success := new(big.Int).SetString(splitArgs[2].Text, 10)
	if success == false {
		return 0, "", nil, fmt.Errorf("fail to parse number %s", splitArgs[2].Text)
	}

	return int32(version), splitArgs[1].Text, amount, nil
}

func (c *injectAccount) Run(args string) (string, uint64, []*types.Event, error) {
	version, accountName, amount, _ := c.parse(args)

	err := context.Get().ConnectBlock(version, contract.NewLuaTxAccountBig(accountName, amount))

	if err != nil {
		return "", 0, nil, err
	}

	Index(context.AccountSymbol, accountName)

	return "inject an account successfully", 0, nil, nil
}
