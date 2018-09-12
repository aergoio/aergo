package types

import "github.com/mr-tron/base58/base58"

//NewAccount alloc new account object
func NewAccount(addr []byte) *Account {
	return &Account{
		Address: addr,
	}
}

//ToAddress return byte array of given address string
func ToAddress(addr string) []byte {
	ret, err := base58.Decode(addr)
	if err != nil {
		return nil
	}
	return ret
}

//ToString return base64 encoded string of address
func (a *Account) ToString() string {
	return base58.Encode(a.Address)
}

//NewAccountList alloc new account list
func NewAccountList(accounts []*Account) *AccountList {
	return &AccountList{
		Accounts: accounts,
	}
}
