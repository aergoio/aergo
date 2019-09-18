package context

var Comment = "#"
var (
	PathSymbol         = "<path>"
	ContractSymbol     = "<contract>"
	AccountSymbol      = "<account>"
	AmountSymbol       = "<amount>"
	ContractArgsSymbol = "<contract_args>"
	ExpectedSymbol     = "<expected>"
	ExpectedErrSymbol  = "<expected_err>"
	FunctionSymbol     = "<function>"
	CommandSymbol      = "[command]"
)

// reprenestation and description map of all symbols
var Symbols = make(map[string]string)

func init() {
	Symbols[PathSymbol] = "folder or file location"
	Symbols[ContractSymbol] = "contract address"
	Symbols[AccountSymbol] = "account address"
	Symbols[ContractArgsSymbol] = "an array of argments to call a contract"
	Symbols[AmountSymbol] = "amount of aergo to send"
	Symbols[ExpectedSymbol] = "expected result"
	Symbols[ExpectedErrSymbol] = "expected error"
	Symbols[FunctionSymbol] = "smart contract function name"
}
