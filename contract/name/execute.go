package name

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
)

func ExecuteNameTx(bs *state.BlockState, scs *statedb.ContractState, txBody *types.TxBody,
	sender, receiver *state.AccountState, blockInfo *types.BlockHeaderInfo) ([]*types.Event, error) {

	ci, err := ValidateNameTx(txBody, sender, scs)
	if err != nil {
		return nil, err
	}

	var nameState *state.AccountState
	owner := getOwner(scs, []byte(types.AergoName), false)
	if owner != nil {
		if bytes.Equal(sender.ID(), owner) {
			nameState = sender
		} else {
			if nameState, err = state.GetAccountState(owner, bs.LuaStateDB, bs.EthStateDB); err != nil {
				return nil, err
			}
		}
	} else {
		nameState = receiver
	}

	var events []*types.Event
	switch ci.Name {
	case types.NameCreate:
		nameArg := ci.Args[0].(string)
		if err = CreateName(scs, txBody, sender, nameState, nameArg); err != nil {
			return nil, err
		}
		jsonArgs := ""
		if blockInfo.ForkVersion < 2 {
			jsonArgs = `{"name":"` + nameArg + `"}`
		} else {
			jsonArgs = `["` + nameArg + `"]`
		}
		events = append(events, &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "create name",
			JsonArgs:        jsonArgs,
		})
	case types.NameUpdate:
		nameArg := ci.Args[0].(string)
		toArg := ci.Args[1].(string)
		if err = UpdateName(bs, scs, txBody, sender, nameState, nameArg, toArg); err != nil {
			return nil, err
		}
		jsonArgs := ""
		if blockInfo.ForkVersion < 2 {
			jsonArgs = `{"name":"` + nameArg + `","to":"` + toArg + `"}`
		} else {
			jsonArgs = `["` + nameArg + `","` + toArg + `"]`
		}
		events = append(events, &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "update name",
			JsonArgs:        jsonArgs,
		})
	case types.SetContractOwner:
		ownerArg := ci.Args[0].(string)
		ownerState, err := SetContractOwner(bs, scs, ownerArg, nameState)
		if err != nil {
			return nil, err
		}
		ownerState.PutState()
	}

	nameState.PutState()

	return events, nil
}

func ValidateNameTx(tx *types.TxBody, sender *state.AccountState, scs *statedb.ContractState) (*types.CallInfo, error) {
	if sender != nil && sender.Balance().Cmp(tx.GetAmountBigInt()) < 0 {
		return nil, types.ErrInsufficientBalance
	}

	var ci types.CallInfo
	if err := json.Unmarshal(tx.Payload, &ci); err != nil {
		return nil, err
	}

	nameArg := ci.Args[0].(string)
	switch ci.Name {
	case types.NameCreate:
		if system.GetNamePrice().Cmp(tx.GetAmountBigInt()) > 0 {
			return nil, types.ErrTooSmallAmount
		}
		if owner := getOwner(scs, []byte(nameArg), false); owner != nil {
			return nil, fmt.Errorf("aleady occupied %s", string(nameArg))
		}
	case types.NameUpdate:
		if system.GetNamePrice().Cmp(tx.GetAmountBigInt()) > 0 {
			return nil, types.ErrTooSmallAmount
		}
		if (!bytes.Equal(tx.Account, []byte(nameArg))) &&
			(!bytes.Equal(tx.Account, getOwner(scs, []byte(nameArg), false))) {
			return nil, fmt.Errorf("owner not matched : %s", nameArg)
		}
	case types.SetContractOwner:
		if owner := getOwner(scs, []byte(types.AergoName), false); owner != nil {
			return nil, fmt.Errorf("owner aleady set to %s", types.EncodeAddress(owner))
		}
	default:
		return nil, errors.New("could not execute unknown cmd")
	}

	return &ci, nil
}

func SetContractOwner(bs *state.BlockState, scs *statedb.ContractState,
	address string, nameState *state.AccountState) (*state.AccountState, error) {

	rawaddr, err := types.DecodeAddress(address)
	if err != nil {
		return nil, err
	}

	ownerState, err := state.GetAccountState(rawaddr, bs.LuaStateDB, bs.EthStateDB)
	if err != nil {
		return nil, err
	}

	if err = state.SendBalance(nameState, ownerState, nameState.Balance()); err != nil {
		return nil, err
	}

	name := []byte(types.AergoName)
	if err = registerOwner(scs, name, rawaddr, name); err != nil {
		return nil, err
	}

	return ownerState, nil
}
