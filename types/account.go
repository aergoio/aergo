package types

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/aergoio/aergo/v2/internal/enc/base58check"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/ethereum/go-ethereum/common"
)

const AddressLength = 33
const NameLength = 12
const EncodedAddressLength = 52

// NewAccount alloc new account object
func NewAccount(addr []byte) *Account {
	return &Account{
		Address: addr,
	}
}

// ToAddress return byte array of given base58check encoded address string
func ToAddress(addr string) []byte {
	ret, err := DecodeAddress(addr)
	if err != nil {
		return nil
	}
	return ret
}

// ToString return base58check encoded string of address
func (a *Account) ToString() string {
	return EncodeAddress(a.Address)
}

// NewAccountList alloc new account list
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
	encoded, _ := base58check.Encode(fmt.Sprintf("%x", AddressVersion), hex.Encode(addr))
	return encoded
}

const allowed = "abcdefghijklmnopqrstuvwxyz1234567890."

func DecodeAddress(encodedAddr string) (Address, error) {
	if IsSpecialAccount([]byte(encodedAddr)) {
		return []byte(encodedAddr), nil
	} else if len(encodedAddr) <= NameLength { // name address
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
	decodedBytes, err := hex.Decode(decodedString)
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
	encoded, _ := base58check.Encode(fmt.Sprintf("%x", PrivKeyVersion), hex.Encode(key))
	return encoded
}

func DecodePrivKey(encodedKey string) ([]byte, error) {
	decodedString, err := base58check.Decode(encodedKey)
	if err != nil {
		return nil, err
	}
	decodedBytes, err := hex.Decode(decodedString)
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

//------------------------------------------------------------------------------------------//
// special accounts

const (
	AergoSystem     = "aergo.system"
	AergoName       = "aergo.name"
	AergoEnterprise = "aergo.enterprise"
	AergoVault      = "aergo.vault" // For community reward program (i.e. voting reward)

	MaxCandidates = 30
)

// too few accounts to use map
var (
	specialAccounts          [][]byte
	specialAccountEth        map[string]common.Address
	specialAccountEthReverse map[common.Address]string
)

func init() {
	specialAccounts = make([][]byte, 0, 4)
	specialAccounts = append(specialAccounts, []byte(AergoSystem))
	specialAccounts = append(specialAccounts, []byte(AergoName))
	specialAccounts = append(specialAccounts, []byte(AergoEnterprise))
	specialAccounts = append(specialAccounts, []byte(AergoVault))

	specialAccountEth = make(map[string]common.Address)
	specialAccountEth[AergoSystem] = common.BigToAddress(big.NewInt(1))     // 0x0000000000000000000000000000000000000001
	specialAccountEth[AergoName] = common.BigToAddress(big.NewInt(2))       // 0x0000000000000000000000000000000000000002
	specialAccountEth[AergoEnterprise] = common.BigToAddress(big.NewInt(3)) // 0x0000000000000000000000000000000000000003
	specialAccountEth[AergoVault] = common.BigToAddress(big.NewInt(4))      // 0x0000000000000000000000000000000000000004

	specialAccountEthReverse = make(map[common.Address]string)
	specialAccountEthReverse[specialAccountEth[AergoSystem]] = AergoSystem
	specialAccountEthReverse[specialAccountEth[AergoName]] = AergoName
	specialAccountEthReverse[specialAccountEth[AergoEnterprise]] = AergoEnterprise
	specialAccountEthReverse[specialAccountEth[AergoVault]] = AergoVault

}

func GetSpecialAccountEth(name []byte) common.Address {
	return specialAccountEth[string(name)]
}

func GetSpecialAccountEthReverse(addr common.Address) string {
	return specialAccountEthReverse[addr]
}

// IsSpecialAccount check if name is the one of special account names.
func IsSpecialAccount(name []byte) bool {
	for _, b := range specialAccounts {
		if bytes.Compare(name, b) == 0 {
			return true
		}
	}
	return false
}
