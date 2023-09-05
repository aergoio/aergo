package name

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

func ExecuteNameTx(bs *state.BlockState, scs *state.ContractState, txBody *types.TxBody,
	sender, receiver *state.V, blockInfo *types.BlockHeaderInfo) ([]*types.Event, error) {

	systemContractState, err := bs.StateDB.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))

	ci, err := ValidateNameTx(txBody, sender, scs, systemContractState)
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
		if err = CreateName(scs, txBody, sender, nameState,
			ci.Args[0].(string)); err != nil {
			return nil, err
		}
		jsonArgs := ""
		if blockInfo.ForkVersion < 2 {
			jsonArgs = `{"name":"` + ci.Args[0].(string) + `"}`
		} else {
			jsonArgs = `["` + ci.Args[0].(string) + `"]`
		}
		events = append(events, &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "create name",
			JsonArgs:        jsonArgs,
		})
	case types.NameUpdate:
		if err = UpdateName(bs, scs, txBody, sender, nameState,
			ci.Args[0].(string), ci.Args[1].(string)); err != nil {
			return nil, err
		}
		jsonArgs := ""
		if blockInfo.ForkVersion < 2 {
			jsonArgs = `{"name":"` + ci.Args[0].(string) +
				`","to":"` + ci.Args[1].(string) + `"}`
		} else {
			jsonArgs = `["` + ci.Args[0].(string) + `","` + ci.Args[1].(string) + `"]`
		}
		events = append(events, &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "update name",
			JsonArgs:        jsonArgs,
		})
	case types.SetContractOwner:
		ownerState, err := SetContractOwner(bs, scs, ci.Args[0].(string), nameState)
		if err != nil {
			return nil, err
		}
		ownerState.PutState()
	}

	nameState.PutState()

	return events, nil
}

func ValidateNameTx(tx *types.TxBody, sender *state.V,
	scs, systemcs *state.ContractState) (*types.CallInfo, error) {

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
		namePrice := system.GetNamePriceFromState(systemcs)
		if namePrice.Cmp(tx.GetAmountBigInt()) > 0 {
			return nil, types.ErrTooSmallAmount
		}
		owner := getOwner(scs, []byte(name), false)
		if owner != nil {
			return nil, fmt.Errorf("aleady occupied %s", string(name))
		}
	case types.NameUpdate:
		namePrice := system.GetNamePriceFromState(systemcs)
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

func SetContractOwner(bs *state.BlockState, scs *state.ContractState,
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

	ownerState.AddBalance(nameState.Balance())
	nameState.SubBalance(nameState.Balance())

	if err = registerOwner(scs, name, rawaddr, name); err != nil {
		return nil, err
	}

	return ownerState, nil
}
