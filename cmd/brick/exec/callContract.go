package exec

import (
	"fmt"
	"math/big"

	"github.com/rs/zerolog"

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
	return fmt.Sprintf("%s %s %s %s %s %s", context.AccountSymbol,
		context.AmountSymbol, context.ContractSymbol,
		context.FunctionSymbol, context.ContractArgsSymbol, context.ExpectedErrSymbol)
}

func (c *callContract) Usage() string {
	return fmt.Sprintf("call <sender_name> <amount> <contract_name> <func_name> `[call_json_str]` `[expected_error_str]`")
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

func (c *callContract) parse(args string) (string, *big.Int, string, string, string, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 4 {
		return "", nil, "", "", "", "", fmt.Errorf("need at least 4 arguments. usage: %s", c.Usage())
	}

	amount, success := new(big.Int).SetString(splitArgs[1].Text, 10)
	if success == false {
		return "", nil, "", "", "", "", fmt.Errorf("fail to parse number %s", splitArgs[1].Text)
	}

	callCode := "[]"
	if len(splitArgs) >= 5 {
		callCode = splitArgs[4].Text
	}

	expectedError := ""
	if len(splitArgs) == 6 {
		expectedError = splitArgs[5].Text
	} else if len(splitArgs) > 6 {
		return "", nil, "", "", "", "", fmt.Errorf("too many arguments. usage: %s", c.Usage())
	}

	return splitArgs[0].Text, //accountName
		amount, //amount
		splitArgs[2].Text, //contractName
		splitArgs[3].Text, //funcName
		callCode, //callCode
		expectedError, //expectedError
		nil
}

func (c *callContract) Run(args string) (string, error) {

	accountName, amount, contractName, funcName, callCode, expectedError, _ := c.parse(args)

	formattedQuery := fmt.Sprintf("{\"name\":\"%s\",\"args\":%s}", funcName, callCode)

	callTx := contract.NewLuaTxCallBig(accountName, contractName, amount, formattedQuery)

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
		return "", err
	}

	if expectedError == "" {
		events := context.Get().GetEvents(callTx)
		for _, event := range events {
			logger.Info().Str("args", event.GetJsonArgs()).Msg(event.GetEventName())
		}
	} else {
		Index(context.ExpectedErrSymbol, expectedError)
	}

	return "call a smart contract successfully", nil
}
