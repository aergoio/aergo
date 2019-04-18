/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"math/big"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

type SystemContext struct {
	BlockNo  uint64
	Call     *types.CallInfo
	Args     []string
	Staked   *types.Staking
	Vote     *types.Vote
	Agenda   *types.Agenda
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
		types.VoteAgenda:
		event, err = voting(txBody, sender, receiver, scs, blockNo, context)
	case types.Unstake:
		event, err = unstaking(txBody, sender, receiver, scs, blockNo, context)
	case types.CreateAgenda:
		event, err = createAgenda(txBody, sender, receiver, scs, blockNo, context)
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
