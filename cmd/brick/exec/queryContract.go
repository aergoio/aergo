package exec

import (
	"fmt"

	"github.com/aergoio/aergo/cmd/brick/context"
)

func init() {
	registerExec(&queryContract{})
}

type queryContract struct{}

func (c *queryContract) Command() string {
	return "query"
}

func (c *queryContract) Syntax() string {
	return fmt.Sprintf("%s %s %s %s", context.ContractSymbol, context.FunctionSymbol,
		context.ContractArgsSymbol, context.ExpectedSymbol)
}

func (c *queryContract) Usage() string {
	return fmt.Sprintf("query <contract_name> <func_name> `<query_json_str>` `[expected_query_result]`")
}

func (c *queryContract) Describe() string {
	return "query a smart contract"
}

func (c *queryContract) Validate(args string) error {

	// is chain is loaded?
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, _, _, err := c.parse(args)

	return err
}

func (c *queryContract) parse(args string) (string, string, string, string, error) {

	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 3 {
		return "", "", "", "", fmt.Errorf("need at least 3 arguments. usage: %s", c.Usage())
	}

	queryCodeAndExpected := splitArgs[2]

	expectedResult := ""
	if len(splitArgs) == 4 {
		expectedResult = splitArgs[3]
	}

	return splitArgs[0], // contractName
		splitArgs[1], //funcName
		queryCodeAndExpected, //queryCode
		expectedResult, //expectedResult
		nil
}

func (c *queryContract) Run(args string) (string, error) {
	contractName, funcName, queryCode, expectedResult, _ := c.parse(args)

	formattedQuery := fmt.Sprintf("{\"name\":\"%s\",\"args\":%s}", funcName, queryCode)

	if expectedResult == "" {
		// there is no expected result
		result, err := context.Get().QueryOnly(contractName, formattedQuery)

		if err != nil {
			return "", err
		}

		return result, nil
	}
	// there is expected result
	err := context.Get().Query(contractName, formattedQuery, "", expectedResult)

	if err != nil {
		return "", err
	}

	Index(context.ExpectedSymbol, expectedResult)

	return "query compare successfully", nil
}
