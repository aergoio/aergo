package exec

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
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

// Set to track imported files
var importedFiles = make(map[string]bool)

// ProcessLines processes a string containing contract code, handling imports
func (c *deployContract) processLines(input string) (string, error) {

	// Create a string builder for the output
	var output strings.Builder

	// Process each line
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()

		// Check if line starts with "import "
		if strings.HasPrefix(line, "import ") {
			// Extract the file name from import statement
			importLine := strings.TrimSpace(line[6:]) // Remove "import " prefix

			// Check if it has minimum length for a valid import
			if len(importLine) < 3 {
				return "", fmt.Errorf("invalid import format: %s", line)
			}

			// Check if it starts with a valid quote character
			quoteChar := importLine[0]
			if quoteChar != '"' && quoteChar != '\'' {
				return "", fmt.Errorf("import statement must use quotes: %s", line)
			}

			// Check if it ends with the same quote character
			if importLine[len(importLine)-1] != quoteChar {
				return "", fmt.Errorf("mismatched quotes in import: %s", line)
			}

			// Extract the file path between quotes
			importFile := importLine[1 : len(importLine)-1]

			// Get absolute path to check for circular imports
			absImportPath, err := filepath.Abs(importFile)
			if err != nil {
				return "", fmt.Errorf("error getting absolute path: %w", err)
			}

			// Skip if already imported
			if importedFiles[absImportPath] {
				continue
			}

			// Mark as imported
			importedFiles[absImportPath] = true

			// Read the imported file
			importContent, err := c.readContractFile(importFile)
			if err != nil {
				return "", fmt.Errorf("error importing file '%s': %w", importFile, err)
			}

			// Process the imported content recursively
			processedImport, err := c.processLines(importContent)
			if err != nil {
				return "", err
			}

			// Add the processed import to output
			output.WriteString(processedImport)
			output.WriteString("\n")
		} else {
			// Regular line, add to output
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning input: %w", err)
	}

	return output.String(), nil
}

func (c *deployContract) readContractFile(filePath string) (string, error) {
	// if the file path is a url, read it from the web
	if strings.HasPrefix(filePath, "http") {
		// search in the web
		req, err := http.NewRequest("GET", filePath, nil)
		if err != nil {
			return "", err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		fileBytes, _ := ioutil.ReadAll(resp.Body)
		return string(fileBytes), nil
	}

	// search in the local file system
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", err
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

func (c *deployContract) readContract(filePath string) (string, error) {
	// Reset imported files tracking for each new processing
	importedFiles = make(map[string]bool)
	// read the contract file
	output, err := c.readContractFile(filePath)
	if err != nil {
		return "", err
	}
	// process the contract file for import statements
	return c.processLines(output)
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
	if _, err := c.readContract(defPath); err != nil {
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

	defByte, err := c.readContract(defPath)
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
