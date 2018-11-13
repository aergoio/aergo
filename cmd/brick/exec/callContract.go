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
	return fmt.Sprintf("call <sender_name> <amount> <contract_name> <func_name> `<call_json_str>`")
}

func (c *callContract) Describe() string {
	return "call to execute a smart contract"
}

func (c *callContract) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, _, _, _, err := c.parse(args)

	return err
}

func (c *callContract) parse(args string) (string, uint64, string, string, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 5 {
		return "", 0, "", "", "", fmt.Errorf("need 5 arguments. usage: %s", c.Usage())
	}

	amount, err := strconv.ParseUint(splitArgs[1], 10, 64)
	if err != nil {
		return "", 0, "", "", "", fmt.Errorf("fail to parse number %s: %s", splitArgs[1], err.Error())
	}

	callCode := splitArgs[4]
	return splitArgs[0], //accountName
		amount, //amount
		splitArgs[2], //contractName
		splitArgs[3], //funcName
		callCode, //callCode
		nil
}

func (c *callContract) Run(args string) (string, error) {

	accountName, amount, contractName, funcName, callCode, _ := c.parse(args)

	formattedQuery := fmt.Sprintf("{\"name\":\"%s\",\"args\":%s}", funcName, callCode)

	err := context.Get().ConnectBlock(
		contract.NewLuaTxCall(accountName, contractName, amount, formattedQuery),
	)

	if err != nil {
		return "", err
	}

	return "call a smart contract successfully", nil
}
