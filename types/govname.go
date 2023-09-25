package types

import (
	"encoding/json"
	"fmt"
)

//governance type transaction which has aergo.system in recipient

const (
	SetContractOwner = "v1setOwner"
	NameCreate       = "v1createName"
	NameUpdate       = "v1updateName"

	TxMaxSize = 200 * 1024
)

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
