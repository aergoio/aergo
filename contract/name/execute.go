package name

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aergoio/aergo/contract/system"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func ExecuteNameTx(bs *state.BlockState, scs *state.ContractState, txBody *types.TxBody,
	sender, receiver *state.V, blockNo types.BlockNo) ([]*types.Event, error) {

	systemContractState, err := bs.StateDB.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))

	ci, err := ValidateNameTx(txBody, sender, scs, systemContractState)
	if err != nil {
		return nil, err
	}
	var events []*types.Event
	switch ci.Name {
	case types.NameCreate:
		if err = CreateName(scs, txBody, sender, receiver,
			ci.Args[0].(string)); err != nil {
			return nil, err
		}
		events = append(events, &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "create name",
			JsonArgs:        `{"name":"` + ci.Args[0].(string) + `"}`,
		})
	case types.NameUpdate:
		if err = UpdateName(bs, scs, txBody, sender, receiver,
			ci.Args[0].(string), ci.Args[1].(string)); err != nil {
			return nil, err
		}
		events = append(events, &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "update name",
			JsonArgs: `{"name":"` + ci.Args[0].(string) +
				`","to":"` + ci.Args[1].(string) + `"}`,
		})
	}
	return events, nil
}

func ValidateNameTx(tx *types.TxBody, sender *state.V,
	scs, systemcs *state.ContractState) (*types.CallInfo, error) {

	namePrice := system.GetNamePrice(systemcs)
	if namePrice.Cmp(tx.GetAmountBigInt()) > 0 {
		return nil, types.ErrTooSmallAmount
	}

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
