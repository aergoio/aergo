package key

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func NewAddressEth(pubkey []byte) common.Address {
	unCompressedPubkey := ConvAddressUncompressed(pubkey)
	if unCompressedPubkey == nil {
		return common.Address{}
	}
	return common.BytesToAddress(unCompressedPubkey)
}

func NewContractEth(from common.Address, nonce uint64) common.Address {
	return crypto.CreateAddress(from, nonce)
}
