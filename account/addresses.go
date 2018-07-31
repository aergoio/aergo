package account

import (
	"io/ioutil"
	"os"

	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
)

const addressLength = 20

type Addresses struct {
	*log.Logger
	addrPath string
}

func NewAddresses(logger *log.Logger, addressPath string) *Addresses {
	return &Addresses{
		Logger:   logger,
		addrPath: addressPath,
	}
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
