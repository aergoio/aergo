//go:build !Debug
// +build !Debug

package vm_dummy

import (
	"math/big"

	"github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/contract"
)

func NewLuaTxDeployBig(sender, recipient string, amount *big.Int, code string) *luaTxDeploy {
	byteCode, err := contract.Compile(code, nil)
	if err != nil {
		return &luaTxDeploy{cErr: err}
	}
	return &luaTxDeploy{
		luaTxContractCommon: luaTxContractCommon{
			_sender:    contract.StrHash(sender),
			_recipient: contract.StrHash(recipient),
			_payload:   util.NewLuaCodePayload(byteCode, nil),
			_amount:    amount,
			txId:       newTxId(),
		},
		cErr: nil,
	}
}
