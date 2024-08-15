package exec

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/contract/vm_dummy"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&multicall{})
}

type multicall struct{}

func (c *multicall) Command() string {
	return "multicall"
}

func (c *multicall) Syntax() string {
	return fmt.Sprintf("%s %s %s %s", context.AccountSymbol,
		context.MulticallSymbol,
		context.ExpectedErrSymbol, context.ExpectedSymbol)
}

func (c *multicall) Usage() string {
	return fmt.Sprintf("multicall <sender> `[commands_json_str]` `[expected_error_str]` `[expected_result_str]`")
}

func (c *multicall) Describe() string {
	return "composable call to multiple smart contracts"
}

func (c *multicall) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, _, _, err := c.parse(args)

	return err
}

func (c *multicall) parse(args string) (string, string, string, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 2 {
		return "", "", "", "", fmt.Errorf("need at least 2 arguments. usage: %s", c.Usage())
	}

	callCode := splitArgs[1].Text
	expectedError := ""
	expectedRes := ""

	if len(splitArgs) >= 3 {
		expectedError = splitArgs[2].Text
	}
	if len(splitArgs) == 4 {
		expectedRes = splitArgs[3].Text
	} else if len(splitArgs) > 4 {
		return "", "", "", "", fmt.Errorf("too many arguments. usage: %s", c.Usage())
	}

	return splitArgs[0].Text, //accountName
		callCode,
		expectedError,
		expectedRes,
		nil
}

func (c *multicall) Run(args string) (string, uint64, []*types.Event, error) {

	accountName, payload, expectedError, expectedRes, _ := c.parse(args)

	multicallTx := vm_dummy.NewLuaTxMultiCall(accountName, payload)

	logLevel := zerolog.GlobalLevel()

	if expectedError != "" {
		multicallTx.Fail(expectedError)
		zerolog.SetGlobalLevel(zerolog.ErrorLevel) // turn off log
	}
	err := context.Get().ConnectBlock(multicallTx)

	if expectedError != "" {
		zerolog.SetGlobalLevel(logLevel) // restore log level
	}
	if err != nil {
		return "", 0, nil, err
	}

	if expectedError != "" {
		Index(context.ExpectedErrSymbol, expectedError)
		return "call a smart contract successfully", 0, nil, nil
	}

	receipt := context.Get().GetReceipt(multicallTx.Hash())

	if expectedRes != "" && expectedRes != receipt.Ret {
		err = fmt.Errorf("expected: %s, but got: %s", expectedRes, receipt.Ret)
		return "", 0, nil, err
	}

	result := "success"
	if expectedRes == "" && len(receipt.Ret) > 0 {
		result += ": " + receipt.Ret
	}
	return result, receipt.GasUsed, context.Get().GetEvents(multicallTx.Hash()), nil

}
