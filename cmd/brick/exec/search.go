package exec

import (
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
					contractName = argsArray[i-1]
				} else if symbol == context.FunctionSymbol {
					funcName = argsArray[i-1]
				}
			}
			if contractName != "" && funcName != "" {
				return searchAbiHint(contractName, funcName)
			}
		}
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

func Index(symbol, text string) {
	if _, ok := index[symbol]; !ok {
		index[symbol] = make(map[string]string)
	}
	index[symbol][text] = symbol
}
