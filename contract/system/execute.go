/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"encoding/json"
	"math/big"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func ExecuteSystemTx(scs *state.ContractState, txBody *types.TxBody, sender *state.V,
	blockNo types.BlockNo) error {

	ci, err := ValidateSystemTx(sender.ID(), txBody, sender, scs, blockNo)
	if err != nil {
		return err
	}

	switch ci.Name {
	case types.Stake:
		err = staking(txBody, sender, scs, blockNo)
	case types.VoteBP:
		err = voting(txBody, sender, scs, blockNo, ci)
	case types.Unstake:
		err = unstaking(txBody, sender, scs, blockNo, ci)
	}
	if err != nil {
		return err
	}

	return nil
}

func ValidateSystemTx(account []byte, txBody *types.TxBody, sender *state.V,
	scs *state.ContractState, blockNo uint64) (*types.CallInfo, error) {
	var ci types.CallInfo
	if err := json.Unmarshal(txBody.Payload, &ci); err != nil {
		return nil, types.ErrTxInvalidPayload
	}
	var err error
	switch ci.Name {
	case types.Stake:
		if sender != nil && sender.Balance().Cmp(txBody.GetAmountBigInt()) < 0 {
			return nil, types.ErrInsufficientBalance
		}
	case types.VoteBP:
		staked, err := getStaking(scs, account)
		if err != nil {
			return nil, err
		}
		if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
			return nil, types.ErrMustStakeBeforeVote
		}
		oldvote, err := getVote(scs, account)
		if err != nil {
			return nil, err
		}
		if oldvote.Amount != nil && staked.GetWhen()+VotingDelay > blockNo {
			return nil, types.ErrLessTimeHasPassed
		}
	case types.Unstake:
		_, err = validateForUnstaking(account, txBody, scs, blockNo)
	}
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

func validateForUnstaking(account []byte, txBody *types.TxBody, scs *state.ContractState, blockNo uint64) (*types.Staking, error) {
	staked, err := getStaking(scs, account)
	if err != nil {
		return nil, err
	}
	if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
		return nil, types.ErrMustStakeBeforeUnstake
	}
	if staked.GetWhen()+StakingDelay > blockNo {
		return nil, types.ErrLessTimeHasPassed
	}
	return staked, nil
}
