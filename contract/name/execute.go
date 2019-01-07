package name

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func ExecuteNameTx(scs *state.ContractState, txBody *types.TxBody) error {
	nameCmd, err := getNameCmd(txBody.GetPayload())

	switch nameCmd {
	case 'c':
		err = CreateName(scs, txBody)
	case 'u':
		err = UpdateName(scs, txBody)
	default:
		err = errors.New("could not execute unknown cmd")
	}
	if err != nil {
		return err
	}

	return nil
}

func getNameCmd(payload []byte) (byte, error) {
	if len(payload) <= 0 {
		return 0, types.ErrTxFormatInvalid
	}
	return payload[0], nil
}

const allowed = "abcdefghijklmnopqrstuvwxyz1234567890"

func validateAllowedChar(param []byte) error {
	if param == nil {
		return fmt.Errorf("invalid parameter in NameTx")
	}
	for _, char := range string(param) {
		if !strings.Contains(allowed, strings.ToLower(string(char))) {
			return fmt.Errorf("not allowed character in %s", string(param))
		}
	}
	return nil
}

func ValidateNameTx(tx *types.TxBody, scs *state.ContractState) error {
	switch tx.Payload[0] {
	case 'c':
		name := tx.Payload[1:]
		if len(name) > types.NameLength {
			return fmt.Errorf("too long name %s", string(tx.GetPayload()))
		}
		if err := validateAllowedChar(name); err != nil {
			return err
		}
		if len(getAddress(scs, name)) > types.NameLength {
			return fmt.Errorf("aleady occupied %s", string(name))
		}
		if len(name) != types.NameLength {
			return fmt.Errorf("not supported yet")
		}

	case 'u':
		name, to := parseUpdatePayload(tx.Payload[1:])
		if len(name) > types.NameLength {
			return fmt.Errorf("too long name %s", string(tx.GetPayload()))
		}
		if len(name) != types.NameLength {
			return fmt.Errorf("not supported yet")
		}
		if err := validateAllowedChar(name); err != nil {
			return err
		}
		if len(to) > types.AddressLength {
			return fmt.Errorf("too long name %s", string(tx.GetPayload()))
		}
	}
	return nil
}

func parseUpdatePayload(p []byte) ([]byte, []byte) {
	if len(p) <= types.NameLength && p[12] != ',' {
		return nil, nil
	}
	comma := strings.IndexByte(string(p), ',')
	if comma < 0 {
		return nil, nil
	}
	name := p[:comma]
	to := p[comma+1:]
	return []byte(name), []byte(to)
}
