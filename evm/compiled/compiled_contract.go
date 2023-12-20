package compiled

import (
	_ "embed"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// CompiledContract contains compiled bytecode and abi
type CompiledContract struct {
	ABI abi.ABI
	Bin HexString
}

func (c *CompiledContract) DeployData(args ...interface{}) ([]byte, error) {
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
