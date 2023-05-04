//go:build !Debug
// +build !Debug

package contract

import (
	"math/big"

	"github.com/aergoio/aergo/cmd/aergoluac/util"
)

func NewLuaTxDefBig(sender, contract string, amount *big.Int, code string) *luaTxDef {
	byteCode, err := compile(code, nil)
	if err != nil {
		return &luaTxDef{cErr: err}
	}
	return &luaTxDef{
		luaTxContractCommon: luaTxContractCommon{
			_sender:   strHash(sender),
			_contract: strHash(contract),
			_code:     util.NewLuaCodePayload(byteCode, nil),
			_amount:   amount,
			txId:      newTxId(),
		},
		cErr: nil,
	}
}
