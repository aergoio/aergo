package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/anaskhan96/base58check"
)

const AddressLength = 33
const NameLength = 12
const EncodedAddressLength = 52

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

const AddressVersion = 0x42
const PrivKeyVersion = 0xAA

func EncodeAddress(addr Address) string {
	if len(addr) != AddressLength {
		return string(addr)
	}
	encoded, _ := base58check.Encode(fmt.Sprintf("%x", AddressVersion), hex.EncodeToString(addr))
	return encoded
}

const allowed = "abcdefghijklmnopqrstuvwxyz1234567890."

func DecodeAddress(encodedAddr string) (Address, error) {
	if IsSpecialAccount([]byte(encodedAddr)) {
		return []byte(encodedAddr), nil
	} else if len(encodedAddr) <= NameLength  { // name address
		name := encodedAddr
		for _, char := range string(name) {
			if !strings.Contains(allowed, strings.ToLower(string(char))) {
				return nil, fmt.Errorf("not allowed character for address in %s", string(name))
			}
		}
		return []byte(name), nil
	}
	decodedString, err := base58check.Decode(encodedAddr)
	if err != nil {
		return nil, err
	}
	decodedBytes, err := hex.DecodeString(decodedString)
	if err != nil {
		return nil, err
	}
	return DecodeAddressBytes(decodedBytes)
}

func DecodeAddressBytes(decodedBytes []byte) (Address, error) {
	var decoded []byte
	version := decodedBytes[0]
	switch version {
	case AddressVersion:
		decoded = decodedBytes[1:]
		if len(decoded) != AddressLength {
			return nil, errors.New("Invalid address length")
		}
	default:
		return nil, errors.New("Invalid address version")
	}
	return decoded, nil
}

func EncodePrivKey(key []byte) string {
	encoded, _ := base58check.Encode(fmt.Sprintf("%x", PrivKeyVersion), hex.EncodeToString(key))
	return encoded
}

func DecodePrivKey(encodedKey string) ([]byte, error) {
	decodedString, err := base58check.Decode(encodedKey)
	if err != nil {
		return nil, err
	}
	decodedBytes, err := hex.DecodeString(decodedString)
	if err != nil {
		return nil, err
	}
	version := decodedBytes[0]
	if version != PrivKeyVersion {
		return nil, errors.New("Invalid private key version")
	}
	decoded := decodedBytes[1:]
	return decoded, nil
}
