/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
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
		ret = append(ret, v.GetAmount()...)
		ret = append(ret, vsep)
		when := make([]byte, 8)
		binary.LittleEndian.PutUint64(when, v.GetWhen())
		ret = append(ret, when...)
	}
	return ret
}

func deserializeStaking(data []byte) *types.Staking {
	datas := bytes.Split(data, []byte{vsep})
	if len(datas[1]) != 8 {
		panic("staking data corruption")
	}
	when := binary.LittleEndian.Uint64(datas[1])
	return &types.Staking{Amount: datas[0], When: when}
}
