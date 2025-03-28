package exec

import (
	"fmt"
	"math/big"
	"path/filepath"
	"strings"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
	"github.com/aergoio/aergo/v2/contract/vm_dummy"
	"github.com/aergoio/aergo/v2/types"
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

func (c *deployContract) parse(args string) (string, *big.Int, string, string, string, error) {
	splitArgs := context.SplitSpaceAndAccent(args, false)
	if len(splitArgs) < 4 {
		return "", nil, "", "", "", fmt.Errorf("need 4 arguments. usage: %s", c.Usage())
	}

	amountStr := context.ParseDecimalAmount(splitArgs[1].Text, 18)
	amount, success := new(big.Int).SetString(amountStr, 10)
	if success == false {
		return "", nil, "", "", "", fmt.Errorf("fail to parse number %s", splitArgs[1].Text)
	}

	defPath := splitArgs[3].Text
	if _, err := readContract(defPath); err != nil {
		return "", nil, "", "", "", fmt.Errorf("fail to read a contract def file %s: %s", splitArgs[3].Text, err.Error())
	}

	constuctorArg := "[]"
	if len(splitArgs) == 5 {
		constuctorArg = splitArgs[4].Text
	} else if len(splitArgs) > 5 {
		return "", nil, "", "", "", fmt.Errorf("too many arguments. usage: %s", c.Usage())
	}

	return splitArgs[0].Text, //accountName
		amount, // amount
		splitArgs[2].Text, // contractName
		defPath, // defPath
		constuctorArg,
		nil
}

func (c *deployContract) Run(args string) (string, uint64, []*types.Event, error) {
	accountName, amount, contractName, defPath, constuctorArg, _ := c.parse(args)

	defByte, err := readContract(defPath)
	if err != nil {
		return "", 0, nil, err
	}

	updateContractInfoInterface(contractName, defPath)

	tx := vm_dummy.NewLuaTxDeployBig(accountName, contractName, amount, string(defByte)).Constructor(constuctorArg)
	err = context.Get().ConnectBlock(tx)

	if enableWatch && !strings.HasPrefix(defPath, "http") {
		absPath, _ := filepath.Abs(defPath)
		watcher.Add(absPath)
	}

	if err != nil {
		return "", 0, nil, err
	}

	Index(context.ContractSymbol, contractName)
	Index(context.AccountSymbol, contractName)

	return "deploy a smart contract successfully",
		context.Get().GetReceipt(tx.Hash()).GasUsed,
		context.Get().GetEvents(tx.Hash()),
		nil
}
