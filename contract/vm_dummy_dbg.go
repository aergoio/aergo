//go:build Debug
// +build Debug

package contract

import (
	luacUtil "github.com/aergoio/aergo/cmd/aergoluac/util"
	"math/big"
)

func getCompiledABI(code string) ([]byte, error) {
	byteCodeAbi, err := compile(code, nil)
	if err != nil {
		return nil, err
	}
	return byteCodeAbi.ABI(), nil
}

func NewLuaTxDefBig(sender, contract string, amount *big.Int, code string) *luaTxDef {
	abi, err := getCompiledABI(code)
	if err != nil {
		return &luaTxDef{cErr: err}
	}
	return &luaTxDef{
		luaTxContractCommon: luaTxContractCommon{
			_sender:   strHash(sender),
			_contract: strHash(contract),
			_code:     luacUtil.NewLuaCodePayload(luacUtil.NewLuaCode([]byte(code), abi), nil),
			_amount:   amount,
			txId:      newTxId(),
		},
		cErr: nil,
	}
}
