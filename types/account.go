package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/anaskhan96/base58check"
)

//NewAccount alloc new account object
func NewAccount(addr []byte) *Account {
	return &Account{
		Address: addr,
	}
}

//ToAddress return byte array of given base58check encoded address string
func ToAddress(addr string) []byte {
	ret, err := DecodeAddress(addr)
	if err != nil {
		return nil
	}
	return ret
}

//ToString return base58check encoded string of address
func (a *Account) ToString() string {
	return EncodeAddress(a.Address)
}

//NewAccountList alloc new account list
func NewAccountList(accounts []*Account) *AccountList {
	return &AccountList{
		Accounts: accounts,
	}
}

type Address = []byte
const AddressVersion = 0x17

func EncodeAddress(addr Address) (string) {
	encoded, _ := base58check.Encode(fmt.Sprintf("%x", AddressVersion), hex.EncodeToString(addr))
	return encoded
}

func DecodeAddress(encodedAddr string) (Address, error) {
	decodedString, err := base58check.Decode(encodedAddr)
	if err != nil {
		return nil, err
	}
	decodedBytes, err := hex.DecodeString(decodedString)
	if err != nil {
		return nil, err
	}
	version := decodedBytes[0]
	if version != AddressVersion {
		return nil, errors.New("Invalid address version")
	}
	decoded := decodedBytes[1:]
	return decoded, nil
}
