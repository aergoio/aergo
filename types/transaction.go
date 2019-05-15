package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/aergoio/aergo/fee"
	"github.com/gogo/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/mr-tron/base58/base58"
)

//governance type transaction which has aergo.system in recipient

const Stake = "v1stake"
const Unstake = "v1unstake"
const SetContractOwner = "v1setOwner"
const NameCreate = "v1createName"
const NameUpdate = "v1updateName"

const TxMaxSize = 200 * 1024

type Transaction interface {
	GetTx() *Tx
	GetBody() *TxBody
	GetHash() []byte
	CalculateTxHash() []byte
	Validate([]byte, bool) error
	ValidateWithSenderState(senderState *State) error
	HasVerifedAccount() bool
	GetVerifedAccount() Address
	SetVerifedAccount(account Address) bool
	RemoveVerifedAccount() bool
	GetMaxFee() *big.Int
}

type transaction struct {
	Tx              *Tx
	VerifiedAccount Address
}

var _ Transaction = (*transaction)(nil)

func NewTransaction(tx *Tx) Transaction {
	return &transaction{Tx: tx}
}

func (tx *transaction) GetTx() *Tx {
	if tx != nil {
		return tx.Tx
	}
	return nil
}

func (tx *transaction) GetBody() *TxBody {
	return tx.Tx.Body
}

func (tx *transaction) GetHash() []byte {
	return tx.Tx.Hash
}

func (tx *transaction) CalculateTxHash() []byte {
	return tx.Tx.CalculateTxHash()
}

func (tx *transaction) Validate(chainidhash []byte, isPublic bool) error {
	if tx.GetTx() == nil || tx.GetTx().GetBody() == nil {
		return ErrTxFormatInvalid
	}

	if !bytes.Equal(chainidhash, tx.GetTx().GetBody().GetChainIdHash()) {
		return ErrTxInvalidChainIdHash
	}
	if proto.Size(tx.GetTx()) > TxMaxSize {
		return ErrTxInvalidSize
	}

	account := tx.GetBody().GetAccount()
	if account == nil {
		return ErrTxFormatInvalid
	}
	if !bytes.Equal(tx.GetHash(), tx.CalculateTxHash()) {
		return ErrTxHasInvalidHash
	}

	amount := tx.GetBody().GetAmountBigInt()
	if amount.Cmp(MaxAER) > 0 {
		return ErrTxInvalidAmount
	}

	gasprice := tx.GetBody().GetGasPriceBigInt()
	if gasprice.Cmp(MaxAER) > 0 {
		return ErrTxInvalidPrice
	}

	if len(tx.GetBody().GetAccount()) > AddressLength {
		return ErrTxInvalidAccount
	}

	if len(tx.GetBody().GetRecipient()) > AddressLength {
		return ErrTxInvalidRecipient
	}

	switch tx.GetBody().Type {
	case TxType_NORMAL:
		if tx.GetBody().GetRecipient() == nil && len(tx.GetBody().GetPayload()) == 0 {
			//contract deploy
			return ErrTxInvalidRecipient
		}
	case TxType_GOVERNANCE:
		if len(tx.GetBody().GetPayload()) <= 0 {
			return ErrTxFormatInvalid
		}
		switch string(tx.GetBody().GetRecipient()) {
		case AergoSystem:
			return ValidateSystemTx(tx.GetBody())
		case AergoName:
			return validateNameTx(tx.GetBody())
		case AergoEnterprise:
			if isPublic {
				return ErrTxInvalidRecipient
			}
		default:
			return ErrTxInvalidRecipient
		}
	default:
		return ErrTxInvalidType
	}
	return nil
}

func ValidateSystemTx(tx *TxBody) error {
	var ci CallInfo
	if err := json.Unmarshal(tx.Payload, &ci); err != nil {
		return ErrTxInvalidPayload
	}
	switch ci.Name {
	case Stake,
		Unstake:
	case VoteBP:
		unique := map[string]int{}
		for i, v := range ci.Args {
			if i >= MaxCandidates {
				return ErrTxInvalidPayload
			}
			encoded, ok := v.(string)
			if !ok {
				return ErrTxInvalidPayload
			}
			if unique[encoded] != 0 {
				return ErrTxInvalidPayload
			}
			unique[encoded]++
			candidate, err := base58.Decode(encoded)
			if err != nil {
				return ErrTxInvalidPayload
			}
			_, err = peer.IDFromBytes(candidate)
			if err != nil {
				return ErrTxInvalidPayload
			}
		}
		/* TODO: will be changed
		case VoteNumBP,
			VoteGasPrice,
			VoteNamePrice,
			VoteMinStaking:
			for i, v := range ci.Args {
				if i > 1 {
					return ErrTxInvalidPayload
				}
				vstr, ok := v.(string)
				if !ok {
					return ErrTxInvalidPayload
				}
				if _, ok := new(big.Int).SetString(vstr, 10); !ok {
					return ErrTxInvalidPayload
				}
			}
		*/
	default:
		return ErrTxInvalidPayload
	}
	return nil
}

