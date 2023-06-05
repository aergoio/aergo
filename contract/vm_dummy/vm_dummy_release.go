//go:build !Debug
// +build !Debug

package vm_dummy

import (
	"math/big"

	"github.com/aergoio/aergo/cmd/aergoluac/util"
	"github.com/aergoio/aergo/contract"
)

func NewLuaTxDefBig(sender, contractMsg string, amount *big.Int, code string) *luaTxDef {
	byteCode, err := contract.Compile(code, nil)
	if err != nil {
		return &luaTxDef{cErr: err}
	}
	return &luaTxDef{
		luaTxContractCommon: luaTxContractCommon{
			_sender:   contract.StrHash(sender),
			_contract: contract.StrHash(contractMsg),
			_code:     util.NewLuaCodePayload(byteCode, nil),
			_amount:   amount,
			txId:      newTxId(),
		},
		cErr: nil,
	}
}
