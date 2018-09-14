package key

import (
	"crypto/ecdsa"
	"io/ioutil"
	"os"
	"crypto/sha256"
)

type Address = []byte

const addressLength = 20

func GenerateAddress(pubkey *ecdsa.PublicKey) []byte {
	h := sha256.New()
	h.Write(pubkey.X.Bytes())
	h.Write(pubkey.Y.Bytes())
	return h.Sum(nil)[:addressLength]
}

func (ks *Store) SaveAddress(addr Address) error {
	f, err := os.OpenFile(ks.addresses, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(addr); err != nil {
		return err
	}
	return nil
}

func (ks *Store) GetAddresses() ([]Address, error) {
	b, err := ioutil.ReadFile(ks.addresses)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ret []Address
	for i := 0; i < len(b); i += addressLength {
		ret = append(ret, b[i:i+addressLength])
	}
	return ret, nil
}
