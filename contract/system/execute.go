/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58"
)

//SystemContext is context of executing aergo.system transaction and filled after validation.
type SystemContext struct {
	BlockNo  uint64
	Call     *types.CallInfo
	Args     []string
	Staked   *types.Staking
	Vote     *types.Vote
	Proposal *types.Proposal
	Sender   *state.V
	Receiver *state.V
}

func ExecuteSystemTx(scs *state.ContractState, txBody *types.TxBody,
	sender, receiver *state.V, blockNo types.BlockNo) ([]*types.Event, error) {

	context, err := ValidateSystemTx(sender.ID(), txBody, sender, scs, blockNo)
	if err != nil {
		return nil, err
	}
	context.Receiver = receiver

	var event *types.Event
	switch context.Call.Name {
	case types.Stake:
		event, err = staking(txBody, sender, receiver, scs, blockNo, context)
	case types.VoteBP,
		types.VoteProposal:
		event, err = voting(txBody, sender, receiver, scs, blockNo, context)
	case types.Unstake:
		event, err = unstaking(txBody, sender, receiver, scs, blockNo, context)
	case types.CreateProposal:
		event, err = createProposal(txBody, sender, receiver, scs, blockNo, context)
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
	//	votelist, err := getVoteResult(scs, []byte(types.VoteNamePrice[2:]), 1)
	//	if err != nil {
	//		panic("could not get vote result for min staking")
	//	}
	//	if len(votelist.Votes) == 0 {
	//		return types.NamePrice
	//	}
	//	return new(big.Int).SetBytes(votelist.Votes[0].GetCandidate())
	return types.NamePrice
}

func GetMinimumStaking(scs *state.ContractState) *big.Int {
	//votelist, err := getVoteResult(scs, []byte(types.VoteMinStaking[2:]), 1)
	//if err != nil {
	//	panic("could not get vote result for min staking")
	//}
	//if len(votelist.Votes) == 0 {
	//	return types.StakingMinimum
	//}
	//minimumStaking, ok := new(big.Int).SetString(string(votelist.Votes[0].GetCandidate()), 10)
	//if !ok {
	//	panic("could not get vote result for min staking")
	//}
	//return minimumStaking
	return types.StakingMinimum
}

func GetVotes(scs *state.ContractState, address []byte) ([]*types.VoteInfo, error) {
	votes := getProposalHistory(scs, address)
	var results []*types.VoteInfo
	votes = append(votes, []byte(defaultVoteKey))
	for _, key := range votes {
		id := types.ProposalIDfromKey(key)
		result := &types.VoteInfo{Id: id}
		v, err := getVote(scs, key, address)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(key, defaultVoteKey) {
			for offset := 0; offset < len(v.Candidate); offset += PeerIDLength {
				candi := base58.Encode(v.Candidate[offset : offset+PeerIDLength])
				result.Candidates = append(result.Candidates, candi)
			}
		} else {
			err := json.Unmarshal(v.Candidate, &result.Candidates)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", err.Error(), string(v.Candidate))
			}
		}
		result.Amount = new(big.Int).SetBytes(v.Amount).String()
		results = append(results, result)
	}
	return results, nil
}
