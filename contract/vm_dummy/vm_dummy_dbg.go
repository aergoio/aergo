//go:build Debug
// +build Debug

package vm_dummy

import (
	luacUtil "github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/contract"
	"math/big"
)

func getCompiledABI(code string) ([]byte, error) {
	byteCodeAbi, err := contract.Compile(code, nil)
	if err != nil {
		return nil, err
	}
	return byteCodeAbi.ABI(), nil
}

func NewLuaTxDeployBig(sender, recipient string, amount *big.Int, code string) *luaTxDeploy {
	abi, err := getCompiledABI(code)
	if err != nil {
		return &luaTxDeploy{cErr: err}
	}
	return &luaTxDeploy{
		luaTxContractCommon: luaTxContractCommon{
			_sender:    contract.StrHash(sender),
			_recipient: contract.StrHash(recipient),
			_payload:   luacUtil.NewLuaCodePayload(luacUtil.NewLuaCode([]byte(code), abi), nil),
			_amount:    amount,
			txId:       newTxId(),
		},
		cErr: nil,
	}
}
