package key

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"io/ioutil"
	"os"
)

type Address = []byte

func GenerateAddress(pubkey *ecdsa.PublicKey) []byte {
	addr := new(bytes.Buffer)
	binary.Write(addr, binary.LittleEndian, pubkey.X.Bytes())
	binary.Write(addr, binary.LittleEndian, pubkey.Y.Bytes())
	return addr.Bytes()[:20] //TODO: ADDRESSLENGTH ?
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
	const addressLength = 20
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
