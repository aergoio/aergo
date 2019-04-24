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

var stakingKey = []byte("staking")
var stakingTotalKey = []byte("stakingtotal")

const StakingDelay = 60 * 60 * 24 //block interval
//const StakingDelay = 5

func staking(txBody *types.TxBody, sender, receiver *state.V,
	scs *state.ContractState, blockNo types.BlockNo, context *SystemContext) (*types.Event, error) {
	staked := context.Staked
	beforeStaked := staked.GetAmountBigInt()
	amount := txBody.GetAmountBigInt()
	staked.Amount = new(big.Int).Add(beforeStaked, amount).Bytes()
	staked.When = blockNo
	if err := setStaking(scs, sender.ID(), staked); err != nil {
		return nil, err
	}
	if err := addTotal(scs, amount); err != nil {
		return nil, err
	}
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       "stake",
		JsonArgs: `{"who":"` +
			types.EncodeAddress(sender.ID()) +
			`", "amount":"` + txBody.GetAmountBigInt().String() + `"}`,
	}, nil
}

func unstaking(txBody *types.TxBody, sender, receiver *state.V, scs *state.ContractState,
	blockNo types.BlockNo, context *SystemContext) (*types.Event, error) {
	staked := context.Staked
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
	staked.When = blockNo

	if err := setStaking(scs, sender.ID(), staked); err != nil {
		return nil, err
	}
	if err := refreshAllVote(txBody, scs, context); err != nil {
		return nil, err
	}
	if err := subTotal(scs, backToBalance); err != nil {
		return nil, err
	}
	sender.AddBalance(backToBalance)
	receiver.SubBalance(backToBalance)
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       "unstake",
		JsonArgs: `{"who":"` +
			types.EncodeAddress(sender.ID()) +
			`", "amount":"` + txBody.GetAmountBigInt().String() + `"}`,
	}, nil
}

func setStaking(scs *state.ContractState, who []byte, staking *types.Staking) error {
	key := append(stakingKey, who...)
	return scs.SetData(key, serializeStaking(staking))
}

func getStaking(scs *state.ContractState, who []byte) (*types.Staking, error) {
	key := append(stakingKey, who...)
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

func getTotal(ar AccountStateReader) (*big.Int, error) {
	scs, err := ar.GetSystemAccountState()
	if err != nil {
		return nil, err
	}
	return getStakingTotal(scs)
}

func getStakingTotal(scs *state.ContractState) (*big.Int, error) {
	data, err := scs.GetData(stakingTotalKey)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(data), nil
}

func addTotal(scs *state.ContractState, amount *big.Int) error {
	data, err := scs.GetData(stakingTotalKey)
	if err != nil {
		return err
	}
	total := new(big.Int).SetBytes(data)
	return scs.SetData(stakingTotalKey, new(big.Int).Add(total, amount).Bytes())
}

func subTotal(scs *state.ContractState, amount *big.Int) error {
	data, err := scs.GetData(stakingTotalKey)
	if err != nil {
		return err
	}
	total := new(big.Int).SetBytes(data)
	return scs.SetData(stakingTotalKey, new(big.Int).Sub(total, amount).Bytes())
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
