/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"errors"
	"math/big"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

var stakingkey = []byte("staking")

const StakingDelay = 60 * 60 * 24 //block interval

func staking(txBody *types.TxBody, sender *state.V,
	scs *state.ContractState, blockNo types.BlockNo) error {

	err := validateForStaking(txBody, scs, blockNo)
	if err != nil {
		return err
	}

	staked, err := getStaking(scs, sender.ID())
	if err != nil {
		return err
	}
	beforeStaked := staked.GetAmountBigInt()
	amount := txBody.GetAmountBigInt()
	staked.Amount = new(big.Int).Add(beforeStaked, amount).Bytes()
	staked.When = blockNo
	err = setStaking(scs, sender.ID(), staked)
	if err != nil {
		return err
	}
	sender.SubBalance(amount)
	return nil
}

func unstaking(txBody *types.TxBody, sender *state.V, scs *state.ContractState, blockNo types.BlockNo) error {
	staked, err := validateForUnstaking(sender.ID(), txBody, scs, blockNo)
	if err != nil {
		return err
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
		return err
	}
	err = voting(txBody, sender, scs, blockNo)
	if err != nil {
		return err
	}
	sender.AddBalance(backToBalance)
	return nil
}

func setStaking(scs *state.ContractState, who []byte, staking *types.Staking) error {
	key := append(stakingkey, who...)
	data, err := proto.Marshal(staking)
	if err != nil {
		return err
	}
	return scs.SetData(key, data)
}

func getStaking(scs *state.ContractState, who []byte) (*types.Staking, error) {
	key := append(stakingkey, who...)
	data, err := scs.GetData(key)
	if err != nil {
		return nil, err
	}
	var staking types.Staking
	if len(data) != 0 {
		err := proto.Unmarshal(data, &staking)
		if err != nil {
			return nil, err
		}
	}
	return &staking, nil
}

func GetStaking(scs *state.ContractState, address []byte) (*types.Staking, error) {
	if address != nil {
		return getStaking(scs, address)
	}
	return nil, errors.New("invalid argument: address should not be nil")
}
