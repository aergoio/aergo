/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"bytes"
	"encoding/gob"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var stakingkey = []byte("staking")

const StakingDelay = 10

func staking(txBody *types.TxBody, senderState *types.State,
	scs *state.ContractState, blockNo types.BlockNo) error {

	err := validateForStaking(txBody, scs, blockNo)
	if err != nil {
		return err
	}

	staked, err := getStaking(scs, txBody.Account)
	if err != nil {
		return err
	}
	staked.Amount += txBody.Amount
	staked.When = blockNo
	err = setStaking(scs, txBody.Account, staked)
	if err != nil {
		return err
	}

	senderState.Balance = senderState.Balance - txBody.Amount
	return nil
}

func unstaking(txBody *types.TxBody, senderState *types.State, scs *state.ContractState, blockNo types.BlockNo) error {
	staked, err := validateForUnstaking(txBody, scs, blockNo)
	if err != nil {
		return err
	}
	amount := txBody.Amount
	var backToBalance uint64
	if staked.GetAmount() < amount {
		amount = 0
		backToBalance = staked.GetAmount()
	} else {
		amount = staked.GetAmount() - txBody.Amount
		backToBalance = txBody.Amount
	}
	staked.Amount = amount
	//blockNo will be updated in voting
	staked.When = 0 /*blockNo*/

	err = setStaking(scs, txBody.Account, staked)
	if err != nil {
		return err
	}
	err = voting(txBody, scs, blockNo)
	if err != nil {
		return err
	}

	senderState.Balance = senderState.Balance + backToBalance
	return nil
}

func setStaking(scs *state.ContractState, who []byte, staking *types.Staking) error {
	key := append(stakingkey, who...)
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err := enc.Encode(staking)
	if err != nil {
		return err
	}
	return scs.SetData(key, data.Bytes())
}

func getStaking(scs *state.ContractState, who []byte) (*types.Staking, error) {
	key := append(stakingkey, who...)
	data, err := scs.GetData(key)
	if err != nil {
		return nil, err
	}
	var staking types.Staking
	if len(data) != 0 {
		dec := gob.NewDecoder(bytes.NewBuffer(data))
		err = dec.Decode(&staking)
		if err != nil {
			return nil, err
		}
	}
	return &staking, nil
}