func validateNameTx(tx *TxBody) error {
	var ci CallInfo
	if err := json.Unmarshal(tx.Payload, &ci); err != nil {
		return ErrTxInvalidPayload
	}
	switch ci.Name {
	case NameCreate:
		if err := _validateNameTx(tx, &ci); err != nil {
			return err
		}
		if len(ci.Args) != 1 {
			return fmt.Errorf("invalid arguments in %s", ci)
		}
	case NameUpdate:
		if err := _validateNameTx(tx, &ci); err != nil {
			return err
		}
		if len(ci.Args) != 2 {
			return fmt.Errorf("invalid arguments in %s", ci)
		}
		to, err := DecodeAddress(ci.Args[1].(string))
		if err != nil {
			return fmt.Errorf("invalid receiver in %s", ci)
		}
		if len(to) > AddressLength {
			return fmt.Errorf("too long name %s", string(tx.GetPayload()))
		}
	case SetContractOwner:
		owner, ok := ci.Args[0].(string)
		if !ok {
			return fmt.Errorf("invalid arguments in %s", owner)
		}
		_, err := DecodeAddress(owner)
		if err != nil {
			return fmt.Errorf("invalid new owner %s", err.Error())
		}
	default:
		return ErrTxInvalidPayload
	}
	return nil
}

func _validateNameTx(tx *TxBody, ci *CallInfo) error {
	if len(ci.Args) < 1 {
		return fmt.Errorf("invalid arguments in %s", ci)
	}
	nameParam, ok := ci.Args[0].(string)
	if !ok {
		return fmt.Errorf("invalid arguments in %s", nameParam)
	}

	if len(nameParam) > NameLength {
		return fmt.Errorf("too long name %s", string(tx.GetPayload()))
	}
	if len(nameParam) != NameLength {
		return fmt.Errorf("not supported yet")
	}
	if err := validateAllowedChar([]byte(nameParam)); err != nil {
		return err
	}
	if new(big.Int).SetUint64(1000000000000000000).Cmp(tx.GetAmountBigInt()) > 0 {
		return ErrTooSmallAmount
	}
	return nil

}

func (tx *transaction) ValidateWithSenderState(senderState *State) error {
	if (senderState.GetNonce() + 1) > tx.GetBody().GetNonce() {
		return ErrTxNonceTooLow
	}
	amount := tx.GetBody().GetAmountBigInt()
	balance := senderState.GetBalanceBigInt()
	switch tx.GetBody().GetType() {
	case TxType_NORMAL:
		spending := new(big.Int).Add(amount, tx.GetMaxFee())
		if spending.Cmp(balance) > 0 {
			return ErrInsufficientBalance
		}
	case TxType_GOVERNANCE:
		switch string(tx.GetBody().GetRecipient()) {
		case AergoSystem:
			var ci CallInfo
			if err := json.Unmarshal(tx.GetBody().GetPayload(), &ci); err != nil {
				return ErrTxInvalidPayload
			}
			if ci.Name == Stake &&
				amount.Cmp(balance) > 0 {
				return ErrInsufficientBalance
			}
		case AergoName:
		case AergoEnterprise:
		default:
			return ErrTxInvalidRecipient
		}
	}
	if (senderState.GetNonce() + 1) < tx.GetBody().GetNonce() {
		return ErrTxNonceToohigh
	}
	return nil
}

//TODO : refoctor after ContractState move to types
func (tx *Tx) ValidateWithContractState(contractState *State) error {
	//in system.ValidateSystemTx
	//in name.ValidateNameTx
	return nil
}

func (tx *transaction) GetVerifedAccount() Address {
	return tx.VerifiedAccount
}

func (tx *transaction) SetVerifedAccount(account Address) bool {
	tx.VerifiedAccount = account
	return true
}

func (tx *transaction) HasVerifedAccount() bool {
	return len(tx.VerifiedAccount) != 0
}

func (tx *transaction) RemoveVerifedAccount() bool {
	return tx.SetVerifedAccount(nil)
}

func (tx *transaction) Clone() *transaction {
	if tx == nil {
		return nil
	}
	if tx.GetBody() == nil {
		return &transaction{}
	}
	body := &TxBody{
		Nonce:     tx.GetBody().Nonce,
		Account:   Clone(tx.GetBody().Account).([]byte),
		Recipient: Clone(tx.GetBody().Recipient).([]byte),
		Amount:    Clone(tx.GetBody().Amount).([]byte),
		Payload:   Clone(tx.GetBody().Payload).([]byte),
		GasLimit:  tx.GetBody().GasLimit,
		GasPrice:  Clone(tx.GetBody().GasPrice).([]byte),
		Type:      tx.GetBody().Type,
		Sign:      Clone(tx.GetBody().Sign).([]byte),
	}
	res := &transaction{
		Tx: &Tx{Body: body},
	}
	res.Tx.Hash = res.CalculateTxHash()
	return res
}

func (tx *transaction) GetMaxFee() *big.Int {
	return fee.MaxPayloadTxFee(len(tx.GetBody().GetPayload()))
}

const allowedNameChar = "abcdefghijklmnopqrstuvwxyz1234567890"

func validateAllowedChar(param []byte) error {
	if param == nil {
		return fmt.Errorf("invalid parameter in NameTx")
	}
	for _, char := range string(param) {
		if !strings.Contains(allowedNameChar, strings.ToLower(string(char))) {
			return fmt.Errorf("not allowed character in %s", string(param))
		}
	}
	return nil
}
