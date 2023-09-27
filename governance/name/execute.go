package name

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

func ExecuteNameTx(bs *state.BlockState, scs *state.ContractState, txBody *types.TxBody,
	sender, receiver *state.V, blockInfo *types.BlockHeaderInfo, names *Names, namePrice *big.Int) ([]*types.Event, error) {

	ci, err := ValidateNameTx(txBody, sender, scs, namePrice)
	if err != nil {
		return nil, err
	}

	var events []*types.Event

	var nameState *state.V
	owner := getOwner(scs, []byte(types.AergoName), false)
	if owner != nil {
		if bytes.Equal(sender.ID(), owner) {
			nameState = sender
		} else {
			nameState, err = bs.GetAccountStateV(owner)
			if err != nil {
				return nil, err
			}
		}
	} else {
		nameState = receiver
	}

	switch ci.Name {
	case types.NameCreate:
		name := ci.Args[0].(string)
		amount := txBody.GetAmountBigInt()

		if err = CreateName(names, scs, txBody, sender, nameState, name, amount); err != nil {
			return nil, err
		}

		jsonArgs := ""
		if blockInfo.ForkVersion < 2 {
			jsonArgs = `{"name":"` + name + `"}`
		} else {
			jsonArgs = `["` + name + `"]`
		}
		events = append(events, &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "create name",
			JsonArgs:        jsonArgs,
		})
	case types.NameUpdate:
		amount := txBody.GetAmountBigInt()
		name := ci.Args[0].(string)
		to := ci.Args[1].(string)

		if err = UpdateName(names, bs, scs, txBody, sender, nameState, name, to, amount); err != nil {
			return nil, err
		}

		// return event
		jsonArgs := ""
		if blockInfo.ForkVersion < 2 {
			jsonArgs = `{"name":"` + name +
				`","to":"` + to + `"}`
		} else {
			jsonArgs = `["` + name + `","` + to + `"]`
		}
		events = append(events, &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "update name",
			JsonArgs:        jsonArgs,
		})
	case types.SetContractOwner:
		name := ci.Args[0].(string)
		ownerState, err := SetContractOwner(names, bs, scs, name, nameState)
		if err != nil {
			return nil, err
		}
		ownerState.PutState()
	}

	nameState.PutState()

	return events, nil
}

func ValidateNameTx(tx *types.TxBody, sender *state.V,
	scs *state.ContractState, namePrice *big.Int) (*types.CallInfo, error) {

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
		if namePrice.Cmp(tx.GetAmountBigInt()) > 0 {
			return nil, types.ErrTooSmallAmount
		}
		owner := getOwner(scs, []byte(name), false)
		if owner != nil {
			return nil, fmt.Errorf("aleady occupied %s", string(name))
		}
	case types.NameUpdate:
		if namePrice.Cmp(tx.GetAmountBigInt()) > 0 {
			return nil, types.ErrTooSmallAmount
		}
		if (!bytes.Equal(tx.Account, []byte(name))) &&
			(!bytes.Equal(tx.Account, getOwner(scs, []byte(name), false))) {
			return nil, fmt.Errorf("owner not matched : %s", name)
		}
	case types.SetContractOwner:
		owner := getOwner(scs, []byte(types.AergoName), false)
		if owner != nil {
			return nil, fmt.Errorf("owner aleady set to %s", types.EncodeAddress(owner))
		}
	default:
		return nil, errors.New("could not execute unknown cmd")
	}

	return &ci, nil
}

func SetContractOwner(names *Names, bs *state.BlockState, scs *state.ContractState,
	address string, nameState *state.V) (*state.V, error) {

	name := []byte(types.AergoName)

	rawaddr, err := types.DecodeAddress(address)
	if err != nil {
		return nil, err
	}

	ownerState, err := bs.GetAccountStateV(rawaddr)
	if err != nil {
		return nil, err
	}
	// update balance
	ownerState.AddBalance(nameState.Balance())
	nameState.SubBalance(nameState.Balance())

	// set owner to state
	if err = registerOwner(scs, name, rawaddr, name); err != nil {
		return nil, err
	}
	// set owner to memory
	names.Set(name, rawaddr, name)

	return ownerState, nil
}
