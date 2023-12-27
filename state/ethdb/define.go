package ethdb

import (
	key "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/types"
	"github.com/ethereum/go-ethereum/common"
)

func GetAddressEth(id []byte) common.Address {
	if types.IsSpecialAccount(id) {
		return types.GetSpecialAccountEth(id)
	}
	addr, err := key.NewAccountEth(id)
	if err != nil {
		return common.BytesToAddress(id)
	}
	return addr
}
