package exec

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aergoio/aergo/v2/cmd/brick/context"
)

var index = make(map[string]map[string]string)

func Candidates(cmd string, chunks []context.Chunk, current int, symbol string) map[string]string {
	if ret := search(cmd, chunks, current, symbol); ret != nil {
		return ret
	}

	ret := make(map[string]string)

	// if there is no suggestion, then return general
	if descr, ok := context.Symbols[symbol]; ok {
		ret[symbol] = descr
	}

	return ret
}

// search contract args using get abi
func extractContractAndFuncName(cmd string, chunks []context.Chunk) (string, string) {
	var contractName string
	var funcName string

	executor := GetExecutor(cmd)
	if executor != nil {
		symbols := strings.Fields(executor.Syntax())

		for i, symbol := range symbols {
			if len(chunks) <= i {
				break
			}
			if symbol == context.ContractSymbol {
				// compare with symbol in syntax and extract contract name
				contractName = chunks[i].Text
			} else if symbol == context.FunctionSymbol {
				// extract function name
				funcName = chunks[i].Text
			}
		}
	}

	return contractName, funcName
}

func search(cmd string, chunks []context.Chunk, current int, symbol string) map[string]string {
	if keywords, ok := index[symbol]; ok {
		return keywords
	}

	if symbol == context.FunctionSymbol {
		contractName, _ := extractContractAndFuncName(cmd, chunks)
		if contractName != "" {
			return searchFuncHint(contractName)
		}
	} else if symbol == context.ContractArgsSymbol {
		contractName, funcName := extractContractAndFuncName(cmd, chunks)
		if contractName != "" && funcName != "" {
			// search abi using contract and function name
			return searchAbiHint(contractName, funcName)
		}
	} else if symbol == context.PathSymbol {
		if len(chunks) <= current { //there is no word yet
			return searchInPath(context.Chunk{Text: ".", Accent: false})
		}
		return searchInPath(chunks[current])
	}

	return nil
}

func searchFuncHint(contractName string) map[string]string {
	// read receipt and extract abi functions
	ret := make(map[string]string)

	abi, err := context.Get().GetABI(contractName)
	if err != nil {
		ret["<error>"] = err.Error()
		return ret
	}
	for _, contractFunc := range abi.Functions {
		// gather functions
		ret[contractFunc.Name] = ""
	}

	return ret
}

func searchAbiHint(contractName, funcName string) map[string]string {
	abi, err := context.Get().GetABI(contractName)
	if err != nil {
		return nil
	}

	for _, contractFunc := range abi.Functions {
		if contractFunc.Name == funcName {
			argsHint := "`["
			for i, funcArg := range contractFunc.GetArguments() {
				argsHint += funcArg.Name
				if i+1 != len(contractFunc.GetArguments()) {
					argsHint += ", "
				}
			}
			argsHint += "]`"

			ret := make(map[string]string)
			ret[argsHint] = context.Symbols[context.ContractArgsSymbol]
			return ret
		}
	}

	return nil
}

func searchInPath(chunk context.Chunk) map[string]string {

	if strings.HasSuffix(chunk.Text, ".") {
		// attach file sperator, to get files in this relative path
		chunk.Text = fmt.Sprintf("%s%c", chunk.Text, filepath.Separator)
	}
	ret := make(map[string]string)

	// extract parent directory path
	dir := filepath.Dir(chunk.Text)

	// navigate file list in the parent directory
	fileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		ret[err.Error()] = ""
		return ret
	}

	// detatch last base path
	// other function internally use filepath.Clean() that remove text . or ..
	// it makes prompt filter hard to match suggestions and the input

	currentDir, _ := filepath.Split(chunk.Text)

	for _, file := range fileInfo {
		// generate suggestion text
		fullPath := currentDir + file.Name()

		if file.IsDir() {
			fullPath += string(os.PathSeparator)
		}
		if chunk.Accent {
			fullPath = "`" + fullPath // attach accent again
		}
		// if contains white space...
		if wsIdx := strings.LastIndex(fullPath, " "); wsIdx != -1 {
			// cut it because auto completer will switch text only after whitespace
			fullPath = fullPath[wsIdx+1:]
		}

		ret[fullPath] = ""
	}

	return ret
}

func Index(symbol, text string) {
	if _, ok := index[symbol]; !ok {
		index[symbol] = make(map[string]string)
	}
	index[symbol][text] = symbol
}
