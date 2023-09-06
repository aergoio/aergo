package system

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

var ErrTxSystemOperatorIsNotSet = errors.New("operator is not set")

func ValidateSystemTx(account []byte, txBody *types.TxBody, sender *state.V,
	scs *state.ContractState, blockInfo *types.BlockHeaderInfo) (*SystemContext, error) {
	var ci types.CallInfo
	if err := json.Unmarshal(txBody.Payload, &ci); err != nil {
		return nil, types.ErrTxInvalidPayload
	}
	blockNo := blockInfo.No

	context := &SystemContext{Call: &ci, Sender: sender, BlockInfo: blockInfo, op: types.GetOpSysTx(ci.Name), scs: scs, txBody: txBody}

	switch context.op {
	case types.Opstake:
		if sender != nil && sender.Balance().Cmp(txBody.GetAmountBigInt()) < 0 {
			return nil, types.ErrInsufficientBalance
		}
		staked, err := validateForStaking(account, txBody, scs, blockNo)
		if err != nil {
			return nil, err
		}
		context.Staked = staked
	case types.OpvoteBP:
		staked, oldvote, err := validateForVote(account, txBody, scs, blockNo, []byte(context.op.ID()))
		if err != nil {
			return nil, err
		}
		context.Staked = staked
		context.Vote = oldvote
	case types.Opunstake:
		staked, err := validateForUnstaking(account, txBody, scs, blockNo)
		if err != nil {
			return nil, err
		}
		context.Staked = staked
	case types.OpvoteDAO:
		if blockInfo.ForkVersion < 2 {
			return nil, fmt.Errorf("not supported operation")
		}
		id, err := parseIDForProposal(&ci)
		if err != nil {
			return nil, err
		}
		proposal, err := getProposal(id)
		if proposal == nil {
			return nil, err
		}
		if blockNo < proposal.Blockfrom {
			return nil, fmt.Errorf("the voting begins at %d", proposal.Blockfrom)
		}
		if proposal.Blockto != 0 && blockNo > proposal.Blockto {
			return nil, fmt.Errorf("the voting was already done at %d", proposal.Blockto)
		}
		candis := ci.Args[1:]
		if int64(len(candis)) > int64(proposal.MultipleChoice) {
			return nil, fmt.Errorf("too many candidates arguments (max : %d)", proposal.MultipleChoice)
		}
		for _, c := range candis {
			candidate, ok := c.(string)
			if !ok {
				return nil, fmt.Errorf("include invalid character")
			}
			candidateNumber, ok := new(big.Int).SetString(candidate, 10)
			if !ok {
				return nil, fmt.Errorf("include invalid number")
			}
			if !validateById(id, candidateNumber) {
				return nil, fmt.Errorf("include invalid number range")
			}
		}
		sort.Slice(proposal.Candidates, func(i, j int) bool {
			return proposal.Candidates[i] <= proposal.Candidates[j]
		})
		if len(proposal.Candidates) != 0 {
			for _, c := range candis {
				candidate, _ := c.(string) //already checked
				i := sort.SearchStrings(proposal.Candidates, candidate)
				if i < len(proposal.Candidates) && proposal.Candidates[i] == candidate {
					//fmt.Printf("Found %s at index %d in %v.\n", x, i, a)
				} else {
					return nil, fmt.Errorf("candidate should be in %v", proposal.Candidates)
				}
			}
		}

		staked, oldvote, err := validateForVote(account, txBody, scs, blockNo, proposal.GetKey())
		if err != nil {
			return nil, err
		}
		context.Proposal = proposal
		context.Staked = staked
		context.Vote = oldvote
	default:
		return nil, types.ErrTxInvalidPayload
	}
	return context, nil
}

func checkStakingBefore(account []byte, scs *state.ContractState) (*types.Staking, error) {
	staked, err := getStaking(scs, account)
	if err != nil {
		return nil, err
	}
	if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
		return nil, fmt.Errorf("not staking before")
	}
	return staked, nil
}

func validateForStaking(account []byte, txBody *types.TxBody, scs *state.ContractState, blockNo uint64) (*types.Staking, error) {
	staked, err := getStaking(scs, account)
	if err != nil {
		return nil, err
	}
	if staked.GetAmount() != nil && staked.GetWhen()+StakingDelay > blockNo {
		return nil, types.ErrLessTimeHasPassed
	}
	toBe := new(big.Int).Add(staked.GetAmountBigInt(), txBody.GetAmountBigInt())
	stakingMin := GetStakingMinimumFromState(scs)
	if stakingMin.Cmp(toBe) > 0 {
		return nil, types.ErrTooSmallAmount
	}
	return staked, nil
}

func validateForVote(account []byte, txBody *types.TxBody, scs *state.ContractState, blockNo uint64, voteKey []byte) (*types.Staking, *types.Vote, error) {
	staked, err := checkStakingBefore(account, scs)
	if err != nil {
		return nil, nil, types.ErrMustStakeBeforeVote
	}
	oldvote, err := GetVote(scs, account, voteKey)
	if err != nil {
		return nil, nil, err
	}
	if oldvote.Amount != nil && staked.GetWhen()+VotingDelay > blockNo {
		return nil, nil, types.ErrLessTimeHasPassed
	}
	return staked, oldvote, nil
}

func validateForUnstaking(account []byte, txBody *types.TxBody, scs *state.ContractState, blockNo uint64) (*types.Staking, error) {
	staked, err := checkStakingBefore(account, scs)
	if err != nil {
		return nil, types.ErrMustStakeBeforeUnstake
	}
	if staked.GetAmountBigInt().Cmp(txBody.GetAmountBigInt()) < 0 {
		return nil, types.ErrExceedAmount
	}
	if staked.GetWhen()+StakingDelay > blockNo {
		return nil, types.ErrLessTimeHasPassed
	}
	toBe := new(big.Int).Sub(staked.GetAmountBigInt(), txBody.GetAmountBigInt())
	stakingMin := GetStakingMinimumFromState(scs)
	if toBe.Cmp(big.NewInt(0)) != 0 && stakingMin.Cmp(toBe) > 0 {
		return nil, types.ErrTooSmallAmount
	}
	return staked, nil
}

func parseIDForProposal(ci *types.CallInfo) (string, error) {
	//length should be checked before this function
	id, ok := ci.Args[0].(string)
	if !ok || len(id) < 1 || !isValidID(id) {
		return "", fmt.Errorf("args[%d] invalid id", 0)
	}
	return strings.ToUpper(id), nil
}

func validateById(id string, candidate *big.Int) bool {
	if big.NewInt(0).Cmp(candidate) == 0 {
		return false
	}
	switch id {
	case bpCount.ID():
		if big.NewInt(100).Cmp(candidate) < 0 {
			return false
		}
	case stakingMin.ID(),
		gasPrice.ID(),
		namePrice.ID():
		if types.MaxAER.Cmp(candidate) < 0 {
			return false
		}
	}
	return true
}
