package p2putil

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// ConvertPKToBTCEC return nil if conversion is failed
func ConvertPKToBTCEC(pk crypto.PrivKey) *secp256k1.PrivateKey {
	raw, err := pk.Raw()
	if err != nil {
		return nil
	}
	return secp256k1.PrivKeyFromBytes(raw)
}

// ConvertPubKeyToBTCEC return nil if conversion is failed
func ConvertPubKeyToBTCEC(pk crypto.PubKey) *secp256k1.PublicKey {
	raw, err := pk.Raw()
	if err != nil {
		return nil
	}
	pub, _ := secp256k1.ParsePubKey(raw)
	return pub
}

// ConvertPKToLibP2P return nil if conversion is failed
func ConvertPKToLibP2P(pk *secp256k1.PrivateKey) crypto.PrivKey {
	libp2pKey, err := crypto.UnmarshalSecp256k1PrivateKey(pk.Serialize())
	if err != nil {
		return nil
	}
	return libp2pKey
}

// ConvertPubToLibP2P return nil if conversion is failed
func ConvertPubToLibP2P(pk *secp256k1.PublicKey) crypto.PubKey {
	libp2pKey, err := crypto.UnmarshalSecp256k1PublicKey(pk.SerializeCompressed())
	if err != nil {
		return nil
	}
	return libp2pKey
}
