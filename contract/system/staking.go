/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var stakingkey = []byte("staking")

const StakingDelay = 60 * 60 * 24 //block interval

func staking(txBody *types.TxBody, sender, receiver *state.V,
	scs *state.ContractState, blockNo types.BlockNo) (*types.Event, error) {

	staked, err := getStaking(scs, sender.ID())
	if err != nil {
		return nil, err
	}
	beforeStaked := staked.GetAmountBigInt()
	amount := txBody.GetAmountBigInt()
	staked.Amount = new(big.Int).Add(beforeStaked, amount).Bytes()
	staked.When = blockNo
	err = setStaking(scs, sender.ID(), staked)
	if err != nil {
		return nil, err
	}
	sender.SubBalance(amount)

	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       "stake",
		JsonArgs: `{"who":"` +
			types.EncodeAddress(txBody.Account) +
			`", "amount":"` + txBody.GetAmountBigInt().String() + `"}`,
	}, nil
}

func unstaking(txBody *types.TxBody, sender, receiver *state.V, scs *state.ContractState,
	blockNo types.BlockNo, ci *types.CallInfo) (*types.Event, error) {
	staked, err := getStaking(scs, sender.ID())
	if err != nil {
		return nil, err
	}
	amount := txBody.GetAmountBigInt()
	var backToBalance *big.Int
	if staked.GetAmountBigInt().Cmp(amount) < 0 {
		amount = new(big.Int).SetUint64(0)
		backToBalance = staked.GetAmountBigInt()
	} else {
		amount = new(big.Int).Sub(staked.GetAmountBigInt(), txBody.GetAmountBigInt())
		backToBalance = txBody.GetAmountBigInt()
	}
	staked.Amount = amount.Bytes()
	//blockNo will be updated in voting
	staked.When = 0 /*blockNo*/

	err = setStaking(scs, sender.ID(), staked)
	if err != nil {
		return nil, err
	}
	_, err = voting(txBody, sender, receiver, scs, blockNo, ci)
	if err != nil {
		return nil, err
	}
	sender.AddBalance(backToBalance)
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       "unstake",
		JsonArgs: `{"who":"` +
			types.EncodeAddress(txBody.Account) +
			`", "amount":"` + txBody.GetAmountBigInt().String() + `"}`,
	}, nil
}

func setStaking(scs *state.ContractState, who []byte, staking *types.Staking) error {
	key := append(stakingkey, who...)
	return scs.SetData(key, serializeStaking(staking))
}

func getStaking(scs *state.ContractState, who []byte) (*types.Staking, error) {
	key := append(stakingkey, who...)
	data, err := scs.GetData(key)
	if err != nil {
		return nil, err
	}
	var staking types.Staking
	if len(data) != 0 {
		return deserializeStaking(data), nil
	}
	return &staking, nil
}

func GetStaking(scs *state.ContractState, address []byte) (*types.Staking, error) {
	if address != nil {
		return getStaking(scs, address)
	}
	return nil, errors.New("invalid argument: address should not be nil")
}

func serializeStaking(v *types.Staking) []byte {
	var ret []byte
	if v != nil {
		when := make([]byte, 8)
		binary.LittleEndian.PutUint64(when, v.GetWhen())
		ret = append(ret, when...)
		ret = append(ret, v.GetAmount()...)
	}
	return ret
}

func deserializeStaking(data []byte) *types.Staking {
	when := binary.LittleEndian.Uint64(data[:8])
	amount := data[8:]
	return &types.Staking{Amount: amount, When: when}
}
