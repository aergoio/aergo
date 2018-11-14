package exec

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/aergoio/aergo/cmd/brick/context"
	"github.com/aergoio/aergo/contract"
)

func init() {
	registerExec(&deployContract{})
}

type deployContract struct{}

func (c *deployContract) Command() string {
	return "deploy"
}

func (c *deployContract) Syntax() string {
	return fmt.Sprintf("%s %s %s %s %s", context.AccountSymbol, context.AmountSymbol,
		context.ContractSymbol, context.PathSymbol, context.ContractArgsSymbol)
}

func (c *deployContract) Usage() string {
	return fmt.Sprintf("deploy <sender_name> <amount> <contract_name> `<definition_file_path>` `[contructor_json_arg]`")
}

func (c *deployContract) Describe() string {
	return "deploy a smart contract"
}

func (c *deployContract) Validate(args string) error {

	// check whether chain is loaded
	if context.Get() == nil {
		return fmt.Errorf("load chain first")
	}

	_, _, _, _, _, err := c.parse(args)

	return err
}

func (c *deployContract) parse(args string) (string, uint64, string, string, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 4 {
		return "", 0, "", "", "", fmt.Errorf("need 4 arguments. usage: %s", c.Usage())
	}
	amount, err := strconv.ParseUint(splitArgs[1].Text, 10, 64)
	if err != nil {
		return "", 0, "", "", "", fmt.Errorf("fail to parse number %s: %s", splitArgs[1].Text, err.Error())
	}
	defPath := splitArgs[3].Text
	if _, err := os.Stat(defPath); os.IsNotExist(err) {
		return "", 0, "", "", "", fmt.Errorf("fail to read a contrat def file %s: %s", splitArgs[3].Text, err.Error())
	}

	constuctorArg := ""
	if len(splitArgs) == 5 {
		constuctorArg = splitArgs[4].Text
	}

	return splitArgs[0].Text, //accountName
		amount, // amount
		splitArgs[2].Text, // contractName
		defPath, // defPath
		constuctorArg,
		nil
}

func (c *deployContract) Run(args string) (string, error) {
	accountName, amount, contractName, defPath, constuctorArg, _ := c.parse(args)

	defByte, err := ioutil.ReadFile(defPath)
	if err != nil {
		return "", err
	}

	err = context.Get().ConnectBlock(
		contract.NewLuaTxDef(accountName, contractName, amount, string(defByte)).Constructor(constuctorArg),
	)

	if err != nil {
		return "", err
	}

	Index(context.ContractSymbol, contractName)
	Index(context.AccountSymbol, contractName)

	// read receipt and extract abi functions
	abi, err := context.Get().GetABI(contractName)
	if err != nil {
		return "", err
	}
	for _, contractFunc := range abi.Functions {
		// indexing functions
		Index(context.FunctionSymbol, contractFunc.Name)
	}

	return "deploy a smart contract successfully", nil
}
