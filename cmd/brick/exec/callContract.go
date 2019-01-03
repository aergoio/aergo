package exec

import (
	"fmt"
	"strconv"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/aergoio/aergo/contract"
)

func init() {
	registerExec(&callContract{})
}

type callContract struct{}

func (c *callContract) Command() string {
	return "call"
}

func (c *callContract) Syntax() string {
	return fmt.Sprintf("%s %s %s %s %s", context.AccountSymbol,
		context.AmountSymbol, context.ContractSymbol,
		context.FunctionSymbol, context.ContractArgsSymbol)
}

func (c *callContract) Usage() string {
	return fmt.Sprintf("call <sender_name> <amount> <contract_name> <func_name> `[call_json_str]`")
}

func (c *callContract) Describe() string {
	return "call to execute a smart contract"
}

func (c *callContract) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, _, _, _, _, err := c.parse(args)

	return err
}

func (c *callContract) parse(args string) (string, uint64, string, string, string, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 4 {
		return "", 0, "", "", "", "", fmt.Errorf("need at least 4 arguments. usage: %s", c.Usage())
	}

	amount, err := strconv.ParseUint(splitArgs[1].Text, 10, 64)
	if err != nil {
		return "", 0, "", "", "", "", fmt.Errorf("fail to parse number %s: %s", splitArgs[1].Text, err.Error())
	}

	callCode := "[]"
	if len(splitArgs) >= 5 {
		callCode = splitArgs[4].Text
	}

	expectedResult := ""
	if len(splitArgs) == 6 {
		expectedResult = splitArgs[5].Text
	}

	return splitArgs[0].Text, //accountName
		amount, //amount
		splitArgs[2].Text, //contractName
		splitArgs[3].Text, //funcName
		callCode, //callCode
		expectedResult, //expectedResult
		nil
}

func (c *callContract) Run(args string) (string, error) {

	accountName, amount, contractName, funcName, callCode, expectedResult, _ := c.parse(args)

	formattedQuery := fmt.Sprintf("{\"name\":\"%s\",\"args\":%s}", funcName, callCode)

	callTx := contract.NewLuaTxCall(accountName, contractName, amount, formattedQuery)
	if expectedResult != "" {
		callTx.Fail(expectedResult)
	}
	err := context.Get().ConnectBlock(callTx)

	if err != nil {
		return "", err
	}

	return "call a smart contract successfully", nil
}
