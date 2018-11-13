package exec

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/aergoio/aergo/cmd/brick/context"
)

var index = make(map[string]map[string]string)

func Candidates(cmd, args, symbol string) map[string]string {

	if ret := search(cmd, args, symbol); ret != nil {
		return ret
	}

	ret := make(map[string]string)

	// if there is no suggestion, then return general
	if descr, ok := context.Symbols[symbol]; ok {
		ret[symbol] = descr
	}

	return ret
}

func search(cmd, args, symbol string) map[string]string {
	if keywords, ok := index[symbol]; ok {
		return keywords
	}

	// search contract args using get abi
	if symbol == context.ContractArgsSymbol {
		executor := GetExecutor(cmd)
		if executor != nil {
			argsArray := strings.Fields(args)
			symbols := strings.Fields(executor.Syntax())
			var contractName string
			var funcName string
			for i, symbol := range symbols {
				if len(argsArray) < i {
					break
				}
				if symbol == context.ContractSymbol {
					// compare with symbol in syntax and extract contract name
					contractName = argsArray[i-1]
				} else if symbol == context.FunctionSymbol {
					// extract function name
					funcName = argsArray[i-1]
				}
			}

			if contractName != "" && funcName != "" {
				// search abi using contract and function name
				return searchAbiHint(contractName, funcName)
			}
		}
	} else if symbol == context.PathSymbol {
		// search in file path
		argsArray := strings.Fields(args)
		lastArg := "."
		if len(argsArray) != 0 {
			// get last word
			lastArg = argsArray[len(argsArray)-1]
		}

		return searchInPath(lastArg)
	}

	return nil
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

func searchInPath(currentPathStr string) map[string]string {

	if strings.HasSuffix(currentPathStr, ".") {
		// attach file sperator, to get files in this relative path
		currentPathStr = fmt.Sprintf("%s%c", currentPathStr, filepath.Separator)
	}
	ret := make(map[string]string)

	// extract parent directory path
	dir := filepath.Dir(currentPathStr)

	// navigate file list in the parent directory
	fileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		ret[err.Error()] = ""
		return ret
	}

	// detatch last base path
	// other function internally use filepath.Clean() that remove text . or ..
	// it makes prompt filter hard to match suggestions and the input
	currentDir, _ := filepath.Split(currentPathStr)

	for _, file := range fileInfo {
		// generate suggestion text
		ret[currentDir+file.Name()] = ""
	}

	return ret
}

func Index(symbol, text string) {
	if _, ok := index[symbol]; !ok {
		index[symbol] = make(map[string]string)
	}
	index[symbol][text] = symbol
}
