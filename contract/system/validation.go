package system

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func ValidateSystemTx(account []byte, txBody *types.TxBody, sender *state.V,
	scs *state.ContractState, blockNo uint64) (*SystemContext, error) {
	var ci types.CallInfo
	context := &SystemContext{Call: &ci, Sender: sender, BlockNo: blockNo}

	if err := json.Unmarshal(txBody.Payload, &ci); err != nil {
		return nil, types.ErrTxInvalidPayload
	}
	switch ci.Name {
	case types.Stake:
		if sender != nil && sender.Balance().Cmp(txBody.GetAmountBigInt()) < 0 {
			return nil, types.ErrInsufficientBalance
		}
		staked, err := validateForStaking(account, txBody, scs, blockNo)
		if err != nil {
			return nil, err
		}
		context.Staked = staked
	case types.VoteBP:
		staked, oldvote, err := validateForVote(account, txBody, scs, blockNo, []byte(ci.Name[2:]))
		if err != nil {
			return nil, err
		}
		context.Staked = staked
		context.Vote = oldvote
	case types.Unstake:
		staked, err := validateForUnstaking(account, txBody, scs, blockNo)
		if err != nil {
			return nil, err
		}
		context.Staked = staked
	case types.CreateAgenda:
		name, version, err := parseNameVersionForAgenda(&ci)
		if err != nil {
			return nil, err
		}
		agenda, err := getAgenda(scs, name, version)
		if err != nil {
			return nil, err
		}
		if agenda != nil {
			return nil, fmt.Errorf("already created agenda name:%s, version:%s", name, version)
		}
		if len(ci.Args) != 7 {
			return nil, fmt.Errorf("the request should be have 7 arguments")
		}
		start, ok := ci.Args[2].(string)
		if !ok {
			return nil, fmt.Errorf("could not parse the start block number %v", ci.Args[2])
		}
		blockfrom, err := strconv.ParseUint(start, 10, 64)
		if err != nil {
			return nil, err
		}
		end, ok := ci.Args[3].(string)
		if !ok {
			return nil, fmt.Errorf("could not parse the start block number %v", ci.Args[3])
		}
		blockto, err := strconv.ParseUint(end, 10, 64)
		if err != nil {
			return nil, err
		}
		max := ci.Args[4].(string)
		if !ok {
			return nil, fmt.Errorf("could not parse the max")
		}
		maxVote, err := strconv.ParseUint(max, 10, 32)
		if err != nil {
			return nil, err
		}
		desc, ok := ci.Args[5].(string)
		if !ok {
			return nil, fmt.Errorf("could not parse the desc")
		}
		candis, ok := ci.Args[6].([]interface{})
		if !ok {
			return nil, fmt.Errorf("could not parse the candidates %v %v", ci.Args[6], reflect.TypeOf(ci.Args[6]))
		}
		var candidates []string
		for _, candi := range candis {
			c, ok := candi.(string)
			if !ok {
				return nil, fmt.Errorf("could not parse the candidates")
			}
			candidates = append(candidates, c)
		}
		context.Agenda = &types.Agenda{
			Name:        name,
			Version:     version,
			Blockfrom:   blockfrom,
			Blockto:     blockto,
			Maxvote:     uint32(maxVote),
			Description: desc,
			Candidates:  candidates,
		}
	case types.VoteAgenda:
		name, version, err := parseNameVersionForAgenda(&ci)
		if err != nil {
			return nil, err
		}
		agenda, err := getAgenda(scs, name, version)
		if err != nil {
			return nil, err
		}
		if agenda == nil {
			return nil, fmt.Errorf("the agenda is not created (%s, %s)", name, version)
		}
		if blockNo < agenda.Blockfrom {
			return nil, fmt.Errorf("the voting begins at %d", agenda.Blockfrom)
		}
		if blockNo > agenda.Blockto {
			return nil, fmt.Errorf("the voting was already done at %d", agenda.Blockto)
		}
		candis := ci.Args[2:]
		if int64(len(candis)) > int64(agenda.Maxvote) {
			return nil, fmt.Errorf("too many candidates arguments (max : %d)", agenda.Maxvote)
		}
		sort.Slice(agenda.Candidates, func(i, j int) bool {
			return agenda.Candidates[i] <= agenda.Candidates[j]
		})
		if len(agenda.GetCandidates()) != 0 {
			for _, c := range ci.Args[2:] {
				candidate, ok := c.(string)
				if !ok {
					return nil, fmt.Errorf("include invalid candidate")
				}
				i := sort.SearchStrings(agenda.GetCandidates(), candidate)
				if i < len(agenda.Candidates) && agenda.Candidates[i] == candidate {
					//fmt.Printf("Found %s at index %d in %v.\n", x, i, a)
				} else {
					return nil, fmt.Errorf("candidate should be in %v", agenda.GetCandidates())
				}
			}
		}

		staked, oldvote, err := validateForVote(account, txBody, scs, blockNo, agenda.GetKey())
		if err != nil {
			return nil, err
		}
		context.Agenda = agenda
		context.Staked = staked
		context.Vote = oldvote
	default:
		return nil, types.ErrTxInvalidPayload
	}
	return context, nil
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
	if GetMinimumStaking(scs).Cmp(toBe) > 0 {
		return nil, types.ErrTooSmallAmount
	}
	return staked, nil
}
func validateForVote(account []byte, txBody *types.TxBody, scs *state.ContractState, blockNo uint64, voteKey []byte) (*types.Staking, *types.Vote, error) {
	staked, err := getStaking(scs, account)
	if err != nil {
		return nil, nil, err
	}
	if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
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
	staked, err := getStaking(scs, account)
	if err != nil {
		return nil, err
	}
	if staked.GetAmountBigInt().Cmp(big.NewInt(0)) == 0 {
		return nil, types.ErrMustStakeBeforeUnstake
	}
	if staked.GetAmountBigInt().Cmp(txBody.GetAmountBigInt()) < 0 {
		return nil, types.ErrExceedAmount
	}
	if staked.GetWhen()+StakingDelay > blockNo {
		return nil, types.ErrLessTimeHasPassed
	}
	toBe := new(big.Int).Sub(staked.GetAmountBigInt(), txBody.GetAmountBigInt())
	if toBe.Cmp(big.NewInt(0)) != 0 && GetMinimumStaking(scs).Cmp(toBe) > 0 {
		return nil, types.ErrTooSmallAmount
	}
	return staked, nil
}
func parseNameVersionForAgenda(ci *types.CallInfo) (string, string, error) {
	if len(ci.Args) < 2 {
		return "", "", types.ErrTxInvalidPayload
	}
	name, ok := ci.Args[0].(string)
	if !ok {
		return "", "", types.ErrTxInvalidPayload
	}
	version, ok := ci.Args[1].(string)
	if !ok {
		return "", "", types.ErrTxInvalidPayload
	}
	return name, version, nil
}
