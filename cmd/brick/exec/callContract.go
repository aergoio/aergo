package exec

import (
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/contract/vm_dummy"
	"github.com/aergoio/aergo/v2/types"
	"github.com/rs/zerolog"
)

func init() {
	registerExec(&callContract{})
}

type callContract struct{}

func (c *callContract) Command() string {
	return "call"
}

func (c *callContract) Syntax() string {
	return fmt.Sprintf("%s %s %s %s %s %s %s", context.AccountSymbol,
		context.AmountSymbol, context.ContractSymbol,
		context.FunctionSymbol, context.ContractArgsSymbol,
		context.ExpectedErrSymbol, context.ExpectedSymbol)
}

func (c *callContract) Usage() string {
	return fmt.Sprintf("call <sender_name> <amount> <contract_name> <func_name> `[call_json_str]` `[expected_error_str]` `[expected_result_str]`")
}

func (c *callContract) Describe() string {
	return "call to execute a smart contract"
}

func (c *callContract) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, _, _, _, _, _, err := c.parse(args)

	return err
}

func (c *callContract) parse(args string) (string, *big.Int, string, string, string, string, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 4 {
		return "", nil, "", "", "", "", "", fmt.Errorf("need at least 4 arguments. usage: %s", c.Usage())
	}

	amount, success := new(big.Int).SetString(splitArgs[1].Text, 10)
	if success == false {
		return "", nil, "", "", "", "", "", fmt.Errorf("fail to parse number %s", splitArgs[1].Text)
	}

	callCode := "[]"
	if len(splitArgs) >= 5 {
		callCode = splitArgs[4].Text
	}

	expectedError := ""
	expectedRes := ""

	if len(splitArgs) >= 6 {
		expectedError = splitArgs[5].Text
	}
	if len(splitArgs) == 7 {
		expectedRes = splitArgs[6].Text
	} else if len(splitArgs) > 7 {
		return "", nil, "", "", "", "", "", fmt.Errorf("too many arguments. usage: %s", c.Usage())
	}

	return splitArgs[0].Text, //accountName
		amount, //amount
		splitArgs[2].Text, //contractName
		splitArgs[3].Text, //funcName
		callCode, //callCode
		expectedError, //expectedError
		expectedRes, //expectedRes
		nil
}

func (c *callContract) Run(args string) (string, uint64, []*types.Event, error) {

	accountName, amount, contractName, funcName, callCode, expectedError, expectedRes, _ := c.parse(args)

	formattedQuery := fmt.Sprintf("{\"name\":\"%s\",\"args\":%s}", funcName, callCode)

	callTx := vm_dummy.NewLuaTxCallBig(accountName, contractName, amount, formattedQuery)

	logLevel := zerolog.GlobalLevel()

	if expectedError != "" {
		callTx.Fail(expectedError)
		zerolog.SetGlobalLevel(zerolog.ErrorLevel) // turn off log
	}
	err := context.Get().ConnectBlock(callTx)

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

	receipt := context.Get().GetReceipt(callTx.Hash())

	if expectedRes != "" && expectedRes != receipt.Ret {
		err = fmt.Errorf("expected: %s, but got: %s", expectedRes, receipt.Ret)
		return "", 0, nil, err
	}

	result := "success"
	if expectedRes == "" && len(receipt.Ret) > 0 {
		result += ": " + receipt.Ret
	}
	return result, receipt.GasUsed, context.Get().GetEvents(callTx.Hash()), nil

}
