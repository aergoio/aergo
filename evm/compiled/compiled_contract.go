package compiled

import (
	_ "embed"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	//go:embed ERC20Contract.json
	erc20JSON []byte
	//go:embed HelloWorldContract.json
	HelloWorldJSON []byte

	ERC20Contract CompiledContract
)

func init() {
	err := json.Unmarshal(erc20JSON, &ERC20Contract)
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
	if len(args) >= 2 {
		owner := args[0].(common.Address)
		supply := args[1].(*big.Int)
		ctorArgs, err = c.ABI.Pack("", owner, supply)
	} else {
		ctorArgs, err = c.ABI.Pack("")
	}
	if err != nil {
		return nil, err
	}
	data := append(c.Bin, ctorArgs...)
	return data, nil
}
