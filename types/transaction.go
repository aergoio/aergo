package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/aergoio/aergo/v2/fee"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/proto"
)

//governance type transaction which has aergo.system in recipient

const SetContractOwner = "v1setOwner"
const NameCreate = "v1createName"
const NameUpdate = "v1updateName"

const TxMaxSize = 200 * 1024

type validator func(tx *TxBody) error

var govValidators map[string]validator

func InitGovernance(consensus string, isPublic bool) {
	sysValidator := ValidateSystemTx
	if consensus != "dpos" {
		sysValidator = func(tx *TxBody) error {
			return ErrTxInvalidType
		}
	}

	govValidators = map[string]validator{
		AergoSystem: sysValidator,
		AergoName:   validateNameTx,
		AergoEnterprise: func(tx *TxBody) error {
			if isPublic {
				return ErrTxOnlySupportedInPriv
			}
			return nil
		},
	}
}

type Transaction interface {
	GetTx() *Tx
	GetBody() *TxBody
	GetHash() []byte
	CalculateTxHash() []byte
	Validate([]byte, bool) error
	ValidateWithSenderState(senderState *State, gasPrice *big.Int, version int32) error
	ValidateMaxFee(balance, gasPrice *big.Int, version int32) error
	HasVerifedAccount() bool
	GetVerifedAccount() Address
	SetVerifedAccount(account Address) bool
	RemoveVerifedAccount() bool
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
	case TxType_REDEPLOY:
		if isPublic {
			return ErrTxInvalidType
		}
		if tx.GetBody().GetRecipient() == nil {
			return ErrTxInvalidRecipient
		}
		fallthrough
	case TxType_NORMAL:
		if tx.GetBody().GetRecipient() == nil && len(tx.GetBody().GetPayload()) == 0 {
			//contract deploy
			return ErrTxInvalidRecipient
		}
	case TxType_GOVERNANCE:
		if len(tx.GetBody().GetPayload()) <= 0 {
			return ErrTxFormatInvalid
		}

		if err := validate(tx.GetBody()); err != nil {
			return err
		}
	case TxType_FEEDELEGATION:
		if tx.GetBody().GetRecipient() == nil {
			return ErrTxInvalidRecipient
		}
		if len(tx.GetBody().GetPayload()) <= 0 {
			return ErrTxFormatInvalid
		}
	case TxType_TRANSFER, TxType_CALL:
		if tx.GetBody().GetRecipient() == nil {
			return ErrTxInvalidRecipient
		}
	case TxType_DEPLOY:
		if tx.GetBody().GetRecipient() != nil {
			return ErrTxInvalidRecipient
		}
		if len(tx.GetBody().GetPayload()) == 0 {
			return ErrTxFormatInvalid
		}
	default:
		return ErrTxInvalidType
	}
	return nil
}

func validate(tx *TxBody) error {
	if val, exist := govValidators[string(tx.GetRecipient())]; exist {
		return val(tx)
	}

	return ErrTxInvalidRecipient
}

func ValidateSystemTx(tx *TxBody) error {
	var ci CallInfo
	if err := json.Unmarshal(tx.Payload, &ci); err != nil {
		return ErrTxInvalidPayload
	}
	op := GetOpSysTx(ci.Name)
	switch op {
	case Opstake,
		Opunstake:
	case OpvoteBP:
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
			_, err = IDFromBytes(candidate)
			if err != nil {
				return ErrTxInvalidPayload
			}
		}
	case OpvoteDAO:
		if len(ci.Args) < 1 {
			return fmt.Errorf("the number of args less then 1")
		}
		unique := map[string]int{}
		for _, v := range ci.Args {
			encoded, ok := v.(string)
			if !ok {
				return ErrTxInvalidPayload
			}
			if unique[encoded] != 0 {
				return ErrTxInvalidPayload
			}
			unique[encoded]++
		}
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
	return nil
}

func (tx *transaction) ValidateWithSenderState(senderState *State, gasPrice *big.Int, version int32) error {
	if (senderState.GetNonce() + 1) > tx.GetBody().GetNonce() {
		return ErrTxNonceTooLow
	}
	amount := tx.GetBody().GetAmountBigInt()
	balance := senderState.GetBalanceBigInt()
	switch tx.GetBody().GetType() {
	case TxType_NORMAL, TxType_REDEPLOY, TxType_TRANSFER, TxType_CALL, TxType_DEPLOY:
		b := new(big.Int).Sub(balance, amount)
		if b.Sign() < 0 {
			return ErrInsufficientBalance
		}
		err := tx.ValidateMaxFee(b, gasPrice, version)
		if err != nil {
			return err
		}
	case TxType_GOVERNANCE:
		switch string(tx.GetBody().GetRecipient()) {
		case AergoSystem:
			var ci CallInfo
			if err := json.Unmarshal(tx.GetBody().GetPayload(), &ci); err != nil {
				return ErrTxInvalidPayload
			}
			if ci.Name == Opstake.Cmd() &&
				amount.Cmp(balance) > 0 {
				return ErrInsufficientBalance
			}
		case AergoName:
		case AergoEnterprise:
		default:
			return ErrTxInvalidRecipient
		}
	case TxType_FEEDELEGATION:
		if amount.Cmp(balance) > 0 {
			return ErrInsufficientBalance
		}
	}
	if (senderState.GetNonce() + 1) < tx.GetBody().GetNonce() {
		return ErrTxNonceToohigh
	}
	return nil
}

// TODO : refoctor after ContractState move to types
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
		Account:   tx.GetBody().Account,
		Recipient: tx.GetBody().Recipient,
		Amount:    tx.GetBody().Amount,
		Payload:   tx.GetBody().Payload,
		GasLimit:  tx.GetBody().GasLimit,
		GasPrice:  tx.GetBody().GasPrice,
		Type:      tx.GetBody().Type,
		Sign:      tx.GetBody().Sign,
	}
	res := &transaction{
		Tx: &Tx{Body: body},
	}
	res.Tx.Hash = res.CalculateTxHash()
	return res
}

func (tx *transaction) ValidateMaxFee(balance, gasPrice *big.Int, version int32) error {
	var (
		lenPayload = len(tx.GetBody().GetPayload())
		gasLimit   = tx.GetBody().GetGasLimit()
	)

	maxFee, err := fee.TxMaxFee(version, lenPayload, gasLimit, balance, gasPrice)
	if err != nil {
		return err
	} else if maxFee.Cmp(balance) > 0 {
		return ErrInsufficientBalance
	}
	return nil
}

const allowedNameChar = "abcdefghijklmnopqrstuvwxyz1234567890"

func validateAllowedChar(param []byte) error {
	if param == nil {
		return fmt.Errorf("not allowed character : nil")
	}
	for _, char := range string(param) {
		if !strings.Contains(allowedNameChar, strings.ToLower(string(char))) {
			return fmt.Errorf("not allowed character in %s", string(param))
		}
	}
	return nil
}
