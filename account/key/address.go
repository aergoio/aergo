package key

import (
	"crypto/ecdsa"
	"io/ioutil"
	"os"
	"encoding/binary"
	"bytes"
)

type Address = []byte

const addressLength = 33

func GenerateAddress(pubkey *ecdsa.PublicKey) []byte {
	addr := new(bytes.Buffer)
	// Compressed pubkey
	binary.Write(addr, binary.LittleEndian, pubkey.X.Bytes())
	binary.Write(addr, binary.LittleEndian, uint8(0x2 + pubkey.Y.Bit(0)))  // 0x2 for even, 0x3 for odd Y
	return addr.Bytes() // 33 bytes
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
