package exec

import (
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/contract/vm_dummy"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&getStateAccount{})
}

type getStateAccount struct{}

func (c *getStateAccount) Command() string {
	return "getstate"
}

func (c *getStateAccount) Syntax() string {
	return fmt.Sprintf("%s", context.AccountSymbol)
}

func (c *getStateAccount) Usage() string {
	return fmt.Sprintf("getstate <account_name> `[expected_balance]`")
}

func (c *getStateAccount) Describe() string {
	return "get the current state of an account"
}

func (c *getStateAccount) Validate(args string) error {
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, err := c.parse(args)

	return err
}

func (c *getStateAccount) parse(args string) (string, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 1 {
		return "", "", fmt.Errorf("missing arguments. usage: %s", c.Usage())
	}

	expectedResult := ""
	if len(splitArgs) == 2 {
		expectedResult = splitArgs[1].Text
	} else if len(splitArgs) > 2 {
		return "", "", fmt.Errorf("too many arguments. usage: %s", c.Usage())
	}

	return splitArgs[0].Text, expectedResult, nil
}

func (c *getStateAccount) Run(args string) (string, uint64, []*types.Event, error) {
	accountName, expectedResult, _ := c.parse(args)

	state, err := context.Get().GetAccountState(accountName)

	if err != nil {
		return "", 0, nil, err
	}
	if expectedResult == "" {
		return fmt.Sprintf("%s = %d", vm_dummy.StrToAddress(accountName), new(big.Int).SetBytes(state.GetBalance())), 0, nil, nil
	} else {
		strRet := fmt.Sprintf("%d", new(big.Int).SetBytes(state.GetBalance()))
		if expectedResult == strRet {
			return "state compare successfully", 0, nil, nil
		} else {
			return "", 0, nil, fmt.Errorf("state compare failed. Expected: %s, Actual: %s", expectedResult, strRet)
		}
	}
}
