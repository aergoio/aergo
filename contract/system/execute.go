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

func ExecuteSystemTx(scs *state.ContractState, txBody *types.TxBody,
	sender, receiver *state.V, blockNo types.BlockNo) ([]*types.Event, error) {

	ci, err := ValidateSystemTx(sender.ID(), txBody, sender, scs, blockNo)
	if err != nil {
		return nil, err
	}
	var event *types.Event
	switch ci.Name {
	case types.Stake:
		event, err = staking(txBody, sender, receiver, scs, blockNo)
	case types.VoteBP,
		types.VoteGasPrice,
		types.VoteNumBP,
		types.VoteNamePrice,
		types.VoteMinStaking:
		event, err = voting(txBody, sender, receiver, scs, blockNo, ci)
	case types.Unstake:
		event, err = unstaking(txBody, sender, receiver, scs, blockNo, ci)
	default:
		err = types.ErrTxInvalidPayload
	}
	if err != nil {
		return nil, err
	}
	var events []*types.Event
	events = append(events, event)
	return events, nil
}

func GetNamePrice(scs *state.ContractState) *big.Int {
	votelist, err := getVoteResult(scs, []byte(types.VoteNamePrice[2:]), 1)
	if err != nil {
		panic("could not get vote result for min staking")
	}
	if len(votelist.Votes) == 0 {
		return types.NamePrice
	}
	return new(big.Int).SetBytes(votelist.Votes[0].GetCandidate())
}

func GetMinimumStaking(scs *state.ContractState) *big.Int {
	votelist, err := getVoteResult(scs, []byte(types.VoteMinStaking[2:]), 1)
	if err != nil {
		panic("could not get vote result for min staking")
	}
	if len(votelist.Votes) == 0 {
		return types.StakingMinimum
	}
	minimumStaking, ok := new(big.Int).SetString(string(votelist.Votes[0].GetCandidate()), 10)
	if !ok {
		panic("could not get vote result for min staking")
	}
	return minimumStaking
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
		amount := txBody.GetAmountBigInt()

		if amount.Cmp(GetMinimumStaking(scs)) < 0 {
			return nil, types.ErrTooSmallAmount
		}
		if sender != nil && sender.Balance().Cmp(amount) < 0 {
			return nil, types.ErrInsufficientBalance
		}
	case types.VoteBP,
		types.VoteGasPrice,
		types.VoteNumBP,
		types.VoteNamePrice,
		types.VoteMinStaking:
		staked, err := getStaking(scs, account)
		if err != nil {
			return nil, err
		}
		if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
			return nil, types.ErrMustStakeBeforeVote
		}
		oldvote, err := GetVote(scs, account, []byte(ci.Name[2:]))
		if err != nil {
			return nil, err
		}
		if oldvote.Amount != nil && staked.GetWhen()+VotingDelay > blockNo {
			return nil, types.ErrLessTimeHasPassed
		}
	case types.Unstake:
		amount := txBody.GetAmountBigInt()
		if amount.Cmp(GetMinimumStaking(scs)) < 0 {
			return nil, types.ErrTooSmallAmount
		}
		_, err = validateForUnstaking(account, txBody, scs, blockNo)
	default:
		return nil, types.ErrTxInvalidPayload
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
