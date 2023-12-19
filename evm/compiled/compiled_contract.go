package compiled

import (
	_ "embed"
	"encoding/json"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	//go:embed HelloWorld.json
	HelloWorldJSON     []byte
	HelloWorldContract CompiledContract

	//go:embed ERC20.json
	erc20JSON     []byte
	ERC20Contract CompiledContract
)

func init() {
	// init hello world contract
	err := json.Unmarshal(HelloWorldJSON, &HelloWorldContract)
	if err != nil {
		panic(err)
	} else if len(HelloWorldContract.Bin) == 0 {
		panic("load contract failed")
	}

	// init erc20 contract
	err = json.Unmarshal(erc20JSON, &ERC20Contract)
	if err != nil {
		panic(err)
	} else if len(ERC20Contract.Bin) == 0 {
		panic("load contract failed")
	}
}

// CompiledContract contains compiled bytecode and abi
type CompiledContract struct {
	ABI abi.ABI
	Bin HexString
}

func (c *CompiledContract) Data(args ...interface{}) ([]byte, error) {
	var ctorArgs []byte
	var err error
	ctorArgs, err = c.ABI.Pack("", args...)
	if err != nil {
		return nil, err
	}
	data := append(c.Bin, ctorArgs...)
	return data, nil
}

func (c *CompiledContract) CallData(funcName string, args ...interface{}) ([]byte, error) {
	return c.ABI.Pack(funcName, args...)
}
