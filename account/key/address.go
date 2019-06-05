package key

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"

	"github.com/aergoio/aergo/types"
)

type Address = []byte

var addresses = []byte("ADDRESSES")

func GenerateAddress(pubkey *ecdsa.PublicKey) []byte {
	if pubkey == nil {
		return nil
	}
	addr := new(bytes.Buffer)
	// Compressed pubkey
	binary.Write(addr, binary.LittleEndian, uint8(0x2+pubkey.Y.Bit(0))) // 0x2 for even, 0x3 for odd Y
	keyLength := len(pubkey.X.Bytes())
	if keyLength < 32 { //add padding
		for i := 1; i < types.AddressLength-keyLength; i++ {
			binary.Write(addr, binary.LittleEndian, uint8(0))
		}
	}
	binary.Write(addr, binary.LittleEndian, pubkey.X.Bytes())
	return addr.Bytes() // 33 bytes
}

func (ks *Store) SaveAddress(addr Address) error {
	if len(addr) != types.AddressLength {
		return errors.New("invalid address length")
	}

	ks.RWMutex.Lock()
	defer ks.RWMutex.Unlock()

	addrs := append(ks.storage.Get(addresses), addr...)
	ks.storage.Set(addresses, addrs)
	return nil
}

func (ks *Store) GetAddresses() ([]Address, error) {
	ks.RWMutex.RLock()
	defer ks.RWMutex.RUnlock()

	b := ks.storage.Get(addresses)
	var ret []Address
	for i := 0; i < len(b); i += types.AddressLength {
		ret = append(ret, b[i:i+types.AddressLength])
	}
	return ret, nil
}
