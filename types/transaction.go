package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

//governance type transaction which has aergo.system in recipient
const BPVote = "BPVoteV1"
const NameCreate = "createNameV1"
const NameUpdate = "updateNameV1"

type Transaction interface {
	GetTx() *Tx
	GetBody() *TxBody
	GetHash() []byte
	CalculateTxHash() []byte
	Validate() error
	ValidateWithSenderState(senderState *State, fee *big.Int) error
	HasVerifedAccount() bool
	GetVerifedAccount() Address
	SetVerifedAccount(account Address) bool
	RemoveVerifedAccount() bool
}

type transaction struct {
	Tx              *Tx
	VerifiedAccount Address
}

func NewTransaction(tx *Tx) Transaction {
	return &transaction{Tx: tx}
}

func (tx *transaction) GetTx() *Tx {
	return tx.Tx
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

func (tx *transaction) Validate() error {
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

	price := tx.GetBody().GetPriceBigInt()
	if price.Cmp(MaxAER) > 0 {
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
		if len(tx.GetBody().Payload) <= 0 {
			return ErrTxFormatInvalid
		}
		switch string(tx.GetBody().GetRecipient()) {
		case AergoSystem:
			if (tx.GetBody().GetPayload()[0] == 's' || tx.GetBody().GetPayload()[0] == 'u') &&
				amount.Cmp(StakingMinimum) < 0 {
				return ErrTooSmallAmount
			}
		case AergoName:
			return validateNameTx(tx.GetBody())
		default:
			return ErrTxInvalidRecipient
		}
	default:
		return ErrTxInvalidType
	}
	return nil
}

func validateNameTx(tx *TxBody) error {
	var ci CallInfo
	if err := json.Unmarshal(tx.Payload, &ci); err != nil {
		return ErrTxInvalidPayload
	}
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
	switch ci.Name {
	case NameCreate:
		if len(ci.Args) != 1 {
			return fmt.Errorf("invalid arguments in %s", ci)
		}
	case NameUpdate:
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
	default:
		return ErrTxInvalidPayload
	}
	if new(big.Int).SetUint64(1000000000000000000).Cmp(tx.GetAmountBigInt()) > 0 {
		return ErrTooSmallAmount
	}
	return nil
}

func (tx *transaction) ValidateWithSenderState(senderState *State, fee *big.Int) error {
	if (senderState.GetNonce() + 1) > tx.GetBody().GetNonce() {
		return ErrTxNonceTooLow
	}
	amount := tx.GetBody().GetAmountBigInt()
	balance := senderState.GetBalanceBigInt()
	switch tx.GetBody().GetType() {
	case TxType_NORMAL:
		spending := new(big.Int).Add(amount, fee)
		if spending.Cmp(balance) > 0 {
			return ErrInsufficientBalance
		}
	case TxType_GOVERNANCE:
		switch string(tx.GetBody().GetRecipient()) {
		case AergoSystem:
			if tx.GetBody().GetPayload()[0] == 's' &&
				amount.Cmp(balance) > 0 {
				return ErrInsufficientBalance
			}
		case AergoName:
			return validateNameTxWithSenderState(senderState, tx.GetBody())
		default:
			return ErrTxInvalidRecipient
		}
	}
	if (senderState.GetNonce() + 1) < tx.GetBody().GetNonce() {
		return ErrTxNonceToohigh
	}
	return nil
}

func validateNameTxWithSenderState(s *State, tx *TxBody) error {
	if tx.GetAmountBigInt().Cmp(s.GetBalanceBigInt()) > 0 {
		return ErrInsufficientBalance
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
		Limit:     tx.GetBody().Limit,
		Price:     Clone(tx.GetBody().Price).([]byte),
		Type:      tx.GetBody().Type,
		Sign:      Clone(tx.GetBody().Sign).([]byte),
	}
	res := &transaction{
		Tx: &Tx{Body: body},
	}
	res.Tx.Hash = res.CalculateTxHash()
	return res
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
