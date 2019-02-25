package name

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func ExecuteNameTx(bs *state.BlockState, scs *state.ContractState, txBody *types.TxBody, sender, receiver *state.V, blockNo types.BlockNo) error {
	ci, err := ValidateNameTx(txBody, sender, scs)
	if err != nil {
		return err
	}
	switch ci.Name {
	case types.NameCreate:
		err = CreateName(scs, txBody, sender, receiver,
			ci.Args[0].(string))
	case types.NameUpdate:
		err = UpdateName(bs, scs, txBody, sender, receiver,
			ci.Args[0].(string), ci.Args[1].(string))
	}
	if err != nil {
		return err
	}
	return nil
}

func ValidateNameTx(tx *types.TxBody, sender *state.V, scs *state.ContractState) (*types.CallInfo, error) {
	if sender != nil && sender.Balance().Cmp(tx.GetAmountBigInt()) < 0 {
		return nil, types.ErrInsufficientBalance
	}

	var ci types.CallInfo
	if err := json.Unmarshal(tx.Payload, &ci); err != nil {
		return nil, err
	}
	name := ci.Args[0].(string)
	switch ci.Name {
	case types.NameCreate:
		owner := getOwner(scs, []byte(name), false)
		if owner != nil {
			return nil, fmt.Errorf("aleady occupied %s", string(name))
		}
	case types.NameUpdate:
		if (!bytes.Equal(tx.Account, []byte(name))) &&
			(!bytes.Equal(tx.Account, getOwner(scs, []byte(name), false))) {
			return nil, fmt.Errorf("owner not matched : %s", name)
		}
	default:
		return nil, errors.New("could not execute unknown cmd")
	}
	return &ci, nil
}
