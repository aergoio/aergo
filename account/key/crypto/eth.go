package key

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func NewAddressEth(pubKey []byte) common.Address {
	if !btcec.IsCompressedPubKey(pubKey) {
		return common.Address{}
	}
	ecdsaPubKey, err := btcec.ParsePubKey(pubKey, btcec.S256())
	if err != nil {
		return common.Address{}
	}
	return crypto.PubkeyToAddress(*ecdsaPubKey.ToECDSA())
}

func NewContractEth(from common.Address, nonce uint64) common.Address {
	return crypto.CreateAddress(from, nonce)
}
