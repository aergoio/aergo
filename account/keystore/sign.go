package keystore

import (
	"github.com/btcsuite/btcd/btcec"
)

//Sign return sign with key
func (ks *KeyStore) Sign(address Address, pass string, hash []byte) ([]byte, error) {
	k, err := ks.getKey(address, pass)
	if k == nil {
		return nil, err
	}
	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), k)
	return btcec.SignCompact(btcec.S256(), key, hash, true)
}
