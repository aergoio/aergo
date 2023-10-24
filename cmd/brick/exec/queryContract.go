package exec

import (
	"fmt"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/types"
)

func init() {
	registerExec(&queryContract{})
}

type queryContract struct{}

func (c *queryContract) Command() string {
	return "query"
}

func (c *queryContract) Syntax() string {
	return fmt.Sprintf("%s %s %s %s %s", context.ContractSymbol, context.FunctionSymbol,
		context.ContractArgsSymbol, context.ExpectedSymbol, context.ExpectedErrSymbol)
}

func (c *queryContract) Usage() string {
	return fmt.Sprintf("query <contract_name> <func_name> `[query_json_str]` `[expected_query_result]` `[expected_error_str]`")
}

func (c *queryContract) Describe() string {
	return "query a smart contract"
}

func (c *queryContract) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, _, _, _, err := c.parse(args)

	return err
}

func (c *queryContract) parse(args string) (string, string, string, string, string, error) {

	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 2 {
		return "", "", "", "", "", fmt.Errorf("need at least 2 arguments. usage: %s", c.Usage())
	}

	queryCode := "[]"

	if len(splitArgs) >= 3 {
		queryCode = splitArgs[2].Text
	}

	expectedResult := ""
	expectedError := ""
	if len(splitArgs) == 4 {
		expectedResult = splitArgs[3].Text
	} else if len(splitArgs) == 5 {
		expectedResult = splitArgs[3].Text
		expectedError = splitArgs[4].Text
	} else if len(splitArgs) > 5 {
		return "", "", "", "", "", fmt.Errorf("too many arguments. usage: %s", c.Usage())
	}

	return splitArgs[0].Text, // contractName
		splitArgs[1].Text, //funcName
		queryCode, //queryCode
		expectedResult, //expectedResult
		expectedError,
		nil
}

func (c *queryContract) Run(args string) (string, uint64, []*types.Event, error) {
	contractName, funcName, queryCode, expectedResult, expectedError, _ := c.parse(args)

	formattedQuery := fmt.Sprintf("{\"name\":\"%s\",\"args\":%s}", funcName, queryCode)

	if expectedResult == "" {
		// there is no expected result
		isTestPassed, result, err := context.Get().QueryOnly(contractName, formattedQuery, expectedError)

		if err != nil {
			return "", 0, nil, err
		} else if isTestPassed {
			return "query to a smart contract successfully", 0, nil, nil
		}

		return result, 0, nil, nil
	}
	// there is expected result
	err := context.Get().Query(contractName, formattedQuery, expectedError, expectedResult)

	if err != nil {
		return "", 0, nil, err
	}

	Index(context.ExpectedSymbol, expectedResult)
	if expectedError != "" {
		Index(context.ExpectedErrSymbol, expectedError)
	}

	return "query to a smart contract successfully", 0, nil, nil
}
