package account

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"io/ioutil"
	"os"

	"github.com/aergoio/aergo/types"
)

const addressLength = 20

type Addresses struct {
	addrPath string
}

func NewAddresses(addressPath string) *Addresses {
	return &Addresses{
		addrPath: addressPath,
	}
}

func generateAddress(pubkey *ecdsa.PublicKey) []byte {
	addr := new(bytes.Buffer)
	binary.Write(addr, binary.LittleEndian, pubkey.X.Bytes())
	binary.Write(addr, binary.LittleEndian, pubkey.Y.Bytes())
	return addr.Bytes()[:20] //TODO: ADDRESSLENGTH ?
}

func (a *Addresses) addAddress(address []byte) {
	f, err := os.OpenFile(a.addrPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.Write(address); err != nil {
		panic(err)
	}
}

func (a *Addresses) getAccounts() ([]*types.Account, error) {
	b, err := ioutil.ReadFile(a.addrPath)
	if err != nil {
		return nil, err
	}
	var ret []*types.Account
	for i := 0; i < len(b); i += addressLength {
		ret = append(ret, &types.Account{Address: b[i : i+addressLength]})
	}
	return ret, nil
}
